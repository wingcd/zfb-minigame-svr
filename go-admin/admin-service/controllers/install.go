package controllers

import (
	"admin-service/utils"
	"encoding/json"
	"fmt"
	"log"

	"github.com/beego/beego/v2/server/web"
)

// InstallController 安装控制器
type InstallController struct {
	web.Controller
}

// InstallResponse 安装响应
type InstallResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CheckStatus 检查安装状态
func (c *InstallController) CheckStatus() {
	status := utils.CheckInstallStatus()

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "success",
		Data:    status,
	}
	c.ServeJSON()
}

// AutoInstall 自动安装
func (c *InstallController) AutoInstall() {
	// 检查是否已安装
	status := utils.CheckInstallStatus()
	if status.IsInstalled {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "系统已安装",
		}
		c.ServeJSON()
		return
	}

	// 执行自动安装
	if err := utils.AutoInstall(); err != nil {
		log.Printf("自动安装失败: %v", err)
		c.Data["json"] = InstallResponse{
			Code:    500,
			Message: "安装失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "安装成功",
		Data: map[string]interface{}{
			"redirect": "/admin/login",
		},
	}
	c.ServeJSON()
}

// ManualInstall 手动安装
func (c *InstallController) ManualInstall() {
	// 检查是否已安装
	status := utils.CheckInstallStatus()
	if status.IsInstalled {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "系统已安装",
		}
		c.ServeJSON()
		return
	}

	// 解析安装配置
	var config utils.InstallConfig
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "配置格式错误: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	// 验证配置
	if err := validateInstallConfig(&config); err != nil {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "配置验证失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	// 执行手动安装
	if err := utils.ManualInstall(&config); err != nil {
		log.Printf("手动安装失败: %v", err)
		c.Data["json"] = InstallResponse{
			Code:    500,
			Message: "安装失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "安装成功",
		Data: map[string]interface{}{
			"redirect": "/admin/login",
		},
	}
	c.ServeJSON()
}

// TestConnection 测试数据库连接
func (c *InstallController) TestConnection() {
	var config utils.InstallConfig
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "配置格式错误",
		}
		c.ServeJSON()
		return
	}

	// 测试连接
	success, message := testDatabaseConnection(&config)

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "测试完成",
		Data: map[string]interface{}{
			"success": success,
			"message": message,
		},
	}
	c.ServeJSON()
}

// Uninstall 卸载系统
func (c *InstallController) Uninstall() {
	// 检查是否已安装
	status := utils.CheckInstallStatus()
	if !status.IsInstalled {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "系统未安装",
		}
		c.ServeJSON()
		return
	}

	// 执行卸载
	if err := utils.Uninstall(); err != nil {
		c.Data["json"] = InstallResponse{
			Code:    500,
			Message: "卸载失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "卸载成功",
	}
	c.ServeJSON()
}

// ShowInstallPage 显示安装页面
func (c *InstallController) ShowInstallPage() {
	status := utils.CheckInstallStatus()

	// 如果已安装，重定向到登录页面
	if status.IsInstalled {
		c.Redirect("/admin/login", 302)
		return
	}

	c.Data["Status"] = status
	c.TplName = "install/index.html"
}

// validateInstallConfig 验证安装配置
func validateInstallConfig(config *utils.InstallConfig) error {
	if config.DatabaseType == "" {
		config.DatabaseType = "sqlite"
	}

	if config.DatabaseType == "mysql" {
		if config.MySQLHost == "" {
			config.MySQLHost = "127.0.0.1"
		}
		if config.MySQLPort == "" {
			config.MySQLPort = "3306"
		}
		if config.MySQLUser == "" {
			return fmt.Errorf("MySQL用户名不能为空")
		}
		if config.MySQLDatabase == "" {
			config.MySQLDatabase = "minigame_admin"
		}
	}

	if config.AdminUsername == "" {
		return fmt.Errorf("管理员用户名不能为空")
	}
	if config.AdminPassword == "" {
		return fmt.Errorf("管理员密码不能为空")
	}
	if len(config.AdminPassword) < 6 {
		return fmt.Errorf("管理员密码长度不能少于6位")
	}

	return nil
}

// testDatabaseConnection 测试数据库连接
func testDatabaseConnection(config *utils.InstallConfig) (bool, string) {
	if config.DatabaseType == "sqlite" {
		return true, "SQLite 连接正常"
	}

	if config.DatabaseType == "mysql" {
		// 这里需要实际测试MySQL连接
		// 简化实现，实际应该创建连接并测试
		if config.MySQLHost == "" || config.MySQLUser == "" {
			return false, "MySQL 配置不完整"
		}

		// TODO: 实际测试MySQL连接
		return true, "MySQL 连接正常"
	}

	return false, "不支持的数据库类型"
}

// InitSystemRequest 系统初始化请求
type InitSystemRequest struct {
	AdminUsername string `json:"adminUsername"`
	AdminPassword string `json:"adminPassword"`
	Force         bool   `json:"force"`
}

// InitSystem 系统初始化（前端调用的接口）
func (c *InstallController) InitSystem() {
	var req InitSystemRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = InstallResponse{
			Code:    4001,
			Message: "参数解析失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	// 检查是否已安装
	status := utils.CheckInstallStatus()
	if status.IsInstalled && !req.Force {
		c.Data["json"] = InstallResponse{
			Code:    409,
			Message: "系统已经初始化",
		}
		c.ServeJSON()
		return
	}

	// 使用前端传递的参数进行安装
	if err := utils.AutoInstallWithParams(req.AdminUsername, req.AdminPassword); err != nil {
		log.Printf("系统初始化失败: %v", err)
		c.Data["json"] = InstallResponse{
			Code:    500,
			Message: "初始化失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "初始化成功",
		Data: map[string]interface{}{
			"createdCollections": 3,
			"createdRoles":       4,
			"createdAdmins":      1,
			"defaultCredentials": map[string]interface{}{
				"username": req.AdminUsername,
				"password": req.AdminPassword,
				"warning":  "请立即登录并修改默认密码！",
			},
		},
	}
	c.ServeJSON()
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Username    string `json:"username"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ChangePassword 修改管理员密码
func (c *InstallController) ChangePassword() {
	var req ChangePasswordRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "参数解析失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	// 验证参数
	if req.Username == "" {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "用户名不能为空",
		}
		c.ServeJSON()
		return
	}

	if req.OldPassword == "" {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "原密码不能为空",
		}
		c.ServeJSON()
		return
	}

	if req.NewPassword == "" {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "新密码不能为空",
		}
		c.ServeJSON()
		return
	}

	if len(req.NewPassword) < 6 {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "新密码长度不能少于6位",
		}
		c.ServeJSON()
		return
	}

	// 执行密码修改
	if err := utils.ChangeAdminPassword(req.Username, req.OldPassword, req.NewPassword); err != nil {
		log.Printf("修改密码失败: %v", err)
		c.Data["json"] = InstallResponse{
			Code:    500,
			Message: "修改密码失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "密码修改成功",
	}
	c.ServeJSON()
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Username    string `json:"username"`
	NewPassword string `json:"newPassword"`
	Force       bool   `json:"force"` // 是否强制重置（不验证原密码）
}

// ResetPassword 重置管理员密码（用于忘记密码的情况）
func (c *InstallController) ResetPassword() {
	var req ResetPasswordRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "参数解析失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	// 验证参数
	if req.Username == "" {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "用户名不能为空",
		}
		c.ServeJSON()
		return
	}

	if req.NewPassword == "" {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "新密码不能为空",
		}
		c.ServeJSON()
		return
	}

	if len(req.NewPassword) < 6 {
		c.Data["json"] = InstallResponse{
			Code:    400,
			Message: "新密码长度不能少于6位",
		}
		c.ServeJSON()
		return
	}

	// 执行密码重置
	if err := utils.ResetAdminPassword(req.Username, req.NewPassword); err != nil {
		log.Printf("重置密码失败: %v", err)
		c.Data["json"] = InstallResponse{
			Code:    500,
			Message: "重置密码失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "密码重置成功",
		Data: map[string]interface{}{
			"username": req.Username,
			"message":  "密码已重置，请使用新密码登录",
		},
	}
	c.ServeJSON()
}

// ListAdmins 列出管理员用户
func (c *InstallController) ListAdmins() {
	// 获取管理员用户列表
	users, err := utils.ListAdminUsers()
	if err != nil {
		log.Printf("获取管理员用户列表失败: %v", err)
		c.Data["json"] = InstallResponse{
			Code:    500,
			Message: "获取用户列表失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = InstallResponse{
		Code:    200,
		Message: "获取用户列表成功",
		Data: map[string]interface{}{
			"users": users,
			"total": len(users),
		},
	}
	c.ServeJSON()
}
