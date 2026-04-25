package upload_file

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	"douyin-backend/app/utils/auth"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	VideoUploadTaskStatusPending    = "pending"
	VideoUploadTaskStatusProcessing = "processing"
	VideoUploadTaskStatusSucceeded  = "succeeded"
	VideoUploadTaskStatusFailed     = "failed"

	defaultVideoUploadMaxRetries               = 3
	defaultVideoUploadProcessingTimeoutSeconds = 900
)

// PersistedVideoUploadTask stores the post-merge video processing task.
type PersistedVideoUploadTask struct {
	ID               int64  `gorm:"column:id" json:"-"`
	TaskID           string `gorm:"column:task_id" json:"taskId"`
	UploadID         string `gorm:"column:upload_id" json:"uploadId"`
	UID              int64  `gorm:"column:uid" json:"uid"`
	Status           string `gorm:"column:status" json:"status"`
	RetryCount       int    `gorm:"column:retry_count" json:"retryCount"`
	MaxRetries       int    `gorm:"column:max_retries" json:"maxRetries"`
	NextRetryAt      int64  `gorm:"column:next_retry_at" json:"nextRetryAt"`
	ErrorMessage     string `gorm:"column:error_message" json:"errorMessage"`
	VideoFilePath    string `gorm:"column:video_file_path" json:"-"`
	CoverFilePath    string `gorm:"column:cover_file_path" json:"-"`
	VideoRelativeDir string `gorm:"column:video_relative_dir" json:"-"`
	CoverRelativeDir string `gorm:"column:cover_relative_dir" json:"-"`
	VideoFileName    string `gorm:"column:video_file_name" json:"-"`
	CoverFileName    string `gorm:"column:cover_file_name" json:"-"`
	ContentType      string `gorm:"column:content_type" json:"-"`
	VideoDesc        string `gorm:"column:video_desc" json:"videoDesc"`
	PrivateStatus    int    `gorm:"column:private_status" json:"privateStatus"`
	PlayAddr         string `gorm:"column:play_addr" json:"playAddr"`
	CoverAddr        string `gorm:"column:cover_addr" json:"coverAddr"`
	CreatedAt        int64  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt        int64  `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 执行对象方法逻辑。
func (PersistedVideoUploadTask) TableName() string {
	return "video_upload_tasks"
}

// videoUploadTaskDB 执行业务处理。
func videoUploadTaskDB() *gorm.DB {
	return model.UseDbConn("")
}

// EnsureVideoUploadTaskTable 确保视频上传补偿任务表已创建。
func EnsureVideoUploadTaskTable() error {
	db := videoUploadTaskDB()
	if db == nil {
		return errors.New("database is unavailable")
	}

	return db.Exec(`
CREATE TABLE IF NOT EXISTS video_upload_tasks (
  id bigint NOT NULL AUTO_INCREMENT,
  task_id varchar(64) NOT NULL,
  upload_id varchar(128) NOT NULL,
  uid bigint NOT NULL,
  status varchar(32) NOT NULL DEFAULT 'pending',
  retry_count int NOT NULL DEFAULT 0,
  max_retries int NOT NULL DEFAULT 3,
  next_retry_at bigint NOT NULL DEFAULT 0,
  error_message text,
  video_file_path varchar(512) NOT NULL,
  cover_file_path varchar(512) NOT NULL,
  video_relative_dir varchar(255) NOT NULL,
  cover_relative_dir varchar(255) NOT NULL,
  video_file_name varchar(255) NOT NULL,
  cover_file_name varchar(255) NOT NULL,
  content_type varchar(128) NOT NULL,
  video_desc text,
  private_status int NOT NULL DEFAULT 0,
  play_addr text,
  cover_addr text,
  created_at bigint NOT NULL,
  updated_at bigint NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_video_upload_tasks_task_id (task_id),
  UNIQUE KEY uk_video_upload_tasks_upload_uid (upload_id, uid),
  KEY idx_video_upload_tasks_status_retry (status, next_retry_at, retry_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`).Error
}

// videoUploadMaxRetries 执行业务处理。
func videoUploadMaxRetries() int {
	maxRetries := variable.ConfigYml.GetInt("FileUploadSetting.VideoAsync.MaxRetries")
	if maxRetries <= 0 {
		return defaultVideoUploadMaxRetries
	}
	return maxRetries
}

// videoUploadProcessingTimeoutSeconds 执行业务处理。
func videoUploadProcessingTimeoutSeconds() int64 {
	timeoutSeconds := variable.ConfigYml.GetInt64("FileUploadSetting.VideoAsync.ProcessingTimeoutSeconds")
	if timeoutSeconds <= 0 {
		return defaultVideoUploadProcessingTimeoutSeconds
	}
	return timeoutSeconds
}

// createVideoUploadTask 执行业务处理。
func createVideoUploadTask(task VideoUploadTask) (*PersistedVideoUploadTask, error) {
	db := videoUploadTaskDB()
	if db == nil {
		return nil, errors.New("database is unavailable")
	}

	now := time.Now().Unix()
	record := PersistedVideoUploadTask{
		TaskID:           task.TaskID,
		UploadID:         task.UploadID,
		UID:              task.UID,
		Status:           VideoUploadTaskStatusPending,
		RetryCount:       0,
		MaxRetries:       videoUploadMaxRetries(),
		NextRetryAt:      now,
		VideoFilePath:    task.VideoFilePath,
		CoverFilePath:    task.CoverFilePath,
		VideoRelativeDir: task.VideoRelativeDir,
		CoverRelativeDir: task.CoverRelativeDir,
		VideoFileName:    task.VideoFileName,
		CoverFileName:    task.CoverFileName,
		ContentType:      task.ContentType,
		VideoDesc:        task.VideoDesc,
		PrivateStatus:    task.PrivateStatus,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var existing PersistedVideoUploadTask
		result := tx.Where("upload_id = ? AND uid = ?", task.UploadID, task.UID).First(&existing)
		if result.Error == nil {
			record = existing
			return nil
		}
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		return tx.Create(&record).Error
	})

	if err != nil {
		return nil, err
	}
	return &record, nil
}

// getVideoUploadTaskForUser 执行业务处理。
func getVideoUploadTaskForUser(taskID string, uid int64) (*PersistedVideoUploadTask, error) {
	db := videoUploadTaskDB()
	if db == nil {
		return nil, errors.New("database is unavailable")
	}

	var task PersistedVideoUploadTask
	err := db.Where("task_id = ? AND uid = ?", taskID, uid).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// getVideoUploadTaskByUploadIDForUser 执行业务处理。
func getVideoUploadTaskByUploadIDForUser(uploadID string, uid int64) (*PersistedVideoUploadTask, error) {
	db := videoUploadTaskDB()
	if db == nil {
		return nil, errors.New("database is unavailable")
	}

	var task PersistedVideoUploadTask
	err := db.Where("upload_id = ? AND uid = ?", uploadID, uid).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// buildVideoUploadTaskResponse 执行业务处理。
func buildVideoUploadTaskResponse(task *PersistedVideoUploadTask) gin.H {
	return gin.H{
		"taskId":        task.TaskID,
		"uploadId":      task.UploadID,
		"status":        task.Status,
		"retryCount":    task.RetryCount,
		"maxRetries":    task.MaxRetries,
		"nextRetryAt":   task.NextRetryAt,
		"errorMessage":  task.ErrorMessage,
		"videoDesc":     task.VideoDesc,
		"privateStatus": task.PrivateStatus,
		"playAddr":      task.PlayAddr,
		"coverAddr":     task.CoverAddr,
		"createdAt":     task.CreatedAt,
		"updatedAt":     task.UpdatedAt,
	}
}

// GetVideoUploadTaskStatus 执行业务处理。
func GetVideoUploadTaskStatus(context *gin.Context) (bool, interface{}, string) {
	taskID := strings.TrimSpace(context.GetString(consts.ValidatorPrefix + "task_id"))
	if taskID == "" {
		return false, nil, "task_id is required"
	}

	task, err := getVideoUploadTaskForUser(taskID, auth.GetUidFromToken(context))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, "video upload task not found"
		}
		return false, nil, "failed to get video upload task"
	}

	return true, buildVideoUploadTaskResponse(task), ""
}

// claimNextVideoUploadTask 执行业务处理。
func claimNextVideoUploadTask() (*PersistedVideoUploadTask, error) {
	db := videoUploadTaskDB()
	if db == nil {
		return nil, errors.New("database is unavailable")
	}

	now := time.Now().Unix()
	staleProcessingBefore := now - videoUploadProcessingTimeoutSeconds()
	var task PersistedVideoUploadTask
	err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where(`(
					(status = ? OR status = ?) AND retry_count < max_retries AND next_retry_at <= ?
				) OR (
					status = ? AND retry_count < max_retries AND updated_at <= ?
				)`,
				VideoUploadTaskStatusPending,
				VideoUploadTaskStatusFailed,
				now,
				VideoUploadTaskStatusProcessing,
				staleProcessingBefore,
			).
			Order("next_retry_at ASC, id ASC").
			First(&task)
		if result.Error != nil {
			return result.Error
		}

		updated := tx.Model(&PersistedVideoUploadTask{}).
			Where(`id = ? AND (
				status = ? OR status = ? OR (status = ? AND updated_at <= ?)
			)`,
				task.ID,
				VideoUploadTaskStatusPending,
				VideoUploadTaskStatusFailed,
				VideoUploadTaskStatusProcessing,
				staleProcessingBefore,
			).
			Updates(map[string]interface{}{
				"status":        VideoUploadTaskStatusProcessing,
				"error_message": "",
				"updated_at":    now,
			})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		task.Status = VideoUploadTaskStatusProcessing
		task.ErrorMessage = ""
		task.UpdatedAt = now
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// markVideoUploadTaskSucceeded 执行业务处理。
func markVideoUploadTaskSucceeded(task *PersistedVideoUploadTask, playAddr, coverAddr string) error {
	db := videoUploadTaskDB()
	if db == nil {
		return errors.New("database is unavailable")
	}

	return db.Model(&PersistedVideoUploadTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]interface{}{
			"status":        VideoUploadTaskStatusSucceeded,
			"play_addr":     playAddr,
			"cover_addr":    coverAddr,
			"error_message": "",
			"updated_at":    time.Now().Unix(),
		}).Error
}

// markVideoUploadTaskFailed 执行业务处理。
func markVideoUploadTaskFailed(task *PersistedVideoUploadTask, err error) error {
	db := videoUploadTaskDB()
	if db == nil {
		return errors.New("database is unavailable")
	}

	now := time.Now().Unix()
	retryCount := task.RetryCount + 1
	status := VideoUploadTaskStatusFailed
	nextRetryAt := now + int64(1<<minInt(retryCount, 6))*60
	if retryCount >= task.MaxRetries {
		nextRetryAt = 0
	}

	return db.Model(&PersistedVideoUploadTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]interface{}{
			"status":        status,
			"retry_count":   retryCount,
			"next_retry_at": nextRetryAt,
			"error_message": err.Error(),
			"updated_at":    now,
		}).Error
}

// minInt 执行业务处理。
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
