package payload

import (
	"encoding/json"
	"errors"
	"github.com/hibiken/asynq"
)

// InsertDFPayload 数据入库任务的参数
type InsertDFPayload struct {
	DBResourceCSVName     string `json:"db_resource_csv_name"`     // 数据库资源CSV名
	DataDictionaryCSVName string `json:"data_dictionary_csv_name"` // 数据字典CSV名
	TaskID                string `json:"task_id"`                  // 关联的DictionaryTask ID
}

// NewInsertDFTask 封装Payload为Asynq任务
func NewInsertDFTask(dbCSV, dictCSV, taskID string) (*asynq.Task, error) {
	p := InsertDFPayload{
		DBResourceCSVName:     dbCSV,
		DataDictionaryCSVName: dictCSV,
		TaskID:                taskID,
	}
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("task:insert_df", payloadBytes), nil
}

// ParseInsertDFPayload 解析任务参数
func ParseInsertDFPayload(task *asynq.Task) (*InsertDFPayload, error) {
	var p InsertDFPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return nil, errors.New("解析insert_df任务参数失败: " + err.Error())
	}
	return &p, nil
}
