package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func Open(ctx context.Context, redisURL string) (*Redis, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}

	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}
	return &Redis{Client: client}, nil
}

func (r *Redis) Ping(ctx context.Context) error { return r.Client.Ping(ctx).Err() }
func (r *Redis) Close() error                   { return r.Client.Close() }
