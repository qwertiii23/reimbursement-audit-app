package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// clientImpl HTTP客户端实现
type clientImpl struct {
	httpClient *http.Client
	config     *Config
}

// NewClient 创建HTTP客户端实例
func NewClient(config *Config) (Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxConnsPerHost,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// 设置是否跟随重定向
	if !config.FollowRedirects {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &clientImpl{
		httpClient: httpClient,
		config:     config,
	}, nil
}

// Get 发送GET请求
func (c *clientImpl) Get(url string, headers map[string]string) (*Response, error) {
	req := NewRequest("GET", url, nil)
	req.SetHeaders(headers)
	return c.Do(req)
}

// Post 发送POST请求
func (c *clientImpl) Post(url string, body interface{}, headers map[string]string) (*Response, error) {
	req := NewRequest("POST", url, body)
	req.SetHeaders(headers)
	return c.Do(req)
}

// Put 发送PUT请求
func (c *clientImpl) Put(url string, body interface{}, headers map[string]string) (*Response, error) {
	req := NewRequest("PUT", url, body)
	req.SetHeaders(headers)
	return c.Do(req)
}

// Delete 发送DELETE请求
func (c *clientImpl) Delete(url string, headers map[string]string) (*Response, error) {
	req := NewRequest("DELETE", url, nil)
	req.SetHeaders(headers)
	return c.Do(req)
}

// Do 发送自定义请求
func (c *clientImpl) Do(req *Request) (*Response, error) {
	// 构建请求URL
	requestURL := req.URL
	if len(req.Params) > 0 {
		requestURL = BuildURL(requestURL, req.Params)
	}

	// 准备请求体
	var bodyReader io.Reader
	if req.Body != nil {
		switch v := req.Body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		case io.Reader:
			bodyReader = v
		default:
			// 尝试将body转换为JSON
			jsonData, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewReader(jsonData)
		}
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(req.Context, req.Method, requestURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// 设置默认Content-Type（如果有请求体且未设置Content-Type）
	if bodyReader != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 设置User-Agent
	if httpReq.Header.Get("User-Agent") == "" && c.config.UserAgent != "" {
		httpReq.Header.Set("User-Agent", c.config.UserAgent)
	}

	// 发送请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// 构建响应对象
	response := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       respBody,
		Request:    httpReq,
	}

	return response, nil
}

// SetTimeout 设置超时时间
func (c *clientImpl) SetTimeout(timeout time.Duration) {
	c.config.Timeout = timeout
	c.httpClient.Timeout = timeout
}

// GetTimeout 获取超时时间
func (c *clientImpl) GetTimeout() time.Duration {
	return c.config.Timeout
}

// BuildURL 构建带参数的URL
func BuildURL(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	query := u.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()

	return u.String()
}

// ParseURL 解析URL
func ParseURL(urlStr string) (scheme, host, path string, params map[string]string, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", "", nil, err
	}

	params = make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return u.Scheme, u.Host, u.Path, params, nil
}

// IsURL 检查是否为有效的URL
func IsURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// JoinURL 连接URL路径
func JoinURL(baseURL, path string) string {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return baseURL + path
}

// EncodeURL 编码URL
func EncodeURL(urlStr string) string {
	return url.QueryEscape(urlStr)
}

// DecodeURL 解码URL
func DecodeURL(urlStr string) (string, error) {
	return url.QueryUnescape(urlStr)
}