package main

import (
	"customs/api/router"
	"customs/infrastructure/db"
	"customs/infrastructure/minio"
	"customs/infrastructure/redis"
	"customs/repository"
	"customs/service"
	"customs/task"
	"log"
)

func main() {
	// 1. 初始化基础设施层
	mysqlClient, err := db.NewMySQLClient("root:123456@tcp(127.0.0.1:3306)/customs?parseTime=true&charset=utf8mb4")
	if err != nil {
		log.Fatal("MySQL初始化失败:", err)
	}
	minioClient, err := minio.NewMinioClient("127.0.0.1:9000", "minioadmin", "minioadmin", false)
	if err != nil {
		log.Fatal("MinIO初始化失败:", err)
	}
	redisClient := redis.NewRedisClient("127.0.0.1:6379", "", 0)

	// 2. 初始化Repository
	repoContainer := repository.NewRepositoryContainer(mysqlClient)

	// 3. 初始化Task
	taskClient := task.NewClient("127.0.0.1:6379", "", 0)
	taskInspector := task.NewInspector("127.0.0.1:6379", "", 0)
	defer func() {
		taskClient.Close()
		taskInspector.Close()
	}()

	// 4. 初始化Service
	serviceContainer := service.NewServiceContainer(
		mysqlClient,
		minioClient,
		redisClient,
		taskClient,
		taskInspector,
		repoContainer,
	)

	// 5. 初始化路由并启动HTTP服务
	r := router.NewRouter(serviceContainer)
	log.Println("HTTP服务启动成功，监听端口: 8080")
	if err := r.Run(":8080"); err != nil { // 监听8080端口
		log.Fatal("服务启动失败:", err)
	}
}
