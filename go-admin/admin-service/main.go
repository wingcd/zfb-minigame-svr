package main

import (
	_ "admin-service/routers"
	"admin-service/utils"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/server/web"
)

var (
	version   = "1.0.0"
	buildTime string
)

func main() {
	// 命令行参数
	var (
		showVersion = flag.Bool("version", false, "显示版本信息")
		showHelp    = flag.Bool("help", false, "显示帮助信息")
		autoInstall = flag.Bool("install", false, "自动安装")
		checkStatus = flag.Bool("status", false, "检查安装状态")
		uninstall   = flag.Bool("uninstall", false, "卸载系统")
	)
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Minigame Admin Service %s\n", version)
		if buildTime != "" {
			fmt.Printf("Build Time: %s\n", buildTime)
		}
		os.Exit(0)
	}

	// 显示帮助信息
	if *showHelp {
		showHelpInfo()
		os.Exit(0)
	}

	// 检查安装状态
	if *checkStatus {
		status := utils.CheckInstallStatus()
		fmt.Printf("安装状态: %v\n", status.IsInstalled)
		fmt.Printf("数据库类型: %s\n", status.DatabaseType)
		fmt.Printf("数据库状态: %s\n", status.DatabaseStatus)
		fmt.Printf("管理员存在: %v\n", status.AdminExists)
		if status.InstallTime != "" {
			fmt.Printf("安装时间: %s\n", status.InstallTime)
		}
		os.Exit(0)
	}

	// 自动安装
	if *autoInstall {
		fmt.Println("开始自动安装...")
		if err := utils.AutoInstall(); err != nil {
			log.Fatalf("自动安装失败: %v", err)
		}
		fmt.Println("自动安装完成！")
		os.Exit(0)
	}

	// 卸载系统
	if *uninstall {
		fmt.Print("确定要卸载系统吗？这将删除所有数据 [y/N]: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			if err := utils.Uninstall(); err != nil {
				log.Fatalf("卸载失败: %v", err)
			}
			fmt.Println("卸载完成！")
		} else {
			fmt.Println("取消卸载")
		}
		os.Exit(0)
	}

	// 启动前检查
	fmt.Printf("🚀 启动 Minigame Admin Service %s\n", version)

	// 检查安装状态
	status := utils.CheckInstallStatus()
	if !status.IsInstalled {
		fmt.Println("⚠️  系统未安装，将启用安装模式")

		// 检查是否启用自动安装
		if checkAutoInstallConfig() {
			fmt.Println("🔧 检测到自动安装配置，开始自动安装...")
			if err := utils.AutoInstall(); err != nil {
				fmt.Printf("❌ 自动安装失败: %v\n", err)
				fmt.Println("请访问 http://localhost:8080/install 进行手动安装")
			} else {
				fmt.Println("✅ 自动安装完成！")
			}
		} else {
			fmt.Println("请访问 http://localhost:8080/install 进行系统安装")
		}
	} else {
		fmt.Printf("✅ 系统已安装 (数据库: %s)\n", status.DatabaseType)
	}

	// 读取配置
	loadConfig()

	// 启动服务
	fmt.Printf("🌐 服务启动在端口: %d\n", web.BConfig.Listen.HTTPPort)
	fmt.Printf("📊 管理界面: http://localhost:%d\n", web.BConfig.Listen.HTTPPort)

	web.Run()
}

// showHelpInfo 显示帮助信息
func showHelpInfo() {
	fmt.Printf(`Minigame Admin Service %s

用法: %s [选项]

选项:
    -version        显示版本信息
    -help          显示此帮助信息
    -install       自动安装系统
    -status        检查安装状态
    -uninstall     卸载系统

示例:
    %s                    # 启动服务
    %s -install          # 自动安装
    %s -status           # 检查状态

更多信息请访问: https://github.com/your-repo/minigame-server

`, version, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

// checkAutoInstallConfig 检查是否启用自动安装
func checkAutoInstallConfig() bool {
	configPath := utils.FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)
	if err != nil {
		return false
	}

	autoInstall, _ := appconf.Bool("auto_install")
	return autoInstall
}

// loadConfig 加载配置
func loadConfig() {
	// 从配置文件读取端口设置
	configPath := utils.FindConfigFile()
	if appconf, err := config.NewConfig("ini", configPath); err == nil {
		if port, err := appconf.Int("httpport"); err == nil && port > 0 {
			web.BConfig.Listen.HTTPPort = port
		}

		// 设置运行模式
		if runmode, _ := appconf.String("runmode"); runmode != "" {
			web.BConfig.RunMode = runmode
		}
	}
}
