package video

import (
	"douyin-backend/app/model"
	"gorm.io/gorm"
	"strings"
	"time"
)

// ShareModel 封装视频分享记录的数据库操作。
type ShareModel struct {
	*gorm.DB   `gorm:"-" json:"-"`
	DiggID     int64 `json:"digg_id"`     // bigint
	UID        int64 `json:"uid"`         // bigint
	AwemeID    int64 `json:"aweme_id"`    // bigint
	CreateTime int   `json:"create_time"` // int
}

// CreateShareFactory 创建带数据库连接的分享模型实例。
func CreateShareFactory(sqlType string) *ShareModel {
	return &ShareModel{DB: model.UseDbConn(sqlType)}
}

// VideoShare 为分享列表中的每个目标用户写入一条分享记录。
func (s *ShareModel) VideoShare(uid, awemeID int64, message string, shareUidList string) bool {
	currentTime := time.Now().Unix()
	sql := `
		INSERT INTO tb_shares (src_uid, dst_uid, aweme_id, message, create_time) VALUES (?, ?, ?, ?, ?);`
	cnt := 0
	DstUidList := strings.Split(shareUidList, ",")
	for _, DstUid := range DstUidList {
		result := s.Exec(sql, uid, DstUid, awemeID, message, currentTime)
		if result.RowsAffected > 0 {
			cnt++
		}
	}
	if cnt == len(DstUidList) {
		return true
	} else {
		return false
	}
}
