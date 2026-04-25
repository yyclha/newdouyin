package video

import (
	"douyin-backend/app/model"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CommentModel 封装评论相关的数据库读写操作。
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

// CreateCommentFactory 创建带数据库连接的评论模型实例。
func CreateCommentFactory(sqlType string) *CommentModel {
	return &CommentModel{DB: model.UseDbConn(sqlType)}
}

// commentSelectFields 执行业务处理。
func commentSelectFields() string {
	return `
		SELECT
			tc.comment_id,
			tc.create_time,
			tc.ip_location,
			tc.aweme_id,
			tc.content,
			tc.is_author_digged,
			tc.is_folded,
			tc.is_hot,
			tc.user_buried,
			tc.user_digged,
			tc.digg_count,
			tc.user_id,
			tc.sec_uid,
			COALESCE(NULLIF(tu.short_id, 0), tc.short_user_id, 0) AS short_user_id,
			COALESCE(NULLIF(tu.unique_id, ''), tc.user_unique_id, '') AS user_unique_id,
			COALESCE(NULLIF(tu.signature, ''), tc.user_signature, '') AS user_signature,
			COALESCE(NULLIF(tu.nickname, ''), tc.nickname, '') AS nickname,
			CASE
				WHEN tu.avatar_small IS NOT NULL AND tu.avatar_small <> '' THEN
					CASE
						WHEN JSON_VALID(tu.avatar_small) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(tu.avatar_small, '$.url_list[0]')), '')
						ELSE tu.avatar_small
					END
				WHEN tc.avatar IS NOT NULL AND tc.avatar <> '' AND tc.avatar NOT LIKE '%/aweme/v1/play/%' THEN tc.avatar
				ELSE ''
			END AS avatar,
			tc.sub_comment_count,
			tc.last_modify_ts
		FROM tb_comments AS tc
		LEFT JOIN tb_users AS tu ON tc.user_id = tu.uid
	`
}

// GetComments 分页查询目标视频的评论列表。
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
	}

	sql := commentSelectFields() + `
		WHERE tc.aweme_id = ?
		ORDER BY tc.create_time DESC
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

// loadRecentCommentsForCache 执行对象方法逻辑。
func (c *CommentModel) loadRecentCommentsForCache(awemeID int64) (comments []Comment, ok bool) {
	sql := commentSelectFields() + `
		WHERE tc.aweme_id = ?
		ORDER BY tc.create_time DESC
		LIMIT 500;
	`
	comments = []Comment{}
	result := c.Raw(sql, awemeID).Scan(&comments)
	if result.Error != nil {
		return nil, false
	}
	return comments, true
}

// commentAuthorProfile 定义业务数据结构。
type commentAuthorProfile struct {
	ShortID   int64  `json:"short_id"`
	UniqueID  string `json:"unique_id"`
	Signature string `json:"signature"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar_small"`
}

// avatarPayload 定义业务数据结构。
type avatarPayload struct {
	URLList []string `json:"url_list"`
}

// loadCommentAuthorProfile 执行对象方法逻辑。
func (c *CommentModel) loadCommentAuthorProfile(uid int64) (profile commentAuthorProfile, ok bool) {
	sql := `
		SELECT
			COALESCE(short_id, 0) AS short_id,
			COALESCE(unique_id, '') AS unique_id,
			COALESCE(signature, '') AS signature,
			COALESCE(nickname, '') AS nickname,
			COALESCE(avatar_small, '') AS avatar_small
		FROM tb_users
		WHERE uid = ?
		LIMIT 1;
	`
	if err := c.Raw(sql, uid).Scan(&profile).Error; err != nil {
		return commentAuthorProfile{}, false
	}
	profile.Avatar = parseAvatarURL(profile.Avatar)
	return profile, true
}

// parseAvatarURL 执行业务处理。
func parseAvatarURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}

	payload := avatarPayload{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return ""
	}
	if len(payload.URLList) == 0 {
		return ""
	}
	return strings.TrimSpace(payload.URLList[0])
}

// markCommentsUserDigged 执行对象方法逻辑。
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

// VideoComment 写入一条评论，并同步更新视频评论数。
func (c *CommentModel) VideoComment(uid, awemeID int64, ipLocation, content, shortID, uniqueID, signature, nickname, avatar string) (commentID int64, ok bool) {
	currentTime := time.Now().Unix()

	var shortIDInt int64
	if shortID != "" {
		parsedID, err := strconv.ParseInt(shortID, 10, 64)
		if err == nil {
			shortIDInt = parsedID
		}
	}
	if profile, loaded := c.loadCommentAuthorProfile(uid); loaded {
		if profile.ShortID > 0 {
			shortIDInt = profile.ShortID
		}
		if profile.UniqueID != "" {
			uniqueID = profile.UniqueID
		}
		if profile.Signature != "" {
			signature = profile.Signature
		}
		if profile.Nickname != "" {
			nickname = profile.Nickname
		}
		if profile.Avatar != "" {
			avatar = profile.Avatar
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
	cache.invalidateStats(awemeID)
	cache.invalidateCommentList(awemeID)

	return commentID, true
}
