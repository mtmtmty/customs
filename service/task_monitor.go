package service

import (
	"context"
	"customs/model"
	"time"
)

// MonitorInsertTask 监控入库任务状态（后台goroutine执行）
func (s *DataDictionaryService) MonitorInsertTask(ctx context.Context, dictTaskID string) {
	// 循环查询状态，直到任务完成
	for {
		// 步骤1：查询数据库任务记录
		dictTask, err := s.dictRepo.GetByID(ctx, dictTaskID)
		if err != nil {
			break // 任务不存在，退出监控
		}
		if dictTask.InsertDFTaskID == "" {
			time.Sleep(10 * time.Second)
			continue
		}

		// 步骤2：查询Asynq任务状态
		taskStatus, err := s.taskInspector.GetTaskStatus(ctx, dictTask.InsertDFTaskID)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		// 步骤3：更新任务状态
		dictTask.UpdateInsertDFStatus(taskStatus)
		s.dictRepo.Update(ctx, dictTask)

		// 步骤4：任务完成（成功/失败），退出循环
		if taskStatus == model.TaskStatusSucceeded || taskStatus == model.TaskStatusFailed {
			break
		}

		// 每10秒查询一次
		time.Sleep(10 * time.Second)
	}
}
