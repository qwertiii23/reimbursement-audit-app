package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"text/template"
	"time"

	"reimbursement-audit/internal/pkg/logger"
)

// PromptBuilder Prompt构造器
type PromptBuilder struct {
	logger          logger.Logger
	systemTemplates map[string]string
	userTemplates   map[string]string
}

// NewPromptBuilder 创建Prompt构造器实例
func NewPromptBuilder(log logger.Logger) *PromptBuilder {
	builder := &PromptBuilder{
		logger:          log,
		systemTemplates: make(map[string]string),
		userTemplates:   make(map[string]string),
	}
	builder.initDefaultTemplates()
	return builder
}

// initDefaultTemplates 初始化默认模板
func (pb *PromptBuilder) initDefaultTemplates() {
	pb.systemTemplates["default"] = `你是一个专业的报销审核助手，能够根据报销制度文档对报销单据进行审核和分析。
请基于提供的报销制度文档内容，对用户的报销问题进行准确、详细的回答。
回答时请注意：
1. 严格依据报销制度文档中的规定
2. 引用具体的条款和标准
3. 如果文档中没有相关信息，请明确说明
4. 提供清晰、有条理的回答`

	pb.systemTemplates["audit"] = `你是一个专业的报销审核专家，负责审核员工的报销申请。
请根据提供的报销制度文档，对报销申请进行严格审核。
审核要点：
1. 检查报销金额是否符合标准
2. 检查报销类型是否在允许范围内
3. 检查审批流程是否完整
4. 检查附件是否齐全
5. 给出明确的审核结论（通过/驳回/需补充材料）`

	pb.systemTemplates["query"] = `你是一个报销制度查询助手，帮助用户快速了解报销政策和规定。
请基于提供的报销制度文档，准确回答用户关于报销政策的问题。
回答要求：
1. 准确引用相关条款
2. 提供具体的标准和限额
3. 说明适用的条件和场景
4. 如有例外情况，请一并说明`

	pb.userTemplates["rag_query"] = `基于以下报销制度文档内容，回答用户的问题：

【报销制度文档】
{{range .Documents}}
文档标题：{{.Title}}
文档内容：
{{.Content}}
{{end}}

【用户问题】
{{.Query}}

请基于上述文档内容，准确回答用户的问题。如果文档中没有相关信息，请明确说明。`

	pb.userTemplates["audit"] = `请审核以下报销申请：

【报销制度文档】
{{range .Documents}}
文档标题：{{.Title}}
文档内容：
{{.Content}}
{{end}}

【报销申请信息】
{{.ReimbursementInfo}}

请根据报销制度文档，对上述报销申请进行审核，并给出审核结论和理由。`

	pb.userTemplates["simple_query"] = `用户问题：{{.Query}}

请回答这个问题。`
}

// RegisterSystemTemplate 注册系统提示词模板
func (pb *PromptBuilder) RegisterSystemTemplate(name, template string) {
	pb.systemTemplates[name] = template
}

// RegisterUserTemplate 注册用户提示词模板
func (pb *PromptBuilder) RegisterUserTemplate(name, template string) {
	pb.userTemplates[name] = template
}

// BuildSystemPrompt 构造系统提示词
func (pb *PromptBuilder) BuildSystemPrompt(templateName string, variables map[string]interface{}) (string, error) {
	templateContent, ok := pb.systemTemplates[templateName]
	if !ok {
		templateContent = pb.systemTemplates["default"]
	}

	if len(variables) == 0 {
		return templateContent, nil
	}

	return pb.renderTemplate(templateContent, variables)
}

// BuildUserPrompt 构造用户提示词
func (pb *PromptBuilder) BuildUserPrompt(templateName string, variables map[string]interface{}) (string, error) {
	templateContent, ok := pb.userTemplates[templateName]
	if !ok {
		templateContent = pb.userTemplates["simple_query"]
	}

	return pb.renderTemplate(templateContent, variables)
}

// BuildRAGPrompt 构造RAG查询提示词
func (pb *PromptBuilder) BuildRAGPrompt(ctx context.Context, query string, documents []*Document, chunks []*DocumentChunk) (*Prompt, error) {
	systemPrompt, err := pb.BuildSystemPrompt("query", nil)
	if err != nil {
		pb.logger.Error("构造系统提示词失败", logger.NewField("error", err))
		return nil, errors.New("构造系统提示词失败")
	}

	variables := map[string]interface{}{
		"Query":     query,
		"Documents": documents,
		"Chunks":    chunks,
	}

	userPrompt, err := pb.BuildUserPrompt("rag_query", variables)
	if err != nil {
		pb.logger.Error("构造用户提示词失败", logger.NewField("error", err))
		return nil, errors.New("构造用户提示词失败")
	}

	prompt := &Prompt{
		ID:        generatePromptID(),
		Name:      "RAG查询提示词",
		Template:  "rag_query",
		Content:   userPrompt,
		Type:      "rag",
		Variables: variables,
		Tokens:    pb.estimateTokens(systemPrompt + userPrompt),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
		Tags:      []string{"rag", "query"},
	}

	return prompt, nil
}

// BuildAuditPrompt 构造审核提示词
func (pb *PromptBuilder) BuildAuditPrompt(ctx context.Context, reimbursementInfo string, documents []*Document) (*Prompt, error) {
	systemPrompt, err := pb.BuildSystemPrompt("audit", nil)
	if err != nil {
		pb.logger.Error("构造系统提示词失败", logger.NewField("error", err))
		return nil, errors.New("构造系统提示词失败")
	}

	variables := map[string]interface{}{
		"ReimbursementInfo": reimbursementInfo,
		"Documents":         documents,
	}

	userPrompt, err := pb.BuildUserTemplate("audit", variables)
	if err != nil {
		pb.logger.Error("构造用户提示词失败", logger.NewField("error", err))
		return nil, errors.New("构造用户提示词失败")
	}

	prompt := &Prompt{
		ID:        generatePromptID(),
		Name:      "报销审核提示词",
		Template:  "audit",
		Content:   userPrompt,
		Type:      "audit",
		Variables: variables,
		Tokens:    pb.estimateTokens(systemPrompt + userPrompt),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
		Tags:      []string{"audit", "reimbursement"},
	}

	return prompt, nil
}

// BuildUserTemplate 构造用户模板
func (pb *PromptBuilder) BuildUserTemplate(templateName string, variables map[string]interface{}) (string, error) {
	templateContent, ok := pb.userTemplates[templateName]
	if !ok {
		pb.logger.Error("模板不存在", logger.NewField("template_name", templateName))
		return "", errors.New("模板不存在")
	}

	return pb.renderTemplate(templateContent, variables)
}

// BuildConversationMessages 构造对话消息列表
func (pb *PromptBuilder) BuildConversationMessages(systemPrompt, userPrompt string) []*ConversationMessage {
	messages := []*ConversationMessage{
		{
			Role:      "system",
			Content:   systemPrompt,
			Timestamp: time.Now(),
		},
		{
			Role:      "user",
			Content:   userPrompt,
			Timestamp: time.Now(),
		},
	}
	return messages
}

// BuildConversationWithHistory 构造带历史记录的对话消息
func (pb *PromptBuilder) BuildConversationWithHistory(systemPrompt string, history []*ConversationMessage, newMessage string) []*ConversationMessage {
	messages := []*ConversationMessage{
		{
			Role:      "system",
			Content:   systemPrompt,
			Timestamp: time.Now(),
		},
	}

	messages = append(messages, history...)

	messages = append(messages, &ConversationMessage{
		Role:      "user",
		Content:   newMessage,
		Timestamp: time.Now(),
	})

	return messages
}

// renderTemplate 渲染模板
func (pb *PromptBuilder) renderTemplate(templateContent string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateContent)
	if err != nil {
		pb.logger.Error("解析模板失败", logger.NewField("error", err))
		return "", errors.New("解析模板失败")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		pb.logger.Error("渲染模板失败", logger.NewField("error", err))
		return "", errors.New("渲染模板失败")
	}

	return buf.String(), nil
}

// estimateTokens 估算Token数量
func (pb *PromptBuilder) estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}

// FormatDocuments 格式化文档列表
func (pb *PromptBuilder) FormatDocuments(documents []*Document) string {
	if len(documents) == 0 {
		return "无相关文档"
	}

	var builder strings.Builder
	for i, doc := range documents {
		builder.WriteString("【文档")
		builder.WriteString(strconv.Itoa(i + 1))
		builder.WriteString("】\n")
		builder.WriteString("标题：")
		builder.WriteString(doc.Title)
		builder.WriteString("\n")
		builder.WriteString("内容：")
		builder.WriteString(doc.Content)
		builder.WriteString("\n\n")
	}
	return builder.String()
}

// FormatChunks 格式化分片列表
func (pb *PromptBuilder) FormatChunks(chunks []*DocumentChunk) string {
	if len(chunks) == 0 {
		return "无相关内容"
	}

	var builder strings.Builder
	for i, chunk := range chunks {
		builder.WriteString("【内容片段")
		builder.WriteString(strconv.Itoa(i + 1))
		builder.WriteString("】\n")
		builder.WriteString("内容：")
		builder.WriteString(chunk.Content)
		builder.WriteString("\n\n")
	}
	return builder.String()
}

// FormatReimbursementInfo 格式化报销信息
func (pb *PromptBuilder) FormatReimbursementInfo(info map[string]interface{}) string {
	if len(info) == 0 {
		return "无报销信息"
	}

	jsonBytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		pb.logger.Error("序列化报销信息失败", logger.NewField("error", err))
		return "无法格式化报销信息"
	}
	return string(jsonBytes)
}

// BuildPromptFromTemplate 从模板构建Prompt
func (pb *PromptBuilder) BuildPromptFromTemplate(templateName, promptType string, variables map[string]interface{}) (*Prompt, error) {
	var content string
	var err error

	switch promptType {
	case "system":
		content, err = pb.BuildSystemPrompt(templateName, variables)
	case "user":
		content, err = pb.BuildUserPrompt(templateName, variables)
	default:
		pb.logger.Error("不支持的提示词类型", logger.NewField("prompt_type", promptType))
		return nil, errors.New("不支持的提示词类型")
	}

	if err != nil {
		return nil, err
	}

	prompt := &Prompt{
		ID:        generatePromptID(),
		Name:      templateName + "提示词",
		Template:  templateName,
		Content:   content,
		Type:      promptType,
		Variables: variables,
		Tokens:    pb.estimateTokens(content),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}

	return prompt, nil
}

// ValidatePrompt 验证Prompt
func (pb *PromptBuilder) ValidatePrompt(prompt *Prompt) error {
	if prompt == nil {
		pb.logger.Error("Prompt不能为空")
		return errors.New("Prompt不能为空")
	}
	if prompt.Content == "" {
		pb.logger.Error("Prompt内容不能为空")
		return errors.New("Prompt内容不能为空")
	}
	if prompt.Tokens > 4000 {
		pb.logger.Error("Prompt长度超过限制", logger.NewField("tokens", prompt.Tokens))
		return errors.New("Prompt长度超过限制")
	}
	return nil
}

// OptimizePrompt 优化Prompt（减少长度）
func (pb *PromptBuilder) OptimizePrompt(prompt *Prompt, maxTokens int) (*Prompt, error) {
	if prompt.Tokens <= maxTokens {
		return prompt, nil
	}

	ratio := float64(maxTokens) / float64(prompt.Tokens)
	newLength := int(float64(len(prompt.Content)) * ratio * 0.9)

	if newLength < 100 {
		pb.logger.Error("优化后的Prompt太短", logger.NewField("new_length", newLength))
		return nil, errors.New("优化后的Prompt太短")
	}

	optimizedContent := prompt.Content[:newLength] + "..."

	optimizedPrompt := &Prompt{
		ID:        prompt.ID,
		Name:      prompt.Name + "（优化后）",
		Template:  prompt.Template,
		Content:   optimizedContent,
		Type:      prompt.Type,
		Variables: prompt.Variables,
		Tokens:    pb.estimateTokens(optimizedContent),
		CreatedAt: prompt.CreatedAt,
		UpdatedAt: time.Now(),
		Version:   prompt.Version,
		Tags:      prompt.Tags,
	}

	return optimizedPrompt, nil
}

// generatePromptID 生成Prompt ID
func generatePromptID() string {
	return "prompt_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// GetSystemTemplate 获取系统模板
func (pb *PromptBuilder) GetSystemTemplate(name string) (string, bool) {
	template, ok := pb.systemTemplates[name]
	return template, ok
}

// GetUserTemplate 获取用户模板
func (pb *PromptBuilder) GetUserTemplate(name string) (string, bool) {
	template, ok := pb.userTemplates[name]
	return template, ok
}

// ListSystemTemplates 列出所有系统模板
func (pb *PromptBuilder) ListSystemTemplates() []string {
	templates := make([]string, 0, len(pb.systemTemplates))
	for name := range pb.systemTemplates {
		templates = append(templates, name)
	}
	return templates
}

// ListUserTemplates 列出所有用户模板
func (pb *PromptBuilder) ListUserTemplates() []string {
	templates := make([]string, 0, len(pb.userTemplates))
	for name := range pb.userTemplates {
		templates = append(templates, name)
	}
	return templates
}
