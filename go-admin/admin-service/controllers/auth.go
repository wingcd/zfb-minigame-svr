package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"time"

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
	token, err := utils.GenerateJWT(admin.ID, admin.Username, admin.RoleId)
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

	// 更新登录时间
	models.UpdateAdminUserFields(admin.ID, map[string]interface{}{
		"lastLoginAt": time.Now(),
	})

	// 记录登录日志
	models.LogAdminOperation(admin.ID, admin.Username, "LOGIN_SUCCESS", "AUTH", map[string]interface{}{
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
				"id":          admin.ID,
				"username":    admin.Username,
				"nickname":    admin.Nickname,
				"role":        role.RoleCode,
				"roleName":    role.RoleName,
				"permissions": permissions,
				"email":       admin.Email,
				"phone":       admin.Phone,
				"lastLoginAt": admin.LastLoginAt.Format("2006-01-02 15:04:05"),
				"createdAt":   admin.CreatedAt,
			},
		},
	}
	c.ServeJSON()
}

// 旧的Login方法已合并到AdminLogin

// 旧的认证方法已合并到新的统一接口中

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
				"id":          admin.ID,
				"username":    admin.Username,
				"nickname":    admin.Nickname,
				"role":        role.RoleCode,
				"roleName":    role.RoleName,
				"permissions": permissions,
				"email":       admin.Email,
				"phone":       admin.Phone,
				"lastLoginAt": admin.LastLoginAt.Format("2006-01-02 15:04:05"),
				"createdAt":   admin.CreatedAt,
			},
		},
	}
	c.ServeJSON()
}

// Logout 登出（对齐云函数logout接口）
func (c *AuthController) LogoutAdmin() {
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "登出成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"message": "登出成功",
		},
	}
	c.ServeJSON()
}

// GetProfile 获取当前用户资料（对齐云函数getProfile接口）
func (c *AuthController) GetAdminProfile() {
	var req struct {
		Token string `json:"token"`
	}

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

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"id":          admin.ID,
				"username":    admin.Username,
				"nickname":    admin.Nickname,
				"role":        role.RoleCode,
				"roleName":    role.RoleName,
				"permissions": permissions,
				"email":       admin.Email,
				"phone":       admin.Phone,
				"lastLoginAt": admin.LastLoginAt.Format("2006-01-02 15:04:05"),
				"createdAt":   admin.CreatedAt,
			},
		},
	}
	c.ServeJSON()
}

// UpdateProfile 更新用户资料（对齐云函数updateProfile接口）
func (c *AuthController) UpdateAdminProfile() {
	var req struct {
		Token string `json:"token"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		Role  string `json:"role"`
	}

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

	// 更新用户资料
	updateFields := make(map[string]interface{})
	if req.Email != "" {
		updateFields["email"] = req.Email
	}
	if req.Phone != "" {
		updateFields["phone"] = req.Phone
	}
	if req.Role != "" {
		updateFields["role"] = req.Role
	}

	if err := models.UpdateAdminUserFields(claims.UserID, updateFields); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取更新后的用户信息
	admin, _ := models.GetAdminUserById(claims.UserID)

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"id":       admin.ID,
				"username": admin.Username,
				"nickname": admin.Role,
				"email":    admin.Email,
				"phone":    admin.Phone,
			},
		},
	}
	c.ServeJSON()
}

// ChangePassword 修改密码（对齐云函数changePassword接口）
func (c *AuthController) ChangeAdminPassword() {
	var req struct {
		Token       string `json:"token"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

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

	if req.OldPassword == "" || req.NewPassword == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "原密码和新密码不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if len(req.NewPassword) < 6 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "新密码长度至少6位",
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

	// 修改密码
	err = models.ChangeAdminPassword(claims.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "修改密码失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(claims.UserID, claims.Username, "CHANGE_PASSWORD", "AUTH", map[string]interface{}{
		"success": true,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "修改成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"message": "修改成功",
		},
	}
	c.ServeJSON()
}
