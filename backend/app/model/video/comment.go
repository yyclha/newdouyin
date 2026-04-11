package video

import (
	"douyin-backend/app/model"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type CommentModel struct {
	*gorm.DB        `gorm:"-" json:"-"`
	CommentID       int64  `json:"comment_id"`        // bigint
	CreateTime      int    `json:"create_time"`       // int
	IPLocation      string `json:"ip_location"`       // varchar(100)
	AwemeID         int64  `json:"aweme_id"`          // bigint
	Content         string `json:"content"`           // text
	IsAuthorDigged  bool   `json:"is_author_digged"`  // tinyint(1)
	IsFolded        bool   `json:"is_folded"`         // tinyint(1)
	IsHot           bool   `json:"is_hot"`            // tinyint(1)
	UserBuried      bool   `json:"user_buried"`       // tinyint(1)
	UserDigged      int    `json:"user_digged"`       // int
	DiggCount       int64  `json:"digg_count"`        // bigint
	UserID          int64  `json:"user_id"`           // bigint
	SecUID          string `json:"sec_uid"`           // text
	ShortUserID     int64  `json:"short_user_id"`     // bigint
	UserUniqueID    string `json:"user_unique_id"`    // varchar(255)
	UserSignature   string `json:"user_signature"`    // text
	Nickname        string `json:"nickname"`          // varchar(100)
	Avatar          string `json:"avatar"`            // text
	SubCommentCount int64  `json:"sub_comment_count"` // bigint
	LastModifyTS    int64  `json:"last_modify_ts"`    // bigint
}

func CreateCommentFactory(sqlType string) *CommentModel {
	return &CommentModel{DB: model.UseDbConn(sqlType)}
}

func (c *CommentModel) GetComments(awemeID, currentUID, pageNo, pageSize int64) (comments []Comment, total int64, hasMore bool, ok bool) {
	if pageNo < 0 {
		pageNo = 0
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	if err := c.Raw(`SELECT COUNT(1) FROM tb_comments WHERE aweme_id = ?`, awemeID).Scan(&total).Error; err != nil {
		return nil, 0, false, false
	}
	if total == 0 {
		return []Comment{}, 0, false, true
	}

	offset := pageNo * pageSize
	if offset >= total {
		return []Comment{}, total, false, true
	}

	cache := newInteractionCache()
	if offset < videoCommentIndexLimit {
		if cachedComments, hit := cache.getCommentsPage(awemeID, offset, pageSize, total); hit {
			c.markCommentsUserDigged(cachedComments, currentUID)
			hasMore = offset+int64(len(cachedComments)) < total
			return cachedComments, total, hasMore, true
		}
		if recentComments, loaded := c.loadRecentCommentsForCache(awemeID); loaded {
			cache.setComments(awemeID, recentComments)
			if cachedComments, hit := cache.getCommentsPage(awemeID, offset, pageSize, total); hit {
				c.markCommentsUserDigged(cachedComments, currentUID)
				hasMore = offset+int64(len(cachedComments)) < total
				return cachedComments, total, hasMore, true
			}
		}
	}

	sql := `
		SELECT
			comment_id,
			create_time,
			ip_location,
			aweme_id,
			content,
			is_author_digged,
			is_folded,
			is_hot,
			user_buried,
			user_digged,
			digg_count,
			user_id,
			sec_uid,
			short_user_id,
			user_unique_id,
			user_signature,
			nickname,
			avatar,
			sub_comment_count,
			last_modify_ts
		FROM tb_comments
		WHERE aweme_id = ?
		ORDER BY create_time DESC
		LIMIT ? OFFSET ?;
	`
	comments = []Comment{}
	result := c.Raw(sql, awemeID, pageSize, offset).Scan(&comments)
	if result.Error != nil {
		return nil, 0, false, false
	}

	c.markCommentsUserDigged(comments, currentUID)
	if offset < videoCommentIndexLimit {
		if recentComments, loaded := c.loadRecentCommentsForCache(awemeID); loaded {
			cache.setComments(awemeID, recentComments)
		}
	}
	hasMore = offset+int64(len(comments)) < total
	return comments, total, hasMore, true
}

func (c *CommentModel) loadRecentCommentsForCache(awemeID int64) (comments []Comment, ok bool) {
	sql := `
		SELECT
			comment_id,
			create_time,
			ip_location,
			aweme_id,
			content,
			is_author_digged,
			is_folded,
			is_hot,
			user_buried,
			user_digged,
			digg_count,
			user_id,
			sec_uid,
			short_user_id,
			user_unique_id,
			user_signature,
			nickname,
			avatar,
			sub_comment_count,
			last_modify_ts
		FROM tb_comments
		WHERE aweme_id = ?
		ORDER BY create_time DESC
		LIMIT 500;
	`
	comments = []Comment{}
	result := c.Raw(sql, awemeID).Scan(&comments)
	if result.Error != nil {
		return nil, false
	}
	return comments, true
}

func (c *CommentModel) markCommentsUserDigged(comments []Comment, currentUID int64) {
	if currentUID <= 0 || len(comments) == 0 {
		return
	}

	commentIDs := make([]int64, 0, len(comments))
	for _, comment := range comments {
		if comment.CommentID > 0 {
			commentIDs = append(commentIDs, comment.CommentID)
		}
	}
	if len(commentIDs) == 0 {
		return
	}

	var diggedRows []struct {
		CommentID int64 `json:"comment_id"`
	}
	if err := c.Table("tb_comment_diggs").
		Select("comment_id").
		Where("uid = ? AND comment_id IN ?", currentUID, commentIDs).
		Scan(&diggedRows).Error; err != nil {
		return
	}

	diggedMap := make(map[int64]struct{}, len(diggedRows))
	for _, row := range diggedRows {
		diggedMap[row.CommentID] = struct{}{}
	}

	for i := range comments {
		if _, exists := diggedMap[comments[i].CommentID]; exists {
			comments[i].UserDigged = 1
		} else {
			comments[i].UserDigged = 0
		}
	}
}

func (c *CommentModel) VideoComment(uid, awemeID int64, ipLocation, content, shortID, uniqueID, signature, nickname, avatar string) (commentID int64, ok bool) {
	currentTime := time.Now().Unix()

	var shortIDInt int64
	if shortID != "" {
		parsedID, err := strconv.ParseInt(shortID, 10, 64)
		if err == nil {
			shortIDInt = parsedID
		}
	}

	tx := c.Begin()
	if tx.Error != nil {
		return 0, false
	}

	insertCommentSQL := `
		INSERT INTO tb_comments (
			create_time,
			ip_location,
			aweme_id,
			content,
			user_id,
			short_user_id,
			user_unique_id,
			user_signature,
			nickname,
			avatar,
			last_modify_ts
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	insertResult := tx.Exec(insertCommentSQL, currentTime, ipLocation, awemeID, content, uid, shortIDInt, uniqueID, signature, nickname, avatar, currentTime)
	if insertResult.Error != nil || insertResult.RowsAffected == 0 {
		tx.Rollback()
		return 0, false
	}

	if err := tx.Raw(`SELECT LAST_INSERT_ID()`).Scan(&commentID).Error; err != nil || commentID == 0 {
		tx.Rollback()
		return 0, false
	}

	updateStatisticsSQL := `
		UPDATE tb_statistics
		SET comment_count = COALESCE(comment_count, 0) + 1
		WHERE id = ?;
	`
	updateResult := tx.Exec(updateStatisticsSQL, awemeID)
	if updateResult.Error != nil || updateResult.RowsAffected == 0 {
		tx.Rollback()
		return 0, false
	}

	if err := tx.Commit().Error; err != nil {
		return 0, false
	}

	cache := newInteractionCache()
	cache.incrStat(awemeID, "comment_count", 1)
	cache.prependComment(awemeID, buildCommentForCache(uid, awemeID, ipLocation, content, shortIDInt, uniqueID, signature, nickname, avatar, currentTime, commentID))

	return commentID, true
}
