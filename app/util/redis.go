package util

import (
	"time"

	"github.com/atrariksa/kenalan-core/config"
	"github.com/redis/go-redis/v9"
)

var ViewProfileDataDuration = time.Hour * 24

func GetRedisClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConfig.Address,
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
	})
	return client
}
