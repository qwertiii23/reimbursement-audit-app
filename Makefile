# 报销审核系统 Makefile

# 变量定义
APP_NAME := reimbursement-audit
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD)

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 目录定义
BIN_DIR := bin
DIST_DIR := dist
COVERAGE_DIR := coverage

# Go相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Docker相关变量
DOCKER_IMAGE := $(APP_NAME)
DOCKER_REGISTRY := your-registry.com
DOCKER_TAG := $(VERSION)

# Kubernetes相关变量
K8S_NAMESPACE := reimbursement-audit
K8S_DEPLOY_DIR := deploy/k8s

# 默认目标
.PHONY: help
help: ## 显示帮助信息
	@echo "报销审核系统 Makefile"
	@echo ""
	@echo "用法:"
	@echo "  make [命令]"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 依赖管理
.PHONY: deps
deps: ## 安装依赖
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: update-deps
update-deps: ## 更新依赖
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

# 代码质量
.PHONY: fmt
fmt: ## 格式化代码
	$(GOCMD) fmt ./...

.PHONY: vet
vet: ## 代码检查
	$(GOCMD) vet ./...

.PHONY: lint
lint: ## 代码检查 (需要安装golangci-lint)
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
	fi

# 测试
.PHONY: test
test: ## 运行测试
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "覆盖率报告已生成: $(COVERAGE_DIR)/coverage.html"

.PHONY: test-race
test-race: ## 运行竞态检测
	$(GOTEST) -race -v ./...

# 构建
.PHONY: build
build: ## 构建二进制文件
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/server cmd/server/main.go
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/migrate cmd/migrate/main.go

.PHONY: build-all
build-all: ## 构建所有平台的二进制文件
	@mkdir -p $(DIST_DIR)
	@echo "构建多平台版本..."
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 cmd/server/main.go
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-migrate-linux-amd64 cmd/migrate/main.go
	tar -czf $(DIST_DIR)/$(APP_NAME)-linux-amd64.tar.gz -C $(DIST_DIR) $(APP_NAME)-linux-amd64 $(APP_NAME)-migrate-linux-amd64
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-migrate-darwin-amd64 cmd/migrate/main.go
	tar -czf $(DIST_DIR)/$(APP_NAME)-darwin-amd64.tar.gz -C $(DIST_DIR) $(APP_NAME)-darwin-amd64 $(APP_NAME)-migrate-darwin-amd64
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe cmd/server/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-migrate-windows-amd64.exe cmd/migrate/main.go
	zip -j $(DIST_DIR)/$(APP_NAME)-windows-amd64.zip $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(DIST_DIR)/$(APP_NAME)-migrate-windows-amd64.exe

.PHONY: clean
clean: ## 清理构建产物
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf $(COVERAGE_DIR)

# 运行
.PHONY: run
run: ## 运行应用
	$(GOCMD) run cmd/server/main.go

.PHONY: run-migrate
run-migrate: ## 运行迁移工具
	$(GOCMD) run cmd/migrate/main.go

# Docker
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

.PHONY: docker-push
docker-push: ## 推送Docker镜像
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest

# Kubernetes
.PHONY: k8s-namespace
k8s-namespace: ## 创建Kubernetes命名空间
	kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

.PHONY: k8s-deploy
k8s-deploy: ## 部署到Kubernetes
	@echo "部署到Kubernetes..."
	envsubst < $(K8S_DEPLOY_DIR)/configmap.yaml | kubectl apply -n $(K8S_NAMESPACE) -f -
	envsubst < $(K8S_DEPLOY_DIR)/secret.yaml | kubectl apply -n $(K8S_NAMESPACE) -f -
	envsubst < $(K8S_DEPLOY_DIR)/postgres.yaml | kubectl apply -n $(K8S_NAMESPACE) -f -
	envsubst < $(K8S_DEPLOY_DIR)/deployment.yaml | kubectl apply -n $(K8S_NAMESPACE) -f -
	envsubst < $(K8S_DEPLOY_DIR)/service.yaml | kubectl apply -n $(K8S_NAMESPACE) -f -
	envsubst < $(K8S_DEPLOY_DIR)/ingress.yaml | kubectl apply -n $(K8S_NAMESPACE) -f -

.PHONY: k8s-status
k8s-status: ## 查看Kubernetes部署状态
	kubectl get pods -n $(K8S_NAMESPACE)
	kubectl get services -n $(K8S_NAMESPACE)
	kubectl get ingress -n $(K8S_NAMESPACE)

.PHONY: k8s-logs
k8s-logs: ## 查看Kubernetes日志
	kubectl logs -f deployment/$(APP_NAME) -n $(K8S_NAMESPACE)

.PHONY: k8s-shell
k8s-shell: ## 进入Kubernetes容器
	kubectl exec -it deployment/$(APP_NAME) -n $(K8S_NAMESPACE) -- /bin/sh

.PHONY: k8s-cleanup
k8s-cleanup: ## 清理Kubernetes部署
	kubectl delete -n $(K8S_NAMESPACE) -f $(K8S_DEPLOY_DIR)/ --ignore-not-found=true
	kubectl delete namespace $(K8S_NAMESPACE) --ignore-not-found=true

# 开发工具
.PHONY: dev-setup
dev-setup: ## 设置开发环境
	@echo "设置开发环境..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/air-verse/air@latest
	@echo "开发环境设置完成"

.PHONY: dev-run
dev-run: ## 使用Air热重载运行应用
	@if command -v air &> /dev/null; then \
		air; \
	else \
		echo "Air 未安装，请运行 make dev-setup"; \
	fi

# 生成工具
.PHONY: generate
generate: ## 生成代码
	$(GOCMD) generate ./...

.PHONY: mocks
mocks: ## 生成Mock文件
	@if command -v mockgen &> /dev/null; then \
		mockgen -source=internal/domain/reimbursement/service.go -destination=internal/domain/reimbursement/service_mock.go -package=reimbursement; \
	else \
		echo "mockgen 未安装，请运行: go install github.com/golang/mock/mockgen@latest"; \
	fi

# 完整流程
.PHONY: ci
ci: fmt vet test lint ## CI流程

.PHONY: release
release: clean fmt vet test lint build-all docker-build docker-push ## 发布流程

# 默认目标
.DEFAULT_GOAL := help