package providers

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisProvider struct {
	client *redis.Client
}

func NewRedisProvider(url string) *RedisProvider {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "",
		DB:       0,
	})
	return &RedisProvider{client: client}
}

func (p *RedisProvider) SetHash(ctx context.Context, hash string, key string, value string) error {
	_, err := p.client.HSet(ctx, hash, key, value).Result()

	if err != nil {
		return err
	}

	return nil
}

func (p *RedisProvider) GetOneFromHash(ctx context.Context, hash string, key string) (*string, error) {
	data, err := p.client.HGet(ctx, hash, key).Result()

	if err != nil {
		return nil, err
	}

	return &data, err
}

func (p *RedisProvider) GetAllKeysFromHash(ctx context.Context, hash string) ([]string, error) {
	data, err := p.client.HGetAll(ctx, hash).Result()

	if err != nil {
		return nil, err
	}
	var keys []string
	for k, _ := range data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (p *RedisProvider) DeleteFromHash(ctx context.Context, hash string, key string) error {

	_, err := p.client.HDel(ctx, hash, key).Result()

	if err != nil {
		return err
	}
	return nil
}
