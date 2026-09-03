package redis

import (
	"context"
	"cu-timepad-bot/internal/store"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDraftStore struct {
	client *redis.Client
}

func NewDraftStore(client *redis.Client) *RedisDraftStore {
	return &RedisDraftStore{
		client: client,
	}
}

func (rdb *RedisDraftStore) Save(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	err := rdb.client.Set(ctx, key, val, ttl)
	if err.Err() != nil {
		return store.ErrTransient
	}
	return nil
}

func (rdb *RedisDraftStore) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := rdb.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, store.ErrNotFound
		}
		return nil, store.ErrTransient
	}
	return b, nil
}

func (rdb *RedisDraftStore) Delete(ctx context.Context, key string) error {
	err := rdb.client.Del(ctx, key)
	if err.Err() != nil {
		return store.ErrTransient
	}
	return nil
}

func (rdb *RedisDraftStore) Exists(ctx context.Context, key string) bool {
	res := rdb.client.Exists(ctx, key)
	cnt, _ := res.Result()
	return cnt > 0
}
