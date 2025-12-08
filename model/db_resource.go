package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DBResource 数据库资源表（存储资源备注、类型等）
type DBResource struct {
	ID              string         `gorm:"column:id;primaryKey;comment:资源ID" json:"id"`
	ResourceComment string         `gorm:"column:resource_comment;index;comment:资源备注（如：署级系统-下发数据）" json:"resource_comment"`
	ResourceType    string         `gorm:"column:resource_type;comment:资源类型（如：MySQL/Oracle）" json:"resource_type"`
	DBName          string         `gorm:"column:db_name;comment:数据库名" json:"db_name"`
	TableNames      string         `gorm:"column:table_names;comment:关联表名（逗号分隔）" json:"table_names"`
	Creator         string         `gorm:"column:creator;comment:创建人" json:"creator"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"deleted_at,omitempty"`
}

// TableName 指定GORM映射的数据库表名
func (DBResource) TableName() string {
	return TableNameDBResource
}

// NewDBResource 初始化资源记录
func NewDBResource(resourceComment, resourceType, dbName, tableNames, creator string) *DBResource {
	return &DBResource{
		ID:              uuid.New().String(),
		ResourceComment: resourceComment,
		ResourceType:    resourceType,
		DBName:          dbName,
		TableNames:      tableNames,
		Creator:         creator,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
