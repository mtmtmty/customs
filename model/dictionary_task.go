package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DictionaryTask 数据字典任务表
type DictionaryTask struct {
	ID                    string         `gorm:"column:id;primaryKey;comment:任务ID" json:"id"`
	CreateDFTaskID        string         `gorm:"column:create_df_task_id;comment:Asynq创建数据帧任务ID" json:"create_df_task_id"`
	ExcelName             string         `gorm:"column:excel_name;comment:上传的Excel文件名" json:"excel_name"`
	CreateDFTaskStatus    string         `gorm:"column:create_df_task_status;comment:创建数据帧任务状态" json:"create_df_task_status"`
	CreateDFTaskRemark    string         `gorm:"column:create_df_task_remark;comment:创建数据帧任务备注（失败原因）" json:"create_df_task_remark"`
	DBResourceCSVName     string         `gorm:"column:db_resource_csv_name;comment:数据库资源CSV文件名" json:"db_resource_csv_name"`
	DataDictionaryCSVName string         `gorm:"column:data_dictionary_csv_name;comment:数据字典CSV文件名" json:"data_dictionary_csv_name"`
	CSVName               string         `gorm:"column:csv_name;comment:通用CSV文件名" json:"csv_name"`
	InsertDFTaskID        string         `gorm:"column:insert_df_task_id;comment:Asynq插入数据库任务ID" json:"insert_df_task_id"`
	InsertDFTaskStatus    string         `gorm:"column:insert_df_task_status;comment:插入数据库任务状态" json:"insert_df_task_status"`
	InsertDFTaskRemark    string         `gorm:"column:insert_df_task_remark;comment:插入数据库任务备注（失败原因）" json:"insert_df_task_remark"`
	Confirm               bool           `gorm:"column:confirm;default:false;comment:是否确认插入数据库" json:"confirm"`
	CreatedAt             time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"deleted_at,omitempty"`
}

// TableName 指定GORM映射的数据库表名
func (DictionaryTask) TableName() string {
	return TableNameDictionaryTask
}

// NewDictionaryTask 初始化任务
func NewDictionaryTask(excelName, createDFTaskID string) *DictionaryTask {
	return &DictionaryTask{
		ID:                 uuid.New().String(), // 生成UUID作为主键
		ExcelName:          excelName,
		CreateDFTaskID:     createDFTaskID,
		CreateDFTaskStatus: TaskStatusPending, // 默认待执行
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Confirm:            false,
	}
}

// UpdateCreateDFStatus 更新创建数据帧任务状态
func (t *DictionaryTask) UpdateCreateDFStatus(status string, remark ...string) {
	t.CreateDFTaskStatus = status
	if len(remark) > 0 {
		t.CreateDFTaskRemark = remark[0]
	}
	t.UpdatedAt = time.Now()
}

// UpdateInsertDFStatus 更新插入数据库任务状态
func (t *DictionaryTask) UpdateInsertDFStatus(status string, remark ...string) {
	t.InsertDFTaskID = status
	if len(remark) > 0 {
		t.InsertDFTaskRemark = remark[0]
	}
	t.UpdatedAt = time.Now()
}

// ConfirmInsert 标记为确认插入
func (t *DictionaryTask) ConfirmInsert(insertDFTaskID string) {
	t.Confirm = true
	t.InsertDFTaskID = insertDFTaskID
	t.InsertDFTaskStatus = TaskStatusPending
	t.UpdatedAt = time.Now()
}
