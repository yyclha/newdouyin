package video

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	videodiggasync "douyin-backend/app/service"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	videoDiggOutboxStatusPending   = "pending"
	videoDiggOutboxStatusPublished = "published"
	videoDiggOutboxStatusFailed    = "failed"

	defaultVideoDiggOutboxMaxRetries         = 5
	defaultVideoDiggOutboxRetryDelaySeconds  = 10
	defaultVideoDiggOutboxDispatchInterval   = 5
	defaultVideoDiggOutboxDispatchBatchLimit = 50
)

// VideoDiggOutboxEvent 持久化 Redis 点赞成功后的 MQ 待发布事件。
type VideoDiggOutboxEvent struct {
	ID           int64  `gorm:"column:id"`
	EventKey     string `gorm:"column:event_key"`
	UID          int64  `gorm:"column:uid"`
	AwemeID      int64  `gorm:"column:aweme_id"`
	AuthorUID    int64  `gorm:"column:author_uid"`
	Action       bool   `gorm:"column:action"`
	Version      int64  `gorm:"column:version"`
	DiggCount    int64  `gorm:"column:digg_count"`
	OccurredAt   int64  `gorm:"column:occurred_at"`
	Status       string `gorm:"column:status"`
	RetryCount   int    `gorm:"column:retry_count"`
	MaxRetries   int    `gorm:"column:max_retries"`
	NextRetryAt  int64  `gorm:"column:next_retry_at"`
	ErrorMessage string `gorm:"column:error_message"`
	CreatedUnix  int64  `gorm:"column:created_at"`
	UpdatedUnix  int64  `gorm:"column:updated_at"`
}

// TableName 返回点赞 outbox 表名。
func (VideoDiggOutboxEvent) TableName() string {
	return "video_digg_outbox_events"
}

// EnsureVideoDiggOutboxTable 确保点赞 MQ outbox 表存在。
func EnsureVideoDiggOutboxTable() error {
	db := modelDB()
	if db == nil {
		return errors.New("database is unavailable")
	}

	return db.Exec(`
CREATE TABLE IF NOT EXISTS video_digg_outbox_events (
  id bigint NOT NULL AUTO_INCREMENT,
  event_key varchar(128) NOT NULL,
  uid bigint NOT NULL,
  aweme_id bigint NOT NULL,
  author_uid bigint NOT NULL,
  action tinyint(1) NOT NULL,
  version bigint NOT NULL DEFAULT 0,
  digg_count bigint NOT NULL DEFAULT 0,
  occurred_at bigint NOT NULL,
  status varchar(32) NOT NULL DEFAULT 'pending',
  retry_count int NOT NULL DEFAULT 0,
  max_retries int NOT NULL DEFAULT 5,
  next_retry_at bigint NOT NULL DEFAULT 0,
  error_message text,
  created_at bigint NOT NULL,
  updated_at bigint NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_video_digg_outbox_event_key (event_key),
  KEY idx_video_digg_outbox_status_retry (status, next_retry_at, retry_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`).Error
}

// StartVideoDiggOutboxDispatcher 启动点赞 outbox 后台补投递任务。
func StartVideoDiggOutboxDispatcher() {
	go func() {
		interval := variable.ConfigYml.GetDuration("RabbitMq.VideoDigg.OutboxDispatchIntervalSec")
		if interval <= 0 {
			interval = defaultVideoDiggOutboxDispatchInterval
		}

		dispatchVideoDiggOutboxEvents()
		ticker := time.NewTicker(interval * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			dispatchVideoDiggOutboxEvents()
		}
	}()
}

func createVideoDiggOutboxEvent(event videodiggasync.VideoDiggEvent) (*VideoDiggOutboxEvent, error) {
	db := modelDB()
	if db == nil {
		return nil, errors.New("database is unavailable")
	}

	now := time.Now().Unix()
	record := VideoDiggOutboxEvent{
		EventKey:    videoDiggOutboxEventKey(event),
		UID:         event.UID,
		AwemeID:     event.AwemeID,
		AuthorUID:   event.AuthorUID,
		Action:      event.Action,
		Version:     event.Version,
		DiggCount:   event.DiggCount,
		OccurredAt:  event.OccurredAt,
		Status:      videoDiggOutboxStatusPending,
		MaxRetries:  videoDiggOutboxMaxRetries(),
		NextRetryAt: now,
		CreatedUnix: now,
		UpdatedUnix: now,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var existing VideoDiggOutboxEvent
		result := tx.Where("event_key = ?", record.EventKey).First(&existing)
		if result.Error == nil && result.RowsAffected > 0 {
			record = existing
			return nil
		}
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		return tx.Create(&record).Error
	})
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func dispatchVideoDiggOutboxEvents() {
	db := modelDB()
	if db == nil {
		return
	}

	now := time.Now().Unix()
	var events []VideoDiggOutboxEvent
	err := db.Where(
		"(status = ? OR status = ?) AND retry_count < max_retries AND next_retry_at <= ?",
		videoDiggOutboxStatusPending,
		videoDiggOutboxStatusFailed,
		now,
	).Order("id ASC").Limit(videoDiggOutboxDispatchBatchLimit()).Find(&events).Error
	if err != nil {
		variable.ZapLog.Error("failed to load video digg outbox events", zap.Error(err))
		return
	}

	for _, event := range events {
		publishVideoDiggOutboxEvent(event)
	}
}

func publishVideoDiggOutboxEvent(record VideoDiggOutboxEvent) {
	event := videodiggasync.VideoDiggEvent{
		UID:        record.UID,
		AwemeID:    record.AwemeID,
		AuthorUID:  record.AuthorUID,
		Action:     record.Action,
		Version:    record.Version,
		DiggCount:  record.DiggCount,
		OccurredAt: record.OccurredAt,
	}

	if err := videodiggasync.PublishVideoDiggEvent(event); err != nil {
		_ = markVideoDiggOutboxFailed(record.ID, err)
		return
	}
	_ = markVideoDiggOutboxPublished(record.ID)
}

func markVideoDiggOutboxPublished(id int64) error {
	db := modelDB()
	if db == nil {
		return errors.New("database is unavailable")
	}
	now := time.Now().Unix()
	return db.Model(&VideoDiggOutboxEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        videoDiggOutboxStatusPublished,
		"error_message": "",
		"updated_at":    now,
	}).Error
}

func markVideoDiggOutboxFailed(id int64, publishErr error) error {
	db := modelDB()
	if db == nil {
		return errors.New("database is unavailable")
	}

	var record VideoDiggOutboxEvent
	if err := db.Where("id = ?", id).First(&record).Error; err != nil {
		return err
	}

	now := time.Now().Unix()
	retryCount := record.RetryCount + 1
	nextRetryAt := now + videoDiggOutboxRetryDelaySeconds()
	return db.Model(&VideoDiggOutboxEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        videoDiggOutboxStatusFailed,
		"retry_count":   retryCount,
		"next_retry_at": nextRetryAt,
		"error_message": publishErr.Error(),
		"updated_at":    now,
	}).Error
}

func videoDiggOutboxEventKey(event videodiggasync.VideoDiggEvent) string {
	return fmt.Sprintf("%d:%d:%d", event.UID, event.AwemeID, event.Version)
}

func videoDiggOutboxMaxRetries() int {
	maxRetries := variable.ConfigYml.GetInt("RabbitMq.VideoDigg.OutboxMaxRetries")
	if maxRetries <= 0 {
		return defaultVideoDiggOutboxMaxRetries
	}
	return maxRetries
}

func videoDiggOutboxRetryDelaySeconds() int64 {
	delay := variable.ConfigYml.GetInt64("RabbitMq.VideoDigg.OutboxRetryDelaySeconds")
	if delay <= 0 {
		return defaultVideoDiggOutboxRetryDelaySeconds
	}
	return delay
}

func videoDiggOutboxDispatchBatchLimit() int {
	limit := variable.ConfigYml.GetInt("RabbitMq.VideoDigg.OutboxDispatchBatchLimit")
	if limit <= 0 {
		return defaultVideoDiggOutboxDispatchBatchLimit
	}
	return limit
}

func modelDB() *gorm.DB {
	return model.UseDbConn("")
}
