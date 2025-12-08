package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应结构体
type Response struct {
	Code int         `json:"code"` // 业务错误码（0=成功）
	Msg  string      `json:"msg"`  // 提示信息
	Data interface{} `json:"data"` // 业务数据（可选）
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
