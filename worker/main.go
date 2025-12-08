package main

import (
	"customs/infrastructure/db"
	"customs/infrastructure/minio"
	"customs/infrastructure/redis"
	"customs/repository"
	"customs/task/handler"
	"github.com/hibiken/asynq"
	"log"
)

func main() {
	// 初始化依赖
	mysqlClient, _ := db.NewMySQLClient("user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4")
	minioClient, _ := minio.NewMinioClient("127.0.0.1:9000", "AK", "SK", false)
	redisClient := redis.NewRedisClient("127.0.0.1:6379", "", 0)
	repoContainer := repository.NewRepositoryContainer(mysqlClient)

	// 初始化Asynq Worker
	worker := asynq.NewWorker(
		asynq.RedisClientOpt{Addr: "127.0.0.1:6379", Password: "", DB: 0},
		asynq.Config{Concurrency: 5}, // 并发数5
	)

	// 注册任务处理器
	mux := asynq.NewServeMux()
	mux.HandleFunc("task:create_df", func(ctx context.Context, t *asynq.Task) error {
		return handler.CreateDFHandler(ctx, t, minioClient, redisClient, repoContainer.Dictionary, repoContainer.DBResource)
	})
	mux.HandleFunc("task:insert_df", func(ctx context.Context, t *asynq.Task) error {
		return handler.InsertDFHandler(ctx, t, minioClient, repoContainer.Dictionary, repoContainer.DBResource)
	})

	// 启动Worker
	log.Println("Worker启动成功，监听任务队列...")
	if err := worker.Run(mux); err != nil {
		log.Fatal("Worker启动失败:", err)
	}
}
