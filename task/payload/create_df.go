package payload

import (
	"encoding/json"
	"errors"
	"github.com/hibiken/asynq"
)

// CreateDFPayload 解析Excel任务的参数
type CreateDFPayload struct {
	ResourceComment string `json:"resource_comment"` // 资源备注
	ExcelName       string `json:"excel_name"`       // MinIO中的Excel文件名
	TaskID          string `json:"task_id"`          // 关联的DictionaryTask ID
}

// NewCreateDFTask 封装Payload为Asynq任务
func NewCreateDFTask(rc, excelName, taskID string) (*asynq.Task, error) {
	p := CreateDFPayload{
		ResourceComment: rc,
		ExcelName:       excelName,
		TaskID:          taskID,
	}
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	// 任务类型标记为"task:create_df"，用于Worker识别处理器
	return asynq.NewTask("task:create_df", payloadBytes), nil
}

// ParseCreateDFPayload 解析任务参数
func ParseCreateDFPayload(task *asynq.Task) (*CreateDFPayload, error) {
	var p CreateDFPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return nil, errors.New("解析create_df任务参数失败: " + err.Error())
	}
	return &p, nil
}
