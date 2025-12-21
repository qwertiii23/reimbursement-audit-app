#!/bin/bash

# 报销审核系统构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 命令不存在，请先安装"
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    log_info "检查构建依赖..."
    check_command go
    check_command git
    log_info "依赖检查完成"
}

# 清理构建产物
clean() {
    log_info "清理构建产物..."
    rm -rf bin/
    rm -rf dist/
    log_info "清理完成"
}

# 代码格式化
format() {
    log_info "格式化代码..."
    go fmt ./...
    log_info "代码格式化完成"
}

# 代码检查
lint() {
    log_info "代码检查..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
    else
        log_warn "golangci-lint 未安装，跳过代码检查"
    fi
    log_info "代码检查完成"
}

# 运行测试
test() {
    log_info "运行测试..."
    go test -v ./...
    log_info "测试完成"
}

# 运行测试并生成覆盖率报告
test_coverage() {
    log_info "运行测试并生成覆盖率报告..."
    mkdir -p coverage
    go test -v -coverprofile=coverage/coverage.out ./...
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    log_info "测试完成，覆盖率报告已生成: coverage/coverage.html"
}

# 构建二进制文件
build() {
    log_info "构建二进制文件..."
    mkdir -p bin
    
    # 构建服务器
    log_info "构建服务器..."
    go build -o bin/server cmd/server/main.go
    
    # 构建迁移工具
    log_info "构建迁移工具..."
    go build -o bin/migrate cmd/migrate/main.go
    
    log_info "构建完成"
}

# 构建发布版本
build_release() {
    log_info "构建发布版本..."
    mkdir -p dist
    
    # 获取版本信息
    VERSION=${VERSION:-$(git describe --tags --always --dirty)}
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse HEAD)
    
    # 构建多平台版本
    platforms=("linux/amd64" "darwin/amd64" "windows/amd64")
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -ra ADDR <<< "$platform"
        GOOS=${ADDR[0]}
        GOARCH=${ADDR[1]}
        
        output_name="reimbursement-audit-${GOOS}-${GOARCH}"
        if [ $GOOS = "windows" ]; then
            output_name+='.exe'
        fi
        
        log_info "构建 $output_name..."
        
        # 构建服务器
        env GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
            -o dist/${output_name} cmd/server/main.go
            
        # 构建迁移工具
        migrate_name="reimbursement-audit-migrate-${GOOS}-${GOARCH}"
        if [ $GOOS = "windows" ]; then
            migrate_name+='.exe'
        fi
        
        env GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
            -o dist/${migrate_name} cmd/migrate/main.go
            
        # 压缩
        if [ $GOOS = "windows" ]; then
            zip -j dist/${output_name}.zip dist/${output_name} dist/${migrate_name}
        else
            tar -czf dist/${output_name}.tar.gz -C dist ${output_name} ${migrate_name}
        fi
    done
    
    log_info "发布版本构建完成"
}

# Docker构建
docker_build() {
    log_info "构建Docker镜像..."
    
    # 获取版本信息
    VERSION=${VERSION:-$(git describe --tags --always --dirty)}
    
    docker build -t reimbursement-audit:${VERSION} .
    docker tag reimbursement-audit:${VERSION} reimbursement-audit:latest
    
    log_info "Docker镜像构建完成"
}

# Docker推送
docker_push() {
    log_info "推送Docker镜像..."
    
    # 获取版本信息
    VERSION=${VERSION:-$(git describe --tags --always --dirty)}
    REGISTRY=${REGISTRY:-"your-registry.com"}
    
    docker tag reimbursement-audit:${VERSION} ${REGISTRY}/reimbursement-audit:${VERSION}
    docker tag reimbursement-audit:latest ${REGISTRY}/reimbursement-audit:latest
    
    docker push ${REGISTRY}/reimbursement-audit:${VERSION}
    docker push ${REGISTRY}/reimbursement-audit:latest
    
    log_info "Docker镜像推送完成"
}

# 显示帮助信息
show_help() {
    echo "报销审核系统构建脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  check          检查依赖"
    echo "  clean          清理构建产物"
    echo "  format         格式化代码"
    echo "  lint           代码检查"
    echo "  test           运行测试"
    echo "  test-coverage  运行测试并生成覆盖率报告"
    echo "  build          构建二进制文件"
    echo "  build-release  构建发布版本"
    echo "  docker-build   构建Docker镜像"
    echo "  docker-push    推送Docker镜像"
    echo "  help           显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 build"
    echo "  $0 build-release"
    echo "  $0 docker-build"
}

# 主函数
main() {
    case "${1:-help}" in
        check)
            check_dependencies
            ;;
        clean)
            clean
            ;;
        format)
            format
            ;;
        lint)
            lint
            ;;
        test)
            test
            ;;
        test-coverage)
            test_coverage
            ;;
        build)
            check_dependencies
            clean
            format
            lint
            test
            build
            ;;
        build-release)
            check_dependencies
            clean
            format
            lint
            test
            build_release
            ;;
        docker-build)
            docker_build
            ;;
        docker-push)
            docker_push
            ;;
        help|*)
            show_help
            ;;
    esac
}

# 执行主函数
main "$@"