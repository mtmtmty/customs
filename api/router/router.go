package router

import (
	"customs/api/handler"
	"customs/api/middleware"
	"customs/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 初始化路由
func NewRouter(serviceContainer *service.ServiceContainer) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.Cors()) // 跨域

	ddHandler := handler.NewDataDictionaryHandler(serviceContainer.DataDictionary)

	apiGroup := r.Group("/api")
	{
		dictGroup := apiGroup.Group("/data_dictionary")
		{
			dictGroup.GET("/insert", ddHandler.DownloadTemplate)              // 下载模版文件
			dictGroup.POST("/insert", ddHandler.UploadExcel)                  // 上传Excel
			dictGroup.GET("/insert/data", ddHandler.GetParseResult)           // 查询解析结果
			dictGroup.POST("/insert/{id}", ddHandler.ConfirmInsert)           // 确认入库
			dictGroup.GET("/resource_comment", ddHandler.GetResourceComments) // 查询资源备注
		}

		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	return r
}
