package db

import (
	"customs/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLClient 通用 MySQL 客户端（无业务感知）
type MySQLClient struct {
	db *gorm.DB
}

// NewMySQLClient 初始化 MySQL 连接
func NewMySQLClient(dsn string) (*MySQLClient, error) {
	// 配置 GORM（日志级别、连接池等）
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印 SQL 日志（开发环境）
	})
	if err != nil {
		return nil, err
	}

	// 自动创建/更新表结构（基于 Model 定义）
	err = db.AutoMigrate(
		&model.DictionaryTask{},
		&model.DBResource{},
	)
	if err != nil {
		return nil, err
	}

	return &MySQLClient{db: db}, nil
}

// GetDB 暴露底层 GORM DB 实例（给 Repository 层用）
func (c *MySQLClient) GetDB() *gorm.DB {
	return c.db
}

// WithTransaction 开启事务（可选，复杂业务用）
func (c *MySQLClient) WithTransaction(fn func(tx *gorm.DB) error) error {
	tx := c.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
