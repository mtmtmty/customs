package handler

import (
	"context"
	"customs/infrastructure/minio"
	"customs/infrastructure/redis"
	"customs/model"
	"customs/repository"
	"customs/task/payload"
	"github.com/hibiken/asynq"
)

// CreateDFHandler 解析Excel任务的消费逻辑
func CreateDFHandler(
	ctx context.Context,
	task *asynq.Task,
	minioClient *minio.Client,
	redisClient *redis.Client,
	dictRepo *repository.DictionaryRepository,
	dbResRepo *repository.DBResourceRepository,
) error {
	// 1. 解析任务参数
	p, err := payload.ParseCreateDFPayload(task)
	if err != nil {
		return err
	}

	// 2. 从MinIO下载Excel文件
	excelFile, err := minioClient.DownloadFile("excel-bucket", p.ExcelName)
	if err != nil {
		// 更新任务状态为失败
		dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
		dictTask.UpdateCreateDFStatus(model.TaskStatusFailed, "下载Excel失败: "+err.Error())
		dictRepo.Update(ctx, dictTask)
		return err
	}

	// 3. 解析Excel（核心业务逻辑，替换为实际解析代码）
	// file, _ := xlsx.OpenReader(excelFile)
	// ......解析逻辑......
	// 生成DBResourceCSVName、DataDictionaryCSVName、CSVName
	dbResourceCSVName := p.ExcelName + "_db.csv"
	dataDictionaryCSVName := p.ExcelName + "_dict.csv"
	csvName := p.ExcelName + "_all.csv"

	// 4. 将解析结果存入Redis（供API查询）
	redisKey := "dict_task_" + p.TaskID
	// resultJSON := 解析后的结果序列化
	resultJSON := `{"data": "解析结果示例"}`
	if err := redisClient.Set(redisKey, resultJSON, 12*3600); err != nil {
		dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
		dictTask.UpdateCreateDFStatus(model.TaskStatusFailed, "缓存解析结果失败: "+err.Error())
		dictRepo.Update(ctx, dictTask)
		return err
	}

	// 5. 更新任务状态为成功
	dictTask, err := dictRepo.GetByID(ctx, p.TaskID)
	if err != nil {
		return err
	}
	dictTask.DBResourceCSVName = dbResourceCSVName
	dictTask.DataDictionaryCSVName = dataDictionaryCSVName
	dictTask.CSVName = csvName
	dictTask.UpdateCreateDFStatus(model.TaskStatusSucceeded)
	return dictRepo.Update(ctx, dictTask)
}
