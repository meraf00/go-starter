package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/chariotplatform/goapi/config"
	"github.com/chariotplatform/goapi/logger"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewRedis(cfg config.RedisConfig, log logger.Log) (*redisCache, func(), error) {
	options := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		// Connection pool settings
		MinIdleConns: 5,
		PoolSize:     20,
		PoolTimeout:  30 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	client := redis.NewClient(options)

	// Ping Redis to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Connected to Redis at ", cfg.Addr)

	shutdown := func() {
		if err := client.Close(); err != nil {
			log.Error("Failed to close Redis connection: ", err)
		} else {
			log.Info("Redis connection closed.")
		}
	}

	return &redisCache{client}, shutdown, nil
}

func (c *redisCache) Get(ctx context.Context, key string) (interface{}, time.Time, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil && err == redis.Nil {
		return nil, time.Time{}, ErrKeyNotFound
	} else if err != nil {
		return nil, time.Time{}, err
	}

	dur, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, time.Time{}, err
	}
	if dur == -1 {
		return val, time.Unix(1<<63-1, 0), nil
	}
	if dur == -2 {
		return val, time.Time{}, ErrItemExpired
	}

	return val, time.Now().Add(dur), nil
}

func (c *redisCache) Set(ctx context.Context, key string, val any, dur time.Duration) error {
	return c.client.Set(ctx, key, val, dur).Err()
}

func (c *redisCache) SetNX(ctx context.Context, key string, val any, dur time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, val, dur).Result()
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (m *redisCache) String() string {
	return "redis"
}
