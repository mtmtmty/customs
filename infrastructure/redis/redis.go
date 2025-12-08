package redis

import (
	"context"
	"time"
)

// Client 通用 Redis 客户端（v8 版本）
type Client struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient 初始化 Redis 连接（v8 版本）
func NewRedisClient(addr, password string, db int) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// v8 的 Ping 调用方式（和 v9 一致，这里无差异）
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic("Redis 连接失败: " + err.Error())
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

// Get 获取缓存（v8 版本，API 和 v9 一致）
func (c *Client) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

// Set 设置缓存（v8 版本，API 和 v9 一致）
func (c *Client) Set(key string, value interface{}, expireSeconds int) error {
	return c.client.Set(c.ctx, key, value, time.Duration(expireSeconds)*time.Second).Err()
}

// Delete 删除缓存（v8 版本，API 和 v9 一致）
func (c *Client) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// GetClient 暴露底层客户端（给 Asynq 用）
func (c *Client) GetClient() *redis.Client {
	return c.client
}
