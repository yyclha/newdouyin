package video

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	"douyin-backend/app/utils/redis_factory"
	"github.com/goccy/go-json"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

const (
	videoStatsCacheTTLSeconds     = 3600
	videoCommentsCacheTTLSeconds  = 300
	videoCommentItemTTLSeconds    = 300
	videoUserLikesCacheTTLSeconds = 600
	videoCommentIndexLimit        = 500
)

type interactionCache struct {
}

func newInteractionCache() *interactionCache {
	return &interactionCache{}
}

func (c *interactionCache) redisClient() *redis_factory.RedisClient {
	return redis_factory.GetOneRedisClient()
}

func (c *interactionCache) statsKey(awemeID int64) string {
	return "video:stats:" + strconv.FormatInt(awemeID, 10)
}

func (c *interactionCache) commentsIndexKey(awemeID int64) string {
	return "video:comments:index:" + strconv.FormatInt(awemeID, 10)
}

func (c *interactionCache) commentItemKey(commentID int64) string {
	return "video:comments:item:" + strconv.FormatInt(commentID, 10)
}

func (c *interactionCache) commentDiggUsersKey(commentID int64) string {
	return "comment:digg:users:" + strconv.FormatInt(commentID, 10)
}

func (c *interactionCache) diggUsersKey(awemeID int64) string {
	return "video:digg:users:" + strconv.FormatInt(awemeID, 10)
}

func (c *interactionCache) userLikeVideosKey(uid int64) string {
	return "user:likes:" + strconv.FormatInt(uid, 10)
}

func (c *interactionCache) userTotalFavoritedKey(uid int64) string {
	return "user:total_favorited:" + strconv.FormatInt(uid, 10)
}

func (c *interactionCache) getStats(awemeID int64) (model.Statistics, bool) {
	client := c.redisClient()
	if client == nil {
		return model.Statistics{}, false
	}
	defer client.ReleaseOneRedisClient()

	values, err := redis.StringMap(client.Execute("HGETALL", c.statsKey(awemeID)))
	if err != nil || len(values) == 0 {
		return model.Statistics{}, false
	}

	stats := model.Statistics{Id: awemeID}
	stats.AdmireCount = parseInt64(values["admire_count"])
	stats.CommentCount = parseInt64(values["comment_count"])
	stats.DiggCount = parseInt64(values["digg_count"])
	stats.CollectCount = parseInt64(values["collect_count"])
	stats.PlayCount = parseInt64(values["play_count"])
	stats.ShareCount = parseInt64(values["share_count"])
	return stats, true
}

func (c *interactionCache) setStats(stats model.Statistics) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	key := c.statsKey(stats.Id)
	if _, err := client.Execute(
		"HMSET",
		key,
		"admire_count", stats.AdmireCount,
		"comment_count", stats.CommentCount,
		"digg_count", stats.DiggCount,
		"collect_count", stats.CollectCount,
		"play_count", stats.PlayCount,
		"share_count", stats.ShareCount,
	); err != nil {
		variable.ZapLog.Error("failed to cache video stats", zap.Error(err), zap.Int64("aweme_id", stats.Id))
		return
	}
	_, _ = client.Execute("EXPIRE", key, videoStatsCacheTTLSeconds)
}

func (c *interactionCache) loadStatsFromDB(awemeID int64) (model.Statistics, bool) {
	videoModel := CreateVideoFactory("")
	stats := model.Statistics{}
	sql := `SELECT id, admire_count, comment_count, digg_count, collect_count, play_count, share_count FROM tb_statistics WHERE id = ? LIMIT 1`
	if err := videoModel.Raw(sql, awemeID).Scan(&stats).Error; err != nil {
		variable.ZapLog.Error("failed to load video stats from db", zap.Error(err), zap.Int64("aweme_id", awemeID))
		return model.Statistics{}, false
	}
	if stats.Id == 0 {
		stats.Id = awemeID
	}
	c.setStats(stats)
	return stats, true
}

func (c *interactionCache) getOrLoadStats(awemeID int64) (model.Statistics, bool) {
	if stats, ok := c.getStats(awemeID); ok {
		return stats, true
	}
	return c.loadStatsFromDB(awemeID)
}

func (c *interactionCache) incrStat(awemeID int64, field string, delta int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	key := c.statsKey(awemeID)
	exists, err := client.Int(client.Execute("EXISTS", key))
	if err == nil && exists == 0 {
		if stats, ok := c.loadStatsFromDB(awemeID); ok {
			stats.Id = awemeID
		}
	}
	if _, err := client.Execute("HINCRBY", key, field, delta); err != nil {
		variable.ZapLog.Error("failed to incr video stat cache", zap.Error(err), zap.Int64("aweme_id", awemeID), zap.String("field", field))
		return
	}
	_, _ = client.Execute("EXPIRE", key, videoStatsCacheTTLSeconds)
}

func (c *interactionCache) normalizeStats(videoList []model.Video) {
	for i := range videoList {
		awemeID, err := strconv.ParseInt(videoList[i].AwemeID, 10, 64)
		if err != nil {
			continue
		}
		stats, ok := c.getOrLoadStats(awemeID)
		if !ok {
			continue
		}
		raw, err := json.Marshal(stats)
		if err != nil {
			continue
		}
		videoList[i].Statistics = raw
	}
}

func (c *interactionCache) getCommentsPage(awemeID, start, pageSize, total int64) ([]Comment, bool) {
	if start < 0 || pageSize <= 0 {
		return nil, false
	}
	if start >= videoCommentIndexLimit || start+pageSize > videoCommentIndexLimit {
		return nil, false
	}

	client := c.redisClient()
	if client == nil {
		return nil, false
	}
	defer client.ReleaseOneRedisClient()

	stop := start + pageSize - 1
	commentIDs, err := redis.Int64s(client.Execute("ZREVRANGE", c.commentsIndexKey(awemeID), start, stop))
	if err != nil {
		return nil, false
	}
	if len(commentIDs) == 0 {
		if total == 0 || start >= total {
			return []Comment{}, true
		}
		return nil, false
	}

	expectedLen := int(pageSize)
	if total > 0 {
		remaining := total - start
		if remaining < pageSize {
			expectedLen = int(remaining)
		}
	}
	if expectedLen > 0 && len(commentIDs) < expectedLen {
		return nil, false
	}

	for _, commentID := range commentIDs {
		if err := client.Send("HGETALL", c.commentItemKey(commentID)); err != nil {
			return nil, false
		}
	}
	if err := client.Flush(); err != nil {
		return nil, false
	}

	comments := make([]Comment, 0, len(commentIDs))
	for range commentIDs {
		reply, receiveErr := client.Receive()
		values, mapErr := redis.StringMap(reply, receiveErr)
		if mapErr != nil || len(values) == 0 {
			return nil, false
		}
		comments = append(comments, commentFromRedisMap(values))
	}
	return comments, true
}

func (c *interactionCache) getCommentItemWithClient(client *redis_factory.RedisClient, commentID int64) (Comment, bool) {
	values, err := redis.StringMap(client.Execute("HGETALL", c.commentItemKey(commentID)))
	if err != nil || len(values) == 0 {
		return Comment{}, false
	}
	return commentFromRedisMap(values), true
}

func (c *interactionCache) setComments(awemeID int64, comments []Comment) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	indexKey := c.commentsIndexKey(awemeID)
	_, _ = client.Execute("DEL", indexKey)
	limit := len(comments)
	if limit > videoCommentIndexLimit {
		limit = videoCommentIndexLimit
	}
	for i := 0; i < limit; i++ {
		comment := comments[i]
		if err := c.setCommentItemWithClient(client, comment); err != nil {
			variable.ZapLog.Error("failed to cache comment item", zap.Error(err), zap.Int64("comment_id", comment.CommentID))
			return
		}
		if _, err := client.Execute("ZADD", indexKey, comment.CreateTime, comment.CommentID); err != nil {
			variable.ZapLog.Error("failed to cache comment index", zap.Error(err), zap.Int64("aweme_id", awemeID))
			return
		}
	}
	_, _ = client.Execute("EXPIRE", indexKey, videoCommentsCacheTTLSeconds)
}

func (c *interactionCache) setCommentItemWithClient(client *redis_factory.RedisClient, comment Comment) error {
	if _, err := client.Execute(
		"HMSET",
		c.commentItemKey(comment.CommentID),
		"comment_id", comment.CommentID,
		"create_time", comment.CreateTime,
		"ip_location", comment.IPLocation,
		"aweme_id", comment.AwemeID,
		"content", comment.Content,
		"is_author_digged", boolToInt(comment.IsAuthorDigged),
		"is_folded", boolToInt(comment.IsFolded),
		"is_hot", boolToInt(comment.IsHot),
		"user_buried", boolToInt(comment.UserBuried),
		"user_digged", 0,
		"digg_count", comment.DiggCount,
		"user_id", comment.UserID,
		"sec_uid", comment.SecUID,
		"short_user_id", comment.ShortUserID,
		"user_unique_id", comment.UserUniqueID,
		"user_signature", comment.UserSignature,
		"nickname", comment.Nickname,
		"avatar", comment.Avatar,
		"sub_comment_count", comment.SubCommentCount,
		"last_modify_ts", comment.LastModifyTS,
	); err != nil {
		return err
	}
	if _, err := client.Execute("EXPIRE", c.commentItemKey(comment.CommentID), videoCommentItemTTLSeconds); err != nil {
		return err
	}
	return nil
}

func (c *interactionCache) prependComment(awemeID int64, comment Comment) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	indexKey := c.commentsIndexKey(awemeID)
	if err := c.setCommentItemWithClient(client, comment); err != nil {
		return
	}
	if _, err := client.Execute("ZADD", indexKey, comment.CreateTime, comment.CommentID); err != nil {
		return
	}
	_, _ = client.Execute("ZREMRANGEBYRANK", indexKey, 0, -(videoCommentIndexLimit + 1))
	_, _ = client.Execute("EXPIRE", indexKey, videoCommentsCacheTTLSeconds)
}

func (c *interactionCache) removeComment(awemeID, commentID int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	_, _ = client.Execute("ZREM", c.commentsIndexKey(awemeID), commentID)
	_, _ = client.Execute("DEL", c.commentItemKey(commentID))
	_, _ = client.Execute("DEL", c.commentDiggUsersKey(commentID))
	_, _ = client.Execute("EXPIRE", c.commentsIndexKey(awemeID), videoCommentsCacheTTLSeconds)
}

func (c *interactionCache) updateCommentDigg(awemeID, commentID, uid int64, action bool) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	indexExists, err := client.Int(client.Execute("EXISTS", c.commentsIndexKey(awemeID)))
	if err != nil || indexExists == 0 {
		if action {
			c.addCommentDiggUser(commentID, uid)
		} else {
			c.removeCommentDiggUser(commentID, uid)
		}
		return
	}

	comment, ok := c.getCommentItemWithClient(client, commentID)
	if !ok {
		if action {
			c.addCommentDiggUser(commentID, uid)
		} else {
			c.removeCommentDiggUser(commentID, uid)
		}
		return
	}

	if action {
		comment.DiggCount++
		c.addCommentDiggUser(commentID, uid)
	} else {
		if comment.DiggCount > 0 {
			comment.DiggCount--
		}
		c.removeCommentDiggUser(commentID, uid)
	}
	if _, err := client.Execute("HSET", c.commentItemKey(commentID), "digg_count", comment.DiggCount); err == nil {
		_, _ = client.Execute("EXPIRE", c.commentItemKey(commentID), videoCommentItemTTLSeconds)
	}
}

func (c *interactionCache) getUserLikedVideos(uid int64) ([]int64, bool) {
	client := c.redisClient()
	if client == nil {
		return nil, false
	}
	defer client.ReleaseOneRedisClient()

	values, err := client.Strings(client.Execute("SMEMBERS", c.userLikeVideosKey(uid)))
	if err != nil || len(values) == 0 {
		return nil, false
	}

	ids := make([]int64, 0, len(values))
	for _, value := range values {
		id, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr == nil {
			ids = append(ids, id)
		}
	}
	return ids, true
}

func (c *interactionCache) setUserLikedVideos(uid int64, awemeIDs []int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	key := c.userLikeVideosKey(uid)
	_, _ = client.Execute("DEL", key)
	if len(awemeIDs) > 0 {
		args := make([]interface{}, 0, len(awemeIDs)+1)
		args = append(args, key)
		for _, awemeID := range awemeIDs {
			args = append(args, awemeID)
		}
		if _, err := client.Execute("SADD", args...); err != nil {
			variable.ZapLog.Error("failed to cache user like videos", zap.Error(err), zap.Int64("uid", uid))
			return
		}
	}
	_, _ = client.Execute("EXPIRE", key, videoUserLikesCacheTTLSeconds)
}

func (c *interactionCache) addUserLikedVideo(uid, awemeID int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	key := c.userLikeVideosKey(uid)
	if _, err := client.Execute("SADD", key, awemeID); err != nil {
		variable.ZapLog.Error("failed to add user like video cache", zap.Error(err), zap.Int64("uid", uid), zap.Int64("aweme_id", awemeID))
		return
	}
	_, _ = client.Execute("EXPIRE", key, videoUserLikesCacheTTLSeconds)
}

func (c *interactionCache) removeUserLikedVideo(uid, awemeID int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	key := c.userLikeVideosKey(uid)
	if _, err := client.Execute("SREM", key, awemeID); err != nil && err != redis.ErrNil {
		variable.ZapLog.Error("failed to remove user like video cache", zap.Error(err), zap.Int64("uid", uid), zap.Int64("aweme_id", awemeID))
		return
	}
	_, _ = client.Execute("EXPIRE", key, videoUserLikesCacheTTLSeconds)
}

func (c *interactionCache) isVideoLikedByUser(uid, awemeID int64) (bool, bool) {
	client := c.redisClient()
	if client == nil {
		return false, false
	}
	defer client.ReleaseOneRedisClient()

	value, err := client.Int(client.Execute("SISMEMBER", c.userLikeVideosKey(uid), awemeID))
	if err == nil {
		return value == 1, true
	}

	value, err = client.Int(client.Execute("SISMEMBER", c.diggUsersKey(awemeID), uid))
	if err == nil {
		return value == 1, true
	}
	return false, false
}

func (c *interactionCache) addDiggUser(awemeID, uid int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	if _, err := client.Execute("SADD", c.diggUsersKey(awemeID), uid); err != nil {
		variable.ZapLog.Error("failed to add digg user cache", zap.Error(err), zap.Int64("uid", uid), zap.Int64("aweme_id", awemeID))
		return
	}
	_, _ = client.Execute("EXPIRE", c.diggUsersKey(awemeID), videoUserLikesCacheTTLSeconds)
}

func (c *interactionCache) removeDiggUser(awemeID, uid int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	if _, err := client.Execute("SREM", c.diggUsersKey(awemeID), uid); err != nil && err != redis.ErrNil {
		variable.ZapLog.Error("failed to remove digg user cache", zap.Error(err), zap.Int64("uid", uid), zap.Int64("aweme_id", awemeID))
		return
	}
	_, _ = client.Execute("EXPIRE", c.diggUsersKey(awemeID), videoUserLikesCacheTTLSeconds)
}

func (c *interactionCache) addCommentDiggUser(commentID, uid int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	if _, err := client.Execute("SADD", c.commentDiggUsersKey(commentID), uid); err != nil {
		variable.ZapLog.Error("failed to add comment digg user cache", zap.Error(err), zap.Int64("uid", uid), zap.Int64("comment_id", commentID))
		return
	}
	_, _ = client.Execute("EXPIRE", c.commentDiggUsersKey(commentID), videoUserLikesCacheTTLSeconds)
}

func (c *interactionCache) removeCommentDiggUser(commentID, uid int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	if _, err := client.Execute("SREM", c.commentDiggUsersKey(commentID), uid); err != nil && err != redis.ErrNil {
		variable.ZapLog.Error("failed to remove comment digg user cache", zap.Error(err), zap.Int64("uid", uid), zap.Int64("comment_id", commentID))
		return
	}
	_, _ = client.Execute("EXPIRE", c.commentDiggUsersKey(commentID), videoUserLikesCacheTTLSeconds)
}

func (c *interactionCache) loadUserTotalFavorited(uid int64) (int64, bool) {
	videoModel := CreateVideoFactory("")
	var total int64
	sql := `
		SELECT COALESCE(SUM(COALESCE(ts.digg_count, 0)), 0)
		FROM tb_videos AS tv
		LEFT JOIN tb_statistics AS ts ON tv.aweme_id = ts.id
		WHERE tv.author_user_id = ?`
	if err := videoModel.Raw(sql, uid).Scan(&total).Error; err != nil {
		variable.ZapLog.Error("failed to load total favorited from db", zap.Error(err), zap.Int64("uid", uid))
		return 0, false
	}

	client := c.redisClient()
	if client != nil {
		defer client.ReleaseOneRedisClient()
		_, _ = client.Execute("SETEX", c.userTotalFavoritedKey(uid), videoStatsCacheTTLSeconds, total)
	}
	return total, true
}

func (c *interactionCache) getUserTotalFavorited(uid int64) (int64, bool) {
	client := c.redisClient()
	if client == nil {
		return 0, false
	}
	defer client.ReleaseOneRedisClient()

	value, err := client.Int64(client.Execute("GET", c.userTotalFavoritedKey(uid)))
	if err == nil {
		return value, true
	}
	return 0, false
}

func (c *interactionCache) incrUserTotalFavorited(uid, delta int64) {
	client := c.redisClient()
	if client == nil {
		return
	}
	defer client.ReleaseOneRedisClient()

	key := c.userTotalFavoritedKey(uid)
	if _, err := client.Execute("INCRBY", key, delta); err != nil {
		variable.ZapLog.Error("failed to incr total favorited cache", zap.Error(err), zap.Int64("uid", uid))
		return
	}
	_, _ = client.Execute("EXPIRE", key, videoStatsCacheTTLSeconds)
}

func parseInt64(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseInt(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseBool(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "1" || value == "true"
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func commentFromRedisMap(values map[string]string) Comment {
	return Comment{
		CommentID:       parseInt64(values["comment_id"]),
		CreateTime:      parseInt(values["create_time"]),
		IPLocation:      values["ip_location"],
		AwemeID:         parseInt64(values["aweme_id"]),
		Content:         values["content"],
		IsAuthorDigged:  parseBool(values["is_author_digged"]),
		IsFolded:        parseBool(values["is_folded"]),
		IsHot:           parseBool(values["is_hot"]),
		UserBuried:      parseBool(values["user_buried"]),
		UserDigged:      0,
		DiggCount:       parseInt64(values["digg_count"]),
		UserID:          parseInt64(values["user_id"]),
		SecUID:          values["sec_uid"],
		ShortUserID:     parseInt64(values["short_user_id"]),
		UserUniqueID:    values["user_unique_id"],
		UserSignature:   values["user_signature"],
		Nickname:        values["nickname"],
		Avatar:          values["avatar"],
		SubCommentCount: parseInt64(values["sub_comment_count"]),
		LastModifyTS:    parseInt64(values["last_modify_ts"]),
	}
}

func buildCommentForCache(uid, awemeID int64, ipLocation, content string, shortID int64, uniqueID, signature, nickname, avatar string, createTime int64, commentID int64) Comment {
	return Comment{
		CommentID:       commentID,
		CreateTime:      int(createTime),
		IPLocation:      ipLocation,
		AwemeID:         awemeID,
		Content:         content,
		IsAuthorDigged:  false,
		IsFolded:        false,
		IsHot:           false,
		UserBuried:      false,
		UserDigged:      0,
		DiggCount:       0,
		UserID:          uid,
		SecUID:          "",
		ShortUserID:     shortID,
		UserUniqueID:    uniqueID,
		UserSignature:   signature,
		Nickname:        nickname,
		Avatar:          avatar,
		SubCommentCount: 0,
		LastModifyTS:    createTime,
	}
}

func currentUnix() int64 {
	return time.Now().Unix()
}

type UserLikeStatusCache struct {
	cache *interactionCache
}

func NewUserLikeStatusCache() *UserLikeStatusCache {
	return &UserLikeStatusCache{cache: newInteractionCache()}
}

func (c *UserLikeStatusCache) GetUserLikedVideos(uid int64) ([]int64, bool) {
	return c.cache.getUserLikedVideos(uid)
}

func (c *UserLikeStatusCache) SetUserLikedVideos(uid int64, awemeIDs []int64) {
	c.cache.setUserLikedVideos(uid, awemeIDs)
}

func (c *UserLikeStatusCache) GetUserTotalFavorited(uid int64) (int64, bool) {
	return c.cache.getUserTotalFavorited(uid)
}

func (c *UserLikeStatusCache) LoadUserTotalFavorited(uid int64) (int64, bool) {
	return c.cache.loadUserTotalFavorited(uid)
}
