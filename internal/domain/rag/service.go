package rag

// 业务请求（审核报销/查政策）→ 提取核心信息 → 向量+关键词检索知识库 → 拼接制度片段到Prompt → 调用大模型 → 解析结果返回

import (
	"context"
	"encoding/json"
	"errors"
	"reimbursement-audit/internal/pkg/logger"
	"strconv"
	"strings"
	"time"
)

// RAGService RAG服务结构体
type RAGService struct {
	logger            logger.Logger
	llmClient         *LLMClient
	documentProcessor *DocumentProcessor
	vectorStore       *VectorStore
	promptBuilder     *PromptBuilder
}

// NewRAGService 创建RAG服务实例
func NewRAGService(log logger.Logger, llmClient *LLMClient, documentProcessor *DocumentProcessor, vectorStore *VectorStore, promptBuilder *PromptBuilder) *RAGService {
	return &RAGService{
		logger:            log,
		llmClient:         llmClient,
		documentProcessor: documentProcessor,
		vectorStore:       vectorStore,
		promptBuilder:     promptBuilder,
	}
}

// Query 查询报销政策（RAG查询）
func (rs *RAGService) Query(ctx context.Context, query string, topK int) (*RAGResult, error) {
	startTime := time.Now()

	if query == "" {
		rs.logger.Error("查询内容不能为空")
		return nil, errors.New("查询内容不能为空")
	}

	if topK <= 0 {
		topK = 5
	}

	embedding, err := rs.llmClient.GenerateEmbedding(ctx, query)
	if err != nil {
		rs.logger.Error("生成查询向量失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("生成查询向量失败")
	}

	searchResults, err := rs.vectorStore.SearchVector(ctx, embedding, topK)
	if err != nil {
		rs.logger.Error("搜索相关文档失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("搜索相关文档失败")
	}

	if len(searchResults) == 0 {
		rs.logger.Error("未找到相关文档", logger.NewField("query", query))
		return nil, errors.New("未找到相关文档")
	}

	documents := rs.buildDocumentsFromSearchResults(searchResults)
	chunks := rs.buildChunksFromSearchResults(searchResults)

	prompt, err := rs.promptBuilder.BuildRAGPrompt(ctx, query, documents, chunks)
	if err != nil {
		rs.logger.Error("构造提示词失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("构造提示词失败")
	}

	systemPrompt, err := rs.promptBuilder.BuildSystemPrompt("query", nil)
	if err != nil {
		rs.logger.Error("构造系统提示词失败", logger.NewField("error", err))
		return nil, errors.New("构造系统提示词失败")
	}

	messages := rs.promptBuilder.BuildConversationMessages(systemPrompt, prompt.Content)

	llmResponse, err := rs.llmClient.Chat(ctx, rs.convertToChatMessages(messages), 0.7, 2000)
	if err != nil {
		rs.logger.Error("调用大模型失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("调用大模型失败")
	}

	if err := rs.validateLLMResponse(llmResponse); err != nil {
		rs.logger.Error("大模型响应格式校验失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("大模型响应格式校验失败")
	}

	analysisResult := rs.parseAnalysisResult(query, llmResponse, searchResults)

	ragResult := &RAGResult{
		Query:          query,
		Documents:      documents,
		Chunks:         chunks,
		Prompt:         prompt.Content,
		Response:       rs.convertToLLMResponse(llmResponse),
		AnalysisResult: analysisResult,
		ExecutionTime:  time.Since(startTime).Milliseconds(),
		CreatedAt:      time.Now(),
	}

	return ragResult, nil
}

// AuditReimbursement 审核报销申请
func (rs *RAGService) AuditReimbursement(ctx context.Context, reimbursementInfo map[string]interface{}, topK int) (*RAGResult, error) {
	startTime := time.Now()

	// 步骤1：参数校验（报销信息不能为空，topK默认5）
	if len(reimbursementInfo) == 0 {
		rs.logger.Error("报销信息不能为空")
		return nil, errors.New("报销信息不能为空")
	}

	if topK <= 0 {
		topK = 5
	}
	// 步骤2：构建查询文本 → 把报销单信息（类目、金额、类型等）转为自然语言查询（如“差旅费 金额700.00元 住宿费”）
	query := rs.buildQueryFromReimbursementInfo(reimbursementInfo)

	// 步骤3：生成查询向量 → 调用大模型的embedding接口，把query转为向量（用于后续检索）
	embedding, err := rs.llmClient.GenerateEmbedding(ctx, query)
	if err != nil {
		rs.logger.Error("生成查询向量失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("生成查询向量失败")
	}

	// 步骤4：混合检索 → 向量检索+关键词检索，提升检索准确度
	keywords := rs.extractReimbursementKeywords(reimbursementInfo)
	searchResults, err := rs.vectorStore.HybridSearch(ctx, embedding, keywords, topK)
	if err != nil {
		rs.logger.Error("混合检索失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("混合检索失败")
	}

	// 步骤5：构建Prompt → 把报销单信息+检索到的制度片段拼到Prompt里（保证AI只看自有知识库）
	documents := rs.buildDocumentsFromSearchResults(searchResults)

	reimbursementInfoJSON := rs.promptBuilder.FormatReimbursementInfo(reimbursementInfo)
	prompt, err := rs.promptBuilder.BuildAuditPrompt(ctx, reimbursementInfoJSON, documents)
	if err != nil {
		rs.logger.Error("构造提示词失败", logger.NewField("error", err))
		return nil, errors.New("构造提示词失败")
	}

	// 步骤6：调用大模型 → 传入SystemPrompt（审核规则）+ 业务Prompt，获取AI审核结论
	systemPrompt, err := rs.promptBuilder.BuildSystemPrompt("audit", nil)
	if err != nil {
		rs.logger.Error("构造系统提示词失败", logger.NewField("error", err))
		return nil, errors.New("构造系统提示词失败")
	}

	messages := rs.promptBuilder.BuildConversationMessages(systemPrompt, prompt.Content)

	llmResponse, err := rs.llmClient.Chat(ctx, rs.convertToChatMessages(messages), 0.7, 2000)
	if err != nil {
		rs.logger.Error("调用大模型失败", logger.NewField("error", err))
		return nil, errors.New("调用大模型失败")
	}

	if err := rs.validateLLMResponse(llmResponse); err != nil {
		rs.logger.Error("大模型响应格式校验失败", logger.NewField("error", err))
		return nil, errors.New("大模型响应格式校验失败")
	}

	analysisResult := rs.parseAuditResult(query, llmResponse, searchResults)

	// 步骤8：封装返回结果 → 包含查询、制度文档、Prompt、AI响应、审核结论、执行时间等
	ragResult := &RAGResult{
		Query:          query,
		Documents:      documents,
		Prompt:         prompt.Content,
		Response:       rs.convertToLLMResponse(llmResponse),
		AnalysisResult: analysisResult,
		ExecutionTime:  time.Since(startTime).Milliseconds(),
		CreatedAt:      time.Now(),
	}

	return ragResult, nil
}

// IngestDocument 导入文档到RAG系统  解析→分片→向量化→存储
func (rs *RAGService) IngestDocument(ctx context.Context, documentPath string) (*Document, error) {
	document, err := rs.documentProcessor.ProcessDocument(ctx, documentPath)
	if err != nil {
		rs.logger.Error("处理文档失败", logger.NewField("document_path", documentPath), logger.NewField("error", err))
		return nil, errors.New("处理文档失败")
	}

	for _, chunk := range document.Chunks {
		embedding, err := rs.llmClient.GenerateEmbedding(ctx, chunk.Content)
		if err != nil {
			rs.logger.Error("生成向量失败", logger.NewField("document_id", document.ID), logger.NewField("error", err))
			return nil, errors.New("生成向量失败")
		}

		chunk.Vector = embedding

		err = rs.vectorStore.StoreVector(ctx, &Vector{
			ID:         generateVectorID(),
			DocumentID: document.ID,
			ChunkID:    chunk.ID,
			Values:     embedding,
			Dimension:  len(embedding),
			Metadata: map[string]interface{}{
				"document_title": document.Title,
				"chunk_index":    chunk.StartPos,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			rs.logger.Error("存储向量失败", logger.NewField("document_id", document.ID), logger.NewField("error", err))
			return nil, errors.New("存储向量失败")
		}
	}

	return document, nil
}

// BatchIngestDocuments 批量导入文档
func (rs *RAGService) BatchIngestDocuments(ctx context.Context, documentPaths []string) ([]*Document, error) {
	if len(documentPaths) == 0 {
		rs.logger.Error("文档路径列表不能为空")
		return nil, errors.New("文档路径列表不能为空")
	}

	documents := make([]*Document, 0, len(documentPaths))
	errorList := make([]error, 0)

	for _, path := range documentPaths {
		document, err := rs.IngestDocument(ctx, path)
		if err != nil {
			rs.logger.Error("导入文档失败", logger.NewField("path", path), logger.NewField("error", err))
			errorList = append(errorList, err)
			continue
		}
		documents = append(documents, document)
	}

	if len(errorList) > 0 {
		rs.logger.Error("部分文档导入失败", logger.NewField("error_count", len(errorList)))
		return documents, errors.New("部分文档导入失败")
	}

	return documents, nil
}

// DeleteDocument 删除文档
func (rs *RAGService) DeleteDocument(ctx context.Context, documentID string) error {
	if documentID == "" {
		rs.logger.Error("文档ID不能为空")
		return errors.New("文档ID不能为空")
	}

	err := rs.vectorStore.DeleteVectorByDocument(ctx, documentID)
	if err != nil {
		rs.logger.Error("删除文档向量失败", logger.NewField("document_id", documentID), logger.NewField("error", err))
		return errors.New("删除文档向量失败")
	}

	return nil
}

// SearchDocuments 搜索文档
func (rs *RAGService) SearchDocuments(ctx context.Context, query string, topK int) ([]*VectorSearchResult, error) {
	if query == "" {
		rs.logger.Error("查询内容不能为空")
		return nil, errors.New("查询内容不能为空")
	}

	if topK <= 0 {
		topK = 5
	}

	embedding, err := rs.llmClient.GenerateEmbedding(ctx, query)
	if err != nil {
		rs.logger.Error("生成查询向量失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("生成查询向量失败")
	}

	results, err := rs.vectorStore.SearchVector(ctx, embedding, topK)
	if err != nil {
		rs.logger.Error("搜索文档失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("搜索文档失败")
	}

	return results, nil
}

// HybridSearch 混合搜索（向量+关键词）
func (rs *RAGService) HybridSearch(ctx context.Context, query string, topK int, keywordWeight float64) ([]*VectorSearchResult, error) {
	if query == "" {
		rs.logger.Error("查询内容不能为空")
		return nil, errors.New("查询内容不能为空")
	}

	if topK <= 0 {
		topK = 5
	}

	if keywordWeight < 0 || keywordWeight > 1 {
		keywordWeight = 0.5
	}

	embedding, err := rs.llmClient.GenerateEmbedding(ctx, query)
	if err != nil {
		rs.logger.Error("生成查询向量失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("生成查询向量失败")
	}

	keywords := rs.extractKeywords(query)

	results, err := rs.vectorStore.HybridSearch(ctx, embedding, keywords, topK)
	if err != nil {
		rs.logger.Error("混合搜索失败", logger.NewField("query", query), logger.NewField("error", err))
		return nil, errors.New("混合搜索失败")
	}

	return results, nil
}

// GetStatistics 获取RAG系统统计信息
func (rs *RAGService) GetStatistics(ctx context.Context) (*VectorStoreStatistics, error) {
	stats, err := rs.vectorStore.GetStatistics(ctx)
	if err != nil {
		rs.logger.Error("获取统计信息失败", logger.NewField("error", err))
		return nil, errors.New("获取统计信息失败")
	}
	return stats, nil
}

// buildDocumentsFromSearchResults 从搜索结果构建文档列表
func (rs *RAGService) buildDocumentsFromSearchResults(results []*VectorSearchResult) []*Document {
	docMap := make(map[string]*Document)

	for _, result := range results {
		if _, exists := docMap[result.DocumentID]; !exists {
			docMap[result.DocumentID] = &Document{
				ID:      result.DocumentID,
				Title:   result.DocumentID,
				Content: result.Content,
				Type:    "txt",
				Status:  "processed",
			}
		}
	}

	documents := make([]*Document, 0, len(docMap))
	for _, doc := range docMap {
		documents = append(documents, doc)
	}

	return documents
}

// buildChunksFromSearchResults 从搜索结果构建分片列表
func (rs *RAGService) buildChunksFromSearchResults(results []*VectorSearchResult) []*DocumentChunk {
	chunks := make([]*DocumentChunk, 0, len(results))

	for _, result := range results {
		chunk := &DocumentChunk{
			ID:         result.ChunkID,
			DocumentID: result.DocumentID,
			Content:    result.Content,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		chunks = append(chunks, chunk)
	}

	return chunks
}

// buildQueryFromReimbursementInfo 从报销信息构建查询
func (rs *RAGService) buildQueryFromReimbursementInfo(info map[string]interface{}) string {
	var query string

	if reimbursementType, ok := info["type"].(string); ok {
		query += reimbursementType + " "
	}

	if amount, ok := info["amount"].(float64); ok {
		query += "金额" + strconv.FormatFloat(amount, 'f', 2, 64) + " "
	}

	if category, ok := info["category"].(string); ok {
		query += category + " "
	}

	if query == "" {
		query = "报销申请"
	}

	return query
}

// convertToChatMessages 转换为聊天消息格式
func (rs *RAGService) convertToChatMessages(messages []*ConversationMessage) []ChatMessage {
	chatMessages := make([]ChatMessage, len(messages))

	for i, msg := range messages {
		chatMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return chatMessages
}

// convertToLLMResponse 转换为LLM响应格式
func (rs *RAGService) convertToLLMResponse(response *ChatResponse) *LLMResponse {
	if response == nil {
		return nil
	}

	llmResponse := &LLMResponse{
		ID:        response.ID,
		Content:   "",
		Model:     response.Model,
		Tokens:    response.Usage.TotalTokens,
		Cost:      calculateCost(response.Usage.TotalTokens),
		CreatedAt: time.Now(),
	}

	if len(response.Choices) > 0 {
		llmResponse.Content = response.Choices[0].Message.Content
	}

	return llmResponse
}

// parseAnalysisResult 解析分析结果
func (rs *RAGService) parseAnalysisResult(query string, response *ChatResponse, references []*VectorSearchResult) *AnalysisResult {
	if response == nil || len(response.Choices) == 0 {
		return &AnalysisResult{
			ID:         generateAnalysisResultID(),
			Query:      query,
			Conclusion: "无法生成分析结果",
			Reasoning:  "大模型响应为空",
			Confidence: 0,
			CreatedAt:  time.Now(),
		}
	}

	content := response.Choices[0].Message.Content

	return &AnalysisResult{
		ID:         generateAnalysisResultID(),
		Query:      query,
		Conclusion: content,
		Reasoning:  "基于检索到的相关文档内容进行分析",
		Confidence: 0.8,
		Data: map[string]interface{}{
			"references_count": len(references),
		},
		CreatedAt: time.Now(),
	}
}

// parseAuditResult 解析审核结果
func (rs *RAGService) parseAuditResult(query string, response *ChatResponse, references []*VectorSearchResult) *AnalysisResult {
	if response == nil || len(response.Choices) == 0 {
		return &AnalysisResult{
			ID:         generateAnalysisResultID(),
			Query:      query,
			Conclusion: "审核失败",
			Reasoning:  "大模型响应为空",
			Confidence: 0,
			CreatedAt:  time.Now(),
		}
	}

	content := response.Choices[0].Message.Content

	confidence := rs.calculateAuditConfidence(content, references)

	return &AnalysisResult{
		ID:         generateAnalysisResultID(),
		Query:      query,
		Conclusion: content,
		Reasoning:  "基于报销制度文档进行审核",
		Confidence: confidence,
		Data: map[string]interface{}{
			"references_count": len(references),
			"avg_score":        rs.calculateAverageScore(references),
		},
		CreatedAt: time.Now(),
	}
}

// calculateAuditConfidence 计算审核置信度
func (rs *RAGService) calculateAuditConfidence(content string, references []*VectorSearchResult) float64 {
	if len(content) == 0 {
		return 0
	}

	baseConfidence := 0.5

	if len(references) > 0 {
		avgScore := rs.calculateAverageScore(references)
		if avgScore > 0.8 {
			baseConfidence += 0.2
		} else if avgScore > 0.6 {
			baseConfidence += 0.1
		}
	}

	if len(references) >= 3 {
		baseConfidence += 0.1
	} else if len(references) >= 1 {
		baseConfidence += 0.05
	}

	if len(content) > 100 {
		baseConfidence += 0.1
	}

	if strings.Contains(content, "通过") || strings.Contains(content, "不通过") || strings.Contains(content, "驳回") {
		baseConfidence += 0.05
	}

	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

// calculateAverageScore 计算检索结果的平均分数
func (rs *RAGService) calculateAverageScore(references []*VectorSearchResult) float64 {
	if len(references) == 0 {
		return 0
	}

	totalScore := 0.0
	for _, ref := range references {
		totalScore += ref.Score
	}

	return totalScore / float64(len(references))
}

// generateVectorID 生成向量ID
func generateVectorID() string {
	return "vec_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// generateAnalysisResultID 生成分析结果ID
func generateAnalysisResultID() string {
	return "analysis_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// HealthCheck 健康检查
func (rs *RAGService) HealthCheck(ctx context.Context) error {
	if rs.llmClient == nil {
		rs.logger.Error("大模型客户端未初始化")
		return errors.New("大模型客户端未初始化")
	}

	if rs.documentProcessor == nil {
		rs.logger.Error("文档处理器未初始化")
		return errors.New("文档处理器未初始化")
	}

	if rs.vectorStore == nil {
		rs.logger.Error("向量存储未初始化")
		return errors.New("向量存储未初始化")
	}

	if rs.promptBuilder == nil {
		rs.logger.Error("提示词构造器未初始化")
		return errors.New("提示词构造器未初始化")
	}

	return nil
}

// GetServiceInfo 获取服务信息
func (rs *RAGService) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"service_name": "RAGService",
		"version":      "1.0",
		"components": map[string]string{
			"llm_client":         "initialized",
			"document_processor": "initialized",
			"vector_store":       "initialized",
			"prompt_builder":     "initialized",
		},
	}
}

// ExportAnalysisResult 导出分析结果
func (rs *RAGService) ExportAnalysisResult(result *AnalysisResult) (string, error) {
	if result == nil {
		rs.logger.Error("分析结果不能为空")
		return "", errors.New("分析结果不能为空")
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		rs.logger.Error("序列化分析结果失败", logger.NewField("error", err))
		return "", errors.New("序列化分析结果失败")
	}

	return string(jsonBytes), nil
}

// validateLLMResponse 校验大模型响应格式
func (rs *RAGService) validateLLMResponse(response *ChatResponse) error {
	if response == nil {
		return errors.New("大模型响应为空")
	}

	if len(response.Choices) == 0 {
		return errors.New("大模型响应不包含任何选择")
	}

	content := response.Choices[0].Message.Content

	if len(content) == 0 {
		return errors.New("大模型响应内容为空")
	}

	if len(content) < 10 {
		return errors.New("大模型响应内容过短")
	}

	if len(content) > 10000 {
		return errors.New("大模型响应内容过长")
	}

	if response.Usage.TotalTokens == 0 {
		return errors.New("大模型响应token数为0")
	}

	if response.Model == "" {
		return errors.New("大模型响应缺少模型信息")
	}

	return nil
}

// extractReimbursementKeywords 从报销信息中提取关键词
func (rs *RAGService) extractReimbursementKeywords(info map[string]interface{}) []string {
	keywords := make([]string, 0)

	if reimbursementType, ok := info["type"].(string); ok && reimbursementType != "" {
		keywords = append(keywords, reimbursementType)
	}

	if category, ok := info["category"].(string); ok && category != "" {
		keywords = append(keywords, category)
	}

	if amount, ok := info["amount"].(float64); ok {
		if amount > 0 {
			if amount < 500 {
				keywords = append(keywords, "小额")
			} else if amount < 2000 {
				keywords = append(keywords, "中等金额")
			} else {
				keywords = append(keywords, "大额")
			}
		}
	}

	if expenseType, ok := info["expense_type"].(string); ok && expenseType != "" {
		keywords = append(keywords, expenseType)
	}

	if city, ok := info["city"].(string); ok && city != "" {
		keywords = append(keywords, city)
	}

	if len(keywords) > 5 {
		keywords = keywords[:5]
	}

	return keywords
}

// extractKeywords 从查询中提取关键词
func (rs *RAGService) extractKeywords(query string) []string {
	if query == "" {
		return []string{}
	}

	words := strings.Fields(query)
	keywords := make([]string, 0)

	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true,
		"有": true, "和": true, "就": true, "不": true, "人": true,
		"都": true, "一": true, "一个": true, "上": true, "也": true,
		"很": true, "到": true, "说": true, "要": true, "去": true,
		"你": true, "会": true, "着": true, "没有": true, "看": true,
		"好": true, "自己": true, "这": true,
	}

	for _, word := range words {
		if len(word) > 1 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	if len(keywords) > 5 {
		keywords = keywords[:5]
	}

	return keywords
}
