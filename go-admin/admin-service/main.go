package main

import (
	_ "admin-service/routers"
	"admin-service/utils"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
		showVersion    = flag.Bool("version", false, "显示版本信息")
		showHelp       = flag.Bool("help", false, "显示帮助信息")
		autoInstall    = flag.Bool("install", false, "自动安装")
		checkStatus    = flag.Bool("status", false, "检查安装状态")
		migrate        = flag.Bool("migrate", false, "执行数据库迁移")
		uninstall      = flag.Bool("uninstall", false, "卸载系统")
		changePassword = flag.Bool("change-password", false, "修改管理员密码")
		createAdmin    = flag.Bool("create-admin", false, "创建新管理员")
		listAdmins     = flag.Bool("list-admins", false, "列出所有管理员用户")
		adminUsername  = flag.String("username", "", "管理员用户名")
		newPassword    = flag.String("password", "", "新密码")
		adminEmail     = flag.String("email", "", "管理员邮箱")
		nickName       = flag.String("nickName", "", "管理员真实姓名")
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

	// 执行数据库迁移
	if *migrate {
		fmt.Println("开始执行数据库迁移...")
		if err := utils.MigrateDatabase(); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
		fmt.Println("数据库迁移完成！")
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

	// 修改管理员密码
	if *changePassword {
		if *adminUsername == "" || *newPassword == "" {
			fmt.Println("错误: 需要指定用户名和新密码")
			fmt.Println("用法: go run main.go -change-password -username=admin -password=newpassword")
			os.Exit(1)
		}

		fmt.Printf("正在修改用户 '%s' 的密码...\n", *adminUsername)
		if err := utils.ChangeAdminPasswordCLI(*adminUsername, *newPassword); err != nil {
			log.Fatalf("修改密码失败: %v", err)
		}
		fmt.Println("✅ 密码修改成功！")
		os.Exit(0)
	}

	// 列出管理员用户
	if *listAdmins {
		fmt.Println("📋 获取管理员用户列表...")
		users, err := utils.ListAdminUsers()
		if err != nil {
			log.Fatalf("获取管理员列表失败: %v", err)
		}

		if len(users) == 0 {
			fmt.Println("📝 暂无管理员用户")
			os.Exit(0)
		}

		fmt.Printf("\n📊 共找到 %d 个管理员用户:\n", len(users))
		fmt.Println("=" + strings.Repeat("=", 120))
		fmt.Printf("%-5s %-15s %-25s %-15s %-15s %-8s %-20s %-20s %-10s\n",
			"ID", "用户名", "邮箱", "手机", "真实姓名", "状态", "最后登录时间", "最后登录IP", "角色ID")
		fmt.Println("-" + strings.Repeat("-", 120))

		for _, user := range users {
			status := "禁用"
			if user["status"].(int) == 1 {
				status = "启用"
			}

			lastLoginAt := "从未登录"
			if user["lastLoginAt"] != nil {
				lastLoginAt = user["lastLoginAt"].(string)
			}

			fmt.Printf("%-5v %-15s %-25s %-15s %-15s %-8s %-20s %-20s %-10v\n",
				user["id"],
				user["username"],
				user["email"],
				user["phone"],
				user["role"],
				status,
				lastLoginAt,
				user["lastLoginIp"],
				user["roleId"])
		}
		fmt.Println("=" + strings.Repeat("=", 120))
		os.Exit(0)
	}

	// 创建管理员用户
	if *createAdmin {
		if *adminUsername == "" || *newPassword == "" {
			fmt.Println("❌ 创建管理员需要指定用户名和密码")
			fmt.Println("使用方法: -create-admin -username=用户名 -password=密码 [-email=邮箱] [-nickName=真实姓名]")
			os.Exit(1)
		}

		fmt.Printf("👤 正在创建管理员用户 '%s'...\n", *adminUsername)

		// 设置默认值
		email := *adminEmail
		if email == "" {
			email = *adminUsername + "@example.com"
		}

		role := *nickName
		if role == "" {
			role = *adminUsername
		}

		if err := utils.CreateAdminUser(*adminUsername, *newPassword, email, role); err != nil {
			log.Fatalf("创建管理员失败: %v", err)
		}

		fmt.Println("✅ 管理员创建成功！")
		fmt.Printf("   用户名: %s\n", *adminUsername)
		fmt.Printf("   密码: %s\n", *newPassword)
		fmt.Printf("   邮箱: %s\n", email)
		fmt.Printf("   真实姓名: %s\n", role)
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

		// 执行数据库迁移
		fmt.Print("🔄 检查数据库升级...")
		if err := utils.MigrateDatabase(); err != nil {
			fmt.Printf(" ❌ 失败: %v\n", err)
			log.Printf("数据库迁移失败: %v", err)
		} else {
			fmt.Printf(" ✅ 完成\n")
		}
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
    -version              显示版本信息
    -help                显示此帮助信息
    -install             自动安装系统
    -status              检查安装状态
    -migrate             执行数据库迁移
    -uninstall           卸载系统
    -create-admin        创建新管理员用户
    -change-password     修改管理员密码
    -list-admins         列出所有管理员用户
    -username=用户名      指定管理员用户名
    -password=密码        指定新密码

示例:
                                             # 启动服务
    -install                                # 自动安装
    -status                                 # 检查状态
    -migrate                                # 执行数据库迁移
    -list-admins                            # 列出管理员用户
    -change-password -username=admin -password=newpass123  # 修改密码

更多信息请访问: https://github.com/your-repo/minigame-server

`, version, os.Args[0])
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
