package video

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type DiggModel struct {
	*gorm.DB   `gorm:"-" json:"-"`
	DiggID     int64 `json:"digg_id"`     // bigint
	UID        int64 `json:"uid"`         // bigint
	AwemeID    int64 `json:"aweme_id"`    // bigint
	CreateTime int   `json:"create_time"` // int
}

func CreateDiggFactory(sqlType string) *DiggModel {
	return &DiggModel{DB: model.UseDbConn(sqlType)}
}

func (v *DiggModel) VideoDigg(uid, awemeID int64, action bool) bool {
	cache := newInteractionCache()
	tx := v.DB.Begin()
	if tx.Error != nil {
		variable.ZapLog.Error("VideoDigg failed to start transaction", zap.Error(tx.Error))
		return false
	}

	var authorUID int64
	if err := tx.Raw(`SELECT author_user_id FROM tb_videos WHERE aweme_id = ? LIMIT 1`, awemeID).Scan(&authorUID).Error; err != nil {
		tx.Rollback()
		variable.ZapLog.Error("VideoDigg failed to query video author", zap.Error(err))
		return false
	}
	if authorUID == 0 {
		tx.Rollback()
		variable.ZapLog.Error("VideoDigg failed because video author was not found", zap.Int64("aweme_id", awemeID))
		return false
	}

	currentTime := time.Now().Unix()
	diggSql := `INSERT IGNORE INTO tb_diggs (uid, aweme_id, create_time) VALUES (?, ?, ?);`
	undiggSql := `DELETE FROM tb_diggs WHERE uid = ? AND aweme_id = ?;`

	var result *gorm.DB
	if action {
		result = tx.Exec(diggSql, uid, awemeID, currentTime)
		if result.Error != nil {
			tx.Rollback()
			variable.ZapLog.Error("VideoDigg failed to insert digg row", zap.Error(result.Error))
			return false
		}

		if result.RowsAffected == 0 {
			if err := tx.Commit().Error; err != nil {
				variable.ZapLog.Error("VideoDigg failed to commit no-op like", zap.Error(err))
				return false
			}
			return true
		}

		result = tx.Exec(
			`INSERT INTO tb_statistics (id, digg_count) VALUES (?, 1)
			 ON DUPLICATE KEY UPDATE digg_count = COALESCE(digg_count, 0) + 1`,
			awemeID,
		)
		if result.Error != nil {
			tx.Rollback()
			variable.ZapLog.Error("VideoDigg failed to update statistics", zap.Error(result.Error))
			return false
		}

		result = tx.Exec(
			`UPDATE tb_users
			 SET total_favorited = COALESCE(total_favorited, 0) + 1
			 WHERE uid = ?`,
			authorUID,
		)
	} else {
		result = tx.Exec(undiggSql, uid, awemeID)
		if result.Error != nil {
			tx.Rollback()
			variable.ZapLog.Error("VideoDigg failed to delete digg row", zap.Error(result.Error))
			return false
		}

		if result.RowsAffected == 0 {
			if err := tx.Commit().Error; err != nil {
				variable.ZapLog.Error("VideoDigg failed to commit no-op unlike", zap.Error(err))
				return false
			}
			return true
		}

		result = tx.Exec(
			`UPDATE tb_statistics
			 SET digg_count = GREATEST(COALESCE(digg_count, 0) - 1, 0)
			 WHERE id = ?`,
			awemeID,
		)
		if result.Error != nil {
			tx.Rollback()
			variable.ZapLog.Error("VideoDigg failed to update statistics", zap.Error(result.Error))
			return false
		}

		result = tx.Exec(
			`UPDATE tb_users
			 SET total_favorited = GREATEST(COALESCE(total_favorited, 0) - 1, 0)
			 WHERE uid = ?`,
			authorUID,
		)
	}

	if result.Error != nil {
		tx.Rollback()
		variable.ZapLog.Error("VideoDigg failed to update author favorited count", zap.Error(result.Error))
		return false
	}

	if err := tx.Commit().Error; err != nil {
		variable.ZapLog.Error("VideoDigg failed to commit transaction", zap.Error(err))
		return false
	}

	if action {
		cache.incrStat(awemeID, "digg_count", 1)
		cache.addDiggUser(awemeID, uid)
		cache.addUserLikedVideo(uid, awemeID)
		cache.incrUserTotalFavorited(authorUID, 1)
	} else {
		cache.incrStat(awemeID, "digg_count", -1)
		cache.removeDiggUser(awemeID, uid)
		cache.removeUserLikedVideo(uid, awemeID)
		cache.incrUserTotalFavorited(authorUID, -1)
	}

	return true
}
