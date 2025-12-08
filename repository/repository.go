package repository

import "customs/infrastructure/db"

// RepositoryContainer 封装所有仓库实例
type RepositoryContainer struct {
	Dictionary *DictionaryRepository // 任务记录仓库
	DBResource *DBResourceRepository // 资源备注仓库
}

// NewRepositoryContainer 初始化所有仓库（注入 Infrastructure 层的 MySQL 客户端）
func NewRepositoryContainer(mysqlClient *db.MySQLClient) *RepositoryContainer {
	return &RepositoryContainer{
		Dictionary: NewDictionaryRepository(mysqlClient),
		DBResource: NewDBResourceRepository(mysqlClient),
	}
}
