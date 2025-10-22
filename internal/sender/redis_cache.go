package sender

import (
	"case_study/internal/config"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct{ client *redis.Client }

func NewRedisCache(cfg config.Config) (*RedisCache, error) {
	c := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisCache{client: c}, nil
}

func (r *RedisCache) SetSent(messageID string, at time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.client.Set(ctx, "sent:"+messageID, at.Format(time.RFC3339), 30*24*time.Hour).Err()
}

func (r *RedisCache) Close() error { return r.client.Close() }
