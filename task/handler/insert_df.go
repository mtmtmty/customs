package handler

import (
	"context"
	"customs/infrastructure/minio"
	"customs/model"
	"customs/repository"
	"customs/task/payload"
	"github.com/hibiken/asynq"
)

// InsertDFHandler 数据入库任务的消费逻辑
func InsertDFHandler(
	ctx context.Context,
	task *asynq.Task,
	minioClient *minio.Client,
	dictRepo *repository.DictionaryRepository,
	dbResRepo *repository.DBResourceRepository,
) error {
	// 1. 解析任务参数
	p, err := payload.ParseInsertDFPayload(task)
	if err != nil {
		return err
	}

	// 2. 从MinIO下载CSV文件
	dbCSVFile, err := minioClient.DownloadFile("csv-bucket", p.DBResourceCSVName)
	if err != nil {
		dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
		dictTask.UpdateInsertDFStatus(model.TaskStatusFailed, "下载DB CSV失败: "+err.Error())
		dictRepo.Update(ctx, dictTask)
		return err
	}
	dictCSVFile, err := minioClient.DownloadFile("csv-bucket", p.DataDictionaryCSVName)
	if err != nil {
		dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
		dictTask.UpdateInsertDFStatus(model.TaskStatusFailed, "下载Dict CSV失败: "+err.Error())
		dictRepo.Update(ctx, dictTask)
		return err
	}

	// 3. 解析CSV并入库（核心业务逻辑，替换为实际入库代码）
	// ......CSV解析+数据库入库逻辑......

	// 4. 更新任务状态为成功
	dictTask, err := dictRepo.GetByID(ctx, p.TaskID)
	if err != nil {
		return err
	}
	dictTask.UpdateInsertDFStatus(model.TaskStatusSucceeded)
	return dictRepo.Update(ctx, dictTask)
}
