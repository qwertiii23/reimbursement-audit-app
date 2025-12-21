package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"

	"reimbursement-audit/internal/config"
	"reimbursement-audit/internal/infra/storage/mysql"
	mysqlmigration "reimbursement-audit/internal/infra/storage/mysql/migration"
)

var (
	configFile = flag.String("config", "config.yaml", "配置文件路径")
	action     = flag.String("action", "up", "迁移操作 (up/down/status/version)")
	version    = flag.Bool("version", false, "显示版本信息")
	help       = flag.Bool("help", false, "显示帮助信息")
	buildTime  = "unknown" // 构建时间，由编译时设置
)

const (
	AppName    = "reimbursement-audit-migration"
	AppVersion = "1.0.0"
	AppDesc    = "报销审核系统数据库迁移工具"
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

	// 创建MySQL客户端
	dbConfig := mysql.DefaultConfig()
	if cfg != nil {
		// TODO: 从配置中设置数据库参数
	}

	client := mysql.NewClient()
	if err := client.Connect(context.Background(), dbConfig); err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer client.Close()

	// 创建迁移管理器
	mgr := mysqlmigration.NewMigrationManager(client)

	// 执行迁移操作
	switch *action {
	case "up":
		if err := mgr.Up(context.Background()); err != nil {
			log.Fatalf("执行迁移失败: %v", err)
		}
		log.Println("迁移执行成功")
	case "down":
		if err := mgr.Down(context.Background()); err != nil {
			log.Fatalf("回滚迁移失败: %v", err)
		}
		log.Println("迁移回滚成功")
	case "status":
		status, err := mgr.Status(context.Background())
		if err != nil {
			log.Fatalf("获取迁移状态失败: %v", err)
		}
		fmt.Printf("迁移状态: %v\n", status)
	case "version":
		version, err := mgr.Version(context.Background())
		if err != nil {
			log.Fatalf("获取迁移版本失败: %v", err)
		}
		fmt.Printf("当前版本: %v\n", version)
	default:
		log.Fatalf("不支持的操作: %s", *action)
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf(`%s - %s

用法:
  %s [选项]

选项:
  -config string
        配置文件路径 (默认: "config.yaml")
  -action string
        迁移操作 (up/down/status/version) (默认: "up")
  -version
        显示版本信息
  -help
        显示帮助信息

示例:
  %s -action up -config config.yaml
  %s -action down -config config.yaml
  %s -action status -config config.yaml
  %s -action version -config config.yaml
`, AppName, AppDesc, AppName, AppName, AppName, AppName, AppName)
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Printf(`%s %s

构建信息:
  Go版本: %s
  编译时间: %s
`, AppName, AppVersion, runtime.Version(), buildTime)
}
