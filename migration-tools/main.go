package main

import (
	"flag"
	"fmt"
	"log"
	"migration-tools/config"
	"migration-tools/migrators"
	"migration-tools/utils"
	"os"
)

var (
	configFile = flag.String("config", "config/migration.yaml", "配置文件路径")
	mode       = flag.String("mode", "full", "迁移模式: full|incremental|verify|report")
	since      = flag.String("since", "", "增量迁移开始时间 (格式: 2023-10-01 00:00:00)")
	dryRun     = flag.Bool("dry-run", false, "试运行模式，不实际执行迁移")
	verbose    = flag.Bool("verbose", false, "详细输出模式")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置日志级别
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Printf("开始数据迁移，模式: %s", *mode)

	switch *mode {
	case "full":
		err = runFullMigration(cfg)
	case "incremental":
		if *since == "" {
			log.Fatal("增量迁移需要指定 --since 参数")
		}
		err = runIncrementalMigration(cfg, *since)
	case "verify":
		err = runVerification(cfg)
	case "report":
		err = generateReport(cfg)
	default:
		log.Fatalf("不支持的迁移模式: %s", *mode)
	}

	if err != nil {
		log.Fatalf("迁移失败: %v", err)
	}

	log.Println("数据迁移完成")
}

// runFullMigration 执行全量迁移
func runFullMigration(cfg *config.Config) error {
	log.Println("开始全量数据迁移...")

	// 初始化数据库连接
	mongoClient, err := utils.NewMongoClient(cfg.Source.Connection)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}
	defer mongoClient.Close()

	mysqlClient, err := utils.NewMySQLClient(cfg.Target.Connection)
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %v", err)
	}
	defer mysqlClient.Close()

	// 创建迁移器管理器
	manager := migrators.NewMigrationManager(mongoClient, mysqlClient, cfg)

	// 执行系统表迁移
	if err := manager.MigrateSystemTables(); err != nil {
		return fmt.Errorf("系统表迁移失败: %v", err)
	}

	// 执行应用表迁移
	if err := manager.MigrateApplicationTables(); err != nil {
		return fmt.Errorf("应用表迁移失败: %v", err)
	}

	// 执行动态表迁移
	if err := manager.MigrateDynamicTables(); err != nil {
		return fmt.Errorf("动态表迁移失败: %v", err)
	}

	log.Println("全量数据迁移完成")
	return nil
}

// runIncrementalMigration 执行增量迁移
func runIncrementalMigration(cfg *config.Config, since string) error {
	log.Printf("开始增量数据迁移，起始时间: %s", since)

	// 初始化数据库连接
	mongoClient, err := utils.NewMongoClient(cfg.Source.Connection)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}
	defer mongoClient.Close()

	mysqlClient, err := utils.NewMySQLClient(cfg.Target.Connection)
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %v", err)
	}
	defer mysqlClient.Close()

	// 创建迁移器管理器
	manager := migrators.NewMigrationManager(mongoClient, mysqlClient, cfg)

	// 执行增量迁移
	if err := manager.MigrateIncremental(since); err != nil {
		return fmt.Errorf("增量迁移失败: %v", err)
	}

	log.Println("增量数据迁移完成")
	return nil
}

// runVerification 执行数据验证
func runVerification(cfg *config.Config) error {
	log.Println("开始数据验证...")

	// 初始化数据库连接
	mongoClient, err := utils.NewMongoClient(cfg.Source.Connection)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}
	defer mongoClient.Close()

	mysqlClient, err := utils.NewMySQLClient(cfg.Target.Connection)
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %v", err)
	}
	defer mysqlClient.Close()

	// 创建验证器
	validator := utils.NewDataValidator(mongoClient, mysqlClient)

	// 执行验证
	report, err := validator.ValidateAll()
	if err != nil {
		return fmt.Errorf("数据验证失败: %v", err)
	}

	// 输出验证报告
	fmt.Println("\n=== 数据验证报告 ===")
	for table, result := range report {
		status := "✅ 通过"
		if !result.Success {
			status = "❌ 失败"
		}
		fmt.Printf("%s: %s (源: %d, 目标: %d)\n", table, status, result.SourceCount, result.TargetCount)
		if result.Error != "" {
			fmt.Printf("  错误: %s\n", result.Error)
		}
	}

	log.Println("数据验证完成")
	return nil
}

// generateReport 生成迁移报告
func generateReport(cfg *config.Config) error {
	log.Println("开始生成迁移报告...")

	// 初始化数据库连接
	mongoClient, err := utils.NewMongoClient(cfg.Source.Connection)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}
	defer mongoClient.Close()

	mysqlClient, err := utils.NewMySQLClient(cfg.Target.Connection)
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %v", err)
	}
	defer mysqlClient.Close()

	// 创建报告生成器
	reporter := utils.NewReporter(mongoClient, mysqlClient)

	// 生成报告
	report, err := reporter.GenerateReport()
	if err != nil {
		return fmt.Errorf("生成报告失败: %v", err)
	}

	// 保存报告到文件
	reportFile := "migration_report.html"
	if err := os.WriteFile(reportFile, []byte(report), 0644); err != nil {
		return fmt.Errorf("保存报告失败: %v", err)
	}

	log.Printf("迁移报告已保存到: %s", reportFile)
	return nil
}
