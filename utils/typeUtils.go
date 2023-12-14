package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// utils 包含 ConvertToNumber 方法
func ToInt(value interface{}, defaultVal int64) int64 {
	if value == nil || strings.Trim(fmt.Sprintf("%v", value), " ") == "" {
		return defaultVal
	}
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return num
		}
	}
	return defaultVal // 默认值
}
