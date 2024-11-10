package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"vieo/auth/internal/domain/models"

	"github.com/go-redis/redis"
)

type Cache interface {
	SetMovie(ctx context.Context, movie *models.Movie) error
	GetMovie(ctx context.Context, id string) (*models.Movie, error)
	InvalidateMovie(ctx context.Context, id string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{client: client}
}

func (c *RedisCache) SetMovie(_ context.Context, movie *models.Movie) error {
	data, err := json.Marshal(movie)
	if err != nil {
		return err
	}
	return c.client.Set(fmt.Sprintf("movie:%s", movie.ID), data, 1*time.Hour).Err()
}

func (c *RedisCache) GetMovie(_ context.Context, id string) (*models.Movie, error) {
	data, err := c.client.Get(fmt.Sprintf("movie:%s", id)).Bytes()
	if err != nil {
		return nil, err
	}
	var movie models.Movie
	err = json.Unmarshal(data, &movie)
	return &movie, err
}

func (c *RedisCache) InvalidateMovie(_ context.Context, id string) error {
	return c.client.Del(fmt.Sprintf("movie:%s", id)).Err()
}
