package task

import (
	"customs/model"
	"errors"
	"github.com/hibiken/asynq"
	"strconv"
)

// Inspector 任务状态查询器
type Inspector struct {
	inspector *asynq.Inspector
}

// NewInspector 初始化查询器
func NewInspector(redisAddr, redisPassword string, redisDB int) *Inspector {
	return &Inspector{
		inspector: asynq.NewInspector(asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		}),
	}
}

// GetTaskStatus 查询任务状态（适配 v0.24.0）
func (i *Inspector) GetTaskStatus(taskID string) (string, error) {
	// 旧版本需要队列名称，通常使用 "default"
	queue := "default"

	// 旧版本 API：GetTaskInfo(queue, taskID string)
	info, err := i.inspector.GetTaskInfo(queue, taskID)
	if err != nil {
		// 旧版本没有 IsErrTaskNotFound，需要手动判断
		if errors.Is(err, asynq.ErrTaskNotFound) {
			return model.TaskStatusFailed, nil
		}
		return "", err
	}

	// v0.24.0 中的状态常量命名（与新版本一致，但需确保导入正确）
	switch info.State {
	case asynq.TaskStatePending:
		return model.TaskStatusPending, nil
	case asynq.TaskStateActive:
		return model.TaskStatusRunning, nil
	case asynq.TaskStateCompleted:
		return model.TaskStatusSucceeded, nil
	case asynq.TaskStateArchived:
		if info.CompletedAt.IsZero() {
			return model.TaskStatusFailed, nil
		}
		return model.TaskStatusSucceeded, nil
	default:
		return strconv.Itoa(int(info.State)), nil
	}
}

// Close 关闭Inspector
func (i *Inspector) Close() error {
	return i.inspector.Close()
}
