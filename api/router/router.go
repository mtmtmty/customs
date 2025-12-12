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
		dictGroup := apiGroup.Group("/dict")
		{
			dictGroup.POST("/upload", ddHandler.UploadExcel)          // 上传Excel
			dictGroup.GET("/result", ddHandler.GetParseResult)        // 查询解析结果
			dictGroup.POST("/confirm", ddHandler.ConfirmInsert)       // 确认入库
			dictGroup.GET("/comments", ddHandler.GetResourceComments) // 查询资源备注
		}

		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	return r
}
