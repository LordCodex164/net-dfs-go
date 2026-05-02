package redis

import (
	"context"
	"encoding/json"

	"github.com/LordCodex164/net-protocol/types"
	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func SaveMetadata(rClient *redis.Client, meta types.FileMeta, fileId string) error {
	meta_json, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	rClient.Set(Ctx, "file:"+fileId, meta_json, 0)
	return nil
}

func GetMetaData(rClient *redis.Client, fileId string) (types.FileMeta, error) {
	val, err := rClient.Get(Ctx, "file:"+fileId).Result()
	if err != nil {
		return types.FileMeta{}, err
	}

	var meta types.FileMeta
	json.Unmarshal([]byte(val), &meta)

	return meta, nil
}
