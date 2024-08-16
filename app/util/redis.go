package util

import (
	"time"

	"github.com/redis/go-redis/v9"
)

var ViewProfileDataDuration = time.Second * 24 * 7

func GetRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return client
}
