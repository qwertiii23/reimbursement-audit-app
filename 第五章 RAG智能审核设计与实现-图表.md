# 第五章 RAG智能审核设计与实现 - 图表

## 5.1 知识库构建流程图

```mermaid
flowchart TD
    A[原始文档] --> B[文档采集]
    B --> C{文档类型识别}
    C -->|TXT| D[TXT解析器]
    C -->|PDF| E[PDF解析器]
    C -->|Word| F[Word解析器]
    D --> G[文档预处理]
    E --> G
    F --> G

    G --> H[文本清洗]
    H --> I[元数据提取]
    I --> J[文档分片]

    J --> K[滑动窗口分片]
    K --> L[分片优化]
    L --> M[生成DocumentChunk对象]

    M --> N[向量嵌入生成]
    N --> O[调用智谱AI Embedding API]
    O --> P[生成768维向量]

    P --> Q[向量存储]
    Q --> R[存储到PostgreSQL]
    R --> S[创建向量索引]
    S --> T[知识库构建完成]

    style A fill:#e1f5ff
    style T fill:#4caf50
```

## 5.2 文档分片策略示意图

```mermaid
graph LR
    A[原始文档内容] --> B[分词处理]
    B --> C[单词列表]

    C --> D[滑动窗口分片]
    D --> E[分片1<br/>chunkSize=500词<br/>chunkOverlap=50词]
    D --> F[分片2<br/>chunkSize=500词<br/>chunkOverlap=50词]
    D --> G[分片3<br/>chunkSize=500词<br/>chunkOverlap=50词]

    E --> H[分片优化]
    F --> H
    G --> H

    H --> I[空分片过滤]
    I --> J[分片验证]
    J --> K[生成DocumentChunk对象]

    K --> L[向量嵌入]
    L --> M[768维向量]

    style A fill:#e1f5ff
    style M fill:#ff9800
```

## 5.3 向量存储架构图

```mermaid
graph TB
    A[DocumentChunk对象] --> B[VectorStore]
    B --> C[GORM ORM框架]
    C --> D[PostgreSQL数据库]

    D --> E[PGVector扩展]
    E --> F[reimbursement_documents表]

    F --> G[字段结构]
    G --> H[ID: 唯一标识]
    G --> I[FileName: 文档名称]
    G --> J[FileType: 文档类型]
    G --> K[Category: 文档分类]
    G --> L[ChunkID: 分片ID]
    G --> M[ChunkIndex: 分片索引]
    G --> N[ChunkContent: 分片内容]
    G --> O[Embedding: 768维向量]
    G --> P[时间戳: 创建/更新时间]

    O --> Q[向量索引]
    Q --> R[IVFFlat索引]
    Q --> S[HNSW索引]
    Q --> T[余弦相似度索引]

    style A fill:#e1f5ff
    style F fill:#4caf50
    style O fill:#ff9800
```

## 5.4 向量索引结构图

```mermaid
graph TB
    A[向量数据] --> B[IVFFlat索引]
    A --> C[HNSW索引]
    A --> D[余弦相似度索引]

    B --> E[聚类中心1]
    B --> F[聚类中心2]
    B --> G[聚类中心N]

    C --> H[第1层图]
    C --> I[第2层图]
    C --> J[第N层图]

    H --> K[近似最近邻搜索]
    I --> K
    J --> K

    K --> L[Top-K结果]
    L --> M[按距离排序]

    style A fill:#e1f5ff
    style K fill:#ff9800
    style M fill:#4caf50
```

## 5.5 混合检索流程图

```mermaid
flowchart TD
    A[用户查询] --> B[查询向量化]
    B --> C[生成768维查询向量]

    C --> D[向量检索]
    C --> E[关键词提取]

    D --> F[向量相似度检索]
    F --> G[Top-K*2结果]
    G --> H[按相似度排序]

    E --> I[关键词检索]
    I --> J[LIKE模糊匹配]
    J --> K[Top-K*2结果]
    K --> L[按匹配度排序]

    H --> M[结果融合]
    L --> M

    M --> N[结果去重]
    N --> O[分数融合]
    O --> P[结果排序]

    P --> Q[返回Top-K结果]
    Q --> R[最终检索结果]

    style A fill:#e1f5ff
    style R fill:#4caf50
```

## 5.6 相似度计算流程图

```mermaid
flowchart LR
    A[查询向量] --> B[向量归一化]
    B --> C[单位向量]

    D[文档向量] --> E[向量归一化]
    E --> F[单位向量]

    C --> G[余弦相似度计算]
    F --> G

    G --> H[计算点积]
    H --> I[A·B]

    G --> J[计算模长]
    J --> K[||A|| × ||B||]

    I --> L[相似度 = 点积 / 模长]
    K --> L

    L --> M[相似度值<br/>范围: [-1, 1]]
    M --> N[1 = 完全相似]
    M --> O[0 = 不相关]
    M --> P[-1 = 完全相反]

    style A fill:#e1f5ff
    style D fill:#e1f5ff
    style M fill:#ff9800
```

## 5.7 LLM客户端架构图

```mermaid
graph TB
    A[LLMClient结构体] --> B[配置管理]
    A --> C[API调用机制]
    A --> D[数据模型]
    A --> E[错误处理]
    A --> F[性能优化]

    B --> G[API密钥]
    B --> H[基础URL]
    B --> I[模型名称]
    B --> J[HTTP客户端]
    B --> K[超时时间]
    B --> L[日志记录器]
    B --> M[Embedding配置]

    C --> N[聊天接口]
    C --> O[Embedding接口]

    N --> P[Chat方法]
    N --> Q[构建ChatRequest]
    Q --> R[发送HTTP请求]
    R --> S[解析ChatResponse]

    O --> T[GenerateEmbedding方法]
    O --> U[构建EmbeddingRequest]
    U --> V[发送HTTP请求]
    V --> W[解析EmbeddingResponse]

    D --> X[请求模型]
    D --> Y[响应模型]
    D --> Z[消息模型]
    D --> AA[使用模型]

    style A fill:#e1f5ff
    style C fill:#4caf50
```

## 5.8 Prompt构建流程图

```mermaid
flowchart TD
    A[用户请求] --> B{选择模板类型}

    B -->|系统模板| C[BuildSystemPrompt]
    B -->|用户模板| D[BuildUserPrompt]
    B -->|RAG查询| E[BuildRAGPrompt]

    C --> F[获取模板内容]
    F --> G[变量替换]
    G --> H[系统Prompt]

    D --> I[获取模板内容]
    I --> J[变量替换]
    J --> K[用户Prompt]

    E --> L[构建系统Prompt]
    E --> M[构建用户Prompt]
    E --> N[创建Prompt对象]

    L --> O[合并系统Prompt和用户Prompt]
    M --> O

    O --> P[模板渲染]
    P --> Q[Go template引擎]

    Q --> R[变量替换]
    R --> S[循环结构]
    R --> T[条件判断]

    S --> U[最终Prompt]
    T --> U

    U --> V[Token估算]
    V --> W[验证Prompt长度]
    W --> X[优化Prompt]

    X --> Y[最终Prompt输出]

    style A fill:#e1f5ff
    style Y fill:#4caf50
```

## 5.9 上下文管理架构图

```mermaid
graph TB
    A[上下文管理] --> B[对话消息管理]
    A --> C[上下文窗口管理]
    A --> D[检索结果格式化]
    A --> E[上下文一致性保证]

    B --> F[ConversationMessage]
    F --> G[Role: system/user/assistant]
    F --> H[Content: 消息内容]
    F --> I[Timestamp: 创建时间]

    C --> J[Token限制控制]
    J --> K[系统Prompt: 500-1000 tokens]
    J --> L[用户Prompt: 2000-3000 tokens]
    J --> M[历史对话: 动态调整]
    J --> N[检索结果: 按相关性排序]

    D --> O[FormatDocuments]
    D --> P[FormatChunks]
    D --> Q[FormatReimbursementInfo]

    E --> R[版本管理]
    E --> S[状态同步]
    E --> T[错误恢复]

    style A fill:#e1f5ff
    style B fill:#4caf50
    style C fill:#ff9800
```

## 5.10 智能审核流程总图

```mermaid
flowchart TD
    A[报销申请信息] --> B[审核问题生成]

    B --> C[信息提取]
    C --> D[提取报销类型、金额、分类等]
    D --> E[构建审核问题]

    E --> F[关键词提取]
    F --> G[提取多维度关键词]

    E --> H[知识检索]
    G --> H

    H --> I[查询向量化]
    I --> J[生成768维查询向量]

    J --> K[混合检索]
    K --> L[向量检索]
    K --> M[关键词检索]

    L --> N[向量相似度检索]
    N --> O[Top-K结果]

    M --> P[关键词模糊匹配]
    P --> Q[Top-K结果]

    O --> R[结果融合]
    Q --> R

    R --> S[结果去重]
    S --> T[分数融合]
    T --> U[结果排序]

    U --> V[检索到的制度文档]
    V --> W[格式化为文档格式]

    W --> X[大模型推理]
    X --> Y[Prompt构建]

    Y --> Z[构建系统Prompt]
    Y --> AA[构建用户Prompt]

    Z --> AB[合并上下文]
    AA --> AB

    AB --> AC[调用大模型]
    AC --> AD[Chat接口调用]
    AD --> AE[temperature=0.7, maxTokens=2000]

    AE --> AF[大模型响应]
    AF --> AG[响应验证]

    AG --> AH[结果解析]
    AH --> AI[JSON解析]
    AI --> AJ[文本解析]

    AJ --> AK[置信度计算]
    AK --> AL[基于检索结果质量]
    AK --> AM[基于响应长度]
    AK --> AN[基于关键词匹配]

    AL --> AO[审核结论生成]
    AM --> AO
    AN --> AO

    AO --> AP[结构化审核结果]
    AP --> AQ[结论、理由、建议]
    AQ --> AR[置信度评分]

    AR --> AS[最终审核结果]

    style A fill:#e1f5ff
    style H fill:#4caf50
    style X fill:#ff9800
    style AS fill:#4caf50
```

## 5.11 RAG系统整体架构图

```mermaid
graph TB
    subgraph "用户层"
        A[报销申请提交]
        B[审核查询]
    end

    subgraph "应用层"
        C[RAGService]
        D[审核问题生成]
        E[知识检索协调]
        F[大模型推理协调]
        G[结果解析与验证]
    end

    subgraph "服务层"
        H[DocumentProcessor]
        I[文档采集与预处理]
        J[文档分片]

        K[VectorStore]
        L[向量存储与检索]
        M[向量索引管理]
        N[相似度计算]
        O[混合检索]

        P[LLMClient]
        Q[聊天接口调用]
        R[Embedding接口调用]

        S[PromptBuilder]
        T[Prompt构建与管理]
        U[模板渲染]
        V[上下文管理]
    end

    subgraph "数据层"
        W[PostgreSQL数据库]
        X[PGVector扩展]
        Y[reimbursement_documents表]
        Z[向量索引]
    end

    subgraph "外部服务"
        AA[智谱AI Embedding API]
        AB[智谱AI Chat API]
    end

    A --> C
    B --> C

    C --> H
    C --> K
    C --> P
    C --> S

    H --> I
    I --> J
    J --> K

    K --> W
    L --> X
    M --> Y
    N --> Z

    P --> AA
    P --> AB

    S --> T
    T --> U
    U --> V

    C --> D
    C --> E
    C --> F
    C --> G

    E --> K
    F --> P
    G --> C

    style A fill:#e1f5ff
    style C fill:#4caf50
    style W fill:#ff9800
```

## 5.12 数据模型关系图

```mermaid
erDiagram
    DOCUMENT ||--o{ DOCUMENT_CHUNK : "包含"
    DOCUMENT_CHUNK {
        string ID PK
        string DocumentID FK
        int ChunkIndex
        string ChunkContent
        vector Embedding
        datetime CreatedAt
        datetime UpdatedAt
    }

    DOCUMENT {
        string ID PK
        string FileName
        string FileType
        string Category
        datetime CreatedAt
        datetime UpdatedAt
    }

    RAG_RESULT {
        string ID PK
        string Query
        string Documents
        string Prompt
        string Response
        ANALYSIS_RESULT AnalysisResult
        int ExecutionTime
        datetime CreatedAt
    }

    ANALYSIS_RESULT {
        string ID PK
        string Query
        string Conclusion
        string Reasoning
        bool Pass
        string[] Suggestions
        float Confidence
        datetime CreatedAt
    }

    LLM_RESPONSE {
        string ID PK
        string Content
        string Model
        int Tokens
        float Cost
        datetime CreatedAt
    }

    PROMPT {
        string ID PK
        string Name
        string Template
        string Content
        string Type
        map Variables
        int Tokens
        datetime CreatedAt
        datetime UpdatedAt
        string Version
        string[] Tags
    }

    DOCUMENT ||--o{ RAG_RESULT : "引用"
    PROMPT ||--|| RAG_RESULT : "使用"
    LLM_RESPONSE ||--|| RAG_RESULT : "来自"
    ANALYSIS_RESULT ||--|| RAG_RESULT : "包含"
```

## 5.13 向量检索性能对比图

```mermaid
graph LR
    A[检索方式] --> B[向量检索]
    A --> C[关键词检索]
    A --> D[混合检索]

    B --> E[优势]
    E --> F[语义理解]
    E --> G[泛化能力强]
    E --> H[适合模糊查询]

    B --> I[劣势]
    I --> J[精确匹配较弱]
    I --> K[计算复杂度高]

    C --> L[优势]
    L --> M[精确匹配]
    L --> N[响应快速]
    L --> O[可控性强]

    C --> P[劣势]
    P --> Q[语义理解有限]
    P --> R[同义词匹配差]

    D --> S[优势]
    S --> T[互补性强]
    S --> U[鲁棒性好]
    S --> V[灵活性高]

    style A fill:#e1f5ff
    style D fill:#4caf50
    style S fill:#ff9800
```

## 5.14 审核结果置信度计算模型

```mermaid
graph TD
    A[置信度计算] --> B[基础置信度: 0.5]

    B --> C{检索结果质量}
    C -->|平均分数 > 0.8| D[+0.2]
    C -->|平均分数 > 0.6| E[+0.1]
    C -->|平均分数 ≤ 0.6| F[不调整]

    B --> G{检索结果数量}
    G -->|≥ 3个| H[+0.1]
    G -->|≥ 1个| I[+0.05]
    G -->|0个| J[不调整]

    B --> K{响应长度}
    K -->|> 100字符| L[+0.1]
    K -->|≤ 100字符| M[不调整]

    B --> N{关键词匹配}
    N -->|包含审核结论| O[+0.05]
    N -->|不包含| P[不调整]

    D --> Q[置信度调整]
    E --> Q
    F --> Q
    H --> Q
    I --> Q
    J --> Q
    L --> Q
    M --> Q
    O --> Q
    P --> Q

    Q --> R[最终置信度]
    R --> S{置信度 > 1.0}
    S -->|是| T[限制为1.0]
    S -->|否| U[保持原值]

    T --> V[最终置信度值]
    U --> V

    style A fill:#e1f5ff
    style V fill:#4caf50
```

## 5.15 智能审核时序图

```mermaid
sequenceDiagram
    participant User as 用户
    participant RAG as RAG服务
    participant VS as 向量存储
    participant LLM as 大模型
    participant DB as 数据库

    User->>RAG: 提交报销申请
    RAG->>RAG: 生成审核问题
    RAG->>RAG: 提取关键词

    RAG->>LLM: 生成查询向量
    LLM-->>RAG: 返回768维向量

    RAG->>VS: 向量检索
    RAG->>VS: 关键词检索

    VS-->>RAG: 向量检索结果
    VS-->>RAG: 关键词检索结果

    RAG->>RAG: 结果融合
    RAG->>RAG: 结果排序

    RAG->>RAG: 构建Prompt
    RAG->>RAG: 格式化文档
    RAG->>RAG: 格式化报销信息

    RAG->>LLM: 调用Chat接口
    Note over RAG,LLM: temperature=0.7<br/>maxTokens=2000

    LLM-->>RAG: 返回审核响应
    RAG->>RAG: 验证响应格式
    RAG->>RAG: 解析审核结果

    RAG->>RAG: 计算置信度
    RAG->>RAG: 生成结构化结果

    RAG->>DB: 存储审核记录
    RAG-->>User: 返回审核结论

    Note over User,User: 审核完成<br/>用时: < 2秒
```

## 图表说明

本章节为第五章"RAG智能审核设计与实现"提供了15个关键图表，涵盖以下方面：

1. **知识库构建流程图** - 展示从原始文档到知识库的完整处理流程
2. **文档分片策略示意图** - 可视化滑动窗口分片算法
3. **向量存储架构图** - 展示向量存储的数据模型和索引结构
4. **向量索引结构图** - 对比IVFFlat和HNSW两种索引结构
5. **混合检索流程图** - 展示向量检索和关键词检索的融合过程
6. **相似度计算流程图** - 详细展示余弦相似度计算步骤
7. **LLM客户端架构图** - 展示LLM客户端的组件结构
8. **Prompt构建流程图** - 展示从模板到最终Prompt的构建过程
9. **上下文管理架构图** - 展示上下文管理的各个组件
10. **智能审核流程总图** - 展示完整的智能审核业务流程
11. **RAG系统整体架构图** - 展示系统的分层架构和组件关系
12. **数据模型关系图** - 展示核心数据模型之间的关系
13. **向量检索性能对比图** - 对比不同检索方式的优劣势
14. **审核结果置信度计算模型** - 展示置信度计算的决策树
15. **智能审核时序图** - 展示审核流程的时序交互

这些图表使用Mermaid语法编写，可以在支持Mermaid的Markdown编辑器中渲染，如GitHub、GitLab、Typora等工具。
