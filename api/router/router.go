package router

import (
	"customs/api/handler"
	"customs/api/middleware"
	"customs/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 初始化路由
func NewRouter(serviceContainer *service.ServiceContainer) *gin.Engine {
	// 初始化Gin引擎（开发环境用Debug模式，生产用Release）
	r := gin.Default()

	// 注册全局中间件
	r.Use(middleware.Cors())   // 跨域
	r.Use(middleware.Logger()) // 日志

	// 初始化Handler
	ddHandler := handler.NewDataDictionaryHandler(serviceContainer.DataDictionary)

	// 分组路由：API前缀
	apiGroup := r.Group("/api")
	{
		// 数据字典子分组
		dictGroup := apiGroup.Group("/dict")
		{
			dictGroup.POST("/upload", ddHandler.UploadExcel)          // 上传Excel
			dictGroup.GET("/result", ddHandler.GetParseResult)        // 查询解析结果
			dictGroup.POST("/confirm", ddHandler.ConfirmInsert)       // 确认入库
			dictGroup.GET("/comments", ddHandler.GetResourceComments) // 查询资源备注
		}

		// 健康检查接口（可选）
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	return r
}
