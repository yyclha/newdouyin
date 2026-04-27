package video

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	videodiggasync "douyin-backend/app/service"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

var videoDiggRedisScript = redis.NewScript(7, `
local diggUsersKey = KEYS[1]
local userLikesKey = KEYS[2]
local userLikesIndexKey = KEYS[3]
local videoStatsKey = KEYS[4]
local totalFavoritedKey = KEYS[5]
local stateKey = KEYS[6]
local versionKey = KEYS[7]

local uid = ARGV[1]
local awemeID = ARGV[2]
local action = tonumber(ARGV[3])
local createTime = tonumber(ARGV[4])
local setTTL = tonumber(ARGV[5])
local statsTTL = tonumber(ARGV[6])
local totalTTL = tonumber(ARGV[7])
local stateTTL = tonumber(ARGV[8])
local versionTTL = tonumber(ARGV[9])

local currentState = tonumber(redis.call('GET', stateKey) or '0')
local version = tonumber(redis.call('GET', versionKey) or '0')
local changed = 0

if action == 1 then
	if currentState == 0 then
		redis.call('SETEX', stateKey, stateTTL, 1)
		redis.call('SADD', diggUsersKey, uid)
		redis.call('SADD', userLikesKey, awemeID)
		redis.call('ZADD', userLikesIndexKey, createTime, awemeID)
		redis.call('HINCRBY', videoStatsKey, 'digg_count', 1)
		redis.call('INCRBY', totalFavoritedKey, 1)
		version = redis.call('INCR', versionKey)
		changed = 1
		currentState = 1
	end
else
	if currentState == 1 then
		redis.call('SETEX', stateKey, stateTTL, 0)
		redis.call('SREM', diggUsersKey, uid)
		redis.call('SREM', userLikesKey, awemeID)
		redis.call('ZREM', userLikesIndexKey, awemeID)
		local diggCount = tonumber(redis.call('HGET', videoStatsKey, 'digg_count') or '0')
		if diggCount > 0 then
			redis.call('HINCRBY', videoStatsKey, 'digg_count', -1)
		else
			redis.call('HSET', videoStatsKey, 'digg_count', 0)
		end
		local totalFavorited = tonumber(redis.call('GET', totalFavoritedKey) or '0')
		if totalFavorited > 0 then
			redis.call('INCRBY', totalFavoritedKey, -1)
		else
			redis.call('SET', totalFavoritedKey, 0)
		end
		version = redis.call('INCR', versionKey)
		changed = 1
		currentState = 0
	end
end

if redis.call('EXISTS', diggUsersKey) == 1 then
	redis.call('EXPIRE', diggUsersKey, setTTL)
end
if redis.call('EXISTS', userLikesKey) == 1 then
	redis.call('EXPIRE', userLikesKey, setTTL)
end
if redis.call('EXISTS', userLikesIndexKey) == 1 then
	redis.call('EXPIRE', userLikesIndexKey, setTTL)
end
if redis.call('EXISTS', videoStatsKey) == 1 then
	redis.call('EXPIRE', videoStatsKey, statsTTL)
end
if redis.call('EXISTS', totalFavoritedKey) == 1 then
	redis.call('EXPIRE', totalFavoritedKey, totalTTL)
end
if version > 0 then
	redis.call('EXPIRE', versionKey, versionTTL)
end
redis.call('EXPIRE', stateKey, stateTTL)

local finalDiggCount = tonumber(redis.call('HGET', videoStatsKey, 'digg_count') or '0')
return {changed, currentState, version, finalDiggCount}
`)

// DiggModel 封装视频点赞和取消点赞的数据库操作。
type DiggModel struct {
	*gorm.DB   `gorm:"-" json:"-"`
	DiggID     int64 `json:"digg_id"`     // bigint
	UID        int64 `json:"uid"`         // bigint
	AwemeID    int64 `json:"aweme_id"`    // bigint
	CreateTime int   `json:"create_time"` // int
}

// videoDiggRedisResult 表示 Redis 点赞脚本返回的执行结果。
type videoDiggRedisResult struct {
	Changed   bool
	Action    bool
	Version   int64
	DiggCount int64
}

// CreateDiggFactory 创建带数据库连接的点赞模型实例。
func CreateDiggFactory(sqlType string) *DiggModel {
	return &DiggModel{DB: model.UseDbConn(sqlType)}
}

// VideoDigg 对目标视频执行点赞或取消点赞，并保持缓存与数据库状态一致。
func (v *DiggModel) VideoDigg(uid, awemeID int64, action bool) bool {
	authorUID, ok := v.getVideoAuthorUID(awemeID)
	if !ok {
		return false
	}

	cache := newInteractionCache()
	if v.prepareVideoDiggRedis(cache, uid, awemeID, authorUID) {
		result, err := v.applyVideoDiggRedis(cache, uid, awemeID, authorUID, action)
		if err == nil {
			if !result.Changed {
				return true
			}

			event := videodiggasync.VideoDiggEvent{
				UID:        uid,
				AwemeID:    awemeID,
				AuthorUID:  authorUID,
				Action:     result.Action,
				Version:    result.Version,
				DiggCount:  result.DiggCount,
				OccurredAt: time.Now().Unix(),
			}
			outboxEvent, outboxErr := createVideoDiggOutboxEvent(event)
			if outboxErr != nil {
				variable.ZapLog.Error("failed to create video digg outbox event", zap.Error(outboxErr), zap.Int64("uid", uid), zap.Int64("aweme_id", awemeID))
				_, _ = v.applyVideoDiggRedis(cache, uid, awemeID, authorUID, !result.Action)
				return false
			}
			if publishErr := videodiggasync.PublishVideoDiggEvent(event); publishErr == nil {
				_ = markVideoDiggOutboxPublished(outboxEvent.ID)
				return true
			} else {
				variable.ZapLog.Error("failed to publish video digg event", zap.Error(publishErr), zap.Int64("uid", uid), zap.Int64("aweme_id", awemeID))
				_ = markVideoDiggOutboxFailed(outboxEvent.ID, publishErr)
				if v.persistVideoDiggState(uid, awemeID, authorUID, result.Action) {
					return true
				}
				_, _ = v.applyVideoDiggRedis(cache, uid, awemeID, authorUID, !result.Action)
				return false
			}
		}

		variable.ZapLog.Error("failed to update video digg cache", zap.Error(err), zap.Int64("uid", uid), zap.Int64("aweme_id", awemeID))
	}

	return v.persistVideoDiggState(uid, awemeID, authorUID, action)
}

// HandleAsyncDiggEvent 持久化缓存更新成功后发布的异步点赞事件。
func (v *DiggModel) HandleAsyncDiggEvent(event videodiggasync.VideoDiggEvent) error {
	cache := newInteractionCache()
	if version, ok := cache.getDiggVersion(event.UID, event.AwemeID); ok && version > event.Version {
		return nil
	}
	if ok := v.persistVideoDiggState(event.UID, event.AwemeID, event.AuthorUID, event.Action); !ok {
		return errors.New("persist video digg state failed")
	}
	return nil
}

// getVideoAuthorUID 执行对象方法逻辑。
func (v *DiggModel) getVideoAuthorUID(awemeID int64) (int64, bool) {
	var authorUID int64
	if err := v.Raw(`SELECT author_user_id FROM tb_videos WHERE aweme_id = ? LIMIT 1`, awemeID).Scan(&authorUID).Error; err != nil {
		variable.ZapLog.Error("VideoDigg failed to query video author", zap.Error(err))
		return 0, false
	}
	if authorUID == 0 {
		variable.ZapLog.Error("VideoDigg failed because video author was not found", zap.Int64("aweme_id", awemeID))
		return 0, false
	}
	return authorUID, true
}

// prepareVideoDiggRedis 执行对象方法逻辑。
func (v *DiggModel) prepareVideoDiggRedis(cache *interactionCache, uid, awemeID, authorUID int64) bool {
	if _, err := cache.ensureDiggState(uid, awemeID); err != nil {
		return false
	}
	if _, ok := cache.getOrLoadStats(awemeID); !ok {
		return false
	}
	if _, ok := cache.getUserTotalFavorited(authorUID); !ok {
		if _, ok = cache.loadUserTotalFavorited(authorUID); !ok {
			return false
		}
	}
	return true
}

// applyVideoDiggRedis 执行对象方法逻辑。
func (v *DiggModel) applyVideoDiggRedis(cache *interactionCache, uid, awemeID, authorUID int64, action bool) (videoDiggRedisResult, error) {
	client := cache.redisClient()
	if client == nil {
		return videoDiggRedisResult{}, fmt.Errorf("redis client unavailable")
	}
	defer client.ReleaseOneRedisClient()

	actionInt := 0
	if action {
		actionInt = 1
	}

	reply, err := videoDiggRedisScript.Do(
		client,
		cache.diggUsersKey(awemeID),
		cache.userLikeVideosKey(uid),
		cache.userLikeIndexKey(uid),
		cache.statsKey(awemeID),
		cache.userTotalFavoritedKey(authorUID),
		cache.diggStateKey(uid, awemeID),
		cache.diggVersionKey(uid, awemeID),
		uid,
		awemeID,
		actionInt,
		time.Now().Unix(),
		videoUserLikesCacheTTLSeconds,
		videoStatsCacheTTLSeconds,
		videoStatsCacheTTLSeconds,
		videoDiggStateTTLSeconds,
		videoDiggVersionTTLSeconds,
	)
	if err != nil {
		return videoDiggRedisResult{}, err
	}

	values, err := redis.Values(reply, nil)
	if err != nil {
		return videoDiggRedisResult{}, err
	}

	var changedInt int64
	var finalActionInt int64
	var version int64
	var diggCount int64
	if _, err := redis.Scan(values, &changedInt, &finalActionInt, &version, &diggCount); err != nil {
		return videoDiggRedisResult{}, err
	}

	return videoDiggRedisResult{
		Changed:   changedInt == 1,
		Action:    finalActionInt == 1,
		Version:   version,
		DiggCount: diggCount,
	}, nil
}

// persistVideoDiggState 执行对象方法逻辑。
func (v *DiggModel) persistVideoDiggState(uid, awemeID, authorUID int64, action bool) bool {
	tx := v.DB.Begin()
	if tx.Error != nil {
		variable.ZapLog.Error("VideoDigg failed to start transaction", zap.Error(tx.Error))
		return false
	}

	currentTime := time.Now().Unix()
	diggSQL := `INSERT IGNORE INTO tb_diggs (uid, aweme_id, create_time) VALUES (?, ?, ?);`
	undiggSQL := `DELETE FROM tb_diggs WHERE uid = ? AND aweme_id = ?;`

	var result *gorm.DB
	if action {
		result = tx.Exec(diggSQL, uid, awemeID, currentTime)
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
		result = tx.Exec(undiggSQL, uid, awemeID)
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

	return true
}
