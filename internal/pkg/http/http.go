package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Client HTTP客户端接口
type Client interface {
	// Get 发送GET请求
	Get(url string, headers map[string]string) (*Response, error)
	// Post 发送POST请求
	Post(url string, body interface{}, headers map[string]string) (*Response, error)
	// Put 发送PUT请求
	Put(url string, body interface{}, headers map[string]string) (*Response, error)
	// Delete 发送DELETE请求
	Delete(url string, headers map[string]string) (*Response, error)
	// Do 发送自定义请求
	Do(req *Request) (*Response, error)
	// SetTimeout 设置超时时间
	SetTimeout(timeout time.Duration)
	// GetTimeout 获取超时时间
	GetTimeout() time.Duration
}

// Request HTTP请求
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Params  map[string]string `json:"params"`
	Context context.Context   `json:"-"`
}

// Response HTTP响应
type Response struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	Request    *http.Request       `json:"-"`
}

// Config HTTP客户端配置
type Config struct {
	Timeout         time.Duration `json:"timeout"`            // 超时时间
	MaxIdleConns    int           `json:"max_idle_conns"`     // 最大空闲连接数
	MaxConnsPerHost int           `json:"max_conns_per_host"` // 每个主机的最大连接数
	UserAgent       string        `json:"user_agent"`         // 用户代理
	ProxyURL        string        `json:"proxy_url"`          // 代理URL
	FollowRedirects bool          `json:"follow_redirects"`   // 是否跟随重定向
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:         30 * time.Second,
		MaxIdleConns:    10,
		MaxConnsPerHost: 5,
		UserAgent:       "Go-HTTP-Client/1.0",
		FollowRedirects: true,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// TODO: 实现配置验证逻辑
	return nil
}

// NewRequest 创建HTTP请求
func NewRequest(method, url string, body interface{}) *Request {
	return &Request{
		Method:  method,
		URL:     url,
		Body:    body,
		Headers: make(map[string]string),
		Params:  make(map[string]string),
	}
}

// SetHeader 设置请求头
func (r *Request) SetHeader(key, value string) *Request {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
	return r
}

// SetHeaders 设置多个请求头
func (r *Request) SetHeaders(headers map[string]string) *Request {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	for k, v := range headers {
		r.Headers[k] = v
	}
	return r
}

// SetParam 设置请求参数
func (r *Request) SetParam(key, value string) *Request {
	if r.Params == nil {
		r.Params = make(map[string]string)
	}
	r.Params[key] = value
	return r
}

// SetParams 设置多个请求参数
func (r *Request) SetParams(params map[string]string) *Request {
	if r.Params == nil {
		r.Params = make(map[string]string)
	}
	for k, v := range params {
		r.Params[k] = v
	}
	return r
}

// SetContext 设置上下文
func (r *Request) SetContext(ctx context.Context) *Request {
	r.Context = ctx
	return r
}

// ToJSON 将响应转换为JSON对象
func (r *Response) ToJSON(obj interface{}) error {
	return json.Unmarshal(r.Body, obj)
}

// ToString 将响应转换为字符串
func (r *Response) ToString() string {
	return string(r.Body)
}

// IsSuccess 检查响应是否成功
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError 检查是否为客户端错误
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError 检查是否为服务器错误
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500
}

// GetHeader 获取响应头
func (r *Response) GetHeader(key string) string {
	if values, ok := r.Headers[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// GetHeaders 获取响应头的所有值
func (r *Response) GetHeaders(key string) []string {
	if values, ok := r.Headers[key]; ok {
		return values
	}
	return nil
}
