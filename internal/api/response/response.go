// response.go 统一响应格式
// 功能点：
// 1. 定义统一的API响应结构体
// 2. 实现响应成功和失败的方法
// 3. 支持分页响应格式
// 4. 支持响应数据序列化
// 5. 提供响应写入方法
// 6. 支持响应头设置

package response

import (
	"encoding/json"
	"net/http"
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`    // 响应码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// PaginationResponse 分页响应结构体
type PaginationResponse struct {
	Code    int         `json:"code"`    // 响应码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
	Total   int64       `json:"total"`   // 总记录数
	Page    int         `json:"page"`    // 当前页码
	Size    int         `json:"size"`    // 每页大小
}

// Success 成功响应
func Success(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, &Response{
		Code:    CodeSuccess,
		Message: GetMessage(CodeSuccess),
		Data:    data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusOK, &Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(w http.ResponseWriter, code int, message string) {
	// 如果消息为空，使用默认消息
	if message == "" {
		message = GetMessage(code)
	}

	WriteJSON(w, http.StatusOK, &Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// Pagination 分页响应
func Pagination(w http.ResponseWriter, data interface{}, total int64, page, size int) {
	WriteJSON(w, http.StatusOK, &PaginationResponse{
		Code:    CodeSuccess,
		Message: GetMessage(CodeSuccess),
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// WriteJSON 写入JSON响应
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	// 设置响应头
	SetHeaders(w)

	// 设置状态码
	w.WriteHeader(statusCode)

	// 编码并写入JSON数据
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, "编码响应数据失败", http.StatusInternalServerError)
	}
}

// SetHeaders 设置响应头
func SetHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 可以添加其他通用响应头，如CORS等
	// w.Header().Set("Access-Control-Allow-Origin", "*")
}
