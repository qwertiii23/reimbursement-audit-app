package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

type VisionHandler struct {
	logger logger.Logger
}

func NewVisionHandler(log logger.Logger) *VisionHandler {
	return &VisionHandler{
		logger: log,
	}
}

type VisionAnalyzeRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
	Prompt   string `json:"prompt" binding:"required"`
}

type VisionAnalyzeResponse struct {
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

func (h *VisionHandler) Analyze(c *gin.Context) {
	ctx := context.Background()

	var req VisionAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogError(c, "请求参数错误", "error", err.Error())
		response.ErrorResponse(c, response.CodeInvalidParams, "请求参数错误")
		return
	}

	result, err := h.analyzeImage(ctx, req.ImageURL, req.Prompt)
	if err != nil {
		middleware.LogError(c, "图像分析失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "图像分析失败")
		return
	}

	response.SuccessResponse(c, result)
}

func (h *VisionHandler) analyzeImage(ctx context.Context, imageURL, prompt string) (*VisionAnalyzeResponse, error) {
	imageData, err := h.downloadImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("下载图片失败: %w", err)
	}

	result, err := h.callVisionAPI(ctx, imageData, prompt)
	if err != nil {
		return nil, fmt.Errorf("调用视觉API失败: %w", err)
	}

	return result, nil
}

func (h *VisionHandler) downloadImage(ctx context.Context, imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (h *VisionHandler) callVisionAPI(ctx context.Context, imageData []byte, prompt string) (*VisionAnalyzeResponse, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	apiURL := "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	apiKey := "d1f2189e1e6c45fbb522481a0f32437c.tvBH9rduuRgWgxEu"

	payload := map[string]interface{}{
		"model": "glm-4v",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
						},
					},
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
		"temperature": 0.3,
		"max_tokens":  500,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Body = io.NopCloser(bytes.NewReader(jsonData))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API调用失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("API返回空结果")
	}

	return h.parseVisionResult(apiResponse.Choices[0].Message.Content)
}

func (h *VisionHandler) parseVisionResult(content string) (*VisionAnalyzeResponse, error) {
	result := &VisionAnalyzeResponse{
		Confidence: 0.5,
		Reasoning:  content,
	}

	if contains(content, "存在P图") || contains(content, "存在篡改") || contains(content, "存在伪造") {
		result.Confidence = 0.8
	} else if contains(content, "不存在P图") || contains(content, "无明显P图痕迹") {
		result.Confidence = 0.2
	}

	return result, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
