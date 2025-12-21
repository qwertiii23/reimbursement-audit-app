// code.go 错误码定义
// 功能点：
// 1. 定义系统错误码常量
// 2. 定义业务错误码常量
// 3. 提供错误码对应的消息映射
// 4. 支持错误码分类（系统错误、业务错误、第三方错误等）
// 5. 提供错误码查询方法
// 6. 支持国际化错误消息

package response

// 错误码常量定义
const (
	// 成功
	CodeSuccess = 0

	// 系统错误 1000-1999
	CodeInternalError    = 1000 // 内部服务器错误
	CodeInvalidParams    = 1001 // 参数错误
	CodeUnauthorized     = 1002 // 未授权
	CodeForbidden        = 1003 // 禁止访问
	CodeNotFound         = 1004 // 资源不存在
	CodeMethodNotAllowed = 1005 // 方法不允许
	CodeTooManyRequests  = 1006 // 请求过多

	// 业务错误 2000-2999
	CodeUploadFailed         = 2000 // 上传失败
	CodeFileFormatInvalid    = 2001 // 文件格式无效
	CodeFileSizeExceeded     = 2002 // 文件大小超限
	CodeOCRError             = 2003 // OCR解析错误
	CodeAuditFailed          = 2004 // 审核失败
	CodeRuleNotFound         = 2005 // 规则不存在
	CodeRuleValidationFailed = 2006 // 规则校验失败
	CodeReimbursementNotFound = 2007 // 报销单不存在
	CodeInvoiceInvalid       = 2008 // 发票无效

	// 第三方错误 3000-3999
	CodeThirdPartyServiceError = 3000 // 第三方服务错误
	CodeLLMError               = 3001 // 大模型调用错误
	CodeVectorSearchError      = 3002 // 向量搜索错误
)

// 错误码消息映射
var codeMessages = map[int]string{
	CodeSuccess:               "成功",
	CodeInternalError:         "内部服务器错误",
	CodeInvalidParams:         "参数错误",
	CodeUnauthorized:          "未授权",
	CodeForbidden:             "禁止访问",
	CodeNotFound:              "资源不存在",
	CodeMethodNotAllowed:      "方法不允许",
	CodeTooManyRequests:       "请求过多",
	CodeUploadFailed:          "上传失败",
	CodeFileFormatInvalid:     "文件格式无效",
	CodeFileSizeExceeded:      "文件大小超限",
	CodeOCRError:              "OCR解析错误",
	CodeAuditFailed:           "审核失败",
	CodeRuleNotFound:          "规则不存在",
	CodeRuleValidationFailed:  "规则校验失败",
	CodeReimbursementNotFound: "报销单不存在",
	CodeInvoiceInvalid:        "发票无效",
	CodeThirdPartyServiceError: "第三方服务错误",
	CodeLLMError:              "大模型调用错误",
	CodeVectorSearchError:     "向量搜索错误",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code int) string {
	// TODO: 实现获取错误消息逻辑
	return ""
}

// SetMessage 设置错误码对应的消息
func SetMessage(code int, message string) {
	// TODO: 实现设置错误消息逻辑
}