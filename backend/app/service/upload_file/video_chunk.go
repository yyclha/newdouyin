package upload_file

import (
	"crypto/sha256"
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/utils/auth"
	"douyin-backend/app/utils/md5_encrypt"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	videoChunkSessionDirName         = ".chunks"
	defaultVideoChunkStateTTLSeconds = 86400
)

// videoChunkUploadMeta 定义业务数据结构。
type videoChunkUploadMeta struct {
	UploadID      string `json:"upload_id"`
	UID           int64  `json:"uid"`
	FileName      string `json:"file_name"`
	ContentType   string `json:"content_type"`
	FileSize      int64  `json:"file_size"`
	ChunkSize     int64  `json:"chunk_size"`
	TotalChunks   int    `json:"total_chunks"`
	Description   string `json:"description"`
	Tags          string `json:"tags"`
	PrivateStatus int    `json:"private_status"`
	Status        string `json:"status"`
	TaskID        string `json:"task_id"`
	VideoDesc     string `json:"video_desc"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}

// InitVideoChunkUpload 执行业务处理。
func InitVideoChunkUpload(context *gin.Context) (r bool, finalSavePath interface{}, message string) {
	uploadID := strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "upload_id"))
	if uploadID == "" {
		uploadID = generateVideoChunkUploadID()
	}

	meta := videoChunkUploadMeta{
		UploadID:      uploadID,
		UID:           auth.GetUidFromToken(context),
		FileName:      strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "file_name")),
		ContentType:   strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "content_type")),
		FileSize:      int64(context.GetFloat64(consts.ValidatorPrefix + "file_size")),
		ChunkSize:     int64(context.GetFloat64(consts.ValidatorPrefix + "chunk_size")),
		TotalChunks:   int(context.GetFloat64(consts.ValidatorPrefix + "total_chunks")),
		Description:   strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "description")),
		Tags:          strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "tags")),
		PrivateStatus: int(context.GetFloat64(consts.ValidatorPrefix + "private_status")),
	}

	if err := validateVideoChunkMeta(meta); err != nil {
		return false, nil, err.Error()
	}

	sessionDir := videoChunkSessionDir(meta.UploadID, meta.UID)
	chunkDir := filepath.Join(sessionDir, "chunks")
	if err := os.MkdirAll(chunkDir, os.ModePerm); err != nil {
		variable.ZapLog.Error("failed to create chunk session directory: " + err.Error())
		return false, nil, "failed to create upload session"
	}

	stateStore := createVideoChunkStateStore(meta.UploadID, meta.UID)
	if stateStore == nil {
		return false, nil, "failed to connect upload session store"
	}
	defer stateStore.Release()

	metaPath := filepath.Join(sessionDir, "meta.json")
	if existingMeta, err := readVideoChunkMeta(metaPath); err == nil {
		if existingMeta.UID != meta.UID || existingMeta.FileSize != meta.FileSize || existingMeta.TotalChunks != meta.TotalChunks || existingMeta.FileName != meta.FileName {
			return false, nil, "upload session conflicts with another file"
		}
		existingMeta.Description = meta.Description
		existingMeta.Tags = meta.Tags
		existingMeta.PrivateStatus = meta.PrivateStatus
		if meta.ContentType != "" {
			existingMeta.ContentType = meta.ContentType
		}
		existingMeta.ChunkSize = meta.ChunkSize
		existingMeta.UpdatedAt = time.Now().Unix()
		if err = writeVideoChunkMeta(metaPath, existingMeta); err != nil {
			variable.ZapLog.Error("failed to update upload meta: " + err.Error())
			return false, nil, "failed to update upload session"
		}

		uploadedChunks, err := reconcileVideoChunkState(existingMeta, chunkDir, stateStore)
		if err != nil {
			variable.ZapLog.Error("failed to inspect uploaded chunks: " + err.Error())
			return false, nil, "failed to inspect uploaded chunks"
		}
		return true, buildVideoChunkInitResponse(existingMeta, uploadedChunks), ""
	} else if !os.IsNotExist(err) {
		variable.ZapLog.Error("failed to read upload meta: " + err.Error())
		return false, nil, "failed to read upload session"
	}

	now := time.Now().Unix()
	meta.Status = "uploading"
	meta.CreatedAt = now
	meta.UpdatedAt = now
	if err := writeVideoChunkMeta(metaPath, meta); err != nil {
		variable.ZapLog.Error("failed to save upload meta: " + err.Error())
		return false, nil, "failed to create upload session"
	}

	if err := stateStore.RefreshTTL(videoChunkStateTTLSeconds()); err != nil {
		variable.ZapLog.Error("failed to initialize upload chunk state ttl: " + err.Error())
		return false, nil, "failed to initialize upload session"
	}

	return true, buildVideoChunkInitResponse(meta, []int{}), ""
}

// SaveVideoChunk 执行业务处理。
func SaveVideoChunk(context *gin.Context) (r bool, finalSavePath interface{}, message string) {
	uploadID := strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "upload_id"))
	chunkIndex := int(context.GetFloat64(consts.ValidatorPrefix + "chunk_index"))
	totalChunks := int(context.GetFloat64(consts.ValidatorPrefix + "total_chunks"))
	chunkHash := normalizeVideoChunkHash(context.GetString(consts.ValidatorPrefix + "chunk_hash"))

	if uploadID == "" {
		return false, nil, "upload_id is required"
	}
	if chunkIndex < 0 || totalChunks <= 0 || chunkIndex >= totalChunks {
		return false, nil, "invalid chunk range"
	}
	if err := validateVideoChunkHash(chunkHash); err != nil {
		return false, nil, err.Error()
	}

	currentUID := auth.GetUidFromToken(context)
	metaPath := filepath.Join(videoChunkSessionDir(uploadID, currentUID), "meta.json")
	meta, err := readVideoChunkMeta(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, "upload session not found"
		}
		variable.ZapLog.Error("failed to read upload meta: " + err.Error())
		return false, nil, "failed to read upload session"
	}

	if meta.UID != currentUID {
		return false, nil, "upload session does not belong to current user"
	}

	stateStore := createVideoChunkStateStore(uploadID, currentUID)
	if stateStore == nil {
		return false, nil, "failed to connect upload session store"
	}
	defer stateStore.Release()

	if meta.Status == "queued" {
		return true, gin.H{
			"uploadId":   meta.UploadID,
			"chunkIndex": chunkIndex,
			"status":     meta.Status,
			"taskId":     meta.TaskID,
		}, ""
	}
	if totalChunks != meta.TotalChunks {
		return false, nil, "total_chunks does not match upload session"
	}

	file, err := context.FormFile(variable.ConfigYml.GetString("FileUploadSetting.UploadFileField"))
	if err != nil {
		variable.ZapLog.Error("failed to get chunk file: " + err.Error())
		return false, nil, "failed to get chunk file"
	}

	expectedChunkSize := expectedVideoChunkSize(meta, chunkIndex)
	if expectedChunkSize <= 0 {
		return false, nil, "invalid chunk size"
	}
	if file.Size != expectedChunkSize {
		return false, nil, "chunk size does not match upload session"
	}

	chunkDir := filepath.Join(videoChunkSessionDir(uploadID, currentUID), "chunks")
	if err = os.MkdirAll(chunkDir, os.ModePerm); err != nil {
		variable.ZapLog.Error("failed to create chunk directory: " + err.Error())
		return false, nil, "failed to create upload session"
	}

	chunkPath := filepath.Join(chunkDir, videoChunkFileName(chunkIndex))
	tempChunkPath := chunkPath + ".tmp"
	if err = context.SaveUploadedFile(file, tempChunkPath); err != nil {
		variable.ZapLog.Error("failed to save chunk file: " + err.Error())
		return false, nil, "failed to save chunk file"
	}

	actualChunkHash, actualChunkSize, err := calculateFileSHA256(tempChunkPath)
	if err != nil {
		_ = os.Remove(tempChunkPath)
		variable.ZapLog.Error("failed to checksum chunk file: " + err.Error())
		return false, nil, "failed to verify chunk integrity"
	}
	if actualChunkSize != expectedChunkSize {
		_ = os.Remove(tempChunkPath)
		return false, nil, "chunk size does not match upload session"
	}
	if !strings.EqualFold(actualChunkHash, chunkHash) {
		_ = os.Remove(tempChunkPath)
		return false, nil, "chunk integrity check failed"
	}

	wroteNewChunk, err := persistVerifiedChunk(tempChunkPath, chunkPath, actualChunkHash)
	if err != nil {
		_ = os.Remove(tempChunkPath)
		variable.ZapLog.Error("failed to persist verified chunk: " + err.Error())
		return false, nil, "failed to persist verified chunk"
	}

	meta.UpdatedAt = time.Now().Unix()
	if err = writeVideoChunkMeta(metaPath, meta); err != nil {
		if wroteNewChunk {
			_ = os.Remove(chunkPath)
		}
		variable.ZapLog.Error("failed to update upload meta: " + err.Error())
		return false, nil, "failed to update upload session"
	}

	if err = stateStore.SetUploadedChunk(chunkIndex, actualChunkHash, videoChunkStateTTLSeconds()); err != nil {
		if wroteNewChunk {
			_ = os.Remove(chunkPath)
		}
		variable.ZapLog.Error("failed to persist upload chunk state: " + err.Error())
		return false, nil, "failed to persist upload session"
	}

	uploadedChunks, err := reconcileVideoChunkState(meta, chunkDir, stateStore)
	if err != nil {
		variable.ZapLog.Error("failed to inspect uploaded chunks: " + err.Error())
		return false, nil, "failed to inspect uploaded chunks"
	}

	return true, gin.H{
		"uploadId":       meta.UploadID,
		"chunkIndex":     chunkIndex,
		"uploadedChunks": uploadedChunks,
		"status":         meta.Status,
	}, ""
}

// CompleteVideoChunkUpload 执行业务处理。
func CompleteVideoChunkUpload(context *gin.Context, savePath string) (r bool, finalSavePath interface{}, message string) {
	uploadID := strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "upload_id"))
	if uploadID == "" {
		return false, nil, "upload_id is required"
	}

	currentUID := auth.GetUidFromToken(context)
	metaPath := filepath.Join(videoChunkSessionDir(uploadID, currentUID), "meta.json")
	meta, err := readVideoChunkMeta(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, "upload session not found"
		}
		variable.ZapLog.Error("failed to read upload meta: " + err.Error())
		return false, nil, "failed to read upload session"
	}

	if meta.UID != currentUID {
		return false, nil, "upload session does not belong to current user"
	}

	if meta.Status == "queued" {
		return true, gin.H{
			"taskId":        meta.TaskID,
			"status":        meta.Status,
			"videoDesc":     meta.VideoDesc,
			"privateStatus": meta.PrivateStatus,
		}, ""
	}

	stateStore := createVideoChunkStateStore(uploadID, currentUID)
	if stateStore == nil {
		return false, nil, "failed to connect upload session store"
	}
	defer stateStore.Release()

	chunkDir := filepath.Join(videoChunkSessionDir(uploadID, currentUID), "chunks")
	uploadedChunks, err := reconcileVideoChunkState(meta, chunkDir, stateStore)
	if err != nil {
		variable.ZapLog.Error("failed to inspect uploaded chunks: " + err.Error())
		return false, nil, "failed to inspect uploaded chunks"
	}

	missingChunks := missingVideoChunkIndexes(meta.TotalChunks, uploadedChunks)
	if len(missingChunks) > 0 {
		return false, gin.H{
			"missingChunks": missingChunks,
		}, "chunks are missing"
	}

	if existingTask, taskErr := getVideoUploadTaskByUploadIDForUser(uploadID, currentUID); taskErr == nil {
		return true, buildVideoUploadTaskResponse(existingTask), ""
	}

	if err = os.MkdirAll(savePath, os.ModePerm); err != nil {
		variable.ZapLog.Error("failed to create video directory: " + err.Error())
		return false, nil, "failed to create video directory"
	}

	sequence := variable.SnowFlake.GetId()
	if sequence <= 0 {
		return false, nil, "failed to generate video id"
	}

	saveFileName := fmt.Sprintf("%d%s", sequence, meta.FileName)
	saveFileName = md5_encrypt.MD5(saveFileName) + path.Ext(saveFileName)
	videoFilePath := filepath.Join(savePath, saveFileName)
	if err = mergeVideoChunks(videoFilePath, chunkDir, meta.TotalChunks); err != nil {
		variable.ZapLog.Error("failed to merge video chunks: " + err.Error())
		cleanupLocalFiles(videoFilePath)
		return false, nil, "failed to merge video chunks"
	}

	ok, payload, errMessage := enqueuePreparedVideoUpload(preparedVideoUploadInput{
		Sequence:         sequence,
		UploadID:         meta.UploadID,
		UID:              meta.UID,
		VideoFilePath:    videoFilePath,
		VideoRelativeDir: variable.ConfigYml.GetString("FileUploadSetting.VideoUploadFileSavePath"),
		VideoFileName:    saveFileName,
		ContentType:      meta.ContentType,
		Description:      meta.Description,
		Tags:             meta.Tags,
		PrivateStatus:    meta.PrivateStatus,
	})
	if !ok {
		cleanupLocalFiles(videoFilePath)
		return false, payload, errMessage
	}

	meta.Status = "queued"
	meta.TaskID = fmt.Sprintf("%d", sequence)
	meta.VideoDesc = buildVideoDescription(meta.Description, meta.Tags)
	meta.UpdatedAt = time.Now().Unix()
	if err = writeVideoChunkMeta(metaPath, meta); err != nil {
		variable.ZapLog.Error("failed to update upload meta after completion: " + err.Error())
	}
	if err = os.RemoveAll(chunkDir); err != nil {
		variable.ZapLog.Error("failed to cleanup chunk directory: " + err.Error())
	}
	if err = stateStore.Clear(); err != nil {
		variable.ZapLog.Error("failed to cleanup upload chunk state: " + err.Error())
	}

	return true, payload, ""
}

// validateVideoChunkMeta 执行业务处理。
func validateVideoChunkMeta(meta videoChunkUploadMeta) error {
	if meta.FileName == "" {
		return fmt.Errorf("file_name is required")
	}
	if meta.FileSize <= 0 {
		return fmt.Errorf("file_size must be greater than 0")
	}
	if meta.ChunkSize <= 0 {
		return fmt.Errorf("chunk_size must be greater than 0")
	}
	if meta.TotalChunks <= 0 {
		return fmt.Errorf("total_chunks must be greater than 0")
	}
	sizeLimit := variable.ConfigYml.GetInt64("FileUploadSetting.Size") << 20
	if meta.FileSize > sizeLimit {
		return fmt.Errorf("file size exceeds limit")
	}
	if meta.ChunkSize > sizeLimit {
		return fmt.Errorf("chunk size exceeds limit")
	}
	expectedChunks := int((meta.FileSize + meta.ChunkSize - 1) / meta.ChunkSize)
	if expectedChunks != meta.TotalChunks {
		return fmt.Errorf("total_chunks does not match file size")
	}
	return nil
}

// validateVideoChunkHash 执行业务处理。
func validateVideoChunkHash(chunkHash string) error {
	if chunkHash == "" {
		return fmt.Errorf("chunk_hash is required")
	}
	if len(chunkHash) != sha256.Size*2 {
		return fmt.Errorf("chunk_hash must be a sha256 hex string")
	}
	for _, char := range chunkHash {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return fmt.Errorf("chunk_hash must be a sha256 hex string")
		}
	}
	return nil
}

// normalizeVideoChunkHash 执行业务处理。
func normalizeVideoChunkHash(chunkHash string) string {
	return strings.ToLower(strings.TrimSpace(chunkHash))
}

// buildVideoChunkInitResponse 执行业务处理。
func buildVideoChunkInitResponse(meta videoChunkUploadMeta, uploadedChunks []int) gin.H {
	return gin.H{
		"uploadId":       meta.UploadID,
		"fileName":       meta.FileName,
		"fileSize":       meta.FileSize,
		"chunkSize":      meta.ChunkSize,
		"totalChunks":    meta.TotalChunks,
		"uploadedChunks": uploadedChunks,
		"status":         meta.Status,
		"taskId":         meta.TaskID,
	}
}

// videoChunkSessionDir 执行业务处理。
func videoChunkSessionDir(uploadID string, uid int64) string {
	rootPath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath")
	return filepath.Join(rootPath, videoChunkSessionDirName, fmt.Sprintf("%d_%s", uid, md5_encrypt.MD5(uploadID)))
}

// videoChunkFileName 执行业务处理。
func videoChunkFileName(chunkIndex int) string {
	return "chunk-" + fmt.Sprintf("%06d", chunkIndex) + ".part"
}

// generateVideoChunkUploadID 执行业务处理。
func generateVideoChunkUploadID() string {
	if variable.SnowFlake != nil {
		if sequence := variable.SnowFlake.GetId(); sequence > 0 {
			return fmt.Sprintf("upload-%d", sequence)
		}
	}
	return "upload-" + strings.ReplaceAll(uuid.NewString(), "-", "")
}

// readVideoChunkMeta 执行业务处理。
func readVideoChunkMeta(metaPath string) (videoChunkUploadMeta, error) {
	var meta videoChunkUploadMeta
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return meta, err
	}
	err = json.Unmarshal(data, &meta)
	return meta, err
}

// writeVideoChunkMeta 执行业务处理。
func writeVideoChunkMeta(metaPath string, meta videoChunkUploadMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}

// expectedVideoChunkSize 执行业务处理。
func expectedVideoChunkSize(meta videoChunkUploadMeta, chunkIndex int) int64 {
	if chunkIndex < 0 || chunkIndex >= meta.TotalChunks || meta.ChunkSize <= 0 {
		return 0
	}

	lastChunkSize := meta.FileSize - int64(meta.TotalChunks-1)*meta.ChunkSize
	if chunkIndex == meta.TotalChunks-1 && lastChunkSize > 0 {
		return lastChunkSize
	}
	return meta.ChunkSize
}

// listValidUploadedChunkIndexes 执行业务处理。
func listValidUploadedChunkIndexes(meta videoChunkUploadMeta, chunkDir string) ([]int, error) {
	entries, err := os.ReadDir(chunkDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, err
	}

	indexes := make([]int, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "chunk-") || !strings.HasSuffix(name, ".part") {
			continue
		}

		indexText := strings.TrimSuffix(strings.TrimPrefix(name, "chunk-"), ".part")
		index, convErr := strconv.Atoi(indexText)
		if convErr != nil || index < 0 || index >= meta.TotalChunks {
			continue
		}

		info, statErr := entry.Info()
		if statErr != nil {
			return nil, statErr
		}
		if info.Size() != expectedVideoChunkSize(meta, index) {
			continue
		}

		indexes = append(indexes, index)
	}

	sort.Ints(indexes)
	return indexes, nil
}

// reconcileVideoChunkState 执行业务处理。
func reconcileVideoChunkState(meta videoChunkUploadMeta, chunkDir string, stateStore *videoChunkStateStore) ([]int, error) {
	redisChunks, err := stateStore.GetUploadedChunks()
	if err != nil {
		return nil, err
	}

	redisHashes, err := stateStore.GetChunkHashes()
	if err != nil {
		return nil, err
	}

	uploadedChunks, chunkHashes, invalidChunks, err := collectVerifiedVideoChunks(meta, chunkDir, redisHashes)
	if err != nil {
		return nil, err
	}

	validChunkSet := make(map[int]struct{}, len(uploadedChunks))
	for _, chunkIndex := range uploadedChunks {
		validChunkSet[chunkIndex] = struct{}{}
		if err = stateStore.SetUploadedChunk(chunkIndex, chunkHashes[chunkIndex], videoChunkStateTTLSeconds()); err != nil {
			return nil, err
		}
	}

	for _, chunkIndex := range invalidChunks {
		chunkPath := filepath.Join(chunkDir, videoChunkFileName(chunkIndex))
		if removeErr := os.Remove(chunkPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return nil, removeErr
		}
		if err = stateStore.RemoveChunk(chunkIndex); err != nil {
			return nil, err
		}
	}

	redisChunkSet := make(map[int]struct{}, len(redisChunks)+len(redisHashes))
	for _, chunkIndex := range redisChunks {
		redisChunkSet[chunkIndex] = struct{}{}
	}
	for chunkIndex := range redisHashes {
		redisChunkSet[chunkIndex] = struct{}{}
	}

	for chunkIndex := range redisChunkSet {
		if _, ok := validChunkSet[chunkIndex]; !ok {
			if err = stateStore.RemoveChunk(chunkIndex); err != nil {
				return nil, err
			}
		}
	}

	return uploadedChunks, nil
}

// collectVerifiedVideoChunks 执行业务处理。
func collectVerifiedVideoChunks(meta videoChunkUploadMeta, chunkDir string, expectedHashes map[int]string) ([]int, map[int]string, []int, error) {
	uploadedChunks, err := listValidUploadedChunkIndexes(meta, chunkDir)
	if err != nil {
		return nil, nil, nil, err
	}

	verifiedChunks := make([]int, 0, len(uploadedChunks))
	verifiedHashes := make(map[int]string, len(uploadedChunks))
	invalidChunks := make([]int, 0)

	for _, chunkIndex := range uploadedChunks {
		chunkPath := filepath.Join(chunkDir, videoChunkFileName(chunkIndex))
		actualHash, actualSize, hashErr := calculateFileSHA256(chunkPath)
		if hashErr != nil {
			return nil, nil, nil, hashErr
		}
		if actualSize != expectedVideoChunkSize(meta, chunkIndex) {
			invalidChunks = append(invalidChunks, chunkIndex)
			continue
		}

		expectedHash := normalizeVideoChunkHash(expectedHashes[chunkIndex])
		if expectedHash != "" && !strings.EqualFold(actualHash, expectedHash) {
			invalidChunks = append(invalidChunks, chunkIndex)
			continue
		}

		verifiedChunks = append(verifiedChunks, chunkIndex)
		verifiedHashes[chunkIndex] = actualHash
	}

	return verifiedChunks, verifiedHashes, invalidChunks, nil
}

// missingVideoChunkIndexes 执行业务处理。
func missingVideoChunkIndexes(totalChunks int, uploadedChunks []int) []int {
	if totalChunks <= 0 {
		return []int{}
	}
	exists := make(map[int]struct{}, len(uploadedChunks))
	for _, index := range uploadedChunks {
		exists[index] = struct{}{}
	}
	missing := make([]int, 0)
	for index := 0; index < totalChunks; index++ {
		if _, ok := exists[index]; !ok {
			missing = append(missing, index)
		}
	}
	return missing
}

// calculateFileSHA256 执行业务处理。
func calculateFileSHA256(filePath string) (hash string, size int64, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	hasher := sha256.New()
	written, err := io.Copy(hasher, file)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(hasher.Sum(nil)), written, nil
}

// persistVerifiedChunk 执行业务处理。
func persistVerifiedChunk(tempChunkPath, chunkPath, expectedHash string) (bool, error) {
	if _, err := os.Stat(chunkPath); err == nil {
		existingHash, _, hashErr := calculateFileSHA256(chunkPath)
		if hashErr != nil {
			return false, hashErr
		}
		if strings.EqualFold(existingHash, expectedHash) {
			if removeErr := os.Remove(tempChunkPath); removeErr != nil && !os.IsNotExist(removeErr) {
				return false, removeErr
			}
			return false, nil
		}
		if removeErr := os.Remove(chunkPath); removeErr != nil {
			return false, removeErr
		}
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	if err := os.Rename(tempChunkPath, chunkPath); err != nil {
		return false, err
	}
	return true, nil
}

// videoChunkStateTTLSeconds 执行业务处理。
func videoChunkStateTTLSeconds() int64 {
	ttl := variable.ConfigYml.GetInt64("FileUploadSetting.VideoChunkStateTTL")
	if ttl <= 0 {
		return defaultVideoChunkStateTTLSeconds
	}
	return ttl
}

// mergeVideoChunks 执行业务处理。
func mergeVideoChunks(targetPath, chunkDir string, totalChunks int) error {
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = targetFile.Close()
	}()

	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		chunkPath := filepath.Join(chunkDir, videoChunkFileName(chunkIndex))
		chunkFile, openErr := os.Open(chunkPath)
		if openErr != nil {
			return openErr
		}
		if _, copyErr := io.Copy(targetFile, chunkFile); copyErr != nil {
			_ = chunkFile.Close()
			return copyErr
		}
		if closeErr := chunkFile.Close(); closeErr != nil {
			return closeErr
		}
	}

	return nil
}
