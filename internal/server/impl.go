package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"reimbursement-audit/internal/api/handler"
	"reimbursement-audit/internal/api/middleware"
	storage "reimbursement-audit/internal/infra/storage/file"
	mysqlRepo "reimbursement-audit/internal/infra/storage/mysql"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// serverImpl 服务器实现
type serverImpl struct {
	config *Config
	engine *gin.Engine
	server *http.Server
}

// Start 启动服务器
func (s *serverImpl) Start() error {
	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         s.config.GetAddress(),
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// 启动服务器
	if s.config.IsTLS() {
		return s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
	}
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *serverImpl) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

// GetEngine 获取Gin引擎
func (s *serverImpl) GetEngine() *gin.Engine {
	return s.engine
}

// GetConfig 获取服务器配置
func (s *serverImpl) GetConfig() *Config {
	return s.config
}

// SetConfig 设置服务器配置
func (s *serverImpl) SetConfig(config *Config) {
	s.config = config

	// 更新Gin模式
	gin.SetMode(config.Mode)
}

// RegisterRoutes 注册路由
func (s *serverImpl) RegisterRoutes() {
	// 注册trace中间件，用于生成和传播traceId
	s.engine.Use(middleware.TraceMiddleware())

	// 创建日志记录器
	// TODO: 从配置中获取日志配置
	loggerImpl, err := logger.NewLogger(logger.DefaultConfig())
	if err != nil {
		panic(fmt.Sprintf("创建日志记录器失败: %v", err))
	}

	// 注册日志中间件，用于将带有traceId的logger注入到Gin上下文中
	s.engine.Use(middleware.LoggerMiddleware(loggerImpl))

	// 注册健康检查路由
	s.engine.GET("/health", HealthCheck)
	s.engine.GET("/ready", ReadyCheck)
	s.engine.GET("/version", VersionCheck("1.0.0"))

	// 创建MySQL客户端（实际应该从依赖注入获取）
	mysqlClient := mysqlRepo.NewClient()
	// TODO: 这里应该从配置中获取数据库连接信息
	// mysqlClient.Connect(ctx, config)

	// 创建文件存储服务
	// TODO: 从配置中获取存储路径和URL
	localStorage := storage.NewLocalStorage("./uploads", "http://localhost:8080/uploads")
	fileService := storage.NewService(localStorage)

	// 创建上传处理器
	uploadHandler := handler.NewUploadHandler(mysqlClient, fileService)

	// 注册上传相关路由
	s.engine.POST("/api/v1/reimbursement/upload", uploadHandler.UploadReimbursement)
	s.engine.POST("/api/v1/invoices/upload", uploadHandler.UploadInvoices)
	s.engine.POST("/api/v1/invoices/batch-upload", uploadHandler.BatchUpload)

	// TODO: 注册其他路由
	// s.engine.POST("/api/v1/audit", auditHandler)
	// s.engine.GET("/api/v1/query", queryHandler)
	// s.engine.POST("/api/v1/rules", createRuleHandler)
	// s.engine.PUT("/api/v1/rules/:id", updateRuleHandler)
	// s.engine.DELETE("/api/v1/rules/:id", deleteRuleHandler)
	// s.engine.GET("/api/v1/rules", listRulesHandler)
}

// SetupMiddleware 设置中间件
func (s *serverImpl) SetupMiddleware(middlewares ...gin.HandlerFunc) {
	for _, middleware := range middlewares {
		s.engine.Use(middleware)
	}
}

// SetupStaticFiles 设置静态文件服务
func (s *serverImpl) SetupStaticFiles(relativePath, root string) {
	s.engine.Static(relativePath, root)
}

// SetupStaticFS 设置静态文件系统
func (s *serverImpl) SetupStaticFS(relativePath string, fs http.FileSystem) {
	s.engine.StaticFS(relativePath, fs)
}

// SetupTemplate 设置模板
func (s *serverImpl) SetupTemplate(pattern string, obj interface{}) {
	s.engine.LoadHTMLGlob(pattern)
}

// AddRoute 添加路由
func (s *serverImpl) AddRoute(method, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		s.engine.GET(path, handler)
	case "POST":
		s.engine.POST(path, handler)
	case "PUT":
		s.engine.PUT(path, handler)
	case "DELETE":
		s.engine.DELETE(path, handler)
	case "PATCH":
		s.engine.PATCH(path, handler)
	case "HEAD":
		s.engine.HEAD(path, handler)
	case "OPTIONS":
		s.engine.OPTIONS(path, handler)
	default:
		panic(fmt.Sprintf("unsupported HTTP method: %s", method))
	}
}

// AddGroup 添加路由组
func (s *serverImpl) AddGroup(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return s.engine.Group(relativePath, handlers...)
}

// NoRoute 设置404处理
func (s *serverImpl) NoRoute(handlers ...gin.HandlerFunc) {
	s.engine.NoRoute(handlers...)
}

// NoMethod 设置405处理
func (s *serverImpl) NoMethod(handlers ...gin.HandlerFunc) {
	s.engine.NoMethod(handlers...)
}

// RunGraceful 优雅运行服务器
func (s *serverImpl) RunGraceful() error {
	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         s.config.GetAddress(),
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// 启动服务器
	go func() {
		var err error
		if s.config.IsTLS() {
			err = s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// 等待中断信号
	// TODO: 实现信号处理
	<-time.After(time.Hour) // 临时使用，实际应该等待信号

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
