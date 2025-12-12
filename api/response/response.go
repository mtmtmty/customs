package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ErrCodeSuccess       = 0    // 成功（默认）
	ErrCodeInvalidParam  = 1001 // 参数无效
	ErrCodeFileError     = 1002 // 文件相关错误（上传/解析/下载）
	ErrCodeDBError       = 1003 // 数据库错误（查询/插入/更新）
	ErrCodeTaskError     = 1004 // 异步任务错误（生产/查询/执行）
	ErrCodeBusinessError = 1005 // 业务逻辑错误（前置条件不满足等）
	ErrCodeSystemError   = 5000 // 系统内部错误（兜底）
)

// Response 统一响应结构体
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
