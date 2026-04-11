package upload_file

import (
	"douyin-backend/app/utils/md5_encrypt"
	"douyin-backend/app/utils/redis_factory"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"sort"
	"strconv"
)

type videoChunkStateStore struct {
	redisClient       *redis_factory.RedisClient
	uploadedChunksKey string
	chunkHashesKey    string
}

func createVideoChunkStateStore(uploadID string, uid int64) *videoChunkStateStore {
	redisClient := redis_factory.GetOneRedisClient()
	if redisClient == nil {
		return nil
	}

	keyPrefix := fmt.Sprintf("upload:video:chunk:%d:%s", uid, md5_encrypt.MD5(uploadID))
	return &videoChunkStateStore{
		redisClient:       redisClient,
		uploadedChunksKey: keyPrefix + ":uploaded",
		chunkHashesKey:    keyPrefix + ":hashes",
	}
}

func (store *videoChunkStateStore) Release() {
	if store == nil || store.redisClient == nil {
		return
	}
	store.redisClient.ReleaseOneRedisClient()
}

func (store *videoChunkStateStore) SetUploadedChunk(chunkIndex int, chunkHash string, ttlSeconds int64) error {
	if _, err := store.redisClient.Execute("SADD", store.uploadedChunksKey, chunkIndex); err != nil {
		return err
	}
	if chunkHash != "" {
		if _, err := store.redisClient.Execute("HSET", store.chunkHashesKey, chunkIndex, chunkHash); err != nil {
			return err
		}
	}
	return store.RefreshTTL(ttlSeconds)
}

func (store *videoChunkStateStore) GetUploadedChunks() ([]int, error) {
	values, err := store.redisClient.Strings(store.redisClient.Execute("SMEMBERS", store.uploadedChunksKey))
	if err != nil {
		if err == redis.ErrNil {
			return []int{}, nil
		}
		return nil, err
	}

	indexes := make([]int, 0, len(values))
	for _, value := range values {
		index, convErr := strconv.Atoi(value)
		if convErr != nil {
			continue
		}
		indexes = append(indexes, index)
	}

	sort.Ints(indexes)
	return indexes, nil
}

func (store *videoChunkStateStore) GetChunkHashes() (map[int]string, error) {
	values, err := redis.StringMap(store.redisClient.Execute("HGETALL", store.chunkHashesKey))
	if err != nil {
		if err == redis.ErrNil {
			return map[int]string{}, nil
		}
		return nil, err
	}

	hashes := make(map[int]string, len(values))
	for field, value := range values {
		index, convErr := strconv.Atoi(field)
		if convErr != nil {
			continue
		}
		hashes[index] = value
	}

	return hashes, nil
}

func (store *videoChunkStateStore) RemoveChunk(chunkIndex int) error {
	if _, err := store.redisClient.Execute("SREM", store.uploadedChunksKey, chunkIndex); err != nil && err != redis.ErrNil {
		return err
	}
	if _, err := store.redisClient.Execute("HDEL", store.chunkHashesKey, chunkIndex); err != nil && err != redis.ErrNil {
		return err
	}
	return nil
}

func (store *videoChunkStateStore) Clear() error {
	_, err := store.redisClient.Execute("DEL", store.uploadedChunksKey, store.chunkHashesKey)
	return err
}

func (store *videoChunkStateStore) RefreshTTL(ttlSeconds int64) error {
	if ttlSeconds <= 0 {
		return nil
	}
	if _, err := store.redisClient.Execute("EXPIRE", store.uploadedChunksKey, ttlSeconds); err != nil {
		return err
	}
	if _, err := store.redisClient.Execute("EXPIRE", store.chunkHashesKey, ttlSeconds); err != nil && err != redis.ErrNil {
		return err
	}
	return nil
}
