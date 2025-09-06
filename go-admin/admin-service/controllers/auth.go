package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
)

type AuthController struct {
	web.Controller
}

// AdminLoginRequest 管理员登录请求结构（对齐云函数）
type AdminLoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

// AdminLogin 管理员登录（对齐云函数adminLogin接口）
func (c *AuthController) AdminLogin() {
	var req AdminLoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 参数校验
	if req.Username == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "用户名不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.Password == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "密码不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// MD5密码加密（对齐云函数）
	passwordHash := utils.HashPassword(req.Password)

	// 验证登录
	admin, err := models.AdminLoginWithMD5(req.Username, passwordHash)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "用户名或密码错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(admin.Id, admin.Username, admin.RoleId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "生成token失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// JWT token不需要存储在数据库中，过期时间已在token中编码
	clientIP := c.Ctx.Input.IP()

	// 获取角色权限
	role, permissions, err := models.GetAdminRolePermissions(admin.RoleId)
	if err != nil {
		// 如果获取角色失败，使用默认值
		role = &models.AdminRole{RoleName: "未知角色"}
		permissions = []string{}
	}

	// 记录登录日志
	models.LogAdminOperation(admin.Id, admin.Username, "LOGIN_SUCCESS", "AUTH", map[string]interface{}{
		"ip":         clientIP,
		"rememberMe": req.RememberMe,
	})

	// 返回成功结果（对齐云函数格式）
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "登录成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"token": token,
			"adminInfo": map[string]interface{}{
				"id":            admin.Id,
				"username":      admin.Username,
				"nickname":      admin.RealName,
				"role":          role.RoleCode,
				"roleName":      role.RoleName,
				"permissions":   permissions,
				"email":         admin.Email,
				"phone":         admin.Phone,
				"lastLoginTime": admin.LastLoginAt.Format("2006-01-02 15:04:05"),
				"createTime":    admin.CreatedAt,
			},
		},
	}
	c.ServeJSON()
}

// Login 管理员登录（保持原有接口兼容性）
func (c *AuthController) Login() {
	// 获取参数
	username := c.GetString("username")
	password := c.GetString("password")

	if username == "" || password == "" {
		utils.ErrorResponse(&c.Controller, 1002, "用户名和密码不能为空", nil)
		return
	}

	// 验证登录
	admin, err := models.AdminLogin(username, password)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "登录失败: "+err.Error(), nil)
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(admin.Id, admin.Username, admin.RoleId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "生成token失败: "+err.Error(), nil)
		return
	}

	// 记录登录日志
	utils.LogOperation(admin.Id, "登录", "管理员登录")

	result := map[string]interface{}{
		"token": token,
		"admin": map[string]interface{}{
			"id":       admin.Id,
			"username": admin.Username,
			"nickname": admin.RealName,
			"email":    admin.Email,
			"roleId":   admin.RoleId,
		},
	}

	utils.SuccessResponse(&c.Controller, "登录成功", result)
}

// Logout 管理员登出
func (c *AuthController) Logout() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 记录登出日志
	utils.LogOperation(claims.UserID, "登出", "管理员登出")

	utils.SuccessResponse(&c.Controller, "登出成功", nil)
}

// GetProfile 获取管理员信息
func (c *AuthController) GetProfile() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取管理员详情
	admin, err := models.GetAdminById(claims.UserID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取管理员信息失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"id":       admin.Id,
		"username": admin.Username,
		"nickname": admin.RealName,
		"email":    admin.Email,
		"roleId":   admin.RoleId,
		"status":   admin.Status,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// UpdateProfile 更新管理员信息
func (c *AuthController) UpdateProfile() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	nickname := c.GetString("nickname")
	email := c.GetString("email")

	if nickname == "" {
		utils.ErrorResponse(&c.Controller, 1002, "昵称不能为空", nil)
		return
	}

	// 更新管理员信息
	err := models.UpdateAdminProfile(claims.UserID, nickname, email)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "更新资料", "更新管理员资料")

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// ChangePassword 修改密码
func (c *AuthController) ChangePassword() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	oldPassword := c.GetString("oldPassword")
	newPassword := c.GetString("newPassword")

	if oldPassword == "" || newPassword == "" {
		utils.ErrorResponse(&c.Controller, 1002, "原密码和新密码不能为空", nil)
		return
	}

	if len(newPassword) < 6 {
		utils.ErrorResponse(&c.Controller, 1002, "新密码长度至少6位", nil)
		return
	}

	// 修改密码
	err := models.ChangeAdminPassword(claims.UserID, oldPassword, newPassword)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "修改密码失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "修改密码", "修改管理员密码")

	utils.SuccessResponse(&c.Controller, "修改成功", nil)
}

// VerifyTokenRequest Token验证请求结构
type VerifyTokenRequest struct {
	Token string `json:"token"`
}

// VerifyToken Token验证（对齐云函数verifyToken接口）
func (c *AuthController) VerifyToken() {
	var req VerifyTokenRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.Token == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "Token不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 验证JWT Token
	claims, err := utils.ParseJWT(req.Token)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4003,
			"msg":       "Token无效或已过期: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取管理员信息
	admin, err := models.GetAdminUserById(claims.UserID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4003,
			"msg":       "用户不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取角色权限
	role, permissions, err := models.GetAdminRolePermissions(admin.RoleId)
	if err != nil {
		role = &models.AdminRole{RoleName: "未知角色"}
		permissions = []string{}
	}

	// 返回管理员信息（对齐云函数格式）
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "Token验证成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"valid": true,
			"adminInfo": map[string]interface{}{
				"id":            admin.Id,
				"username":      admin.Username,
				"nickname":      admin.RealName,
				"role":          role.RoleCode,
				"roleName":      role.RoleName,
				"permissions":   permissions,
				"email":         admin.Email,
				"phone":         admin.Phone,
				"lastLoginTime": admin.LastLoginAt.Format("2006-01-02 15:04:05"),
				"createTime":    admin.CreatedAt,
			},
		},
	}
	c.ServeJSON()
}
