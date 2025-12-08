package task

import (
	"context"
	"customs/model"
	"github.com/hibiken/asynq"
)

// Inspector 任务状态查询器（完整实现）
type Inspector struct {
	inspector *asynq.Inspector
}

// NewInspector 初始化查询器（复用Redis配置）
func NewInspector(redisAddr, redisPassword string, redisDB int) *Inspector {
	// 创建Asynq Inspector实例（底层依赖Redis）
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	return &Inspector{inspector: inspector}
}

// GetTaskStatus 查询任务状态（转换为Model层的统一状态常量）
func (i *Inspector) GetTaskStatus(ctx context.Context, taskID string) (string, error) {
	// 调用Asynq Inspector获取任务详情
	info, err := i.inspector.GetTaskInfo(ctx, taskID)
	if err != nil {
		// 区分“任务不存在”和“查询失败”
		if asynq.IsTaskNotFoundError(err) {
			return model.TaskStatusFailed, nil // 任务不存在视为失败
		}
		return "", err // 其他错误直接返回
	}

	// 将Asynq的State转换为Model层定义的状态常量（统一语义）
	switch info.State {
	case asynq.StatePending:
		return model.TaskStatusPending, nil
	case asynq.StateProcessing:
		return model.TaskStatusRunning, nil
	case asynq.StateSucceeded:
		return model.TaskStatusSucceeded, nil
	case asynq.StateFailed:
		return model.TaskStatusFailed, nil
	default:
		return string(info.State), nil // 未知状态返回原始值
	}
}

// Close 关闭Inspector（可选，优雅退出）
func (i *Inspector) Close() error {
	return i.inspector.Close()
}
