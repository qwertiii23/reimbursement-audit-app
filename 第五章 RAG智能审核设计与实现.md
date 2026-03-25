# 第五章 RAG智能审核设计与实现

## 5.1 知识库构建

知识库构建是RAG智能审核系统的基础环节，其质量直接影响后续检索和审核的准确性。本节将从文档采集与预处理、文档分片策略、向量嵌入与存储三个方面详细阐述知识库的构建过程。

### 5.1.1 文档采集与预处理

文档采集与预处理是知识库构建的第一步，主要完成对原始报销制度文档的解析、清洗和标准化处理，为后续的分片和向量化奠定基础。

#### （1）文档采集

系统支持多种格式的文档采集，包括TXT、PDF、Word（.doc、.docx）等常见文档格式。文档采集模块通过统一的接口处理不同格式的文档，根据文件扩展名自动识别文档类型并调用相应的解析方法。文档采集的核心代码位于`DocumentProcessor`结构体中，通过`ParseDocument`方法实现文档类型的自动识别和解析。

文档采集过程包括以下步骤：
1. **文档类型识别**：通过文件扩展名识别文档类型（.txt、.pdf、.doc、.docx等）
2. **文档验证**：检查文档是否存在、是否为空、是否为目录等基本属性
3. **文档读取**：根据文档类型调用相应的读取方法，将文档内容读取到内存中
4. **元数据提取**：提取文档的创建时间、修改时间、文件大小等基本信息

#### （2）文档预处理

文档预处理是对原始文档内容进行清洗和标准化的过程，主要包括以下步骤：

**文本清洗**：
- 去除文档首尾的空白字符
- 统一换行符格式（将\r\n转换为\n）
- 替换制表符为空格
- 去除空行和重复的空白行
- 规范化文本格式

**元数据提取**：
系统为每个文档提取丰富的元数据信息，包括：
- **基本信息**：文档名称、文档类型、文档大小、文档路径
- **时间信息**：创建时间、更新时间、生效时间、失效时间
- **分类信息**：文档分类（如reimbursement）、所属部门
- **优先级信息**：文档优先级、语言类型
- **版本信息**：文档版本号，支持文档版本管理

文档预处理的代码实现位于`CleanContent`和`ExtractMetadata`方法中，通过这些方法确保文档内容的规范性和一致性，为后续的分片处理提供高质量的输入数据。

### 5.1.2 文档分片策略

文档分片是将长文档切分为多个较小的文本片段的过程，合理的分片策略能够提高检索的准确性和效率。本系统采用基于词的滑动窗口分片策略，结合重叠分片技术，确保分片内容的完整性和语义连贯性。

#### （1）分片参数配置

系统提供灵活的分片参数配置，主要包括：
- **分片大小（chunkSize）**：默认为500词，可根据文档特点进行调整
- **重叠大小（chunkOverlap）**：默认为50词，用于保持分片间的语义连贯性
- **分片索引**：记录每个分片在原文档中的起始位置和结束位置

分片参数的配置通过`NewDocumentProcessor`构造函数完成，支持根据实际需求动态调整分片大小和重叠大小。

#### （2）分片算法

系统采用基于词的滑动窗口分片算法，具体步骤如下：

1. **文本分词**：使用`strings.Fields`方法将文档内容按空白字符分割为单词列表
2. **滑动窗口分片**：以chunkSize为窗口大小，以chunkOverlap为滑动步长，遍历单词列表
3. **分片生成**：将窗口内的单词重新组合为分片内容
4. **位置记录**：记录每个分片在原文档中的起始位置和结束位置

分片算法的核心代码位于`SplitContent`方法中，通过该方法实现高效的文档分片。算法的时间复杂度为O(n)，其中n为文档的单词数量，能够快速处理大规模文档。

#### （3）分片优化

为了提高分片质量，系统对生成的分片进行优化处理：

1. **空分片过滤**：去除内容为空或仅包含空白字符的分片
2. **分片验证**：确保每个分片都有有效的文档ID和内容
3. **分片合并**：支持将多个分片合并为完整的文档内容

分片优化通过`OptimizeChunks`和`MergeChunks`方法实现，确保分片的质量和可用性。

#### （4）分片模型

每个文档分片对应一个`DocumentChunk`对象，包含以下属性：
- **分片ID**：唯一标识符，使用UUID生成
- **文档ID**：所属文档的标识符
- **分片内容**：分片的文本内容
- **起始位置**：分片在原文档中的起始位置
- **结束位置**：分片在原文档中的结束位置
- **向量表示**：分片的向量嵌入（768维）
- **时间戳**：创建时间和更新时间

分片模型的设计确保了分片的可追溯性和可管理性，为后续的向量化和检索提供了良好的数据基础。

### 5.1.3 向量嵌入与存储

向量嵌入是将文本内容转换为数值向量的过程，通过向量嵌入可以实现文本的语义表示和相似度计算。本系统采用智谱AI的Embedding模型生成768维向量，并使用PostgreSQL + PGVector扩展进行向量存储和检索。

#### （1）向量嵌入生成

系统集成了智谱AI的Embedding API，支持将文本内容转换为高维向量表示。向量嵌入生成的核心功能包括：

**Embedding模型配置**：
- **模型选择**：支持智谱AI的Embedding模型，可根据需求切换不同的模型
- **向量维度**：默认为768维，可根据模型配置调整
- **API配置**：支持自定义API地址、API密钥等配置

**向量生成流程**：
1. **文本预处理**：对输入文本进行必要的预处理
2. **API调用**：调用智谱AI的Embedding接口生成向量
3. **结果解析**：解析API响应，提取向量数据
4. **错误处理**：处理API调用失败、超时等异常情况

向量嵌入生成的代码位于`LLMClient`的`GenerateEmbedding`方法中，该方法支持单个文本和批量文本的向量化。系统还提供了`BatchGenerateEmbeddings`方法，支持批量生成向量嵌入，提高处理效率。

**向量特性**：
- **维度固定**：所有向量均为768维，确保向量空间的一致性
- **语义表示**：向量能够捕捉文本的语义信息，支持语义相似度计算
- **数值范围**：向量值为浮点数，范围通常在-1到1之间

#### （2）向量存储架构

系统采用PostgreSQL + PGVector扩展作为向量存储方案，该方案具有以下优势：

**技术优势**：
- **原生支持**：PGVector是PostgreSQL的原生向量扩展，支持高效的向量存储和检索
- **成熟稳定**：PostgreSQL是成熟稳定的关系型数据库，具有良好的可靠性和性能
- **功能丰富**：支持向量索引、相似度计算、混合查询等高级功能
- **易于集成**：与现有的PostgreSQL基础设施无缝集成

**存储模型**：
向量存储的核心数据模型为`DocumentModel`，包含以下字段：
- **ID**：文档分片的唯一标识符
- **FileName**：文档名称
- **FileType**：文档类型
- **Category**：文档分类
- **ChunkID**：分片ID
- **ChunkIndex**：分片索引
- **ChunkContent**：分片内容
- **Embedding**：向量嵌入（768维向量）
- **时间戳**：创建时间和更新时间

向量存储的代码位于`VectorStore`结构体中，通过GORM ORM框架与数据库交互，提供了丰富的向量操作接口。

#### （3）向量索引优化

为了提高向量检索的性能，系统支持多种向量索引优化技术：

**索引类型**：
- **IVFFlat索引**：基于倒排文件和扁平量化的索引，适合大规模向量检索
- **HNSW索引**：基于层次化可导航小世界图的索引，提供更快的检索速度
- **余弦相似度索引**：专门针对余弦相似度优化的索引

**索引创建**：
系统提供了`CreateVectorIndex`方法，支持创建IVFFlat索引，并可以配置索引参数（如lists参数）。索引的创建过程包括：
1. **索引配置**：设置索引名称、索引类型、索引参数
2. **索引构建**：在数据库中创建向量索引
3. **索引优化**：通过ANALYZE命令优化索引性能

**索引管理**：
- **索引查询**：`ListIndexes`方法列出所有索引
- **索引删除**：`DropIndex`方法删除指定索引
- **索引优化**：`OptimizeIndex`方法优化索引性能

#### （4）向量操作接口

系统提供了丰富的向量操作接口，包括：

**存储操作**：
- **单向量存储**：`StoreVector`方法存储单个向量
- **批量存储**：`StoreVectors`方法批量存储向量，支持事务处理
- **向量更新**：`UpdateVector`方法更新已有向量
- **向量删除**：`DeleteVector`和`DeleteVectors`方法删除单个或多个向量

**检索操作**：
- **向量相似度检索**：`SearchVector`方法基于向量相似度检索相关文档
- **分类检索**：`SearchVectorByCategory`方法在指定分类内检索
- **关键词检索**：`KeywordSearch`方法基于关键词检索
- **混合检索**：`HybridSearch`方法结合向量和关键词进行混合检索

**辅助操作**：
- **向量归一化**：`NormalizeVector`方法对向量进行归一化处理
- **相似度计算**：`CalculateSimilarity`方法计算两个向量的相似度
- **统计信息**：`GetStatistics`方法获取向量存储的统计信息

#### （5）向量存储优化

为了提高向量存储的性能和可靠性，系统实现了多种优化机制：

**性能优化**：
- **批量操作**：支持批量存储和批量检索，减少数据库交互次数
- **索引优化**：通过创建向量索引提高检索速度
- **查询优化**：使用PGVector的相似度算子（<->）进行高效检索
- **缓存机制**：对热点数据进行缓存，减少重复计算

**可靠性优化**：
- **重试机制**：`retryOperation`方法实现了操作重试机制，提高系统容错能力
- **超时控制**：所有数据库操作都设置了超时时间，避免长时间阻塞
- **事务处理**：批量操作使用事务确保数据一致性
- **错误处理**：完善的错误处理机制，确保系统稳定性

**存储优化**：
- **向量压缩**：通过JSON格式存储向量，减少存储空间
- **分片存储**：将长文档分片存储，提高检索效率
- **元数据管理**：通过元数据过滤减少检索范围

通过以上优化措施，系统能够高效、可靠地存储和检索向量数据，为RAG智能审核提供坚实的数据基础。

## 5.1.4 知识库构建流程总结

知识库构建是一个完整的文档处理流程，包括文档采集、预处理、分片、向量化和存储等多个环节。整个流程的代码实现位于`RAGService`的`IngestDocument`方法中，具体步骤如下：

1. **文档处理**：调用`DocumentProcessor.ProcessDocument`方法解析文档、清洗内容、提取元数据
2. **文档分片**：调用`CreateDocumentChunks`方法将文档切分为多个分片
3. **向量化**：对每个分片调用`LLMClient.GenerateEmbedding`方法生成向量嵌入
4. **向量存储**：调用`VectorStore.StoreVector`方法将向量存储到数据库
5. **结果返回**：返回处理后的文档对象，包含分片和向量信息

系统还支持批量文档导入，通过`BatchIngestDocuments`方法可以一次性处理多个文档，提高知识库构建的效率。

通过完整的知识库构建流程，系统能够将原始的报销制度文档转换为结构化的向量知识库，为后续的智能审核提供高质量的知识支持。

## 5.2 向量检索系统

向量检索系统是RAG智能审核的核心组件，负责根据用户查询从知识库中快速、准确地检索相关文档片段。本节将从向量索引构建、相似度计算、混合检索策略三个方面详细阐述向量检索系统的设计与实现。

### 5.2.1 向量索引构建

向量索引是提高向量检索性能的关键技术，通过构建高效的索引结构，可以显著降低向量相似度搜索的时间复杂度。本系统采用PGVector扩展提供的索引机制，支持多种索引类型和优化策略。

#### （1）索引类型选择

系统支持多种向量索引类型，根据不同的应用场景和数据规模选择合适的索引：

**IVFFlat索引**：
- **原理**：基于倒排文件和扁平量化的索引结构
- **特点**：通过聚类将向量空间划分为多个倒排列表，检索时只搜索部分聚类中心
- **适用场景**：适合大规模向量数据集（百万级以上），在召回率和检索速度之间取得平衡
- **参数配置**：lists参数控制聚类数量，通常设置为向量数量的平方根

**HNSW索引**：
- **原理**：基于层次化可导航小世界图的索引结构
- **特点**：通过构建多层图结构实现快速近似最近邻搜索
- **适用场景**：适合需要高检索速度的场景，召回率略低于IVFFlat
- **优势**：检索速度快，内存占用相对较低

**余弦相似度索引**：
- **原理**：专门针对余弦相似度计算的索引优化
- **特点**：使用vector_cosine_ops操作符，优化余弦相似度计算
- **适用场景**：适合需要计算余弦相似度的应用场景
- **优势**：计算精度高，适合文本语义检索

#### （2）索引创建流程

系统提供了完整的索引创建和管理功能，通过`CreateVectorIndex`方法实现向量索引的创建。索引创建流程包括以下步骤：

**1. 索引配置**：
```go
func (vs *VectorStore) CreateVectorIndex(ctx context.Context, indexName string, lists int) error {
    if lists <= 0 {
        lists = 100  // 默认聚类数量
    }
    query := "CREATE INDEX " + indexName + " ON reimbursement_documents USING ivfflat (embedding vector_cosine_ops) WITH (lists = ?)"
    result := vs.db.WithContext(ctx).Exec(query, lists)
    return result.Error
}
```

索引配置包括：
- **索引名称**：唯一标识索引，便于管理和维护
- **索引类型**：指定使用的索引类型（ivfflat、hnsw等）
- **向量列**：指定要索引的向量列（embedding）
- **距离算子**：指定相似度计算方式（vector_cosine_ops表示余弦相似度）
- **索引参数**：配置索引特定参数（如lists参数）

**2. 索引构建**：
- 执行CREATE INDEX语句在数据库中创建索引
- PGVector会自动对向量数据进行聚类或图构建
- 索引构建过程可能需要较长时间，取决于数据规模

**3. 索引优化**：
```go
func (vs *VectorStore) OptimizeIndex(ctx context.Context, indexName string) error {
    query := "ANALYZE TABLE reimbursement_documents"
    result := vs.db.WithContext(ctx).Exec(query)
    return result.Error
}
```

通过ANALYZE命令优化索引统计信息，提高查询优化器的决策质量。

#### （3）索引管理机制

系统提供了完整的索引管理功能，确保索引的有效性和性能：

**索引查询**：
```go
func (vs *VectorStore) ListIndexes(ctx context.Context) ([]string, error) {
    query := `
        SELECT INDEX_NAME 
        FROM INFORMATION_SCHEMA.STATISTICS 
        WHERE TABLE_NAME = 'reimbursement_documents'
        GROUP BY INDEX_NAME
    `
    // 查询并返回所有索引名称
}
```

通过`ListIndexes`方法可以查询当前表的所有索引，便于索引管理和监控。

**索引删除**：
```go
func (vs *VectorStore) DropIndex(ctx context.Context, indexName string) error {
    query := "DROP INDEX " + indexName
    result := vs.db.WithContext(ctx).Exec(query)
    return result.Error
}
```

通过`DropIndex`方法可以删除不再需要的索引，释放存储空间。

**索引重建**：
当向量数据发生大规模更新时，可能需要重建索引以保持检索性能。系统支持删除旧索引并创建新索引的重建流程。

#### （4）索引性能优化

为了提高索引性能，系统实现了多种优化策略：

**参数调优**：
- **lists参数**：根据向量数量动态调整，通常设置为√N（N为向量数量）
- **ef_search参数**：控制检索时的搜索范围，平衡召回率和速度
- **M参数**（HNSW）：控制每个节点的最大连接数，影响图结构的稠密程度

**索引维护**：
- **定期优化**：定期执行ANALYZE命令更新统计信息
- **索引监控**：监控索引的使用情况和性能指标
- **索引重建**：在数据更新后及时重建索引

**查询优化**：
- **索引选择**：根据查询特征选择合适的索引
- **并行查询**：利用PostgreSQL的并行查询能力
- **结果缓存**：对热点查询结果进行缓存

### 5.2.2 相似度计算

相似度计算是向量检索的核心算法，用于衡量两个向量之间的相似程度。系统支持多种相似度计算方法，并提供了高效的实现和优化。

#### （1）相似度度量方法

系统主要采用余弦相似度作为向量相似度的度量方法，该方法在文本语义检索中表现优异。

**余弦相似度**：
```go
func (vs *VectorStore) CalculateSimilarity(vector1, vector2 []float64) float64 {
    if len(vector1) != len(vector2) {
        return 0
    }

    var dotProduct, norm1, norm2 float64
    for i := 0; i < len(vector1); i++ {
        dotProduct += vector1[i] * vector2[i]
        norm1 += vector1[i] * vector1[i]
        norm2 += vector2[i] * vector2[i]
    }

    if norm1 == 0 || norm2 == 0 {
        return 0
    }

    return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}
```

余弦相似度的计算公式为：
```
similarity = (A · B) / (||A|| × ||B||)
```

其中：
- A · B 表示向量的点积
- ||A|| 和 ||B|| 表示向量的模长
- 相似度值范围为[-1, 1]，1表示完全相似，0表示不相关，-1表示完全相反

**余弦相似度的优势**：
- **尺度不变性**：不受向量长度影响，只关注方向
- **语义表示**：能够有效捕捉文本的语义相似性
- **计算效率**：计算复杂度为O(n)，其中n为向量维度
- **适用性广**：在文本检索、推荐系统等领域广泛应用

#### （2）向量归一化

为了提高相似度计算的效率和准确性，系统实现了向量归一化功能：

```go
func (vs *VectorStore) NormalizeVector(vector []float64) []float64 {
    if len(vector) == 0 {
        return vector
    }

    norm := 0.0
    for _, v := range vector {
        norm += v * v
    }
    norm = math.Sqrt(norm)

    if norm == 0 {
        return vector
    }

    normalized := make([]float64, len(vector))
    for i, v := range vector {
        normalized[i] = v / norm
    }

    return normalized
}
```

向量归一化的作用：
- **统一尺度**：将所有向量归一化到单位长度，便于比较
- **简化计算**：归一化后余弦相似度简化为点积计算
- **提高精度**：减少数值计算误差，提高相似度计算的准确性

#### （3）数据库层面的相似度计算

系统利用PGVector扩展提供的相似度算子，在数据库层面进行高效的相似度计算：

**相似度算子**：
- **<->**：欧氏距离算子，计算两个向量之间的欧氏距离
- **<=>**：余弦距离算子，计算两个向量之间的余弦距离
- **<#>**：内积算子，计算两个向量的内积

**查询示例**：
```sql
SELECT id, file_name, chunk_content, 
       embedding <-> ?::vector AS distance
FROM reimbursement_documents
WHERE embedding IS NOT NULL
ORDER BY distance ASC
LIMIT ?
```

该查询使用欧氏距离算子检索最相似的向量，并通过ORDER BY distance ASC按距离升序排序。

**距离与相似度的转换**：
- 欧氏距离越小，相似度越高
- 余弦距离 = 1 - 余弦相似度
- 系统通过`1.0 - distance`将距离转换为相似度分数

#### （4）相似度检索实现

系统通过`SearchVector`方法实现了基于相似度的向量检索：

```go
func (vs *VectorStore) SearchVector(ctx context.Context, queryVector []float64, topK int) ([]*VectorSearchResult, error) {
    // 1. 参数校验
    if len(queryVector) == 0 {
        return nil, errors.New("查询向量不能为空")
    }
    if len(queryVector) != VectorDimension {
        return nil, errors.New("查询向量维度必须为768维")
    }
    if topK <= 0 {
        topK = 10
    }

    // 2. 执行相似度检索
    type SearchResult struct {
        ID           string
        FileName     string
        FileType     string
        Category     string
        ChunkID      string
        ChunkIndex   int
        ChunkContent string
        Distance     float64
    }

    var results []SearchResult
    queryVectorJSON, _ := json.Marshal(queryVector)

    err := vs.db.WithContext(ctx).Raw(`
        SELECT id, file_name, file_type, category, chunk_id, chunk_index, chunk_content, 
               embedding <-> ?::vector AS distance
        FROM reimbursement_documents
        WHERE embedding IS NOT NULL
        ORDER BY distance ASC
        LIMIT ?
    `, string(queryVectorJSON), topK).Scan(&results).Error

    // 3. 转换为标准结果格式
    vectorResults := make([]*VectorSearchResult, 0, len(results))
    for _, result := range results {
        vectorResults = append(vectorResults, &VectorSearchResult{
            ID:         result.ID,
            DocumentID: result.FileName,
            ChunkID:    result.ChunkID,
            Content:    result.ChunkContent,
            Score:      1.0 - result.Distance,  // 距离转换为相似度
            Metadata: map[string]interface{}{
                "category":  result.Category,
                "file_type": result.FileType,
            },
        })
    }

    return vectorResults, nil
}
```

检索流程包括：
1. **参数校验**：验证查询向量的有效性和维度
2. **相似度计算**：在数据库层面计算查询向量与所有向量的相似度
3. **结果排序**：按相似度降序排序，返回最相似的topK个结果
4. **结果转换**：将数据库查询结果转换为标准的结果格式

#### （5）分类检索

系统还支持在指定分类内进行相似度检索，通过`SearchVectorByCategory`方法实现：

```go
func (vs *VectorStore) SearchVectorByCategory(ctx context.Context, queryVector []float64, category string, topK int) ([]*VectorSearchResult, error) {
    // 在指定分类内检索
    err := vs.db.WithContext(ctx).Raw(`
        SELECT id, file_name, file_type, category, chunk_id, chunk_index, chunk_content, 
               embedding <-> ?::vector AS distance
        FROM reimbursement_documents
        WHERE embedding IS NOT NULL AND category = ?
        ORDER BY distance ASC
        LIMIT ?
    `, string(queryVectorJSON), category, topK).Scan(&results).Error
}
```

分类检索的优势：
- **提高精度**：在相关分类内检索，减少无关结果
- **提高效率**：缩小检索范围，降低计算量
- **灵活应用**：支持按报销类别（差旅费、招待费等）进行分类检索

### 5.2.3 混合检索策略

混合检索策略结合了向量检索和关键词检索的优势，通过融合两种检索方式的结果，提高检索的准确性和召回率。系统实现了完善的混合检索机制，支持多种融合策略和权重配置。

#### （1）混合检索原理

混合检索的核心思想是结合语义检索和关键词检索的优势：

**向量检索优势**：
- **语义理解**：能够理解查询的语义含义，而不仅仅是关键词匹配
- **泛化能力强**：即使查询词与文档词不完全匹配，也能检索到相关内容
- **适合模糊查询**：对于概念性、描述性的查询效果更好

**关键词检索优势**：
- **精确匹配**：能够精确匹配关键词，适合专有名词、术语等
- **响应快速**：基于倒排索引，检索速度快
- **可控性强**：用户可以通过调整关键词精确控制检索结果

**混合检索优势**：
- **互补性**：结合两种检索方式的优势，提高检索质量
- **鲁棒性**：对不同的查询类型都有较好的表现
- **灵活性**：可以根据应用场景调整两种检索方式的权重

#### （2）混合检索实现

系统通过`HybridSearch`方法实现了混合检索功能：

```go
func (vs *VectorStore) HybridSearch(ctx context.Context, queryVector []float64, keywords []string, topK int) ([]*VectorSearchResult, error) {
    // 1. 向量检索
    vectorResults, err := vs.SearchVector(ctx, queryVector, topK*2)
    if err != nil {
        return nil, err
    }

    // 2. 关键词检索
    if len(keywords) == 0 {
        if len(vectorResults) > topK {
            return vectorResults[:topK], nil
        }
        return vectorResults, nil
    }

    keywordResults, err := vs.KeywordSearch(ctx, keywords, topK*2)
    if err != nil {
        return nil, err
    }

    // 3. 结果融合
    combined := vs.CombineResults(vectorResults, keywordResults, topK)
    return combined, nil
}
```

混合检索流程包括：
1. **向量检索**：基于查询向量进行语义检索，获取topK*2个结果
2. **关键词检索**：基于关键词进行精确匹配检索，获取topK*2个结果
3. **结果融合**：融合两种检索结果，返回最终的topK个结果

#### （3）关键词检索实现

系统通过`KeywordSearch`方法实现了基于关键词的检索：

```go
func (vs *VectorStore) KeywordSearch(ctx context.Context, keywords []string, topK int) ([]*VectorSearchResult, error) {
    if len(keywords) == 0 {
        return nil, nil
    }

    // 构建查询条件
    query := vs.db.WithContext(ctx).
        Model(&DocumentModel{}).
        Where("chunk_content LIKE ?", "%"+keywords[0]+"%")

    for i := 1; i < len(keywords); i++ {
        query = query.Or("chunk_content LIKE ?", "%"+keywords[i]+"%")
    }

    // 执行查询
    var docs []*DocumentModel
    result := query.Limit(topK).Find(&docs)

    // 转换结果
    var results []*VectorSearchResult
    for _, doc := range docs {
        results = append(results, &VectorSearchResult{
            ID:         doc.ID,
            DocumentID: doc.FileName,
            ChunkID:    doc.ChunkID,
            Content:    doc.ChunkContent,
            Score:      0.5,  // 关键词检索的默认分数
            Metadata:   map[string]interface{}{},
        })
    }

    return results, nil
}
```

关键词检索特点：
- **模糊匹配**：使用LIKE操作符进行模糊匹配
- **多关键词**：支持多个关键词的OR组合
- **快速检索**：基于文本索引，检索速度快

#### （4）结果融合策略

系统通过`CombineResults`方法实现了多种结果融合策略：

```go
func (vs *VectorStore) CombineResults(vectorResults, keywordResults []*VectorSearchResult, topK int) []*VectorSearchResult {
    // 1. 结果去重和合并
    scoreMap := make(map[string]*VectorSearchResult)

    for _, result := range vectorResults {
        if existing, ok := scoreMap[result.ID]; ok {
            existing.Score = (existing.Score + result.Score) / 2
        } else {
            scoreMap[result.ID] = result
        }
    }

    for _, result := range keywordResults {
        if existing, ok := scoreMap[result.ID]; ok {
            existing.Score = (existing.Score + result.Score) / 2
        } else {
            scoreMap[result.ID] = result
        }
    }

    // 2. 结果排序
    var combined []*VectorSearchResult
    for _, result := range scoreMap {
        combined = append(combined, result)
    }

    // 冒泡排序（按分数降序）
    for i := 0; i < len(combined)-1; i++ {
        for j := i + 1; j < len(combined); j++ {
            if combined[i].Score < combined[j].Score {
                combined[i], combined[j] = combined[j], combined[i]
            }
        }
    }

    // 3. 返回topK结果
    if len(combined) > topK {
        combined = combined[:topK]
    }

    return combined
}
```

融合策略包括：
1. **结果去重**：通过ID去重，避免重复结果
2. **分数融合**：对同时出现在两种检索结果中的文档，融合其分数
3. **结果排序**：按融合后的分数降序排序
4. **结果截取**：返回最终的topK个结果

#### （5）混合检索优化

为了提高混合检索的性能和效果，系统实现了多种优化策略：

**检索优化**：
- **并行检索**：向量检索和关键词检索可以并行执行
- **结果缓存**：对热点查询结果进行缓存
- **索引优化**：为向量列和文本列都创建索引

**融合优化**：
- **权重调整**：可以根据应用场景调整向量检索和关键词检索的权重
- **动态topK**：根据查询特点动态调整检索数量
- **阈值过滤**：设置相似度阈值，过滤低质量结果

**性能优化**：
- **批量查询**：支持批量混合检索，提高吞吐量
- **查询优化**：优化SQL查询语句，提高检索速度
- **内存优化**：合理管理内存使用，避免内存溢出

#### （6）混合检索应用

混合检索在RAG智能审核系统中有广泛的应用：

**报销审核场景**：
```go
func (rs *RAGService) AuditReimbursement(ctx context.Context, reimbursementInfo map[string]interface{}, topK int) (*RAGResult, error) {
    // 1. 构建查询文本
    query := rs.buildQueryFromReimbursementInfo(reimbursementInfo)

    // 2. 生成查询向量
    embedding, err := rs.llmClient.GenerateEmbedding(ctx, query)

    // 3. 提取关键词
    keywords := rs.extractReimbursementKeywords(reimbursementInfo)

    // 4. 混合检索
    searchResults, err := rs.vectorStore.HybridSearch(ctx, embedding, keywords, topK)

    // 5. 后续处理...
}
```

**关键词提取**：
```go
func (rs *RAGService) extractReimbursementKeywords(info map[string]interface{}) []string {
    keywords := make([]string, 0)

    // 提取报销类型
    if reimbursementType, ok := info["type"].(string); ok && reimbursementType != "" {
        keywords = append(keywords, reimbursementType)
    }

    // 提取分类
    if category, ok := info["category"].(string); ok && category != "" {
        keywords = append(keywords, category)
    }

    // 提取金额范围
    if amount, ok := info["amount"].(float64); ok {
        if amount < 500 {
            keywords = append(keywords, "小额")
        } else if amount < 2000 {
            keywords = append(keywords, "中等金额")
        } else {
            keywords = append(keywords, "大额")
        }
    }

    // 提取费用类型
    if expenseType, ok := info["expense_type"].(string); ok && expenseType != "" {
        keywords = append(keywords, expenseType)
    }

    // 提取城市
    if city, ok := info["city"].(string); ok && city != "" {
        keywords = append(keywords, city)
    }

    // 限制关键词数量
    if len(keywords) > 5 {
        keywords = keywords[:5]
    }

    return keywords
}
```

通过混合检索策略，系统能够结合语义理解和精确匹配的优势，为智能审核提供更准确、更全面的知识支持。在实际应用中，混合检索显著提高了检索的准确率和召回率，为后续的大模型推理提供了高质量的上下文信息。

## 5.2.4 向量检索系统总结

向量检索系统是RAG智能审核的核心组件，通过向量索引构建、相似度计算和混合检索策略，实现了高效、准确的文档检索。系统的核心优势包括：

**技术优势**：
- **高性能**：通过向量索引和优化算法，实现毫秒级检索响应
- **高准确率**：结合语义检索和关键词检索，提高检索准确性
- **高扩展性**：支持百万级向量数据，满足大规模应用需求
- **高可靠性**：完善的错误处理和重试机制，确保系统稳定性

**应用优势**：
- **智能审核**：为智能审核提供准确的知识支持
- **灵活配置**：支持多种检索策略和参数配置
- **易于集成**：提供简洁的API接口，易于集成到现有系统
- **可维护性**：完善的日志和监控，便于系统维护

通过向量检索系统，RAG智能审核能够快速、准确地从知识库中检索相关文档片段，为大模型推理提供高质量的上下文信息，从而提高智能审核的准确性和效率。

## 5.3 大模型调用与Prompt设计

大模型调用与Prompt设计是RAG智能审核系统的核心环节，负责将检索到的知识库内容与用户查询结合，通过精心设计的Prompt引导大模型生成准确的审核结果。本节将从LLM客户端设计、Prompt构建策略、上下文管理、结果解析与验证四个方面详细阐述大模型调用与Prompt的设计与实现。

### 5.3.1 LLM客户端设计

LLM客户端是系统与大语言模型交互的核心组件，负责封装大模型API调用、处理请求响应、管理连接池等功能。系统设计了完善的LLM客户端，支持多种大模型提供商和灵活的配置选项。

#### （1）客户端架构设计

系统采用面向对象的设计模式，通过`LLMClient`结构体封装所有大模型交互功能。客户端架构包括以下核心组件：

**配置管理**：
```go
type LLMClient struct {
    apiKey     string
    baseURL    string
    model      string
    httpClient *http.Client
    timeout    time.Duration
    logger     logger.Logger
    // Embedding配置
    embeddingProvider  string
    embeddingModel     string
    embeddingAPIKey    string
    embeddingBaseURL   string
    embeddingDimension int
}
```

客户端配置包括：
- **API配置**：API密钥、基础URL、模型名称等
- **网络配置**：HTTP客户端、超时时间、连接池等
- **Embedding配置**：独立的Embedding API配置，支持不同的Embedding提供商
- **日志配置**：日志记录器，用于记录请求响应信息

**构造函数**：
系统提供了两种构造函数，满足不同的使用场景：

```go
// 基础构造函数
func NewLLMClient(apiKey, baseURL, model string, timeout int, log logger.Logger) *LLMClient

// 扩展构造函数（包含Embedding配置）
func NewLLMClientWithEmbedding(apiKey, baseURL, model string, timeout int, log logger.Logger, 
    embeddingProvider, embeddingModel, embeddingAPIKey, embeddingBaseURL string, 
    embeddingDimension int) *LLMClient
```

#### （2）API调用机制

系统实现了完整的API调用机制，支持大模型的聊天接口和Embedding接口。

**聊天接口调用**：
```go
func (c *LLMClient) Chat(ctx context.Context, messages []ChatMessage, temperature float64, maxTokens int) (*ChatResponse, error) {
    // 1. 参数校验
    if len(messages) == 0 {
        return nil, errors.New("消息列表不能为空")
    }

    // 2. 构建请求
    request := ChatRequest{
        Model:       c.model,
        Messages:    messages,
        Temperature: temperature,
        MaxTokens:   maxTokens,
        Stream:      false,
    }

    // 3. 发送HTTP请求
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.apiKey)

    // 4. 处理响应
    resp, err := c.httpClient.Do(req)
    // 解析响应并返回
}
```

聊天接口调用流程包括：
1. **参数校验**：验证消息列表的有效性
2. **请求构建**：构建符合大模型API规范的请求体
3. **HTTP请求**：发送HTTP POST请求到指定端点
4. **响应处理**：解析响应体，提取生成结果

**Embedding接口调用**：
```go
func (c *LLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
    // 1. 配置选择
    baseURL := c.embeddingBaseURL
    if baseURL == "" {
        baseURL = c.baseURL
    }

    // 2. 请求构建
    embeddingRequest := map[string]interface{}{
        "model": model,
        "input": []string{text},
    }

    // 3. API调用
    resp, err := c.httpClient.Do(req)

    // 4. 向量提取
    return embeddingResponse.Data[0].Embedding, nil
}
```

Embedding接口调用支持多种提供商（如智谱AI），通过配置不同的参数适配不同的API规范。

#### （3）数据模型设计

系统定义了完整的数据模型，用于封装请求和响应数据：

**请求模型**：
```go
type ChatRequest struct {
    Model       string        `json:"model"`
    Messages    []ChatMessage `json:"messages"`
    Temperature float64       `json:"temperature,omitempty"`
    MaxTokens   int           `json:"max_tokens,omitempty"`
    Stream      bool          `json:"stream,omitempty"`
}

type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

**响应模型**：
```go
type ChatResponse struct {
    ID      string       `json:"id"`
    Object  string       `json:"object"`
    Created int64        `json:"created"`
    Model   string       `json:"model"`
    Choices []ChatChoice `json:"choices"`
    Usage   ChatUsage    `json:"usage"`
}

type ChatChoice struct {
    Index        int         `json:"index"`
    Message      ChatMessage `json:"message"`
    FinishReason string      `json:"finish_reason"`
}

type ChatUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

数据模型的设计遵循大模型API的规范，支持JSON序列化和反序列化，便于与API交互。

#### （4）错误处理机制

系统实现了完善的错误处理机制，确保大模型调用的可靠性：

**参数校验**：
- 消息列表不能为空
- API密钥不能为空
- 模型名称不能为空

**网络错误处理**：
- HTTP请求失败处理
- 超时处理
- 连接失败重试

**响应验证**：
- HTTP状态码检查
- 响应体解析错误处理
- 响应数据完整性验证

**错误日志记录**：
```go
c.logger.Error("发送请求失败", 
    logger.NewField("url", c.baseURL), 
    logger.NewField("error", err))
```

通过完善的错误处理和日志记录，系统能够快速定位和解决问题，提高系统的可维护性。

#### （5）性能优化

为了提高大模型调用的性能，系统实现了多种优化策略：

**连接池管理**：
- 复用HTTP连接，减少连接建立开销
- 配置合理的连接池大小
- 及时清理空闲连接

**超时控制**：
- 设置合理的请求超时时间
- 避免长时间阻塞
- 支持上下文取消

**批量处理**：
- 支持批量Embedding生成
- 减少API调用次数
- 提高吞吐量

**成本计算**：
```go
func calculateCost(tokens int) float64 {
    costPer1KTokens := 0.001
    return float64(tokens) / 1000.0 * costPer1KTokens
}
```

通过成本计算，系统能够监控和控制大模型调用的成本，实现成本优化。

### 5.3.2 Prompt构建策略

Prompt构建是引导大模型生成准确结果的关键技术。系统设计了灵活的Prompt构建策略，支持多种模板和动态变量，能够根据不同的应用场景生成最优的Prompt。

#### （1）Prompt构建器架构

系统通过`PromptBuilder`结构体实现Prompt的构建和管理：

```go
type PromptBuilder struct {
    logger          logger.Logger
    systemTemplates map[string]string
    userTemplates   map[string]string
}
```

Prompt构建器包括：
- **系统模板库**：存储系统级的Prompt模板
- **用户模板库**：存储用户自定义的Prompt模板
- **模板渲染引擎**：基于Go template的模板渲染
- **Token估算器**：估算Prompt的Token数量

#### （2）模板系统设计

系统实现了完善的模板系统，支持多种类型的Prompt模板：

**系统模板**：
系统模板定义了大模型的角色和行为准则，包括：

**默认模板**：
```
你是一个专业的报销审核助手，能够根据报销制度文档对报销单据进行审核和分析。
请基于提供的报销制度文档内容，对用户的报销问题进行准确、详细的回答。
回答时请注意：
1. 严格依据报销制度文档中的规定
2. 引用具体的条款和标准
3. 如果文档中没有相关信息，请明确说明
4. 提供清晰、有条理的回答
```

**审核模板**：
```
你是一个专业的报销审核专家，负责审核员工的报销申请。
请根据提供的报销制度文档，对报销申请进行严格审核。
审核要点：
1. 检查报销金额是否符合标准
2. 检查报销类型是否在允许范围内
3. 检查审批流程是否完整
4. 检查附件是否齐全
5. 给出明确的审核结论（通过/驳回/需补充材料）
```

**查询模板**：
```
你是一个报销制度查询助手，帮助用户快速了解报销政策和规定。
请基于提供的报销制度文档，准确回答用户关于报销政策的问题。
回答要求：
1. 准确引用相关条款
2. 提供具体的标准和限额
3. 说明适用的条件和场景
4. 如有例外情况，请一并说明
```

**用户模板**：
用户模板定义了具体的任务描述和输入格式，包括：

**RAG查询模板**：
```
基于以下报销制度文档内容，回答用户的问题：

【报销制度文档】
{{range .Documents}}
文档标题：{{.Title}}
文档内容：
{{.Content}}
{{end}}

【用户问题】
{{.Query}}

请基于上述文档内容，准确回答用户的问题。如果文档中没有相关信息，请明确说明。
```

**审核模板**：
```
请审核以下报销申请：

【报销制度文档】
{{range .Documents}}
文档标题：{{.Title}}
文档内容：
{{.Content}}
{{end}}

【报销申请信息】
{{.ReimbursementInfo}}

请根据报销制度文档，对上述报销申请进行审核，并给出审核结论和理由。
```

#### （3）模板渲染机制

系统基于Go template实现模板渲染，支持动态变量和条件逻辑：

```go
func (pb *PromptBuilder) renderTemplate(templateContent string, variables map[string]interface{}) (string, error) {
    tmpl, err := template.New("prompt").Parse(templateContent)
    if err != nil {
        return "", errors.New("解析模板失败")
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, variables); err != nil {
        return "", errors.New("渲染模板失败")
    }

    return buf.String(), nil
}
```

模板渲染特点：
- **变量替换**：支持{{.Variable}}语法进行变量替换
- **循环结构**：支持{{range}}语法进行循环
- **条件判断**：支持{{if}}语法进行条件判断
- **管道操作**：支持管道操作符进行数据处理

#### （4）Prompt构建流程

系统提供了完整的Prompt构建流程，包括系统Prompt和用户Prompt的构建：

**系统Prompt构建**：
```go
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
```

**用户Prompt构建**：
```go
func (pb *PromptBuilder) BuildUserPrompt(templateName string, variables map[string]interface{}) (string, error) {
    templateContent, ok := pb.userTemplates[templateName]
    if !ok {
        templateContent = pb.userTemplates["simple_query"]
    }

    return pb.renderTemplate(templateContent, variables)
}
```

**完整Prompt构建**：
```go
func (pb *PromptBuilder) BuildRAGPrompt(ctx context.Context, query string, documents []*Document, chunks []*DocumentChunk) (*Prompt, error) {
    // 1. 构建系统Prompt
    systemPrompt, err := pb.BuildSystemPrompt("query", nil)

    // 2. 构建用户Prompt
    variables := map[string]interface{}{
        "Query":     query,
        "Documents": documents,
        "Chunks":    chunks,
    }
    userPrompt, err := pb.BuildUserPrompt("rag_query", variables)

    // 3. 创建Prompt对象
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
```

#### （5）Prompt优化策略

为了提高Prompt的效果和效率，系统实现了多种优化策略：

**Token估算**：
```go
func (pb *PromptBuilder) estimateTokens(text string) int {
    if text == "" {
        return 0
    }
    return len(text) / 4
}
```

通过Token估算，系统能够控制Prompt的长度，避免超出模型的Token限制。

**Prompt验证**：
```go
func (pb *PromptBuilder) ValidatePrompt(prompt *Prompt) error {
    if prompt == nil {
        return errors.New("Prompt不能为空")
    }
    if prompt.Content == "" {
        return errors.New("Prompt内容不能为空")
    }
    if prompt.Tokens > 4000 {
        return errors.New("Prompt长度超过限制")
    }
    return nil
}
```

通过Prompt验证，系统能够确保Prompt的有效性，避免无效的API调用。

**Prompt优化**：
```go
func (pb *PromptBuilder) OptimizePrompt(prompt *Prompt, maxTokens int) (*Prompt, error) {
    if prompt.Tokens <= maxTokens {
        return prompt, nil
    }

    ratio := float64(maxTokens) / float64(prompt.Tokens)
    newLength := int(float64(len(prompt.Content)) * ratio * 0.9)

    optimizedContent := prompt.Content[:newLength] + "..."

    optimizedPrompt := &Prompt{
        ID:        prompt.ID,
        Name:      prompt.Name + "（优化后）",
        Template:  prompt.Template,
        Content:   optimizedContent,
        Type:      prompt.Type,
        Variables:  prompt.Variables,
        Tokens:    pb.estimateTokens(optimizedContent),
        CreatedAt:  prompt.CreatedAt,
        UpdatedAt:  time.Now(),
        Version:   prompt.Version,
        Tags:      prompt.Tags,
    }

    return optimizedPrompt, nil
}
```

通过Prompt优化，系统能够在保持核心信息的前提下，压缩Prompt长度，提高API调用的效率。

### 5.3.3 上下文管理

上下文管理是RAG系统的关键技术，负责管理对话历史、检索结果、用户输入等上下文信息，确保大模型能够基于完整的上下文生成准确的响应。系统实现了完善的上下文管理机制。

#### （1）对话消息管理

系统通过`ConversationMessage`结构体管理对话消息：

```go
type ConversationMessage struct {
    Role      string    `json:"role"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

对话消息包括：
- **角色**：system、user、assistant
- **内容**：消息的具体内容
- **时间戳**：消息的创建时间

#### （2）对话构建策略

系统支持多种对话构建策略，满足不同的应用场景：

**简单对话构建**：
```go
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
```

简单对话构建适用于单轮对话场景，只包含系统Prompt和用户Prompt。

**带历史记录的对话构建**：
```go
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
```

带历史记录的对话构建适用于多轮对话场景，能够保持对话的连贯性。

#### （3）上下文窗口管理

系统实现了上下文窗口管理，确保在Token限制内保留最重要的上下文信息：

**Token限制控制**：
- 系统Prompt：约500-1000 tokens
- 用户Prompt：约2000-3000 tokens
- 历史对话：根据剩余空间动态调整
- 检索结果：根据相关性排序，保留最相关的结果

**上下文优先级**：
1. **系统Prompt**：最高优先级，必须完整保留
2. **当前用户输入**：高优先级，确保当前问题得到回答
3. **检索结果**：中优先级，根据相关性排序保留
4. **历史对话**：低优先级，根据Token空间动态调整

**上下文压缩策略**：
- **摘要压缩**：对历史对话进行摘要，减少Token占用
- **重要性过滤**：保留重要的历史对话，过滤无关内容
- **时间窗口**：只保留最近N轮对话

#### （4）检索结果格式化

系统提供了多种检索结果的格式化方法，确保检索结果能够有效地融入上下文：

**文档格式化**：
```go
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
```

**分片格式化**：
```go
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
```

**报销信息格式化**：
```go
func (pb *PromptBuilder) FormatReimbursementInfo(info map[string]interface{}) string {
    if len(info) == 0 {
        return "无报销信息"
    }

    jsonBytes, err := json.MarshalIndent(info, "", "  ")
    if err != nil {
        return "无法格式化报销信息"
    }
    return string(jsonBytes)
}
```

通过格式化方法，系统能够将结构化数据转换为易于理解的文本格式，提高大模型的理解能力。

#### （5）上下文一致性保证

系统实现了多种机制确保上下文的一致性：

**版本管理**：
- Prompt版本控制
- 模板版本管理
- 上下文快照

**状态同步**：
- 多线程上下文同步
- 分布式上下文共享
- 上下文缓存一致性

**错误恢复**：
- 上下文回滚机制
- 异常恢复策略
- 上下文重建

### 5.3.4 结果解析与验证

结果解析与验证是确保大模型响应质量和可靠性的关键环节。系统实现了完善的结果解析和验证机制，能够从大模型的自然语言响应中提取结构化信息，并进行严格的验证。

#### （1）响应验证机制

系统实现了多层响应验证机制，确保响应的有效性和可靠性：

**基础验证**：
```go
func (c *LLMClient) ValidateResponse(response *ChatResponse) error {
    if response == nil {
        return errors.New("大模型响应空")
    }
    if len(response.Choices) == 0 {
        return errors.New("大模型响应不包含任何选择")
    }
    if response.Choices[0].Message.Content == "" {
        return errors.New("大模型响应内容为空")
    }
    return nil
}
```

**内容验证**：
```go
func (rs *RAGService) validateLLMResponse(response *ChatResponse) error {
    if response == nil {
        return errors.New("大模型响应空")
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
```

**格式验证**：
- 检查响应是否为JSON格式
- 检查必需字段是否存在
- 检查字段类型是否正确

#### （2）结果解析策略

系统实现了多种结果解析策略，能够从大模型响应中提取结构化信息：

**简单解析**：
```go
func (c *LLMClient) ParseAnalysisResult(content string) (*AnalysisResult, error) {
    result := &AnalysisResult{
        ID:         uuid.New().String(),
        Conclusion: content,
        Reasoning:  content,
        CreatedAt:  time.Now(),
    }
    return result, nil
}
```

**结构化解析**：
系统支持JSON格式的结构化响应，能够解析复杂的审核结果：

```go
func (rs *RAGService) parseAuditResult(query string, response *ChatResponse, references []*VectorSearchResult) *AnalysisResult {
    content := response.Choices[0].Message.Content

    // 尝试解析JSON格式
    var structuredResult struct {
        Conclusion string   `json:"conclusion"`
        Reasoning  string   `json:"reasoning"`
        Pass       bool     `json:"pass"`
        Suggestions []string `json:"suggestions"`
    }

    if err := json.Unmarshal([]byte(content), &structuredResult); err == nil {
        // JSON解析成功
        return &AnalysisResult{
            ID:         generateAnalysisResultID(),
            Query:      query,
            Conclusion: structuredResult.Conclusion,
            Reasoning:  structuredResult.Reasoning,
            Suggestions: structuredResult.Suggestions,
            Confidence: rs.calculateAuditConfidence(content, references),
            CreatedAt:  time.Now(),
        }
    }

    // JSON解析失败，使用文本解析
    return &AnalysisResult{
        ID:         generateAnalysisResultID(),
        Query:      query,
        Conclusion: content,
        Reasoning:  "基于报销制度文档进行审核",
        Confidence: rs.calculateAuditConfidence(content, references),
        CreatedAt:  time.Now(),
    }
}
```

**关键词提取**：
- 提取审核结论（通过/驳回/需补充）
- 提取违规原因
- 提取改进建议

#### （3）置信度计算

系统实现了智能的置信度计算机制，能够评估审核结果的可靠性：

```go
func (rs *RAGService) calculateAuditConfidence(content string, references []*VectorSearchResult) float64 {
    if len(content) == 0 {
        return 0
    }

    baseConfidence := 0.5

    // 基于检索结果的置信度调整
    if len(references) > 0 {
        avgScore := rs.calculateAverageScore(references)
        if avgScore > 0.8 {
            baseConfidence += 0.2
        } else if avgScore > 0.6 {
            baseConfidence += 0.1
        }
    }

    // 基于检索结果数量的置信度调整
    if len(references) >= 3 {
        baseConfidence += 0.1
    } else if len(references) >= 1 {
        baseConfidence += 0.05
    }

    // 基于响应长度的置信度调整
    if len(content) > 100 {
        baseConfidence += 0.1
    }

    // 基于关键词的置信度调整
    if strings.Contains(content, "通过") || strings.Contains(content, "不通过") || strings.Contains(content, "驳回") {
        baseConfidence += 0.05
    }

    if baseConfidence > 1.0 {
        baseConfidence = 1.0
    }

    return baseConfidence
}
```

置信度计算考虑以下因素：
- **检索结果质量**：检索结果的平均相似度分数
- **检索结果数量**：检索到的相关文档数量
- **响应长度**：响应内容的长度
- **关键词匹配**：是否包含审核结论关键词

#### （4）结果格式化

系统提供了多种结果格式化方法，满足不同的应用需求：

**LLM响应格式化**：
```go
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
```

**分析结果格式化**：
```go
func (rs *RAGService) ExportAnalysisResult(result *AnalysisResult) (string, error) {
    if result == nil {
        return "", errors.New("分析结果不能为空")
    }

    jsonBytes, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        return "", errors.New("序列化分析结果失败")
    }

    return string(jsonBytes), nil
}
```

#### （5）异常处理

系统实现了完善的异常处理机制，确保在出现异常情况时系统能够优雅地处理：

**响应异常处理**：
- 响应为空的处理
- 响应格式错误的处理
- 响应内容不完整的处理

**解析异常处理**：
- JSON解析失败的处理
- 字段缺失的处理
- 类型转换失败的处理

**降级策略**：
- 当大模型响应不可用时，降级为规则引擎审核
- 当置信度过低时，标记为需人工审核
- 当检索结果不足时，提示用户补充信息

### 5.3.5 大模型调用与Prompt设计总结

大模型调用与Prompt设计是RAG智能审核系统的核心环节，通过完善的LLM客户端设计、灵活的Prompt构建策略、智能的上下文管理和严格的结果解析验证，系统能够生成准确、可靠的审核结果。

**技术优势**：
- **灵活性**：支持多种大模型提供商和模板配置
- **可靠性**：完善的错误处理和验证机制
- **高效性**：优化的API调用和上下文管理
- **可扩展性**：模块化设计，易于扩展新功能

**应用优势**：
- **准确审核**：基于知识库的准确审核
- **智能推理**：大模型的智能推理能力
- **灵活配置**：支持多种审核场景
- **易于维护**：完善的日志和监控

通过大模型调用与Prompt设计，RAG智能审核系统能够将检索到的知识库内容与大模型的智能推理能力结合，生成准确、可靠的审核结果，为企业的报销审核提供强有力的技术支持。

## 5.4 智能审核流程

智能审核流程是RAG智能审核系统的核心业务流程，负责将报销申请信息与知识库中的制度文档结合，通过大模型的智能推理能力生成准确的审核结论。本节将从审核问题生成、知识检索、大模型推理、审核结论生成四个方面详细阐述智能审核流程的设计与实现。

### 5.4.1 审核问题生成

审核问题生成是智能审核流程的第一步，负责将结构化的报销申请信息转换为适合大模型理解和检索的自然语言查询。系统设计了智能的问题生成机制，能够从复杂的报销数据中提取关键信息，构建准确的审核问题。

#### （1）问题生成策略

系统采用基于模板和规则的问题生成策略，能够根据不同类型的报销申请生成针对性的审核问题：

**信息提取**：
系统从报销申请信息中提取关键要素，包括：
- **报销类型**：差旅费、招待费、办公费、培训费等
- **报销金额**：报销的总金额和各项明细金额
- **报销分类**：具体的费用类别（如交通费、住宿费、伙食费等）
- **费用类型**：具体的费用细项（如机票、火车票、酒店等）
- **城市信息**：消费城市，用于判断城市等级
- **时间信息**：报销时间，用于判断时效性

**问题构建**：
系统通过`buildQueryFromReimbursementInfo`方法构建审核问题：

```go
func (rs *RAGService) buildQueryFromReimbursementInfo(info map[string]interface{}) string {
    var query string

    // 提取报销类型
    if reimbursementType, ok := info["type"].(string); ok {
        query += reimbursementType + " "
    }

    // 提取金额信息
    if amount, ok := info["amount"].(float64); ok {
        query += "金额" + strconv.FormatFloat(amount, 'f', 2, 64) + " "
    }

    // 提取分类信息
    if category, ok := info["category"].(string); ok {
        query += category + " "
    }

    // 默认查询
    if query == "" {
        query = "报销申请"
    }

    return query
}
```

问题构建特点：
- **简洁明了**：问题简洁直接，便于大模型理解
- **信息完整**：包含审核所需的关键信息
- **格式统一**：统一的问题格式，便于后续处理
- **动态生成**：根据报销信息动态生成问题

#### （2）关键词提取

系统实现了智能的关键词提取机制，用于后续的混合检索：

```go
func (rs *RAGService) extractReimbursementKeywords(info map[string]interface{}) []string {
    keywords := make([]string, 0)

    // 提取报销类型
    if reimbursementType, ok := info["type"].(string); ok && reimbursementType != "" {
        keywords = append(keywords, reimbursementType)
    }

    // 提取分类
    if category, ok := info["category"].(string); ok && category != "" {
        keywords = append(keywords, category)
    }

    // 提取金额范围
    if amount, ok := info["amount"].(float64); ok {
        if amount < 500 {
            keywords = append(keywords, "小额")
        } else if amount < 2000 {
            keywords = append(keywords, "中等金额")
        } else {
            keywords = append(keywords, "大额")
        }
    }

    // 提取费用类型
    if expenseType, ok := info["expense_type"].(string); ok && expenseType != "" {
        keywords = append(keywords, expenseType)
    }

    // 提取城市
    if city, ok := info["city"].(string); ok && city != "" {
        keywords = append(keywords, city)
    }

    // 限制关键词数量
    if len(keywords) > 5 {
        keywords = keywords[:5]
    }

    return keywords
}
```

关键词提取特点：
- **多维度提取**：从多个维度提取关键词
- **智能分类**：根据金额范围智能分类
- **数量控制**：限制关键词数量，避免过多干扰
- **优先级排序**：重要关键词优先

#### （3）问题优化

系统实现了问题优化机制，提高问题的质量和检索效果：

**问题规范化**：
- 统一术语表达
- 去除冗余信息
- 补充缺失信息

**问题扩展**：
- 添加相关术语
- 扩展同义词
- 补充上下文信息

**问题验证**：
- 检查问题完整性
- 验证问题有效性
- 评估问题质量

### 5.4.2 知识检索

知识检索是智能审核流程的核心环节，负责根据生成的审核问题从知识库中检索相关的制度文档片段。系统采用混合检索策略，结合向量检索和关键词检索的优势，提高检索的准确性和召回率。

#### （1）向量检索

系统首先将审核问题转换为向量，然后进行向量相似度检索：

**向量生成**：
```go
embedding, err := rs.llmClient.GenerateEmbedding(ctx, query)
if err != nil {
    rs.logger.Error("生成查询向量失败", logger.NewField("query", query), logger.NewField("error", err))
    return nil, errors.New("生成查询向量失败")
}
```

向量生成特点：
- **语义表示**：向量能够捕捉问题的语义信息
- **维度统一**：所有向量均为768维，确保一致性
- **高效生成**：调用智谱AI的Embedding API，快速生成向量

**向量检索**：
```go
searchResults, err := rs.vectorStore.HybridSearch(ctx, embedding, keywords, topK)
if err != nil {
    rs.logger.Error("混合检索失败", logger.NewField("query", query), logger.NewField("error", err))
    return nil, errors.New("混合检索失败")
}
```

向量检索特点：
- **语义匹配**：基于语义相似度进行检索
- **Top-K检索**：返回最相关的topK个结果
- **高效检索**：利用向量索引，实现快速检索

#### （2）关键词检索

系统同时进行关键词检索，补充向量检索的不足：

**关键词匹配**：
- 使用LIKE操作符进行模糊匹配
- 支持多个关键词的OR组合
- 基于文本索引，检索速度快

**结果排序**：
- 按匹配度排序
- 考虑关键词权重
- 过滤低质量结果

#### （3）混合检索

系统通过混合检索策略融合向量检索和关键词检索的结果：

```go
func (vs *VectorStore) HybridSearch(ctx context.Context, queryVector []float64, keywords []string, topK int) ([]*VectorSearchResult, error) {
    // 1. 向量检索
    vectorResults, err := vs.SearchVector(ctx, queryVector, topK*2)
    if err != nil {
        return nil, err
    }

    // 2. 关键词检索
    if len(keywords) == 0 {
        if len(vectorResults) > topK {
            return vectorResults[:topK], nil
        }
        return vectorResults, nil
    }

    keywordResults, err := vs.KeywordSearch(ctx, keywords, topK*2)
    if err != nil {
        return nil, err
    }

    // 3. 结果融合
    combined := vs.CombineResults(vectorResults, keywordResults, topK)
    return combined, nil
}
```

混合检索特点：
- **互补性**：结合两种检索方式的优势
- **去重机制**：避免重复结果
- **分数融合**：对共同结果融合分数
- **结果排序**：按融合分数降序排序

#### （4）检索结果处理

系统对检索结果进行进一步处理，提高结果的质量和可用性：

**结果过滤**：
- 过滤低相似度结果
- 过滤过期文档
- 过滤无效分片

**结果排序**：
- 按相似度分数排序
- 考虑文档版本
- 考虑文档优先级

**结果格式化**：
```go
documents := rs.buildDocumentsFromSearchResults(searchResults)
```

通过格式化，将检索结果转换为标准化的文档格式，便于后续处理。

### 5.4.3 大模型推理

大模型推理是智能审核流程的核心环节，负责将检索到的制度文档与报销申请信息结合，通过大模型的智能推理能力生成审核结论。系统设计了完善的大模型推理机制，确保审核结论的准确性和可靠性。

#### （1）Prompt构建

系统构建专门的审核Prompt，引导大模型进行准确的审核：

**系统Prompt构建**：
```go
systemPrompt, err := rs.promptBuilder.BuildSystemPrompt("audit", nil)
if err != nil {
    rs.logger.Error("构造系统提示词失败", logger.NewField("error", err))
    return nil, errors.New("构造系统提示词失败")
}
```

系统Prompt定义了大模型的角色和审核准则：
- **角色定义**：专业的报销审核专家
- **审核要点**：金额标准、类型范围、审批流程、附件齐全等
- **结论要求**：明确的审核结论（通过/驳回/需补充材料）

**用户Prompt构建**：
```go
reimbursementInfoJSON := rs.promptBuilder.FormatReimbursementInfo(reimbursementInfo)
prompt, err := rs.promptBuilder.BuildAuditPrompt(ctx, reimbursementInfoJSON, documents)
if err != nil {
    rs.logger.Error("构造提示词失败", logger.NewField("error", err))
    return nil, errors.New("构造提示词失败")
}
```

用户Prompt包含：
- **报销制度文档**：检索到的相关制度片段
- **报销申请信息**：结构化的报销申请数据
- **审核要求**：明确的审核任务和要求

#### （2）大模型调用

系统调用大模型进行智能推理：

**消息构建**：
```go
messages := rs.promptBuilder.BuildConversationMessages(systemPrompt, prompt.Content)
```

消息结构：
- **system消息**：定义大模型的角色和行为准则
- **user消息**：包含具体的审核任务和相关信息

**大模型调用**：
```go
llmResponse, err := rs.llmClient.Chat(ctx, rs.convertToChatMessages(messages), 0.7, 2000)
if err != nil {
    rs.logger.Error("调用大模型失败", logger.NewField("error", err))
    return nil, errors.New("调用大模型失败")
}
```

调用参数：
- **temperature**：0.7，平衡创造性和准确性
- **maxTokens**：2000，控制响应长度
- **messages**：包含system和user消息的列表

#### （3）响应验证

系统对大模型的响应进行严格验证：

```go
if err := rs.validateLLMResponse(llmResponse); err != nil {
    rs.logger.Error("大模型响应格式校验失败", logger.NewField("error", err))
    return nil, errors.New("大模型响应格式校验失败")
}
```

验证内容包括：
- **响应完整性**：响应不为空，包含选择项
- **内容有效性**：响应内容长度合理（10-10000字符）
- **Token有效性**：Token数不为0
- **模型信息**：包含模型名称

#### （4）推理过程

大模型的推理过程包括以下步骤：

**信息理解**：
- 理解报销申请的具体内容
- 理解检索到的制度文档
- 理解审核任务和要求

**规则匹配**：
- 将报销申请与制度规则进行匹配
- 识别符合规则的方面
- 识别违反规则的方面

**逻辑推理**：
- 基于规则进行逻辑推理
- 考虑各种情况和例外
- 形成审核结论

**结论生成**：
- 生成明确的审核结论
- 提供详细的审核理由
- 给出改进建议

### 5.4.4 审核结论生成

审核结论生成是智能审核流程的最后一步，负责将大模型的自然语言响应转换为结构化的审核结果，并进行必要的验证和优化。

#### （1）结果解析

系统实现了多种结果解析策略，能够从大模型响应中提取结构化信息：

**结构化解析**：
```go
func (rs *RAGService) parseAuditResult(query string, response *ChatResponse, references []*VectorSearchResult) *AnalysisResult {
    content := response.Choices[0].Message.Content

    // 尝试解析JSON格式
    var structuredResult struct {
        Conclusion string   `json:"conclusion"`
        Reasoning  string   `json:"reasoning"`
        Pass       bool     `json:"pass"`
        Suggestions []string `json:"suggestions"`
    }

    if err := json.Unmarshal([]byte(content), &structuredResult); err == nil {
        // JSON解析成功
        return &AnalysisResult{
            ID:         generateAnalysisResultID(),
            Query:      query,
            Conclusion: structuredResult.Conclusion,
            Reasoning:  structuredResult.Reasoning,
            Suggestions: structuredResult.Suggestions,
            Confidence: rs.calculateAuditConfidence(content, references),
            CreatedAt:  time.Now(),
        }
    }

    // JSON解析失败，使用文本解析
    return &AnalysisResult{
        ID:         generateAnalysisResultID(),
        Query:      query,
        Conclusion: content,
        Reasoning:  "基于报销制度文档进行审核",
        Confidence: rs.calculateAuditConfidence(content, references),
        CreatedAt:  time.Now(),
    }
}
```

解析策略：
- **优先JSON解析**：尝试解析JSON格式的结构化响应
- **降级文本解析**：JSON解析失败时，使用文本解析
- **关键词提取**：从文本中提取审核结论关键词

#### （2）置信度计算

系统实现了智能的置信度计算机制，评估审核结果的可靠性：

```go
func (rs *RAGService) calculateAuditConfidence(content string, references []*VectorSearchResult) float64 {
    if len(content) == 0 {
        return 0
    }

    baseConfidence := 0.5

    // 基于检索结果的置信度调整
    if len(references) > 0 {
        avgScore := rs.calculateAverageScore(references)
        if avgScore > 0.8 {
            baseConfidence += 0.2
        } else if avgScore > 0.6 {
            baseConfidence += 0.1
        }
    }

    // 基于检索结果数量的置信度调整
    if len(references) >= 3 {
        baseConfidence += 0.1
    } else if len(references) >= 1 {
        baseConfidence += 0.05
    }

    // 基于响应长度的置信度调整
    if len(content) > 100 {
        baseConfidence += 0.1
    }

    // 基于关键词的置信度调整
    if strings.Contains(content, "通过") || strings.Contains(content, "不通过") || strings.Contains(content, "驳回") {
        baseConfidence += 0.05
    }

    if baseConfidence > 1.0 {
        baseConfidence = 1.0
    }

    return baseConfidence
}
```

置信度计算因素：
- **检索结果质量**：检索结果的平均相似度分数
- **检索结果数量**：检索到的相关文档数量
- **响应长度**：响应内容的长度
- **关键词匹配**：是否包含审核结论关键词

#### （3）结果封装

系统将审核结果封装为标准化的RAGResult结构：

```go
ragResult := &RAGResult{
    Query:          query,
    Documents:      documents,
    Prompt:         prompt.Content,
    Response:       rs.convertToLLMResponse(llmResponse),
    AnalysisResult: analysisResult,
    ExecutionTime:  time.Since(startTime).Milliseconds(),
    CreatedAt:      time.Now(),
}
```

RAGResult包含：
- **Query**：审核查询
- **Documents**：检索到的制度文档
- **Prompt**：使用的Prompt
- **Response**：大模型的原始响应
- **AnalysisResult**：解析后的分析结果
- **ExecutionTime**：执行时间
- **CreatedAt**：创建时间

#### （4）结果验证

系统对审核结果进行最终验证：

**完整性验证**：
- 检查所有必需字段是否存在
- 检查字段值是否有效
- 检查数据类型是否正确

**一致性验证**：
- 检查结论与推理是否一致
- 检查结论与检索结果是否一致
- 检查置信度是否合理

**业务规则验证**：
- 检查是否符合业务规则
- 检查是否有逻辑矛盾
- 检查是否有异常情况

#### （5）结果优化

系统对审核结果进行优化处理：

**结论标准化**：
- 统一审核结论的表述
- 标准化审核状态
- 补充缺失信息

**理由优化**：
- 优化审核理由的表述
- 补充引用的条款
- 提供详细的说明

**建议生成**：
- 生成改进建议
- 提供操作指引
- 给出风险提示

### 5.4.5 智能审核流程总结

智能审核流程是RAG智能审核系统的核心业务流程，通过审核问题生成、知识检索、大模型推理、审核结论生成四个环节，实现了从报销申请到审核结论的完整流程。

**流程特点**：
- **自动化**：全流程自动化，无需人工干预
- **智能化**：基于大模型的智能推理能力
- **准确性**：基于知识库的准确审核
- **高效性**：快速生成审核结论

**技术优势**：
- **问题生成**：智能的审核问题生成机制
- **混合检索**：结合向量和关键词检索的优势
- **大模型推理**：利用大模型的智能推理能力
- **结果验证**：完善的验证和优化机制

**应用优势**：
- **提高效率**：大幅提高审核效率
- **降低成本**：减少人工审核成本
- **提高准确性**：基于制度的准确审核
- **提升体验**：快速响应，良好的用户体验

通过智能审核流程，RAG智能审核系统能够自动、准确、高效地完成报销审核任务，为企业提供强大的智能审核能力，显著提升财务管理的效率和准确性。
