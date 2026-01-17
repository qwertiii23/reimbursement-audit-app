package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"reimbursement-audit/internal/api/handler"
	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/application/service"
	"reimbursement-audit/internal/config"
	"reimbursement-audit/internal/domain/audit"
	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/domain/ocr/provider"
	"reimbursement-audit/internal/domain/rag"
	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/domain/rule"
	"reimbursement-audit/internal/domain/user"
	storage "reimbursement-audit/internal/infra/storage/file"
	mysqlRepo "reimbursement-audit/internal/infra/storage/mysql"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// serverImpl 服务器实现
type serverImpl struct {
	config    *Config
	appConfig *config.Config
	engine    *gin.Engine
	server    *http.Server
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

// SetAppConfig 设置应用配置
func (s *serverImpl) SetAppConfig(appConfig *config.Config) {
	s.appConfig = appConfig
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

	// 创建logger实例
	loggerInstance, _ := logger.NewLogger(logger.DefaultConfig())

	// 注册健康检查路由
	s.engine.GET("/health", HealthCheck)
	s.engine.GET("/ready", ReadyCheck)
	s.engine.GET("/version", VersionCheck("1.0.0"))

	// 调试：检查配置是否正确加载
	loggerInstance.Info("数据库配置信息",
		logger.NewField("host", s.appConfig.Database.Host),
		logger.NewField("port", s.appConfig.Database.Port),
		logger.NewField("username", s.appConfig.Database.Username),
		logger.NewField("password", s.appConfig.Database.Password),
		logger.NewField("dbname", s.appConfig.Database.DBName))

	// 创建MySQL客户端（实际应该从依赖注入获取）
	mysqlClient := mysqlRepo.NewClient(loggerInstance)

	// 从配置中获取数据库连接信息并连接
	mysqlConfig := &mysqlRepo.Config{
		Host:            s.appConfig.Database.Host,
		Port:            s.appConfig.Database.Port,
		Username:        s.appConfig.Database.Username,
		Password:        s.appConfig.Database.Password,
		DBName:          s.appConfig.Database.DBName,
		Charset:         "utf8mb4",
		Collation:       "utf8mb4_unicode_ci",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    s.appConfig.Database.MaxOpenConns,
		MaxIdleConns:    s.appConfig.Database.MaxIdleConns,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		EnableLog:       true,
		LogLevel:        "info",
		SlowThreshold:   200,
		MaxRetries:      3,
		RetryDelay:      1000,
	}

	ctx := context.Background()
	if err := mysqlClient.Connect(ctx, mysqlConfig); err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	// 创建文件存储服务
	// TODO: 从配置中获取存储路径和URL
	localStorage := storage.NewLocalStorage("./uploads", "http://localhost:8080/uploads")
	fileService := storage.NewService(localStorage)

	// 创建OCR服务
	// 从配置中获取OCR配置
	var ocrConfig ocr.Config
	if s.appConfig != nil && s.appConfig.OCR.Provider != "" {
		ocrConfig = ocr.Config{
			SecretID:   s.appConfig.OCR.SecretID,
			SecretKey:  s.appConfig.OCR.SecretKey,
			Region:     s.appConfig.OCR.Region,
			Timeout:    s.appConfig.OCR.Timeout,
			MaxRetries: s.appConfig.OCR.MaxRetries,
		}
	} else {
		// 使用默认配置
		ocrConfig = ocr.Config{
			SecretID:   "", // 需要从环境变量或配置文件中获取
			SecretKey:  "", // 需要从环境变量或配置文件中获取
			Region:     "ap-beijing",
			Timeout:    30,
			MaxRetries: 3,
		}
	}
	ocrProvider := provider.NewTencentProvider(ocrConfig, loggerInstance)

	reimbursementRepo := mysqlRepo.NewReimbursementRepository(mysqlClient, loggerInstance)

	ocrRepo := mysqlRepo.NewOCRRepository(mysqlClient, loggerInstance)

	// 创建领域服务
	reimbursementDomainService := reimbursement.NewDomainService(reimbursementRepo, loggerInstance)
	ocrDomainService := ocr.NewParserService(ocrProvider, ocrRepo, loggerInstance, "./uploads")

	// 创建应用服务
	reimbursementAppService := service.NewReimbursementApplicationService(
		reimbursementRepo,
		reimbursementDomainService,
		ocrDomainService,
		ocrRepo,
		fileService,
		loggerInstance,
	)

	// 创建规则仓储和规则引擎
	ruleRepo := mysqlRepo.NewRuleRepository(mysqlClient, loggerInstance)
	ruleEngine := rule.NewGRuleEngine(ruleRepo, loggerInstance)

	// 初始化规则引擎
	if err := ruleEngine.Initialize(ctx); err != nil {
		loggerInstance.Error("规则引擎初始化失败", logger.NewField("error", err.Error()))
	}

	ruleService := rule.NewRuleService(ruleRepo, loggerInstance, ruleEngine)

	// 创建RAG服务
	llmClient := rag.NewLLMClientWithEmbedding(
		s.appConfig.RAG.APIKey,
		s.appConfig.RAG.APIBase,
		s.appConfig.RAG.Model,
		s.appConfig.LLM.Timeout,
		loggerInstance,
		s.appConfig.RAG.EmbeddingProvider,
		s.appConfig.RAG.EmbeddingModel,
		s.appConfig.RAG.EmbeddingAPIKey,
		s.appConfig.RAG.EmbeddingAPIBase,
		s.appConfig.RAG.EmbeddingDimension,
	)

	documentProcessor := rag.NewDocumentProcessor(500, 50, loggerInstance)
	promptBuilder := rag.NewPromptBuilder(loggerInstance)

	// 创建向量存储（使用PostgreSQL+PGVector）
	vectorStore, err := rag.NewVectorStore(
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			s.appConfig.Postgres.Host,
			s.appConfig.Postgres.Port,
			s.appConfig.Postgres.Username,
			s.appConfig.Postgres.Password,
			s.appConfig.Postgres.DBName,
			s.appConfig.Postgres.SSLMode),
		loggerInstance,
	)
	if err != nil {
		loggerInstance.Error("创建向量存储失败", logger.NewField("error", err.Error()))
	}

	userRepo := mysqlRepo.NewUserRepository(mysqlClient, loggerInstance)
	userService := user.NewUserService(userRepo, loggerInstance)

	ragService := rag.NewRAGService(loggerInstance, llmClient, documentProcessor, vectorStore, promptBuilder)

	// 创建审核仓储和审核服务
	auditRepo := mysqlRepo.NewAuditRepository(mysqlClient, loggerInstance)
	auditDomainService := audit.NewService(auditRepo, reimbursementRepo, ruleService, ragService, loggerInstance)
	auditAppService := service.NewAuditApplicationService(auditDomainService, loggerInstance)

	// 创建上传处理器
	uploadHandler := handler.NewUploadHandler(reimbursementAppService)

	// 注册上传相关路由
	s.engine.POST("/api/v1/reimbursement/upload", uploadHandler.UploadReimbursement)
	s.engine.POST("/api/v1/invoices/upload", uploadHandler.UploadInvoices)
	s.engine.POST("/api/v1/invoices/batch-upload", uploadHandler.BatchUpload)
	s.engine.POST("/api/v1/invoices/ocr", uploadHandler.TriggerOCRParsing)
	s.engine.POST("/api/v1/invoices/update-image", uploadHandler.UpdateInvoiceImage)

	// 创建审核处理器
	auditHandler := handler.NewAuditHandler(auditAppService)

	// 注册审核相关路由
	s.engine.POST("/api/v1/audit", auditHandler.StartAudit)
	s.engine.GET("/api/v1/audit/:id/status", auditHandler.GetAuditStatus)
	s.engine.GET("/api/v1/audit/:id/result", auditHandler.GetAuditResult)
	s.engine.POST("/api/v1/audit/:id/retry", auditHandler.RetryAudit)
	s.engine.POST("/api/v1/audit/:id/withdraw", auditHandler.WithdrawAudit)

	// 创建规则处理器
	ruleHandler := handler.NewRuleHandler(ruleService)

	// 注册规则相关路由
	s.engine.POST("/api/v1/rules", ruleHandler.CreateRule)
	s.engine.PUT("/api/v1/rules/:id", ruleHandler.UpdateRule)
	s.engine.DELETE("/api/v1/rules/:id", ruleHandler.DeleteRule)
	s.engine.GET("/api/v1/rules", ruleHandler.GetRules)
	s.engine.PATCH("/api/v1/rules/:id/enable", ruleHandler.EnableRule)
	s.engine.PATCH("/api/v1/rules/:id/disable", ruleHandler.DisableRule)
	s.engine.POST("/api/v1/rules/:id/test", ruleHandler.TestRule)

	userHandler := handler.NewUserHandler(userService)

	s.engine.POST("/api/v1/auth/login", userHandler.Login)

	s.engine.POST("/api/v1/audit/:id/manual-audit", auditHandler.ManualAudit)
	s.engine.GET("/api/v1/audit/flow-logs", auditHandler.GetFlowLogs)

	queryHandler := handler.NewQueryHandler(
		reimbursementAppService,
		reimbursementRepo,
		auditDomainService,
		loggerInstance,
	)

	// 注册查询相关路由
	s.engine.GET("/api/v1/reimbursement/:id", queryHandler.GetReimbursementByID)
	s.engine.PUT("/api/v1/reimbursement/:id", queryHandler.UpdateReimbursement)
	s.engine.GET("/api/v1/reimbursements/user", queryHandler.GetReimbursementsByUserID)
	s.engine.GET("/api/v1/reimbursements/all", queryHandler.GetAllReimbursements)
	s.engine.GET("/api/v1/reimbursements/date-range", queryHandler.GetReimbursementsByDateRange)
	s.engine.GET("/api/v1/reimbursement/:id/audit-report", queryHandler.GetAuditReport)
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
