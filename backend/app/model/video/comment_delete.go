package video

import (
	"douyin-backend/app/global/variable"
	"go.uber.org/zap"
)

// DeleteComment 删除当前用户自己的评论，并更新相关统计数据。
func (c *CommentModel) DeleteComment(uid, commentID int64) bool {
	tx := c.DB.Begin()
	if tx.Error != nil {
		variable.ZapLog.Error("DeleteComment failed to start transaction", zap.Error(tx.Error))
		return false
	}

	var comment struct {
		AwemeID int64 `json:"aweme_id"`
		UserID  int64 `json:"user_id"`
	}
	if err := tx.Raw(`SELECT aweme_id, user_id FROM tb_comments WHERE comment_id = ? LIMIT 1`, commentID).Scan(&comment).Error; err != nil {
		tx.Rollback()
		variable.ZapLog.Error("DeleteComment failed to query comment", zap.Error(err), zap.Int64("comment_id", commentID))
		return false
	}
	if comment.AwemeID == 0 || comment.UserID == 0 || comment.UserID != uid {
		tx.Rollback()
		return false
	}

	if result := tx.Exec(`DELETE FROM tb_comment_diggs WHERE comment_id = ?`, commentID); result.Error != nil {
		tx.Rollback()
		variable.ZapLog.Error("DeleteComment failed to delete comment diggs", zap.Error(result.Error), zap.Int64("comment_id", commentID))
		return false
	}

	result := tx.Exec(`DELETE FROM tb_comments WHERE comment_id = ? AND user_id = ?`, commentID, uid)
	if result.Error != nil || result.RowsAffected == 0 {
		tx.Rollback()
		if result.Error != nil {
			variable.ZapLog.Error("DeleteComment failed to delete comment", zap.Error(result.Error), zap.Int64("comment_id", commentID))
		}
		return false
	}

	updateStats := tx.Exec(`
		UPDATE tb_statistics
		SET comment_count = GREATEST(COALESCE(comment_count, 0) - 1, 0)
		WHERE id = ?`,
		comment.AwemeID,
	)
	if updateStats.Error != nil {
		tx.Rollback()
		variable.ZapLog.Error("DeleteComment failed to update statistics", zap.Error(updateStats.Error), zap.Int64("aweme_id", comment.AwemeID))
		return false
	}

	if err := tx.Commit().Error; err != nil {
		variable.ZapLog.Error("DeleteComment failed to commit transaction", zap.Error(err), zap.Int64("comment_id", commentID))
		return false
	}

	cache := newInteractionCache()
	cache.invalidateStats(comment.AwemeID)
	cache.invalidateCommentList(comment.AwemeID)
	cache.invalidateCommentItem(commentID)

	return true
}
