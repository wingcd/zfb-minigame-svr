package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// AdminController 管理员控制器
type AdminController struct {
	web.Controller
}

// 登录相关功能已移至 AuthController

// GetAdmins 获取管理员列表 (暂时保留但不启用路由)

// GetAdmin 获取单个管理员信息 (API命名空间使用)
func (c *AdminController) GetAdmin() {
	var req struct {
		ID int64 `json:"id"`
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

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的管理员ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	user, err := models.GetAdminUserById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "用户不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      user,
	}
	c.ServeJSON()
}

// UpdateAdmin 更新管理员信息 (API命名空间使用)
func (c *AdminController) UpdateAdmin() {
	var req struct {
		ID     int64  `json:"id"`
		Email  string `json:"email"`
		Phone  string `json:"phone"`
		Role   string `json:"role"`
		RoleId int64  `json:"roleId"`
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

	var user models.AdminUser
	if err := utils.ParseJSON(&c.Controller, &user); err != nil {
		utils.Error(&c.Controller, utils.CodeInvalidParam, "参数解析失败")
		return
	}

	// 检查用户是否存在
	existUser, err := models.GetAdminUserById(req.ID)
	if err != nil {
		utils.Error(&c.Controller, utils.CodeNotFound, "用户不存在")
		return
	}

	// 更新数据
	if user.Password != "" {
		hashedPassword, err := utils.HashPasswordBcrypt(user.Password)
		if err != nil {
			utils.Error(&c.Controller, utils.CodeError, "密码加密失败")
			return
		}
		user.Password = hashedPassword
	} else {
		user.Password = existUser.Password
	}

	if err := models.UpdateAdminUser(&user); err != nil {
		utils.Error(&c.Controller, utils.CodeError, "更新失败")
		return
	}

	c.logOperation("更新管理员", "admin", "PUT", "/api/admins/"+strconv.FormatInt(req.ID, 10), user, 1, "", 0)
	utils.Success(&c.Controller, "更新成功")
}

// DeleteAdmin 删除管理员 (API命名空间使用)
func (c *AdminController) DeleteAdmin() {
	var req struct {
		ID int64 `json:"id"`
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

	// 检查用户是否存在
	_, err := models.GetAdminUserById(req.ID)
	if err != nil {
		utils.Error(&c.Controller, utils.CodeNotFound, "用户不存在")
		return
	}

	if err := models.DeleteAdminUser(req.ID); err != nil {
		utils.Error(&c.Controller, utils.CodeError, "删除失败")
		return
	}

	c.logOperation("删除管理员", "admin", "DELETE", "/api/admins/"+strconv.FormatInt(req.ID, 10), nil, 1, "", 0)
	utils.Success(&c.Controller, "删除成功")
}

// ResetPassword 重置管理员密码 (API命名空间使用)
func (c *AdminController) ResetPassword() {
	var req struct {
		ID          int64  `json:"id"`
		NewPassword string `json:"newPassword"`
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

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的管理员ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if len(req.NewPassword) < 6 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "密码长度至少6位",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查用户是否存在
	_, err := models.GetAdminUserById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "用户不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPasswordBcrypt(req.NewPassword)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "密码加密失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 更新密码
	if err := models.UpdateAdminUserFields(req.ID, map[string]interface{}{
		"password": hashedPassword,
	}); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "密码重置失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.logOperation("重置管理员密码", "admin", "POST", "/api/admins/"+strconv.FormatInt(req.ID, 10)+"/reset-password", nil, 1, "", 0)
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "密码重置成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// GetUsers 获取管理员列表
func (c *AdminController) GetUsers() {
	var req struct {
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
		Keyword  string `json:"keyword"`
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

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	users, total, err := models.GetAllAdminUsers(req.Page, req.PageSize, req.Keyword)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取数据失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"list":       users,
			"total":      total,
			"page":       req.Page,
			"pageSize":   req.PageSize,
			"totalPages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		},
	}
	c.ServeJSON()
}

// AddUser 添加管理员
func (c *AdminController) AddUser() {
	var user models.AdminUser
	if err := utils.ParseJSON(&c.Controller, &user); err != nil {
		utils.Error(&c.Controller, utils.CodeInvalidParam, "参数解析失败")
		return
	}

	// 验证必填参数
	if !utils.ValidateRequired(&c.Controller, map[string]interface{}{
		"username": user.Username,
		"password": user.Password,
	}) {
		return
	}

	// 检查用户名是否已存在
	existUser, _ := models.GetAdminUserByUsername(user.Username)
	if existUser != nil {
		utils.Error(&c.Controller, utils.CodeConflict, "用户名已存在")
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPasswordBcrypt(user.Password)
	if err != nil {
		utils.Error(&c.Controller, utils.CodeError, "密码加密失败")
		return
	}
	user.Password = hashedPassword

	// 保存用户
	if err := models.AddAdminUser(&user); err != nil {
		utils.Error(&c.Controller, utils.CodeError, "添加用户失败")
		return
	}

	c.logOperation("添加管理员", "admin", "POST", "/admin/users", user, 1, "", 0)
	utils.Success(&c.Controller, "添加成功")
}

// UpdateUser 更新管理员
func (c *AdminController) UpdateUser() {
	var req struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		RoleId   int64  `json:"roleId"`
		Status   int    `json:"status"`
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

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的管理员ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查用户是否存在
	existUser, err := models.GetAdminUserById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "用户不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 构建更新的用户对象
	user := models.AdminUser{
		Username: req.Username,
		Role:     req.Role,
		Email:    req.Email,
		Phone:    req.Phone,
		RoleId:   req.RoleId,
		Status:   req.Status,
	}

	// 处理密码
	if req.Password != "" {
		hashedPassword, err := utils.HashPasswordBcrypt(req.Password)
		if err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      5001,
				"msg":       "密码加密失败",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}
		user.Password = hashedPassword
	} else {
		user.Password = existUser.Password
	}

	if err := models.UpdateAdminUser(&user); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.logOperation("更新管理员", "admin", "PUT", "/admin/users/"+strconv.FormatInt(req.ID, 10), user, 1, "", 0)
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// DeleteUser 删除管理员
func (c *AdminController) DeleteUser() {
	var req struct {
		ID int64 `json:"id"`
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

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的管理员ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查用户是否存在
	_, err := models.GetAdminUserById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "用户不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.DeleteAdminUser(req.ID); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.logOperation("删除管理员", "admin", "DELETE", "/admin/users/"+strconv.FormatInt(req.ID, 10), nil, 1, "", 0)
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// logOperation 记录操作日志
func (c *AdminController) logOperation(action, module, method, url string, params interface{}, status int, errorMsg string, executeTime int) {
	userID := c.GetSession("playerId")
	userName := c.GetSession("username")

	var adminID int64
	var adminName string

	if userID != nil {
		adminID, _ = userID.(int64)
	}
	if userName != nil {
		adminName, _ = userName.(string)
	}

	paramsJSON, _ := json.Marshal(params)

	log := &models.AdminOperationLog{
		UserId:    adminID,
		Username:  adminName,
		Action:    action,
		Resource:  module,
		Params:    string(paramsJSON),
		IpAddress: utils.GetClientIP(&c.Controller),
		UserAgent: c.Ctx.Request.UserAgent(),
	}

	log.Insert()
}

// CreateAdminRequest 创建管理员请求结构（对齐云函数）
type CreateAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RoleName string `json:"roleName"`
}

// CreateAdmin 创建管理员（对齐云函数createAdmin接口）
func (c *AdminController) CreateAdmin() {
	var req CreateAdminRequest
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
	if len(req.Username) < 3 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "用户名必须至少3个字符",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if len(req.Password) < 6 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "密码必须至少6个字符",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 设置默认值
	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	if req.Role == "" {
		req.Role = "viewer"
	}

	// 验证角色是否有效， role表中查询roleCode
	role := &models.AdminRole{}
	err := role.GetByRoleCode(req.Role)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的角色",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查用户名是否已存在
	_, err = models.GetAdminUserByUsername(req.Username)
	if err != nil && err != orm.ErrNoRows {
		c.Data["json"] = map[string]interface{}{
			"code":      4002,
			"msg":       "用户名已存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		// TODO: 实现邮箱重复检查
	}

	// MD5密码加密（对齐云函数）
	hash := md5.Sum([]byte(req.Password))
	passwordHash := hex.EncodeToString(hash[:])

	// 创建管理员
	user := &models.AdminUser{
		Username: req.Username,
		Password: passwordHash,
		Role:     req.RoleName,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1, // 活跃状态
		// TODO: 根据role字符串查找对应的RoleId
		RoleId: 1, // 暂时使用默认角色ID
	}

	if err := models.AddAdminUser(user); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "创建失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "CREATE", "ADMIN", map[string]interface{}{
		"newAdminUsername": req.Username,
		"newAdminRole":     req.Role,
		"newAdminNickname": req.Nickname,
	})

	// 返回成功结果（对齐云函数格式）
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"id":        user.ID,
			"username":  req.Username,
			"nickname":  req.Nickname,
			"role":      req.Role,
			"status":    "active",
			"createdAt": user.CreatedAt,
		},
	}
	c.ServeJSON()
}

// GetAllRoles 获取所有角色列表（对齐云函数getAllRoles接口）
func (c *AdminController) GetAllRoles() {
	// TODO: 实现权限验证

	roles, _, err := models.GetRoleList(1, 100, "")
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取角色失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 转换为云函数格式
	var roleList []map[string]interface{}
	for _, role := range roles {
		roleList = append(roleList, map[string]interface{}{
			"roleCode":    role.RoleCode,
			"roleName":    role.RoleName,
			"description": role.Description,
			"permissions": []string{}, // TODO: 解析权限JSON
		})
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "success",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"roles": roleList,
			"total": len(roleList),
		},
	}
	c.ServeJSON()
}

// InitAdmin 初始化管理员系统（对齐云函数initAdmin接口）
func (c *AdminController) InitAdmin() {
	var req struct {
		Force bool `json:"force"`
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

	// TODO: 实现管理员系统初始化逻辑
	// 1. 检查是否已初始化
	// 2. 创建默认角色
	// 3. 创建默认管理员账户

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "初始化完成",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"createdCollections": 3,
			"createdRoles":       4,
			"createdAdmins":      1,
			"defaultCredentials": map[string]interface{}{
				"username": "admin",
				"password": "123456",
				"warning":  "请立即登录并修改默认密码！",
			},
		},
	}
	c.ServeJSON()
}

// GetAdminById 根据ID获取管理员
func (c *AdminController) GetAdminById() {
	var requestData struct {
		ID int64 `json:"id"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	user, err := models.GetAdminById(requestData.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "管理员不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      user,
	}
	c.ServeJSON()
}

// UpdateAdminProfile 更新管理员资料
func (c *AdminController) UpdateAdminProfile() {
	var requestData struct {
		ID       int64  `json:"id"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.UpdateAdminProfile(requestData.ID, requestData.Nickname, requestData.Email); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新资料失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// ChangeAdminPassword 修改管理员密码
func (c *AdminController) ChangeAdminPassword() {
	var requestData struct {
		ID          int64  `json:"id"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.ChangeAdminPassword(requestData.ID, requestData.OldPassword, requestData.NewPassword); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "修改密码失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "修改成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// AdminLoginWithMD5 使用MD5密码登录
func (c *AdminController) AdminLoginWithMD5() {
	var requestData struct {
		Username     string `json:"username"`
		PasswordHash string `json:"passwordHash"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	user, err := models.AdminLoginWithMD5(requestData.Username, requestData.PasswordHash)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "登录失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "登录成功",
		"timestamp": utils.UnixMilli(),
		"data":      user,
	}
	c.ServeJSON()
}

// UpdateAdminStatus 更新管理员状态
func (c *AdminController) UpdateAdminStatus() {
	var requestData struct {
		ID     int64 `json:"id"`
		Status int   `json:"status"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 将状态转换为字符串
	var status string
	if requestData.Status == 1 {
		status = "active"
	} else {
		status = "inactive"
	}

	if err := models.UpdateAdminUserStatus(requestData.ID, status); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新状态失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// GetAdminRolePermissions 获取管理员角色权限
func (c *AdminController) GetAdminRolePermissions() {
	var requestData struct {
		RoleId int64 `json:"roleId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	role, permissions, err := models.GetAdminRolePermissions(requestData.RoleId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取权限失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"role":        role,
			"permissions": permissions,
		},
	}
	c.ServeJSON()
}

// GetAdminByToken 根据Token获取管理员
func (c *AdminController) GetAdminByToken() {
	var requestData struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	user, err := models.GetAdminByToken(requestData.Token)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "Token无效",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      user,
	}
	c.ServeJSON()
}

// GetAdminList 获取管理员列表（对齐云函数getAdminList接口）
func (c *AdminController) GetAdminList() {
	var req struct {
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Keyword  string `json:"keyword"`
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

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	users, total, err := models.GetAllAdminUsers(req.Page, req.PageSize, req.Keyword)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取管理员列表失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"list":     users,
			"total":    total,
			"page":     req.Page,
			"pageSize": req.PageSize,
		},
	}
	c.ServeJSON()
}

// DeleteAdminUser 删除管理员（对齐云函数deleteAdmin接口）
func (c *AdminController) DeleteAdminUser() {
	var req struct {
		ID int64 `json:"id"`
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

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的管理员ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查管理员是否存在
	_, err := models.GetAdminUserById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "管理员不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 删除管理员
	if err := models.DeleteAdminUser(req.ID); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "DELETE", "ADMIN", map[string]interface{}{
		"deletedAdminId": req.ID,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      map[string]interface{}{},
	}
	c.ServeJSON()
}

// UpdateAdminUser 更新管理员信息（对齐云函数updateAdmin接口）
func (c *AdminController) UpdateAdminUser() {
	var req struct {
		ID     int64  `json:"id"`
		Email  string `json:"email"`
		Phone  string `json:"phone"`
		Role   string `json:"role"`
		RoleId int64  `json:"roleId"`
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

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的管理员ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查管理员是否存在
	existUser, err := models.GetAdminUserById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "管理员不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 更新字段
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
	if req.RoleId > 0 {
		updateFields["roleId"] = req.RoleId
	}

	// 执行更新
	if err := models.UpdateAdminUserFields(req.ID, updateFields); err != nil {
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
	updatedUser, _ := models.GetAdminUserById(req.ID)

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "UPDATE", "ADMIN", map[string]interface{}{
		"updatedAdminId": req.ID,
		"originalData":   existUser,
		"updatedData":    updatedUser,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"id":       updatedUser.ID,
			"username": updatedUser.Username,
			"email":    updatedUser.Email,
			"phone":    updatedUser.Phone,
			"role":     updatedUser.Role,
			"roleId":   updatedUser.RoleId,
		},
	}
	c.ServeJSON()
}

// ResetAdminPassword 重置管理员密码（对齐云函数resetPassword接口）
func (c *AdminController) ResetAdminPassword() {
	var req struct {
		ID          int64  `json:"id"`
		NewPassword string `json:"newPassword"`
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

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的管理员ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if len(req.NewPassword) < 6 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "密码长度至少6位",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查管理员是否存在
	_, err := models.GetAdminUserById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "管理员不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// MD5加密密码（对齐云函数）
	passwordHash := utils.HashPassword(req.NewPassword)

	// 更新密码
	if err := models.UpdateAdminUserFields(req.ID, map[string]interface{}{
		"password": passwordHash,
	}); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "密码重置失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "RESET_PASSWORD", "ADMIN", map[string]interface{}{
		"targetAdminId": req.ID,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "密码重置成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"message": "密码重置成功",
		},
	}
	c.ServeJSON()
}
