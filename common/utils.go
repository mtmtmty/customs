package common

import (
	"encoding/json"
	"strings"
)

// IsExcelFile 判断文件是否为Excel格式（.xlsx/.xls）
func IsExcelFile(filename string) bool {
	ext := strings.ToLower(filename)
	return strings.HasSuffix(ext, ".xlsx") || strings.HasSuffix(ext, ".xls")
}

// Paginate 通用分页处理（输入总条数、页码、页大小，返回分页信息）
func Paginate(total, page, size int) (int, int, map[string]int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	offset := (page - 1) * size

	// 返回偏移量、页大小、分页元信息
	return offset, size, map[string]int{
		"page":  page,
		"size":  size,
		"total": total,
	}
}

// JSONStringToMap 将JSON字符串转换为map（用于解析缓存的JSON结果）
func JSONStringToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return result, nil
}
