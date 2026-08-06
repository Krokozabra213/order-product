package redisRepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	prefix string
}

func New(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

func (r *RedisCache) SetProducts(
	ctx context.Context,
	key string,
	value []domain.ProductWithComponents,
	ttl time.Duration,
) error {

	fullKey := r.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal products: %w", err)
	}

	if err := r.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (r *RedisCache) GetProducts(
	ctx context.Context,
	key string,
) ([]domain.ProductWithComponents, error) {

	fullKey := r.prefix + key

	data, err := r.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var products []domain.ProductWithComponents

	if err := json.Unmarshal(data, &products); err != nil {
		return nil, fmt.Errorf("unmarshal products: %w", err)
	}

	return products, nil
}

func (r *RedisCache) SetProduct(
	ctx context.Context,
	key string,
	value *domain.ProductWithComponents,
	ttl time.Duration,
) error {

	fullKey := r.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal product: %w", err)
	}

	if err := r.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (r *RedisCache) GetProduct(
	ctx context.Context,
	key string,
) (*domain.ProductWithComponents, error) {

	fullKey := r.prefix + key

	data, err := r.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var product domain.ProductWithComponents

	if err := json.Unmarshal(data, &product); err != nil {
		return nil, fmt.Errorf("unmarshal product: %w", err)
	}

	return &product, nil
}
