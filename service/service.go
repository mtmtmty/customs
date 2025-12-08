package service

import (
	"customs/infrastructure/db"
	"customs/infrastructure/minio"
	"customs/infrastructure/redis"
	"customs/repository"
	"customs/task"
)

// ServiceContainer 封装所有Service实例
type ServiceContainer struct {
	DataDictionary *DataDictionaryService // 核心：数据字典业务服务
}

// NewServiceContainer 初始化所有Service
func NewServiceContainer(
	mysqlClient *db.MySQLClient,
	minioClient *minio.Client,
	redisClient *redis.Client,
	taskClient *task.Client,
	taskInspector *task.Inspector,
	repoContainer *repository.RepositoryContainer,
) *ServiceContainer {
	return &ServiceContainer{
		DataDictionary: NewDataDictionaryService(
			minioClient,
			redisClient,
			taskClient,
			taskInspector,
			repoContainer.Dictionary,
			repoContainer.DBResource,
		),
	}
}
