package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reimbursement-audit/internal/config"
	"reimbursement-audit/internal/server"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	configFile = flag.String("config", "configs/config.dev.yaml", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
	help       = flag.Bool("help", false, "显示帮助信息")
	buildTime  = "unknown" // 构建时间，由编译时设置
)

const (
	AppName    = "reimbursement-audit"
	AppVersion = "1.0.0"
	AppDesc    = "报销审核系统"
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	// 加载配置
	loader := config.NewLoader(*configFile)
	cfg, err := loader.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 转换服务器配置
	readTimeout, _ := time.ParseDuration(cfg.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(cfg.Server.WriteTimeout)
	idleTimeout, _ := time.ParseDuration(cfg.Server.IdleTimeout)

	serverConfig := &server.Config{
		Host:         cfg.Server.Host,
		Port:         cfg.Server.Port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Mode:         gin.ReleaseMode,
		TLS:          false,
	}

	// 创建服务器
	srv := server.NewServer(serverConfig)

	// 设置应用配置
	srv.SetAppConfig(cfg)

	// 注册路由
	srv.RegisterRoutes()

	// 设置静态文件服务
	srv.SetupStaticFiles("/uploads", "./uploads")
	srv.SetupStaticFiles("/api/v1/files", "./uploads")

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听信号
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("收到退出信号，正在优雅关闭服务器...")
		cancel()
	}()

	// 启动服务器
	go func() {
		address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Printf("启动服务器，监听地址: %s", address)
		if err := srv.Start(); err != nil && err != context.Canceled {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()

	// 优雅关闭服务器
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf(`%s - %s

用法:
  %s [选项]

选项:
  -config string
        配置文件路径 (默认: "config.yaml")
  -version
        显示版本信息
  -help
        显示帮助信息

示例:
  %s -config config.yaml
  %s -version
`, AppName, AppDesc, AppName, AppName, AppName)
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Printf(`%s %s

构建信息:
  Go版本: %s
  编译时间: %s
`, AppName, AppVersion, runtime.Version(), buildTime)
}
