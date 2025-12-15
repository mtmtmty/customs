package handler

import (
	"customs/api/response"
	"customs/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

// DataDictionaryHandler 数据字典接口处理器
type DataDictionaryHandler struct {
	svc *service.DataDictionaryService // 依赖Service层
}

// NewDataDictionaryHandler 初始化处理器（注入Service依赖）
func NewDataDictionaryHandler(svc *service.DataDictionaryService) *DataDictionaryHandler {
	return &DataDictionaryHandler{svc: svc}
}

// DownloadTemplate 下载数据字典Excel模板
// @Summary 下载数据字典Excel模板
// @Description 获取系统数据字典导入模板（system-db.xls）
// @Tags 数据字典
// @Produce application/octet-stream
// @Success 200 {file} file "模板文件流"
// @Failure 400 {object} response.Response "获取模板失败"
// @Router /api/data_dictionary/insert [get]
func (h *DataDictionaryHandler) DownloadTemplate(c *gin.Context) {
	// 调用Service获取文件流和文件名
	fileReader, fileName, err := h.svc.DownloadTemplate(c.Request.Context())
	if err != nil {
		response.Fail(c, response.ErrCodeFileError, err.Error())
		return
	}

	// 设置下载响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))

	// 根据文件后缀设置MIME类型
	var contentType string
	if strings.HasSuffix(fileName, ".xlsx") {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	} else {
		contentType = "application/vnd.ms-excel"
	}
	c.Header("Content-Type", contentType) // 正确设置Content-Type

	// 直接通过流返回文件内容，使用上面定义的contentType变量
	c.DataFromReader(http.StatusOK, -1, contentType, fileReader, nil)
}

// UploadExcel 上传Excel
// @Summary 上传Excel文件并触发解析
// @Description 上传Excel文件，关联资源备注，生产解析异步任务
// @Tags 数据字典
// @Accept multipart/form-data
// @Param resource_comment formData string true "资源备注"
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response{data=model.DictionaryTask}
// @Router /api/data_dictionary/insert [post]
func (h *DataDictionaryHandler) UploadExcel(c *gin.Context) {
	// 步骤1：解析表单参数
	resourceComment := c.PostForm("resource_comment")
	if resourceComment == "" {
		response.Fail(c, response.ErrCodeInvalidParam, "资源备注不能为空")
		return
	}

	// 步骤2：解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, response.ErrCodeFileError, "文件上传失败："+err.Error())
		return
	}

	// 步骤3：调用Service层方法
	dictTask, err := h.svc.UploadExcel(c.Request.Context(), resourceComment, file)
	if err != nil {
		// 根据错误类型返回对应提示（这里简化，直接返回错误信息）
		response.Fail(c, response.ErrCodeBusinessError, err.Error())
		return
	}

	// 步骤4：返回成功响应
	response.Success(c, dictTask)
}

// GetParseResult 查询解析结果接口
// @Summary 查询Excel解析结果
// @Description 根据任务ID查询解析结果（缓存/任务状态）
// @Tags 数据字典
// @Param task_id query string true "字典任务ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页条数" default(10)
// @Success 200 {object} response.Response
// @Router /api/data_dictionary/insert/data [get]
func (h *DataDictionaryHandler) GetParseResult(c *gin.Context) {
	// 步骤1：解析查询参数
	taskID := c.Query("task_id")
	if taskID == "" {
		response.Fail(c, response.ErrCodeInvalidParam, "任务ID不能为空")
		return
	}

	// 步骤2：获取分页参数并转换为int（解决未使用变量问题）
	pageStr := c.DefaultQuery("page", "1")  // 默认页码1
	sizeStr := c.DefaultQuery("size", "10") // 默认每页10条

	// 转换page为int，并校验合法性
	pageInt, err := strconv.Atoi(pageStr)
	if err != nil || pageInt < 1 {
		response.Fail(c, response.ErrCodeInvalidParam, "页码必须为正整数")
		return
	}

	// 转换size为int，并校验合法性（限制最大条数，避免查询过多数据）
	sizeInt, err := strconv.Atoi(sizeStr)
	if err != nil || sizeInt < 1 || sizeInt > 100 {
		response.Fail(c, response.ErrCodeInvalidParam, "每页条数必须为1-100的整数")
		return
	}

	// 步骤3：调用Service层方法（使用转换后的pageInt和sizeInt）
	result, err := h.svc.GetParseResult(c.Request.Context(), taskID, pageInt, sizeInt)
	if err != nil {
		response.Fail(c, response.ErrCodeBusinessError, err.Error())
		return
	}

	// 步骤4：返回成功响应
	response.Success(c, result)
}

// ConfirmInsert 确认入库接口
// @Summary 确认/取消Excel解析结果入库
// @Description 根据任务ID确认入库，生产入库异步任务
// @Tags 数据字典
// @Param task_id query string true "字典任务ID"
// @Param confirm query bool true "是否确认入库"
// @Success 200 {object} response.Response
// @Router /api/data_dictionary/insert/{id} [post]
func (h *DataDictionaryHandler) ConfirmInsert(c *gin.Context) {
	// 步骤1：解析参数
	taskID := c.Param("id")
	if taskID == "" {
		response.Fail(c, response.ErrCodeInvalidParam, "任务ID不能为空")
		return
	}
	confirm := c.Query("confirm") == "true" // 转换为bool

	// 步骤2：调用Service层方法
	err := h.svc.ConfirmInsert(c.Request.Context(), taskID, confirm)
	if err != nil {
		response.Fail(c, response.ErrCodeBusinessError, err.Error())
		return
	}

	// 步骤3：返回成功响应
	response.Success(c, gin.H{"msg": "操作成功"})
}

// GetResourceComments 查询资源备注接口
// @Summary 查询所有资源备注
// @Description 获取去重的资源备注列表
// @Tags 数据字典
// @Success 200 {object} response.Response{data=[]string}
// @Router /api/data_dictionary/resource_comment [get]
func (h *DataDictionaryHandler) GetResourceComments(c *gin.Context) {
	// 步骤1：调用Service层方法
	comments, err := h.svc.GetResourceComments(c.Request.Context())
	if err != nil {
		response.Fail(c, response.ErrCodeDBError, "查询资源备注失败："+err.Error())
		return
	}

	// 步骤2：返回成功响应
	response.Success(c, comments)
}
