package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reimbursement-audit/internal/pkg/logger"
	"time"

	"github.com/google/uuid"
)

// LLMClient 大模型客户端结构体
type LLMClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	timeout    time.Duration
	logger     logger.Logger
}

// NewLLMClient 创建大模型客户端实例
func NewLLMClient(apiKey, baseURL, model string, timeout int, log logger.Logger) *LLMClient {
	return &LLMClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		timeout: time.Duration(timeout) * time.Second,
		logger:  log,
	}
}

// ChatMessage 聊天消息结构体
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求结构体
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatResponse 聊天响应结构体
type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
}

// ChatChoice 聊天选择结构体
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage 使用情况结构体
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Chat 调用大模型聊天接口
func (c *LLMClient) Chat(ctx context.Context, messages []ChatMessage, temperature float64, maxTokens int) (*ChatResponse, error) {
	if len(messages) == 0 {
		c.logger.Error("消息列表不能为空")
		return nil, errors.New("消息列表不能为空")
	}

	request := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		c.logger.Error("序列化请求失败", logger.NewField("error", err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.Error("创建请求失败", logger.NewField("error", err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("发送请求失败", logger.NewField("url", c.baseURL), logger.NewField("error", err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("读取响应失败", logger.NewField("error", err))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("请求失败", logger.NewField("status_code", resp.StatusCode), logger.NewField("response", string(body)))
		return nil, errors.New("请求失败")
	}

	var chatResponse ChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		c.logger.Error("解析响应失败", logger.NewField("error", err))
		return nil, err
	}

	return &chatResponse, nil
}

// GenerateLLMResponse 生成大模型响应
func (c *LLMClient) GenerateLLMResponse(ctx context.Context, prompt string) (*LLMResponse, error) {
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "你是一个专业的报销审核助手，能够根据报销制度文档对报销单据进行审核和分析。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	startTime := time.Now()
	chatResponse, err := c.Chat(ctx, messages, 0.7, 2000)
	if err != nil {
		return nil, err
	}
	duration := time.Since(startTime)

	if len(chatResponse.Choices) == 0 {
		c.logger.Error("响应中没有选择项")
		return nil, errors.New("响应中没有选择项")
	}

	llmResponse := &LLMResponse{
		ID:        chatResponse.ID,
		Content:   chatResponse.Choices[0].Message.Content,
		Model:     chatResponse.Model,
		Tokens:    chatResponse.Usage.TotalTokens,
		Cost:      calculateCost(chatResponse.Usage.TotalTokens),
		Duration:  duration.Milliseconds(),
		CreatedAt: time.Now(),
	}

	return llmResponse, nil
}

// calculateCost 计算成本
func calculateCost(tokens int) float64 {
	costPer1KTokens := 0.001
	return float64(tokens) / 1000.0 * costPer1KTokens
}

// GenerateEmbedding 生成向量嵌入
func (c *LLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	embeddingRequest := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	requestBody, err := json.Marshal(embeddingRequest)
	if err != nil {
		c.logger.Error("序列化请求失败", logger.NewField("error", err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.Error("创建请求失败", logger.NewField("error", err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("发送请求失败", logger.NewField("url", c.baseURL), logger.NewField("error", err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("读取响应失败", logger.NewField("error", err))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("请求失败", logger.NewField("status_code", resp.StatusCode), logger.NewField("response", string(body)))
		return nil, errors.New("请求失败")
	}

	var embeddingResponse struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &embeddingResponse); err != nil {
		c.logger.Error("解析响应失败", logger.NewField("error", err))
		return nil, err
	}

	if len(embeddingResponse.Data) == 0 {
		c.logger.Error("响应中没有嵌入向量")
		return nil, errors.New("响应中没有嵌入向量")
	}

	return embeddingResponse.Data[0].Embedding, nil
}

// BatchGenerateEmbeddings 批量生成向量嵌入
func (c *LLMClient) BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		embedding, err := c.GenerateEmbedding(ctx, text)
		if err != nil {
			c.logger.Error("生成文本嵌入失败", logger.NewField("index", i), logger.NewField("text", text), logger.NewField("error", err))
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

// ValidateResponse 验证响应
func (c *LLMClient) ValidateResponse(response *ChatResponse) error {
	if response == nil {
		c.logger.Error("响应不能为空")
		return errors.New("响应不能为空")
	}
	if len(response.Choices) == 0 {
		c.logger.Error("响应中没有选择项")
		return errors.New("响应中没有选择项")
	}
	if response.Choices[0].Message.Content == "" {
		c.logger.Error("响应内容为空")
		return errors.New("响应内容为空")
	}
	return nil
}

// ParseAnalysisResult 解析分析结果
func (c *LLMClient) ParseAnalysisResult(content string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		ID:         uuid.New().String(),
		Conclusion: content,
		Reasoning:  content,
		CreatedAt:  time.Now(),
	}

	return result, nil
}

// GetModelInfo 获取模型信息
func (c *LLMClient) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"model":      c.model,
		"base_url":   c.baseURL,
		"timeout":    c.timeout.Seconds(),
		"created_at": time.Now(),
	}
}

// HealthCheck 健康检查
func (c *LLMClient) HealthCheck(ctx context.Context) error {
	messages := []ChatMessage{
		{
			Role:    "user",
			Content: "ping",
		},
	}

	_, err := c.Chat(ctx, messages, 0.0, 10)
	if err != nil {
		c.logger.Error("健康检查失败", logger.NewField("error", err))
		return err
	}

	return nil
}

// Close 关闭客户端
func (c *LLMClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
