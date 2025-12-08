package model

// 任务状态常量
const (
	TaskStatusPending   = "PENDING"   // 待执行
	TaskStatusRunning   = "RUNNING"   // 执行中
	TaskStatusSucceeded = "SUCCEEDED" // 执行成功
	TaskStatusFailed    = "FAILED"    // 执行失败
)

// 数据库表名常量
const (
	TableNameDictionaryTask = "dictionary_task"
	TableNameDBResource     = "db_resource"
)
