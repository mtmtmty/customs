package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// Client 通用 Redis 客户端
type Client struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient 初始化 Redis 连接
func NewRedisClient(addr, password string, db int) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Ping 调用方式
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic("Redis 连接失败: " + err.Error())
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

// Get 获取缓存
func (c *Client) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

// Set 设置缓存
func (c *Client) Set(key string, value interface{}, expireSeconds int) error {
	return c.client.Set(c.ctx, key, value, time.Duration(expireSeconds)*time.Second).Err()
}

// Delete 删除缓存
func (c *Client) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// GetClient 暴露底层客户端
func (c *Client) GetClient() *redis.Client {
	return c.client
}
