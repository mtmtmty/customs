package errno

import "fmt"

// Errno 自定义错误类型
type Errno struct {
	Code int
	Msg  string
}

// Error 实现error接口
func (e *Errno) Error() string {
	return e.Msg
}

// 定义通用错误码
var (
	ErrInternalServer = &Errno{Code: 500, Msg: "服务器内部错误"}
	ErrInvalidParam   = &Errno{Code: 400, Msg: "无效的参数"}

	ErrFileOpenFailed    = &Errno{Code: 1001, Msg: "文件打开失败"}
	ErrInvalidFileFormat = &Errno{Code: 1002, Msg: "无效的文件格式"}
	ErrFileUploadFailed  = &Errno{Code: 1003, Msg: "文件上传失败"}

	ErrInvalidFileNameFormat = &Errno{Code: 1004, Msg: "文件名格式错误（需为'系统名-dbname'）"}
	ErrExcelNoSheet          = &Errno{Code: 1005, Msg: "Excel文件无工作表"}
	ErrExcelEmpty            = &Errno{Code: 1006, Msg: "Excel文件内容为空"}
	ErrExcelReadFailed       = &Errno{Code: 1007, Msg: "Excel文件读取失败"}
	ErrExcelColumnMissing    = &Errno{Code: 1008, Msg: "Excel缺少必要列（需包含数据表名称、字段名称等标准列）"}
	ErrExcelOpenFailed       = &Errno{Code: 1009, Msg: "Excel打开失败"}

	ErrDBInsertFailed = &Errno{Code: 2001, Msg: "数据库插入失败"}
	ErrDBUpdateFailed = &Errno{Code: 2002, Msg: "数据库更新失败"}
	ErrDBQueryFailed  = &Errno{Code: 2003, Msg: "数据库查询失败"}

	ErrRedisGetFailed = &Errno{Code: 3001, Msg: "Redis获取失败"}
	ErrRedisSetFailed = &Errno{Code: 3002, Msg: "Redis设置失败"}

	ErrTaskCreateFailed    = &Errno{Code: 4001, Msg: "任务创建失败"}
	ErrTaskQueryFailed     = &Errno{Code: 4002, Msg: "任务状态查询失败"}
	ErrPreTaskNotCompleted = &Errno{Code: 4003, Msg: "前置任务未完成"}

	ErrMinioUploadFailed   = &Errno{Code: 5001, Msg: "MinIO上传失败"}
	ErrMinioDownloadFailed = &Errno{Code: 5002, Msg: "MinIO下载失败"}
)

// WithMessage 为错误添加自定义消息（不改变错误码）
func (e *Errno) WithMessage(msg string) *Errno {
	return &Errno{
		Code: e.Code,
		Msg:  fmt.Sprintf("%s: %s", e.Msg, msg),
	}
}
