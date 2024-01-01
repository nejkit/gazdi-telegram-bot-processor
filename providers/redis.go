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

func (p *RedisProvider) GetAllFromHash(ctx context.Context, hash string) (map[string]string, error) {
	data, err := p.client.HGetAll(ctx, hash).Result()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *RedisProvider) DeleteFromHash(ctx context.Context, hash string, key string) error {

	_, err := p.client.HDel(ctx, hash, key).Result()

	if err != nil {
		return err
	}
	return nil
}

func (p *RedisProvider) AddToSet(ctx context.Context, set string, value string) error {
	_, err := p.client.SAdd(ctx, set, value).Result()

	if err != nil {
		return err
	}
	return nil
}

func (p *RedisProvider) DeleteFromSet(ctx context.Context, set string, value string) error {
	_, err := p.client.SRem(ctx, set, value).Result()
	if err != nil {
		return err
	}
	return nil
}

func (p *RedisProvider) CheckMemberOfHash(ctx context.Context, hash string, key string) (bool, error) {
	exists, err := p.client.HExists(ctx, hash, key).Result()

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (p *RedisProvider) CheckMemberOfSet(ctx context.Context, set string, value interface{}) (bool, error) {
	exists, err := p.client.SIsMember(ctx, set, value).Result()

	if err != nil {
		return false, err
	}

	return exists, nil
}
