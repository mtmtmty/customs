package task

import (
	"context"
	"customs/task/payload" // 替换为你的模块名
	"github.com/hibiken/asynq"
	"time"
)

// Client 异步任务生产者客户端
type Client struct {
	asynqClient *asynq.Client
}

// NewClient 初始化生产者（复用Infrastructure层的Redis配置）
func NewClient(redisAddr, redisPassword string, redisDB int) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	return &Client{asynqClient: client}
}

// CreateDFTask 生产“解析Excel”任务
func (c *Client) CreateDFTask(ctx context.Context, resourceComment, excelName, taskID string) (*asynq.TaskInfo, error) {
	task, err := payload.NewCreateDFTask(resourceComment, excelName, taskID)
	if err != nil {
		return nil, err
	}
	// 入队任务（可配置超时、重试策略）
	return c.asynqClient.EnqueueContext(ctx, task,
		asynq.MaxRetry(3),               // 失败重试3次
		asynq.Timeout(5*60*time.Second), // 超时5分钟
		asynq.Queue("excel"),            // 指定队列（可选，用于任务优先级）
	)
}

// InsertDFTask 生产“数据入库”任务
func (c *Client) InsertDFTask(ctx context.Context, dbCSV, dictCSV, taskID string) (*asynq.TaskInfo, error) {
	task, err := payload.NewInsertDFTask(dbCSV, dictCSV, taskID)
	if err != nil {
		return nil, err
	}
	return c.asynqClient.EnqueueContext(ctx, task,
		asynq.MaxRetry(3),
		asynq.Timeout(10*60*time.Second),
		asynq.Queue("db"),
	)
}

// Close 关闭客户端
func (c *Client) Close() error {
	return c.asynqClient.Close()
}
