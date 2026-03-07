package errors

import "fmt"

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewAppError 创建应用错误
func NewAppError(code int, message string, detail string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// 错误码定义
const (
	// 通用错误码 (1-999)
	Success           = 0
	UnknownError      = 1
	InvalidParams     = 2
	NotFound          = 3
	AlreadyExists     = 4
	PermissionDenied  = 5
	InternalError     = 6
	
	// 订阅相关错误码 (1000-1999)
	SubscriptionNotFound      = 1000
	SubscriptionParseFailed   = 1001
	SubscriptionFetchFailed   = 1002
	SubscriptionDecodeFailed  = 1003
	SubscriptionInvalid       = 1004
	
	// 节点相关错误码 (2000-2999)
	NodeNotFound         = 2000
	NodeTestFailed       = 2001
	NodeConnectFailed    = 2002
	NodeDisconnectFailed = 2003
	NodeAlreadyConnected = 2004
	
	// 规则相关错误码 (3000-3999)
	RuleNotFound   = 3000
	RuleInvalid    = 3001
	RuleConflict   = 3002
	
	// 系统相关错误码 (4000-4999)
	SystemBusy      = 4000
	ConfigLoadError = 4001
	DataSaveError   = 4002
	
	// 集群相关错误码 (5000-5999)
	ServerNotFound       = 5000
	ServerAlreadyExists  = 5001
	ConnectionFailed     = 5002
	ConnectionNotFound   = 5003
	AuthFailed           = 5004
	SessionCreateFailed  = 5005
	FileTransferFailed   = 5006
	DeployFailed         = 5007
	BackupFailed         = 5008
	RestoreFailed        = 5009
	ScaleFailed          = 5010
	GroupNotFound        = 5011
)

// 错误消息映射
var errorMsgs = map[int]string{
	Success:           "成功",
	UnknownError:      "未知错误",
	InvalidParams:     "参数错误",
	NotFound:          "资源不存在",
	AlreadyExists:     "资源已存在",
	PermissionDenied:  "权限不足",
	InternalError:     "内部错误",
	
	SubscriptionNotFound:      "订阅不存在",
	SubscriptionParseFailed:   "订阅解析失败",
	SubscriptionFetchFailed:   "订阅获取失败",
	SubscriptionDecodeFailed:  "订阅解码失败",
	SubscriptionInvalid:       "订阅无效",
	
	NodeNotFound:         "节点不存在",
	NodeTestFailed:       "节点测试失败",
	NodeConnectFailed:    "节点连接失败",
	NodeDisconnectFailed: "节点断开失败",
	NodeAlreadyConnected: "节点已连接",
	
	RuleNotFound:   "规则不存在",
	RuleInvalid:    "规则无效",
	RuleConflict:   "规则冲突",
	
	SystemBusy:      "系统繁忙",
	ConfigLoadError: "配置加载失败",
	DataSaveError:   "数据保存失败",
}

// GetErrorMsg 获取错误消息
func GetErrorMsg(code int) string {
	if msg, ok := errorMsgs[code]; ok {
		return msg
	}
	return "未知错误"
}

// NewError 创建错误
func NewError(code int, detail string) *AppError {
	return &AppError{
		Code:    code,
		Message: GetErrorMsg(code),
		Detail:  detail,
	}
}
