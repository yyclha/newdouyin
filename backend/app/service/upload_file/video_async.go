package upload_file

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model/video"
	"douyin-backend/app/utils/file_storage"
	"errors"
	"runtime"
	"sync"

	"go.uber.org/zap"
)

const (
	defaultVideoUploadWorkerCount = 2
	defaultVideoUploadQueueSize   = 16
)

var (
	videoUploadQueue     chan VideoUploadTask
	videoUploadQueueOnce sync.Once

	errVideoUploadQueueFull = errors.New("video upload queue is full")
)

type VideoUploadTask struct {
	TaskID           string
	UID              int64
	VideoFilePath    string
	CoverFilePath    string
	VideoRelativeDir string
	CoverRelativeDir string
	VideoFileName    string
	CoverFileName    string
	ContentType      string
	VideoDesc        string
	PrivateStatus    int
}

func InitVideoUploadQueue() {
	videoUploadQueueOnce.Do(func() {
		workerCount := variable.ConfigYml.GetInt("FileUploadSetting.VideoAsync.Workers")
		if workerCount <= 0 {
			workerCount = defaultVideoUploadWorkerCount
			if cpuHalf := runtime.NumCPU() / 2; cpuHalf > workerCount && cpuHalf < 5 {
				workerCount = cpuHalf
			}
		}

		queueSize := variable.ConfigYml.GetInt("FileUploadSetting.VideoAsync.QueueSize")
		if queueSize <= 0 {
			queueSize = defaultVideoUploadQueueSize
		}

		videoUploadQueue = make(chan VideoUploadTask, queueSize)
		for idx := 0; idx < workerCount; idx++ {
			go videoUploadWorker(idx + 1)
		}

		variable.ZapLog.Info("video upload queue initialized",
			zap.Int("workers", workerCount),
			zap.Int("queue_size", queueSize),
		)
	})
}

func EnqueueVideoUploadTask(task VideoUploadTask) error {
	InitVideoUploadQueue()

	select {
	case videoUploadQueue <- task:
		return nil
	default:
		return errVideoUploadQueueFull
	}
}

func videoUploadWorker(workerID int) {
	for task := range videoUploadQueue {
		processVideoUploadTask(task, workerID)
	}
}

func processVideoUploadTask(task VideoUploadTask, workerID int) {
	logger := variable.ZapLog.With(
		zap.String("task_id", task.TaskID),
		zap.Int("worker_id", workerID),
		zap.Int64("uid", task.UID),
	)

	if err := extractCoverFrame(task.VideoFilePath, task.CoverFilePath); err != nil {
		logger.Error("failed to extract video cover", zap.Error(err))
		cleanupLocalFiles(task.VideoFilePath, task.CoverFilePath)
		return
	}

	playAddr := buildPublicFileURL(task.VideoRelativeDir, task.VideoFileName)
	coverAddr := buildPublicFileURL(task.CoverRelativeDir, task.CoverFileName)

	if useCOSStorage() {
		var err error
		playAddr, err = uploadLocalFileToCOS(task.VideoFilePath, task.VideoRelativeDir, task.VideoFileName, task.ContentType)
		if err != nil {
			logger.Error("failed to upload video to cos", zap.Error(err))
			cleanupLocalFiles(task.VideoFilePath, task.CoverFilePath)
			return
		}

		coverAddr, err = uploadLocalFileToCOS(task.CoverFilePath, task.CoverRelativeDir, task.CoverFileName, "image/png")
		if err != nil {
			logger.Error("failed to upload cover to cos", zap.Error(err))
			_ = file_storage.DeletePublicResource(playAddr)
			cleanupLocalFiles(task.VideoFilePath, task.CoverFilePath)
			return
		}
	}

	if ok := video.CreateVideoFactory("").InsertVideoByUID(task.UID, playAddr, task.VideoDesc, coverAddr, task.PrivateStatus); !ok {
		logger.Error("failed to persist uploaded video")
		if useCOSStorage() {
			_ = file_storage.DeletePublicResource(playAddr)
			_ = file_storage.DeletePublicResource(coverAddr)
		}
		cleanupLocalFiles(task.VideoFilePath, task.CoverFilePath)
		return
	}

	updateResult := video.CreateVideoFactory("").Exec(
		`UPDATE tb_users SET aweme_count = COALESCE(aweme_count, 0) + 1 WHERE uid = ?`,
		task.UID,
	)
	if updateResult.Error != nil {
		logger.Error("failed to update user aweme_count", zap.Error(updateResult.Error))
	}

	if useCOSStorage() {
		cleanupLocalFiles(task.VideoFilePath, task.CoverFilePath)
	}

	logger.Info("video upload task completed",
		zap.String("play_addr", playAddr),
		zap.String("cover_addr", coverAddr),
	)
}
