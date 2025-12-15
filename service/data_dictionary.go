package service

import (
	"bytes"
	"context"
	"customs/common"
	"customs/common/errno"
	"customs/infrastructure/minio"
	"customs/infrastructure/redis"
	"customs/model"
	"customs/repository"
	"customs/task"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
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

// DownloadTemplate 从MinIO获取Excel模板文件
func (s *DataDictionaryService) DownloadTemplate(ctx context.Context) (io.Reader, string, error) {
	templateName := "system-db.xls" // 模板文件名，与Python原逻辑保持一致

	// 从MinIO下载模板文件
	fileBytes, err := s.minioClient.DownloadFile("sjdt-update-dictionary-config-excel", templateName)
	if err != nil {
		return nil, "", fmt.Errorf("minio下载失败：%w", err)
	}

	return fileBytes, templateName, nil
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
		return nil, errno.ErrFileOpenFailed.WithMessage(err.Error())
	}
	defer func() {
		if err := src.Close(); err != nil {
			// 记录关闭文件失败的日志，不阻断主流程
			log.Printf("关闭文件失败: %v, 文件名: %s", err, file.Filename)
		}
	}()

	// 读取文件内容用于格式验证（提前发现问题，避免无效上传）
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, errno.ErrFileOpenFailed.WithMessage("读取文件内容失败: " + err.Error())
	}

	// 步骤3：预验证Excel格式（文件名+内容结构）
	if err := judgeExcelFormat(file.Filename, content); err != nil {
		return nil, err
	}

	// 重置文件指针，确保后续上传完整
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, errno.ErrFileOpenFailed.WithMessage("重置文件指针失败: " + err.Error())
	}
	// 上传到MinIO的excel-bucket，文件名用原文件名
	err = s.minioClient.UploadFile("sjdt-update-dictionary-config-excel", file.Filename, src, file.Size)
	if err != nil {
		return nil, errno.ErrMinioUploadFailed // 自定义错误码：MinIO上传失败
	}

	// 步骤4：生产Asynq解析任务
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

	// 步骤5：更新任务记录的create_df_task_id
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

	// 步骤2：只检查「解析状态是否成功」
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

// paginateResult 分页处理JSON字符串
func paginateResult(jsonStr string, page, size int) (interface{}, error) {
	// 1. 解析JSON字符串为切片（假设原数据是数组格式）
	var dataList []interface{}
	if err := json.Unmarshal([]byte(jsonStr), &dataList); err != nil {
		return nil, err // 解析失败返回错误
	}

	// 2. 计算总条数
	total := len(dataList)

	// 3. 调用通用分页工具计算偏移量和分页信息
	offset, limit, pageInfo := common.Paginate(total, page, size)

	// 4. 截取分页数据（处理边界情况：偏移量超过总条数时返回空数组）
	var paginatedData []interface{}
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedData = dataList[offset:end]
	} else {
		paginatedData = []interface{}{}
	}

	// 5. 组装分页结果
	return map[string]interface{}{
		"data":  paginatedData,     // 分页后的数据列表
		"page":  pageInfo["page"],  // 当前页码
		"size":  pageInfo["size"],  // 每页条数
		"total": pageInfo["total"], // 总条数
	}, nil
}

// judgeExcelFormat 验证Excel文件名格式和列名是否符合规范
// 文件名要求："系统名-dbname"（用短横线分割为两部分）
// 列名要求：必须包含指定的5个标准列
func judgeExcelFormat(fileName string, fileContent []byte) error {
	// 1. 验证文件名格式（系统名-dbname）
	parts := strings.Split(fileName, "-")
	if len(parts) != 2 {
		return errno.ErrInvalidFileNameFormat // 自定义错误：文件名格式错误
	}

	// 2. 验证Excel列名是否包含所有标准列
	// 打开Excel文件（从字节流读取）
	f, err := excelize.OpenReader(bytes.NewReader(fileContent))
	if err != nil {
		return errno.ErrExcelOpenFailed // 自定义错误：打开Excel失败
	}
	defer f.Close()

	// 获取第一个sheet的列名（默认读取第一个sheet）
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return errno.ErrExcelNoSheet // 自定义错误：Excel无工作表
	}
	rows, err := f.GetRows(sheetList[0])
	if err != nil {
		return errno.ErrExcelReadFailed // 自定义错误：读取Excel失败
	}
	if len(rows) == 0 {
		return errno.ErrExcelEmpty // 自定义错误：Excel内容为空
	}

	// 提取第一行作为列名
	columns := rows[0]
	columnSet := make(map[string]struct{}, len(columns))
	for _, col := range columns {
		columnSet[col] = struct{}{}
	}

	// 标准列名集合（与Python版本保持一致）
	stdColumns := map[string]struct{}{
		"数据表名称（英文）":    {},
		"数据表名称（中文）":    {},
		"字段/数据项名称（英文）": {},
		"字段/数据项名称（中文）": {},
		"字段/数据项说明":     {},
	}

	// 检查是否包含所有标准列
	for col := range stdColumns {
		if _, exists := columnSet[col]; !exists {
			return errno.ErrExcelColumnMissing // 自定义错误：缺少必要列
		}
	}

	return nil
}
