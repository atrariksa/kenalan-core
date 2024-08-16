package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/atrariksa/kenalan-core/app/model"
	"github.com/atrariksa/kenalan-core/app/util"
	"github.com/redis/go-redis/v9"
)

type IRedisCoreRepository interface {
	StoreViewProfile(ctx context.Context, key string, data model.ViewProfile) error
	GetViewProfile(ctx context.Context, key string) (model.ViewProfile, error)
}

type RedisCoreRepository struct {
	RC *redis.Client
}

func NewRedisCoreRepository(rc *redis.Client) *RedisCoreRepository {
	return &RedisCoreRepository{
		RC: rc,
	}
}

func (ar *RedisCoreRepository) StoreViewProfile(ctx context.Context, key string, data model.ViewProfile) error {
	jsonData, _ := json.Marshal(data)
	return ar.RC.Set(ctx, key, jsonData, util.ViewProfileDataDuration).Err()
}

func (ar *RedisCoreRepository) GetViewProfile(ctx context.Context, key string) (model.ViewProfile, error) {
	jsonData, err := ar.RC.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return model.ViewProfile{}, errors.New("internal error")
	}

	var viewProfileData model.ViewProfile
	json.Unmarshal([]byte(jsonData), &viewProfileData)
	return viewProfileData, nil
}
