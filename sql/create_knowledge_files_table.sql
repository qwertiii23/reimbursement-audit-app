-- 创建知识库文件表
CREATE TABLE IF NOT EXISTS knowledge_files (
  id VARCHAR(36) PRIMARY KEY COMMENT '文件ID',
  file_name VARCHAR(255) NOT NULL COMMENT '文件名',
  file_path VARCHAR(500) NOT NULL COMMENT '文件路径',
  file_type VARCHAR(50) NOT NULL COMMENT '文件类型',
  file_size BIGINT COMMENT '文件大小(字节)',
  category VARCHAR(100) COMMENT '文件分类',
  description TEXT COMMENT '文件描述',
  uploaded_by VARCHAR(36) COMMENT '上传者ID',
  uploader_name VARCHAR(100) COMMENT '上传者姓名',
  download_count INT DEFAULT 0 COMMENT '下载次数',
  status VARCHAR(20) DEFAULT 'active' COMMENT '状态',
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  INDEX idx_category (category),
  INDEX idx_status (status),
  INDEX idx_uploaded_by (uploaded_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库文件表';
