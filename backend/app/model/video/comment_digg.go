package video

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CommentDiggModel 封装评论点赞和取消点赞的数据库操作。
type CommentDiggModel struct {
	*gorm.DB `gorm:"-" json:"-"`
}

// CreateCommentDiggFactory 创建带数据库连接的评论点赞模型实例。
func CreateCommentDiggFactory(sqlType string) *CommentDiggModel {
	return &CommentDiggModel{DB: model.UseDbConn(sqlType)}
}

// CommentDigg 对目标评论执行点赞或取消点赞操作。
func (c *CommentDiggModel) CommentDigg(uid, commentID int64, action bool) bool {
	tx := c.DB.Begin()
	if tx.Error != nil {
		variable.ZapLog.Error("CommentDigg failed to start transaction", zap.Error(tx.Error))
		return false
	}

	var comment struct {
		AwemeID int64 `json:"aweme_id"`
	}
	if err := tx.Raw(`SELECT aweme_id FROM tb_comments WHERE comment_id = ? LIMIT 1`, commentID).Scan(&comment).Error; err != nil {
		tx.Rollback()
		variable.ZapLog.Error("CommentDigg failed to query comment", zap.Error(err))
		return false
	}
	if comment.AwemeID == 0 {
		tx.Rollback()
		return false
	}

	var result *gorm.DB
	if action {
		result = tx.Exec(`INSERT IGNORE INTO tb_comment_diggs (uid, comment_id, create_time) VALUES (?, ?, UNIX_TIMESTAMP())`, uid, commentID)
		if result.Error != nil {
			tx.Rollback()
			variable.ZapLog.Error("CommentDigg insert failed", zap.Error(result.Error))
			return false
		}

		if result.RowsAffected == 0 {
			if err := tx.Commit().Error; err != nil {
				variable.ZapLog.Error("CommentDigg commit no-op like failed", zap.Error(err))
				return false
			}
			return true
		}

		result = tx.Exec(`UPDATE tb_comments SET digg_count = COALESCE(digg_count, 0) + 1 WHERE comment_id = ?`, commentID)
	} else {
		result = tx.Exec(`DELETE FROM tb_comment_diggs WHERE uid = ? AND comment_id = ?`, uid, commentID)
		if result.Error != nil {
			tx.Rollback()
			variable.ZapLog.Error("CommentDigg delete failed", zap.Error(result.Error))
			return false
		}

		if result.RowsAffected == 0 {
			if err := tx.Commit().Error; err != nil {
				variable.ZapLog.Error("CommentDigg commit no-op unlike failed", zap.Error(err))
				return false
			}
			return true
		}

		result = tx.Exec(`UPDATE tb_comments SET digg_count = GREATEST(COALESCE(digg_count, 0) - 1, 0) WHERE comment_id = ?`, commentID)
	}

	if result.Error != nil {
		tx.Rollback()
		variable.ZapLog.Error("CommentDigg failed to update comment", zap.Error(result.Error))
		return false
	}

	if err := tx.Commit().Error; err != nil {
		variable.ZapLog.Error("CommentDigg failed to commit transaction", zap.Error(err))
		return false
	}

	cache := newInteractionCache()
	cache.invalidateCommentItem(commentID)
	cache.invalidateCommentList(comment.AwemeID)
	return true
}
