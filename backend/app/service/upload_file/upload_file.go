package upload_file

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/my_errors"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model/video"
	"douyin-backend/app/utils/auth"
	"douyin-backend/app/utils/md5_encrypt"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func UploadVideo(context *gin.Context, savePath string) (r bool, finalSavePath interface{}, message string) {
	uploadField := variable.ConfigYml.GetString("FileUploadSetting.UploadFileField")
	file, err := context.FormFile(uploadField)
	if err != nil {
		variable.ZapLog.Error("failed to get upload file: " + err.Error())
		return false, nil, "failed to get upload file"
	}

	if err = os.MkdirAll(savePath, os.ModePerm); err != nil {
		variable.ZapLog.Error("failed to create video directory: " + err.Error())
		return false, nil, "failed to create video directory"
	}

	sequence := variable.SnowFlake.GetId()
	if sequence <= 0 {
		variable.ZapLog.Error("snowflake failed to generate video id")
		return false, nil, "failed to generate video id"
	}

	saveFileName := fmt.Sprintf("%d%s", sequence, file.Filename)
	saveFileName = md5_encrypt.MD5(saveFileName) + path.Ext(saveFileName)
	videoFilePath := filepath.Join(savePath, saveFileName)

	if saveErr := context.SaveUploadedFile(file, videoFilePath); saveErr != nil {
		variable.ZapLog.Error("failed to save uploaded video: " + saveErr.Error())
		return false, nil, "failed to save uploaded video"
	}

	return enqueuePreparedVideoUpload(preparedVideoUploadInput{
		Sequence:         sequence,
		UID:              auth.GetUidFromToken(context),
		VideoFilePath:    videoFilePath,
		VideoRelativeDir: variable.ConfigYml.GetString("FileUploadSetting.VideoUploadFileSavePath"),
		VideoFileName:    saveFileName,
		ContentType:      file.Header.Get("Content-Type"),
		Description:      strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "description")),
		Tags:             strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "tags")),
		PrivateStatus:    int(context.GetFloat64(consts.ValidatorPrefix + "private_status")),
	})
}

func UploadAvatar(context *gin.Context, savePath string) (r bool, finnalSavePath interface{}) {
	file, _ := context.FormFile(variable.ConfigYml.GetString("FileUploadSetting.UploadFileField"))
	var saveErr error
	if sequence := variable.SnowFlake.GetId(); sequence > 0 {
		saveFileName := fmt.Sprintf("%d%s", sequence, file.Filename)
		saveFileName = md5_encrypt.MD5(saveFileName) + path.Ext(saveFileName)
		filePath := filepath.Join(savePath, saveFileName)
		if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
			variable.ZapLog.Error("failed to create avatar directory: " + err.Error())
			return false, nil
		}
		if saveErr = context.SaveUploadedFile(file, filePath); saveErr == nil {
			urlAddr := buildPublicFileURL(variable.ConfigYml.GetString("FileUploadSetting.AvatarSmallUploadFileSavePath"), saveFileName)
			if useCOSStorage() {
				urlAddr, saveErr = uploadLocalFileToCOS(filePath, variable.ConfigYml.GetString("FileUploadSetting.AvatarSmallUploadFileSavePath"), saveFileName, file.Header.Get("Content-Type"))
				if saveErr != nil {
					variable.ZapLog.Error("failed to upload avatar to cos: " + saveErr.Error())
					_ = os.Remove(filePath)
					return false, nil
				}
			}
			insertStatus := video.CreateVideoFactory("").UpdateAvatar(context, urlAddr)
			if insertStatus {
				finnalSavePath = gin.H{
					"urlAddr": urlAddr,
				}
			}
			if useCOSStorage() {
				_ = os.Remove(filePath)
			}
			return true, finnalSavePath
		}
	} else {
		saveErr = errors.New(my_errors.ErrorsSnowflakeGetIdFail)
		variable.ZapLog.Error("snowflake failed to generate avatar id: " + saveErr.Error())
	}
	return false, nil
}

func UploadCover(context *gin.Context, savePath string) (r bool, finnalSavePath interface{}) {
	file, _ := context.FormFile(variable.ConfigYml.GetString("FileUploadSetting.UploadFileField"))
	var saveErr error
	if sequence := variable.SnowFlake.GetId(); sequence > 0 {
		saveFileName := fmt.Sprintf("%d%s", sequence, file.Filename)
		saveFileName = md5_encrypt.MD5(saveFileName) + path.Ext(saveFileName)
		filePath := filepath.Join(savePath, saveFileName)
		if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
			variable.ZapLog.Error("failed to create cover directory: " + err.Error())
			return false, nil
		}
		if saveErr = context.SaveUploadedFile(file, filePath); saveErr == nil {
			urlAddr := buildPublicFileURL(variable.ConfigYml.GetString("FileUploadSetting.CoverUploadFileSavePath"), saveFileName)
			if useCOSStorage() {
				urlAddr, saveErr = uploadLocalFileToCOS(filePath, variable.ConfigYml.GetString("FileUploadSetting.CoverUploadFileSavePath"), saveFileName, file.Header.Get("Content-Type"))
				if saveErr != nil {
					variable.ZapLog.Error("failed to upload cover to cos: " + saveErr.Error())
					_ = os.Remove(filePath)
					return false, nil
				}
			}
			insertStatus := video.CreateVideoFactory("").UpdateCover(context, urlAddr)
			if insertStatus {
				finnalSavePath = gin.H{
					"urlAddr": urlAddr,
				}
			}
			if useCOSStorage() {
				_ = os.Remove(filePath)
			}
			return true, finnalSavePath
		}
	} else {
		saveErr = errors.New(my_errors.ErrorsSnowflakeGetIdFail)
		variable.ZapLog.Error("snowflake failed to generate cover id: " + saveErr.Error())
	}
	return false, nil
}

func generateYearMonthPath(savePathPre string) (string, string) {
	returnPath := variable.BasePath + variable.ConfigYml.GetString("FileUploadSetting.UploadFileReturnPath")
	curYearMonth := time.Now().Format("2006_01")
	newSavePathPre := savePathPre + curYearMonth
	newReturnPathPre := returnPath + curYearMonth
	if _, err := os.Stat(newSavePathPre); err != nil {
		if err = os.MkdirAll(newSavePathPre, os.ModePerm); err != nil {
			variable.ZapLog.Error("failed to create directory: " + err.Error())
			return "", ""
		}
	}
	return newSavePathPre + "/", newReturnPathPre + "/"
}

type preparedVideoUploadInput struct {
	Sequence         int64
	UID              int64
	VideoFilePath    string
	VideoRelativeDir string
	VideoFileName    string
	ContentType      string
	Description      string
	Tags             string
	PrivateStatus    int
}

func buildVideoDescription(description, tags string) string {
	videoDesc := strings.TrimSpace(description)
	tags = strings.TrimSpace(tags)
	if tags == "" {
		return videoDesc
	}
	if videoDesc != "" {
		videoDesc += "\n"
	}
	return videoDesc + tags
}

func enqueuePreparedVideoUpload(input preparedVideoUploadInput) (bool, interface{}, string) {
	coverSavePath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath") +
		variable.ConfigYml.GetString("FileUploadSetting.VideoCoverUploadFileSavePath")
	if err := os.MkdirAll(coverSavePath, os.ModePerm); err != nil {
		variable.ZapLog.Error("failed to create cover directory: " + err.Error())
		cleanupLocalFiles(input.VideoFilePath)
		return false, nil, "failed to create cover directory"
	}

	saveCoverFileName := strings.TrimSuffix(input.VideoFileName, path.Ext(input.VideoFileName)) + ".png"
	coverFilePath := filepath.Join(coverSavePath, saveCoverFileName)
	videoDesc := buildVideoDescription(input.Description, input.Tags)

	task := VideoUploadTask{
		TaskID:           fmt.Sprintf("%d", input.Sequence),
		UID:              input.UID,
		VideoFilePath:    input.VideoFilePath,
		CoverFilePath:    coverFilePath,
		VideoRelativeDir: input.VideoRelativeDir,
		CoverRelativeDir: variable.ConfigYml.GetString("FileUploadSetting.VideoCoverUploadFileSavePath"),
		VideoFileName:    input.VideoFileName,
		CoverFileName:    saveCoverFileName,
		ContentType:      input.ContentType,
		VideoDesc:        videoDesc,
		PrivateStatus:    input.PrivateStatus,
	}

	if err := EnqueueVideoUploadTask(task); err != nil {
		variable.ZapLog.Error("failed to enqueue video upload task: " + err.Error())
		cleanupLocalFiles(input.VideoFilePath, coverFilePath)
		return false, nil, "failed to queue video upload task"
	}

	return true, gin.H{
		"taskId":        task.TaskID,
		"status":        "queued",
		"videoDesc":     videoDesc,
		"privateStatus": input.PrivateStatus,
	}, ""
}

func extractCoverFrame(videoPath, coverPath string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-frames:v", "1", coverPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract frame using ffmpeg: %v", err)
	}
	return nil
}

func cleanupLocalFiles(paths ...string) {
	for _, filePath := range paths {
		if strings.TrimSpace(filePath) == "" {
			continue
		}
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			variable.ZapLog.Error("failed to remove local file: " + err.Error())
		}
	}
}
