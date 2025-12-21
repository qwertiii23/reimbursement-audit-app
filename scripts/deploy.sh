#!/bin/bash

# 报销审核系统部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 默认配置
ENV=${ENV:-dev}
VERSION=${VERSION:-latest}
REGISTRY=${REGISTRY:-"your-registry.com"}
NAMESPACE=${NAMESPACE:-"reimbursement-audit"}

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
    log_info "检查部署依赖..."
    check_command kubectl
    check_command helm
    
    # 检查kubectl连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到Kubernetes集群"
        exit 1
    fi
    
    log_info "依赖检查完成"
}

# 创建命名空间
create_namespace() {
    log_info "创建命名空间: $NAMESPACE"
    
    kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "命名空间创建完成"
}

# 部署配置
deploy_config() {
    log_info "部署配置..."
    
    # 部署ConfigMap
    envsubst < deploy/k8s/configmap.yaml | kubectl apply -n $NAMESPACE -f -
    
    # 部署Secret
    envsubst < deploy/k8s/secret.yaml | kubectl apply -n $NAMESPACE -f -
    
    log_info "配置部署完成"
}

# 部署数据库
deploy_database() {
    log_info "部署数据库..."
    
    # 部署PostgreSQL
    envsubst < deploy/k8s/postgres.yaml | kubectl apply -n $NAMESPACE -f -
    
    # 等待数据库就绪
    kubectl wait --for=condition=ready pod -l app=postgres -n $NAMESPACE --timeout=300s
    
    # 运行数据库迁移
    kubectl run migrate \
        --image=${REGISTRY}/reimbursement-audit-migrate:${VERSION} \
        --restart=Never \
        --namespace=$NAMESPACE \
        -- \
        -action=up -config=/etc/config/config.yaml
    
    # 等待迁移完成
    kubectl wait --for=condition=complete job/migrate -n $NAMESPACE --timeout=300s
    
    # 清理迁移Job
    kubectl delete job migrate -n $NAMESPACE --ignore-not-found=true
    
    log_info "数据库部署完成"
}

# 部署应用
deploy_app() {
    log_info "部署应用..."
    
    # 部署应用
    envsubst < deploy/k8s/deployment.yaml | kubectl apply -n $NAMESPACE -f -
    
    # 部署服务
    envsubst < deploy/k8s/service.yaml | kubectl apply -n $NAMESPACE -f -
    
    # 部署Ingress
    envsubst < deploy/k8s/ingress.yaml | kubectl apply -n $NAMESPACE -f -
    
    # 等待应用就绪
    kubectl wait --for=condition=available deployment/reimbursement-audit -n $NAMESPACE --timeout=300s
    
    log_info "应用部署完成"
}

# 部署监控
deploy_monitoring() {
    log_info "部署监控..."
    
    # 部署ServiceMonitor
    envsubst < deploy/k8s/monitoring.yaml | kubectl apply -n $NAMESPACE -f -
    
    log_info "监控部署完成"
}

# 回滚部署
rollback() {
    log_info "回滚部署..."
    
    # 回滚应用
    kubectl rollout undo deployment/reimbursement-audit -n $NAMESPACE
    
    # 等待回滚完成
    kubectl rollout status deployment/reimbursement-audit -n $NAMESPACE --timeout=300s
    
    log_info "回滚完成"
}

# 查看部署状态
status() {
    log_info "查看部署状态..."
    
    # 查看Pod状态
    kubectl get pods -n $NAMESPACE
    
    # 查看服务状态
    kubectl get services -n $NAMESPACE
    
    # 查看Ingress状态
    kubectl get ingress -n $NAMESPACE
    
    log_info "部署状态查看完成"
}

# 查看日志
logs() {
    log_info "查看应用日志..."
    
    kubectl logs -f deployment/reimbursement-audit -n $NAMESPACE
}

# 进入容器
shell() {
    log_info "进入应用容器..."
    
    kubectl exec -it deployment/reimbursement-audit -n $NAMESPACE -- /bin/sh
}

# 清理部署
cleanup() {
    log_info "清理部署..."
    
    # 删除所有资源
    kubectl delete -n $NAMESPACE -f deploy/k8s/ --ignore-not-found=true
    
    # 删除命名空间
    kubectl delete namespace $NAMESPACE --ignore-not-found=true
    
    log_info "清理完成"
}

# 显示帮助信息
show_help() {
    echo "报销审核系统部署脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "环境变量:"
    echo "  ENV         部署环境 (dev/staging/prod) (默认: dev)"
    echo "  VERSION     部署版本 (默认: latest)"
    echo "  REGISTRY    镜像仓库 (默认: your-registry.com)"
    echo "  NAMESPACE   命名空间 (默认: reimbursement-audit)"
    echo ""
    echo "命令:"
    echo "  check        检查依赖"
    echo "  namespace    创建命名空间"
    echo "  config       部署配置"
    echo "  database     部署数据库"
    echo "  app          部署应用"
    echo "  monitoring   部署监控"
    echo "  deploy       完整部署 (config + database + app + monitoring)"
    echo "  rollback     回滚部署"
    echo "  status       查看部署状态"
    echo "  logs         查看应用日志"
    echo "  shell        进入应用容器"
    echo "  cleanup      清理部署"
    echo "  help         显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 deploy"
    echo "  $0 app"
    echo "  $0 status"
    echo "  ENV=prod VERSION=v1.0.0 $0 deploy"
}

# 主函数
main() {
    case "${1:-help}" in
        check)
            check_dependencies
            ;;
        namespace)
            create_namespace
            ;;
        config)
            create_namespace
            deploy_config
            ;;
        database)
            create_namespace
            deploy_database
            ;;
        app)
            create_namespace
            deploy_app
            ;;
        monitoring)
            create_namespace
            deploy_monitoring
            ;;
        deploy)
            check_dependencies
            create_namespace
            deploy_config
            deploy_database
            deploy_app
            deploy_monitoring
            ;;
        rollback)
            rollback
            ;;
        status)
            status
            ;;
        logs)
            logs
            ;;
        shell)
            shell
            ;;
        cleanup)
            cleanup
            ;;
        help|*)
            show_help
            ;;
    esac
}

# 执行主函数
main "$@"