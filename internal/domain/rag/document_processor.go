// document_processor.go 报销制度文档处理（解析+分片）
// 功能点：
// 1. 文档内容解析（PDF、Word、TXT等）
// 2. 文档内容清洗和预处理
// 3. 文档分片和分段
// 4. 文档元数据提取
// 5. 文档版本管理
// 6. 文档索引构建

package rag

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

// DocumentProcessor 文档处理器结构体
type DocumentProcessor struct {
	chunkSize    int
	chunkOverlap int
	logger       logger.Logger
}

// NewDocumentProcessor 创建文档处理器实例
func NewDocumentProcessor(chunkSize, chunkOverlap int, log logger.Logger) *DocumentProcessor {
	if chunkSize <= 0 {
		chunkSize = 500
	}
	if chunkOverlap < 0 {
		chunkOverlap = 50
	}
	return &DocumentProcessor{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		logger:       log,
	}
}

// ProcessDocument 处理单个文档
func (dp *DocumentProcessor) ProcessDocument(ctx context.Context, documentPath string) (*Document, error) {
	content, err := dp.ParseDocument(ctx, documentPath)
	if err != nil {
		dp.logger.Error("解析文档失败", logger.NewField("document_path", documentPath), logger.NewField("error", err))
		return nil, err
	}

	cleanedContent := dp.CleanContent(content)

	metadata, err := dp.ExtractMetadata(ctx, documentPath)
	if err != nil {
		dp.logger.Error("提取元数据失败", logger.NewField("document_path", documentPath), logger.NewField("error", err))
		return nil, err
	}

	document := &Document{
		ID:        uuid.New().String(),
		Title:     filepath.Base(documentPath),
		Content:   cleanedContent,
		Type:      dp.GetDocumentType(documentPath),
		Source:    documentPath,
		Path:      documentPath,
		Size:      dp.getFileSize(documentPath),
		Metadata:  metadata,
		Status:    "processed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}

	chunks, err := dp.CreateDocumentChunks(ctx, document)
	if err != nil {
		dp.logger.Error("创建文档分片失败", logger.NewField("document_id", document.ID), logger.NewField("error", err))
		return nil, err
	}
	document.Chunks = chunks

	return document, nil
}

// ProcessDocuments 批量处理文档
func (dp *DocumentProcessor) ProcessDocuments(ctx context.Context, documentPaths []string) ([]*Document, error) {
	documents := make([]*Document, 0, len(documentPaths))
	for _, path := range documentPaths {
		document, err := dp.ProcessDocument(ctx, path)
		if err != nil {
			dp.logger.Error("处理文档失败", logger.NewField("document_path", path), logger.NewField("error", err))
			return nil, err
		}
		documents = append(documents, document)
	}
	return documents, nil
}

// ParseDocument 解析文档内容
func (dp *DocumentProcessor) ParseDocument(ctx context.Context, documentPath string) (string, error) {
	docType := dp.GetDocumentType(documentPath)
	switch strings.ToLower(docType) {
	case "txt":
		return dp.ExtractTextFromText(ctx, documentPath)
	case "pdf":
		return dp.ExtractTextFromPDF(ctx, documentPath)
	case "doc", "docx":
		return dp.ExtractTextFromWord(ctx, documentPath)
	default:
		return dp.ExtractTextFromText(ctx, documentPath)
	}
}

// CleanContent 清洗文档内容
func (dp *DocumentProcessor) CleanContent(content string) string {
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\t", " ")

	lines := strings.Split(content, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// SplitContent 分割文档内容
func (dp *DocumentProcessor) SplitContent(content string, chunkSize int, overlap int) []string {
	if chunkSize <= 0 {
		chunkSize = dp.chunkSize
	}
	if overlap < 0 {
		overlap = dp.chunkOverlap
	}

	var chunks []string
	words := strings.Fields(content)

	for i := 0; i < len(words); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)

		if end >= len(words) {
			break
		}
	}

	return chunks
}

// ExtractMetadata 提取文档元数据
func (dp *DocumentProcessor) ExtractMetadata(ctx context.Context, documentPath string) (*DocumentMetadata, error) {
	fileInfo, err := os.Stat(documentPath)
	if err != nil {
		dp.logger.Error("获取文件信息失败", logger.NewField("document_path", documentPath), logger.NewField("error", err))
		return nil, err
	}

	metadata := &DocumentMetadata{
		CreatedAt:   fileInfo.ModTime(),
		UpdatedAt:   fileInfo.ModTime(),
		Category:    "reimbursement",
		Department:  "finance",
		EffectiveAt: time.Now(),
		ExpiresAt:   time.Now().AddDate(10, 0, 0),
		Priority:    1,
		Language:    "zh-CN",
	}

	return metadata, nil
}

// CreateDocumentChunks 创建文档分片
func (dp *DocumentProcessor) CreateDocumentChunks(ctx context.Context, document *Document) ([]*DocumentChunk, error) {
	chunks := dp.SplitContent(document.Content, dp.chunkSize, dp.chunkOverlap)

	documentChunks := make([]*DocumentChunk, 0, len(chunks))
	position := 0

	for _, chunkContent := range chunks {
		chunk := &DocumentChunk{
			ID:         uuid.New().String(),
			DocumentID: document.ID,
			Content:    chunkContent,
			StartPos:   position,
			EndPos:     position + len(chunkContent),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		documentChunks = append(documentChunks, chunk)
		position += len(chunkContent)
	}

	return documentChunks, nil
}

// ValidateDocument 验证文档
func (dp *DocumentProcessor) ValidateDocument(documentPath string) error {
	if documentPath == "" {
		return errors.New("文档路径不能为空")
	}

	fileInfo, err := os.Stat(documentPath)
	if err != nil {
		dp.logger.Error("文档不存在或无法访问", logger.NewField("document_path", documentPath), logger.NewField("error", err))
		return err
	}

	if fileInfo.IsDir() {
		dp.logger.Error("路径指向的是目录而非文件", logger.NewField("document_path", documentPath))
		return errors.New("路径指向的是目录而非文件")
	}

	if fileInfo.Size() == 0 {
		dp.logger.Error("文档为空", logger.NewField("document_path", documentPath))
		return errors.New("文档为空")
	}

	return nil
}

// GetDocumentType 获取文档类型
func (dp *DocumentProcessor) GetDocumentType(documentPath string) string {
	ext := strings.ToLower(filepath.Ext(documentPath))
	switch ext {
	case ".txt":
		return "txt"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "word"
	default:
		return "unknown"
	}
}

// ConvertToText 转换为纯文本
func (dp *DocumentProcessor) ConvertToText(ctx context.Context, documentPath string) (string, error) {
	return dp.ParseDocument(ctx, documentPath)
}

// ExtractTextFromPDF 从PDF提取文本
func (dp *DocumentProcessor) ExtractTextFromPDF(ctx context.Context, pdfPath string) (string, error) {
	file, err := os.Open(pdfPath)
	if err != nil {
		dp.logger.Error("打开PDF文件失败", logger.NewField("pdf_path", pdfPath), logger.NewField("error", err))
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		dp.logger.Error("读取PDF文件失败", logger.NewField("pdf_path", pdfPath), logger.NewField("error", err))
		return "", err
	}

	return string(content), nil
}

// ExtractTextFromWord 从Word文档提取文本
func (dp *DocumentProcessor) ExtractTextFromWord(ctx context.Context, wordPath string) (string, error) {
	file, err := os.Open(wordPath)
	if err != nil {
		dp.logger.Error("打开Word文件失败", logger.NewField("word_path", wordPath), logger.NewField("error", err))
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		dp.logger.Error("读取Word文件失败", logger.NewField("word_path", wordPath), logger.NewField("error", err))
		return "", err
	}

	return string(content), nil
}

// ExtractTextFromText 从文本文件提取文本
func (dp *DocumentProcessor) ExtractTextFromText(ctx context.Context, textPath string) (string, error) {
	file, err := os.Open(textPath)
	if err != nil {
		dp.logger.Error("打开文本文件失败", logger.NewField("text_path", textPath), logger.NewField("error", err))
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		dp.logger.Error("读取文本文件失败", logger.NewField("text_path", textPath), logger.NewField("error", err))
		return "", err
	}

	return string(content), nil
}

// OptimizeChunks 优化文档分片
func (dp *DocumentProcessor) OptimizeChunks(chunks []*DocumentChunk) []*DocumentChunk {
	if len(chunks) == 0 {
		return chunks
	}

	optimized := make([]*DocumentChunk, 0, len(chunks))
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk.Content) != "" {
			optimized = append(optimized, chunk)
		}
	}

	return optimized
}

// MergeChunks 合并文档分片
func (dp *DocumentProcessor) MergeChunks(chunks []*DocumentChunk) string {
	if len(chunks) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, chunk := range chunks {
		builder.WriteString(chunk.Content)
		builder.WriteString("\n")
	}

	return builder.String()
}

// getFileSize 获取文件大小
func (dp *DocumentProcessor) getFileSize(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

// EncodeVector 编码向量为Base64字符串
func (dp *DocumentProcessor) EncodeVector(vector []float64) string {
	data := make([]byte, len(vector)*8)
	for i, v := range vector {
		bits := uint64(v)
		for j := 0; j < 8; j++ {
			data[i*8+j] = byte(bits >> (j * 8))
		}
	}
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeVector 解码Base64字符串为向量
func (dp *DocumentProcessor) DecodeVector(encoded string) ([]float64, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		dp.logger.Error("解码向量失败", logger.NewField("encoded", encoded), logger.NewField("error", err))
		return nil, err
	}

	vector := make([]float64, len(data)/8)
	for i := 0; i < len(vector); i++ {
		var bits uint64
		for j := 0; j < 8; j++ {
			bits |= uint64(data[i*8+j]) << (j * 8)
		}
		vector[i] = float64(bits)
	}

	return vector, nil
}
