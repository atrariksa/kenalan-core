package repository

import (
	"context"

	"github.com/atrariksa/kenalan-core/app/model"
)

type ICoreRepository interface {
	GetViewProfile(ctx context.Context, key string) (model.ViewProfile, error)
}

type CoreRepository struct {
}

func NewCoreRepository() *CoreRepository {
	return &CoreRepository{}
}

func (c *CoreRepository) GetViewProfile(ctx context.Context, key string) (model.ViewProfile, error) {
	return model.ViewProfile{}, nil
}
