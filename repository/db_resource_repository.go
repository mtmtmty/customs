package repository

import (
	"context"
	"customs/infrastructure/db"
	"customs/model"
)

// DBResourceRepository 处理 DBResource 的 CRUD
type DBResourceRepository struct {
	mysqlClient *db.MySQLClient
}

// NewDBResourceRepository 初始化仓库
func NewDBResourceRepository(mysqlClient *db.MySQLClient) *DBResourceRepository {
	return &DBResourceRepository{mysqlClient: mysqlClient}
}

// GetDistinctResourceComment 查询去重的资源备注（对应 Python 的 get_resource_comment）
func (r *DBResourceRepository) GetDistinctResourceComment(ctx context.Context) ([]string, error) {
	var comments []string
	err := r.mysqlClient.GetDB().WithContext(ctx).
		Model(&model.DBResource{}).
		Distinct("resource_comment").
		Find(&comments).Error
	return comments, err
}

// 可选：其他基础 CRUD
func (r *DBResourceRepository) Create(ctx context.Context, resource *model.DBResource) error {
	return r.mysqlClient.GetDB().WithContext(ctx).Create(resource).Error
}

func (r *DBResourceRepository) GetByComment(ctx context.Context, comment string) (*model.DBResource, error) {
	var resource model.DBResource
	err := r.mysqlClient.GetDB().WithContext(ctx).Where("resource_comment = ?", comment).First(&resource).Error
	return &resource, err
}
