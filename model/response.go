package model

// Response 结构体定义了统一的返回格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SuccessResponse 生成成功时的返回格式
func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    10000,
		Message: "Success",
		Data:    data,
	}
}

// ErrorResponse 生成失败时的返回格式
func ErrorResponse(code int, message string, data interface{}) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
