package handler

import (
	"context"
	"customs/infrastructure/minio"
	"customs/infrastructure/redis"
	"customs/model"
	"customs/repository"
	"customs/task/payload"
	"encoding/json"
	"github.com/hibiken/asynq"
	"github.com/xuri/excelize/v2"
	"log"
	"time"
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
	ctx, cancel := context.WithTimeout(ctx, 290*time.Second)
	defer cancel()
	// 1. 解析任务参数
	p, err := payload.ParseCreateDFPayload(task)
	if err != nil {
		return err
	}

	// 2. 从MinIO下载Excel文件
	excelFileBytes, err := minioClient.DownloadFile("excel-bucket", p.ExcelName)
	if err != nil {
		// 更新任务状态为失败
		dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
		dictTask.UpdateCreateDFStatus(model.TaskStatusFailed, "下载Excel失败: "+err.Error())
		dictRepo.Update(ctx, dictTask)
		return err
	}

	// 3. 解析Excel具体逻辑
	// 3.1 打开Excel文件
	f, err := excelize.OpenReader(excelFileBytes)
	if err != nil {
		dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
		dictTask.UpdateCreateDFStatus(model.TaskStatusFailed, "打开Excel失败: "+err.Error())
		dictRepo.Update(ctx, dictTask)
		return err
	}
	defer f.Close()

	// 3.2 解析Excel数据
	parseResult := make(map[string]interface{}) // 存储最终解析结果

	// 遍历所有sheet
	for _, sheetName := range f.GetSheetList() {
		// 读取当前sheet的所有行
		rows, err := f.GetRows(sheetName)
		if err != nil {
			dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
			dictTask.UpdateCreateDFStatus(model.TaskStatusFailed, "读取sheet["+sheetName+"]失败: "+err.Error())
			dictRepo.Update(ctx, dictTask)
			return err
		}

		// 处理行数据（示例：第一行作为表头，后续行作为数据）
		if len(rows) == 0 {
			parseResult[sheetName] = []interface{}{}
			continue
		}

		header := rows[0]    // 表头行
		dataRows := rows[1:] // 数据行
		sheetData := make([]map[string]string, 0, len(dataRows))

		// 遍历数据行，组装键值对（表头为键，单元格为值）
		for _, row := range dataRows {
			rowData := make(map[string]string)
			for i, cell := range row {
				if i < len(header) { // 防止列数超过表头
					rowData[header[i]] = cell
				}
			}
			sheetData = append(sheetData, rowData)
		}

		// 将当前sheet的结果存入解析结果
		parseResult[sheetName] = sheetData
	}

	// 3.3 生成CSV文件名（保持原有逻辑）
	dbResourceCSVName := p.ExcelName + "_db.csv"
	dataDictionaryCSVName := p.ExcelName + "_dict.csv"
	csvName := p.ExcelName + "_all.csv"

	// 4. 将解析结果存入Redis
	redisKey := "dict_task_" + p.TaskID
	// resultJSON := 解析后的结果序列化
	resultJSON, err := json.Marshal(parseResult)
	if err != nil {
		dictTask, _ := dictRepo.GetByID(ctx, p.TaskID)
		dictTask.UpdateCreateDFStatus(model.TaskStatusFailed, "序列化解析结果失败: "+err.Error())
		dictRepo.Update(ctx, dictTask)
		return err
	}

	if err := redisClient.Set(redisKey, resultJSON, 12*3600); err != nil {
		log.Printf("缓存写入失败：%v, taskID=%s", err, p.TaskID)
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
