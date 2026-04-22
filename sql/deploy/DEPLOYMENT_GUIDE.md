# 智能报销审核系统 - 本地部署指南

## 一、环境要求

### 1.1 硬件要求
- CPU: 2核及以上
- 内存: 4GB及以上
- 磁盘: 20GB及以上

### 1.2 软件要求
- **操作系统**: macOS 10.14+ / Linux (Ubuntu 18.04+) / Windows 10+
- **Go**: 1.21+
- **MySQL**: 5.7+ 或 8.0+
- **PostgreSQL**: 13+ (用于RAG向量存储，可选)
- **Git**: 最新版本

### 1.3 可选组件
- **MinIO**: 用于存储发票图片（也可使用本地文件系统）
- **Redis**: 用于缓存（可选）

---

## 二、数据库配置

### 2.1 MySQL 数据库初始化

#### 步骤1: 登录MySQL
```bash
mysql -u root -p
```

#### 步骤2: 创建数据库
```sql
CREATE DATABASE IF NOT EXISTS reimbursement_audit CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE reimbursement_audit;
```

#### 步骤3: 执行建表脚本
```bash
# 在项目根目录下执行
mysql -u root -p reimbursement_audit < sql/deploy/mysql_schema.sql
```

### 2.2 PostgreSQL 数据库初始化（可选，用于RAG向量存储）

PostgreSQL主要用于RAG智能审核的向量存储。如果不需要RAG功能，可以跳过此步骤。

#### 步骤1: 登录PostgreSQL
```bash
psql -U postgres
```

#### 步骤2: 创建数据库
```sql
CREATE DATABASE reimbursement_audit;
```

#### 步骤3: 启用pgvector扩展
```sql
\c reimbursement_audit
CREATE EXTENSION IF NOT EXISTS vector;
```

#### 步骤4: 执行建表脚本
```bash
psql -U postgres -d reimbursement_audit -f sql/deploy/postgres_schema.sql
```

---

## 三、应用配置

### 3.1 配置文件结构
项目提供多环境配置文件，在 `configs/` 目录下：

```
configs/
├── config.dev.yaml      # 开发环境
├── config.prod.yaml     # 生产环境
└── config.test.yaml     # 测试环境
```

### 3.2 修改数据库连接配置

编辑 `configs/config.dev.yaml`：

```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "your_password"
  name: "reimbursement_audit"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600

postgres:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "your_password"
  name: "reimbursement_audit"
  sslmode: "disable"

app:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

jwt:
  secret: "your-jwt-secret-key-change-in-production"
  expire_hours: 24

storage:
  type: "local"  # local 或 minio
  local:
    path: "./uploads"
  minio:
    endpoint: "localhost:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    bucket: "invoices"
    use_ssl: false

ocr:
  provider: "tencent"
  tencent:
    secret_id: "your-tencent-secret-id"
    secret_key: "your-tencent-secret-key"
    region: "ap-guangzhou"

llm:
  provider: "zhipu"
  zhipu:
    api_key: "your-zhipu-api-key"

rag:
  vector_store: "chroma"  # chroma 或 postgres
  chroma:
    host: "localhost"
    port: 8000
```

---

## 四、项目构建与运行

### 4.1 获取源代码
```bash
git clone <repository-url>
cd reimbursement-audit
```

### 4.2 安装依赖
```bash
go mod download
```

### 4.3 编译项目
```bash
# 编译所有包
go build ./...

# 或使用Makefile
make build
```

### 4.4 运行服务
```bash
# 开发环境运行
go run cmd/server/main.go -config configs/config.dev.yaml

# 或使用编译后的二进制文件
./bin/server -config configs/config.dev.yaml
```

---

## 五、Docker部署（可选）

### 5.1 构建Docker镜像
```bash
docker build -t reimbursement-audit:latest .
```

### 5.2 使用Docker Compose运行

创建 `docker-compose.yml`：
```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: reimbursement_audit
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./sql/deploy/mysql_schema.sql:/docker-entrypoint-initdb.d/1_schema.sql

  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_PASSWORD: postgres_password
      POSTGRES_DB: reimbursement_audit
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  app:
    image: reimbursement-audit:latest
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - postgres
    volumes:
      - ./configs/config.prod.yaml:/app/config.yaml
      - ./uploads:/app/uploads

volumes:
  mysql_data:
  postgres_data:
```

运行：
```bash
docker-compose up -d
```

---

## 六、初始化数据

### 6.1 插入初始用户
```sql
-- 管理员账户 (密码: admin123)
INSERT INTO users (id, username, password, email, real_name, role, status)
VALUES ('550e8400-e29b-41d4-a716-446655440001', 'admin', '$2a$10$Xj9kT8eB8NnRJKpLz1xPQe3V8jF3KqXQvN5JhPf1T8Xw7S5K6eG.', 'admin@example.com', '系统管理员', 'admin', 'active');

-- 财务账户 (密码: finance123)
INSERT INTO users (id, username, password, email, real_name, role, status)
VALUES ('550e8400-e29b-41d4-a716-446655440002', 'finance', '$2a$10$Xj9kT8eB8NnRJKpLz1xPQe3V8jF3KqXQvN5JhPf1T8Xw7S5K6eG.', 'finance@example.com', '财务人员', 'finance', 'active');
```

### 6.2 插入特征数据
```bash
mysql -u root -p reimbursement_audit < sql/insert_features.sql
```

### 6.3 插入规则数据
```bash
# 依次执行各规则SQL文件
mysql -u root -p reimbursement_audit < sql/insert_transportation_rule.sql
mysql -u root -p reimbursement_audit < sql/insert_meal_rule.sql
mysql -u root -p reimbursement_audit < sql/insert_overtime_rule.sql
# ... 其他规则文件
```

---

## 七、验证部署

### 7.1 检查服务是否启动
```bash
curl http://localhost:8080/health
```

### 7.2 登录测试
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 7.3 前端访问
在浏览器中打开：`http://localhost:8080`

---

## 八、常见问题

### 8.1 数据库连接失败
- 检查MySQL服务是否启动
- 检查配置文件中的数据库连接信息是否正确
- 检查防火墙是否开放了数据库端口

### 8.2 OCR功能不可用
- 检查是否配置了腾讯云OCR的SecretID和SecretKey
- 检查网络是否可以访问腾讯云OCR服务

### 8.3 RAG功能不可用
- 如果使用Chroma，确认Chroma服务是否启动
- 如果使用PostgreSQL，确认pgvector扩展是否正确安装

### 8.4 前端无法访问后端API
- 检查CORS配置
- 检查API地址是否正确配置在前端

---

## 九、数据备份

### 9.1 MySQL备份
```bash
mysqldump -u root -p reimbursement_audit > backup_$(date +%Y%m%d).sql
```

### 9.2 PostgreSQL备份
```bash
pg_dump -U postgres reimbursement_audit > backup_$(date +%Y%m%d).sql
```

---

## 十、卸载

### 10.1 停止服务
```bash
pkill -f "reimbursement-audit"
```

### 10.2 删除数据库
```sql
DROP DATABASE reimbursement_audit;
```

### 10.3 删除上传文件
```bash
rm -rf ./uploads
```

---

## 十一、联系方式

如有问题，请联系系统管理员或提交Issue。
