package service

import (
	"context"
	"customs/common/errno"
	"customs/infrastructure/minio"
	"customs/infrastructure/redis"
	"customs/model"
	"customs/repository"
	"customs/task"
	"fmt"
	"mime/multipart"
	"path/filepath"
)

// DataDictionaryService 数据字典核心业务服务
type DataDictionaryService struct {
	minioClient   *minio.Client                    // MinIO工具（上传/下载Excel）
	redisClient   *redis.Client                    // Redis工具（缓存解析结果）
	taskClient    *task.Client                     // 异步任务生产者
	taskInspector *task.Inspector                  // 任务状态查询器
	dictRepo      *repository.DictionaryRepository // 任务记录CRUD
	dbResRepo     *repository.DBResourceRepository // 资源备注CRUD
}

// NewDataDictionaryService 初始化核心服务（依赖注入）
func NewDataDictionaryService(
	minioClient *minio.Client,
	redisClient *redis.Client,
	taskClient *task.Client,
	taskInspector *task.Inspector,
	dictRepo *repository.DictionaryRepository,
	dbResRepo *repository.DBResourceRepository,
) *DataDictionaryService {
	return &DataDictionaryService{
		minioClient:   minioClient,
		redisClient:   redisClient,
		taskClient:    taskClient,
		taskInspector: taskInspector,
		dictRepo:      dictRepo,
		dbResRepo:     dbResRepo,
	}
}

// UploadExcel 上传Excel
func (s *DataDictionaryService) UploadExcel(
	ctx context.Context,
	resourceComment string, // 资源备注
	file *multipart.FileHeader, // 上传的Excel文件
) (*model.DictionaryTask, error) {
	// 步骤1：参数校验
	if resourceComment == "" || file == nil {
		return nil, errno.ErrInvalidParam // 自定义错误码：参数无效
	}
	if !isExcelFile(file.Filename) { // 简单判断文件类型（可抽入common/utils）
		return nil, errno.ErrInvalidFileFormat // 自定义错误码：文件格式错误
	}

	// 步骤2：打开文件并上传到MinIO
	src, err := file.Open()
	if err != nil {
		return nil, errno.ErrFileOpenFailed // 自定义错误码：文件打开失败
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {

		}
	}(src)
	// 上传到MinIO的excel-bucket，文件名用原文件名（也可加前缀避免重复）
	err = s.minioClient.UploadFile("excel-bucket", file.Filename, src, file.Size)
	if err != nil {
		return nil, errno.ErrMinioUploadFailed // 自定义错误码：MinIO上传失败
	}

	// 步骤3：生产Asynq解析任务
	dictTask := model.NewDictionaryTask(file.Filename, "") // 先初始化任务记录（无taskID）
	// 先创建数据库任务记录
	if err := s.dictRepo.Create(ctx, dictTask); err != nil {
		return nil, errno.ErrDBInsertFailed // 自定义错误码：数据库插入失败
	}
	// 生产解析任务（获取Asynq的taskID）
	taskInfo, err := s.taskClient.CreateDFTask(ctx, resourceComment, file.Filename, dictTask.ID)
	if err != nil {
		// 任务生产失败，更新数据库状态
		dictTask.UpdateCreateDFStatus(model.TaskStatusFailed, "生产解析任务失败："+err.Error())
		err := s.dictRepo.Update(ctx, dictTask)
		if err != nil {
			return nil, err
		}
		return nil, errno.ErrTaskCreateFailed // 自定义错误码：任务创建失败
	}

	// 步骤4：更新任务记录的create_df_task_id
	dictTask.CreateDFTaskID = taskInfo.ID
	dictTask.UpdateCreateDFStatus(model.TaskStatusPending) // 状态改为待执行
	if err := s.dictRepo.Update(ctx, dictTask); err != nil {
		return nil, errno.ErrDBUpdateFailed // 自定义错误码：数据库更新失败
	}

	return dictTask, nil
}

// GetParseResult 查询解析结果
func (s *DataDictionaryService) GetParseResult(ctx context.Context, taskID string, page, size int) (interface{}, error) {
	// 步骤1：查询任务记录
	dictTask, err := s.dictRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("任务不存在：%w", err)
	}

	// 步骤2：只检查「解析状态是否成功」（关键修改：去掉对confirm的判定）
	if dictTask.CreateDFTaskStatus != model.TaskStatusSucceeded {
		return nil, fmt.Errorf("任务解析未完成，当前状态：%s", dictTask.CreateDFTaskStatus)
	}

	// 步骤3：读取Redis缓存
	redisKey := "dict_task_" + taskID
	result, err := s.redisClient.Get(redisKey)
	if err != nil {
		return nil, fmt.Errorf("缓存中无解析结果：%w", err)
	}

	return map[string]interface{}{
		"result": result,
		"page":   page,
		"size":   size,
	}, nil
}

// ConfirmInsert 确认入库
func (s *DataDictionaryService) ConfirmInsert(
	ctx context.Context,
	dictTaskID string,
	confirm bool,
) error {
	// 步骤1：查询数据库任务记录
	dictTask, err := s.dictRepo.GetByID(ctx, dictTaskID)
	if err != nil {
		return errno.ErrDBQueryFailed
	}

	// 步骤2：校验前置条件：解析任务必须成功
	if dictTask.CreateDFTaskStatus != model.TaskStatusSucceeded {
		return errno.ErrPreTaskNotCompleted // 自定义错误码：前置任务未完成
	}

	// 步骤3：取消入库（仅更新状态）
	if !confirm {
		dictTask.Confirm = false
		return s.dictRepo.Update(ctx, dictTask)
	}

	// 步骤4：确认入库：生产Asynq入库任务
	taskInfo, err := s.taskClient.InsertDFTask(
		ctx,
		dictTask.DBResourceCSVName,
		dictTask.DataDictionaryCSVName,
		dictTask.ID,
	)
	if err != nil {
		return errno.ErrTaskCreateFailed
	}

	// 步骤5：更新任务记录（标记确认+入库任务ID+状态）
	dictTask.ConfirmInsert(taskInfo.ID) // 调用Model的封装方法
	if err := s.dictRepo.Update(ctx, dictTask); err != nil {
		return errno.ErrDBUpdateFailed
	}

	// 步骤6：启动goroutine监控入库任务状态（对应back_task.py）
	go s.MonitorInsertTask(ctx, dictTaskID)

	return nil
}

// GetResourceComments 查询资源备注
func (s *DataDictionaryService) GetResourceComments(ctx context.Context) ([]string, error) {
	// 调用Repository查询去重的资源备注
	comments, err := s.dbResRepo.GetDistinctResourceComment(ctx)
	if err != nil {
		return nil, errno.ErrDBQueryFailed
	}
	return comments, nil
}

// isExcelFile 判断是否为Excel文件
func isExcelFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".xlsx" || ext == ".xls"
}

// paginateResult 分页处理JSON字符串（示例逻辑，可根据实际调整）
func paginateResult(jsonStr string, page, size int) interface{} {
	// 实际场景：解析JSON为数组，按page/size截取
	// 这里简化返回，仅标记分页信息
	return map[string]interface{}{
		"data":  jsonStr,
		"page":  page,
		"size":  size,
		"total": 100, // 示例总数
	}
}
