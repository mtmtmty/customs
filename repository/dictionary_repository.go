package repository

import (
	"context"
	"customs/infrastructure/db"
	"customs/model"
)

// DictionaryRepository 处理 DictionaryTask 的 CRUD
type DictionaryRepository struct {
	mysqlClient *db.MySQLClient // 依赖 Infrastructure 层的通用 MySQL 能力
}

// NewDictionaryRepository 初始化仓库
func NewDictionaryRepository(mysqlClient *db.MySQLClient) *DictionaryRepository {
	return &DictionaryRepository{mysqlClient: mysqlClient}
}

// Create 创建任务记录（对应 Python 的 add+commit）
func (r *DictionaryRepository) Create(ctx context.Context, task *model.DictionaryTask) error {
	return r.mysqlClient.GetDB().WithContext(ctx).Create(task).Error
}

// GetByID 根据 ID 查询任务（最常用）
func (r *DictionaryRepository) GetByID(ctx context.Context, id string) (*model.DictionaryTask, error) {
	var task model.DictionaryTask
	err := r.mysqlClient.GetDB().WithContext(ctx).Where("id = ?", id).First(&task).Error
	return &task, err
}

// Update 更新任务记录（如状态、CSV 文件名）
func (r *DictionaryRepository) Update(ctx context.Context, task *model.DictionaryTask) error {
	return r.mysqlClient.GetDB().WithContext(ctx).Save(task).Error
}

// GetByCreateDFTaskID 根据 create_df_task_id 查询任务（关联 Asynq 任务）
func (r *DictionaryRepository) GetByCreateDFTaskID(ctx context.Context, taskID string) (*model.DictionaryTask, error) {
	var task model.DictionaryTask
	err := r.mysqlClient.GetDB().WithContext(ctx).Where("create_df_task_id = ?", taskID).First(&task).Error
	return &task, err
}

// GetByInsertDFTaskID 根据 insert_df_task_id 查询任务
func (r *DictionaryRepository) GetByInsertDFTaskID(ctx context.Context, taskID string) (*model.DictionaryTask, error) {
	var task model.DictionaryTask
	err := r.mysqlClient.GetDB().WithContext(ctx).Where("insert_df_task_id = ?", taskID).First(&task).Error
	return &task, err
}
