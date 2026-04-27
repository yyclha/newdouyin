package service

// VideoDiggEvent 定义视频点赞异步事件的消息体结构。
type VideoDiggEvent struct {
	UID        int64 `json:"uid"`
	AwemeID    int64 `json:"aweme_id"`
	AuthorUID  int64 `json:"author_uid"`
	Action     bool  `json:"action"`
	Version    int64 `json:"version"`
	DiggCount  int64 `json:"digg_count"`
	OccurredAt int64 `json:"occurred_at"`
	RetryCount int   `json:"retry_count"`
}
