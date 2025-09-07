package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
)

type PermissionController struct {
	web.Controller
}

// GetRoles 获取角色列表（参考getRoleList云函数）
func (c *PermissionController) GetRoles() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "20")
	roleName := c.GetString("roleName", "") // 对齐云函数参数名

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 20 // 对齐云函数默认值
	}

	// 获取角色列表
	roles, total, err := models.GetRoles(page, pageSize, roleName)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "获取角色列表失败: "+err.Error(), nil)
		return
	}

	// 对齐云函数返回格式
	result := map[string]interface{}{
		"list":     roles,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(&c.Controller, "success", result)
}

// GetRole 获取角色详情
func (c *PermissionController) GetRole() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取角色ID
	roleIdStr := c.Ctx.Input.Param(":id")
	if roleIdStr == "" {
		utils.ErrorResponse(&c.Controller, 4001, "角色ID不能为空", nil)
		return
	}

	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "角色ID格式错误", nil)
		return
	}

	// 获取角色详情
	role, err := models.GetRoleWithPermissions(roleId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "获取角色详情失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", role)
}

// CreateRole 创建角色（参考createRole云函数）
func (c *PermissionController) CreateRole() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	roleCode := c.GetString("roleCode")
	roleName := c.GetString("roleName")
	description := c.GetString("description")
	permissionsStr := c.GetString("permissions")

	// 参数校验
	if roleCode == "" || len(roleCode) < 2 {
		utils.ErrorResponse(&c.Controller, 4001, "角色代码必须至少2个字符", nil)
		return
	}

	if roleName == "" || len(roleName) < 2 {
		utils.ErrorResponse(&c.Controller, 4001, "角色名称必须至少2个字符", nil)
		return
	}

	// 解析权限数组
	var permissions []string
	if permissionsStr != "" {
		err := json.Unmarshal([]byte(permissionsStr), &permissions)
		if err != nil {
			utils.ErrorResponse(&c.Controller, 4001, "权限列表格式错误", nil)
			return
		}
	}

	// 验证权限列表
	validPermissions := []string{
		"admin_manage", "role_manage", "app_manage", "user_manage",
		"leaderboard_manage", "mail_manage", "stats_view", "system_config",
	}

	for _, permission := range permissions {
		valid := false
		for _, validPerm := range validPermissions {
			if permission == validPerm {
				valid = true
				break
			}
		}
		if !valid {
			utils.ErrorResponse(&c.Controller, 4001, "无效的权限: "+permission, nil)
			return
		}
	}

	// 创建角色
	role := &models.AdminRole{
		RoleCode:    roleCode,
		RoleName:    roleName,
		Description: description,
		Permissions: permissionsStr,
		Status:      1,
	}
	role.CreatedAt = time.Now()
	role.UpdatedAt = role.CreatedAt

	err := models.CreateRole(role)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			utils.ErrorResponse(&c.Controller, 4002, "角色代码已存在", nil)
		} else {
			utils.ErrorResponse(&c.Controller, 5001, "创建角色失败: "+err.Error(), nil)
		}
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "创建角色", "创建角色: "+roleName)

	// 对齐云函数返回格式
	result := map[string]interface{}{
		"id":          role.ID,
		"roleCode":    roleCode,
		"roleName":    roleName,
		"description": description,
		"permissions": permissions,
		"createdAt":   role.CreatedAt,
	}

	utils.SuccessResponse(&c.Controller, "创建成功", result)
}

// UpdateRole 更新角色（参考updateRole云函数）
func (c *PermissionController) UpdateRole() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取角色代码（从路径参数或请求体）
	roleCode := c.Ctx.Input.Param(":code")
	if roleCode == "" {
		roleCode = c.GetString("roleCode")
	}

	if roleCode == "" {
		utils.ErrorResponse(&c.Controller, 4001, "角色代码不能为空", nil)
		return
	}

	// 获取参数
	roleName := c.GetString("roleName")
	description := c.GetString("description")
	permissionsStr := c.GetString("permissions")
	status := c.GetString("status")

	// 检查是否为系统预设角色
	systemRoles := []string{"super_admin", "admin", "operator", "viewer"}
	for _, sysRole := range systemRoles {
		if roleCode == sysRole {
			utils.ErrorResponse(&c.Controller, 4005, "不能修改系统预设角色", nil)
			return
		}
	}

	// 验证权限列表（如果提供）
	if permissionsStr != "" {
		var permissions []string
		err := json.Unmarshal([]byte(permissionsStr), &permissions)
		if err != nil {
			utils.ErrorResponse(&c.Controller, 4001, "权限列表格式错误", nil)
			return
		}

		validPermissions := []string{
			"admin_manage", "role_manage", "app_manage", "user_manage",
			"leaderboard_manage", "mail_manage", "stats_view", "system_config",
			"counter_manage",
		}

		for _, permission := range permissions {
			valid := false
			for _, validPerm := range validPermissions {
				if permission == validPerm {
					valid = true
					break
				}
			}
			if !valid {
				utils.ErrorResponse(&c.Controller, 4001, "无效的权限: "+permission, nil)
				return
			}
		}
	}

	// 获取现有角色
	role := &models.AdminRole{}
	err := role.GetByRoleCode(roleCode)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 4004, "角色不存在", nil)
		return
	}

	// 更新字段
	if roleName != "" {
		role.RoleName = roleName
	}
	if description != "" {
		role.Description = description
	}
	if permissionsStr != "" {
		role.Permissions = permissionsStr
	}
	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		role.Status = statusInt
	}
	role.UpdatedAt = time.Now()

	// 更新角色
	err = models.UpdateRole(role)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "更新角色失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "更新角色", "更新角色: "+roleCode)

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// DeleteRole 删除角色（参考deleteRole云函数）
func (c *PermissionController) DeleteRole() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取角色代码
	roleCode := c.Ctx.Input.Param(":code")
	if roleCode == "" {
		roleCode = c.GetString("roleCode")
	}

	if roleCode == "" {
		utils.ErrorResponse(&c.Controller, 4001, "角色代码不能为空", nil)
		return
	}

	// 检查是否为系统预设角色
	systemRoles := []string{"super_admin", "admin", "operator", "viewer"}
	for _, sysRole := range systemRoles {
		if roleCode == sysRole {
			utils.ErrorResponse(&c.Controller, 4005, "不能删除系统预设角色", nil)
			return
		}
	}

	// 获取角色信息
	role := &models.AdminRole{}
	err := role.GetByRoleCode(roleCode)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 4004, "角色不存在", nil)
		return
	}

	// 检查角色是否被使用
	inUse := models.IsRoleInUse(role.ID)
	if inUse {
		utils.ErrorResponse(&c.Controller, 4006, "角色正在使用中，不能删除", nil)
		return
	}

	// 删除角色
	err = models.DeleteRole(role.ID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "删除角色失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "删除角色", "删除角色: "+roleCode)

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}

// GetPermissions 获取权限列表
func (c *PermissionController) GetPermissions() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取权限列表
	permissions, err := models.GetAllPermissions()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "获取权限列表失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", permissions)
}

// GetPermissionTree 获取权限树
func (c *PermissionController) GetPermissionTree() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取权限树
	permissionTree, err := models.GetPermissionTree()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "获取权限树失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", permissionTree)
}

// CreatePermission 创建权限
func (c *PermissionController) CreatePermission() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	permissionName := c.GetString("permissionName")
	permissionCode := c.GetString("permissionCode")
	parentIdStr := c.GetString("parentId", "0")
	sortOrder := c.GetString("sortOrder", "0")
	description := c.GetString("description")

	if permissionName == "" || permissionCode == "" {
		utils.ErrorResponse(&c.Controller, 1002, "权限名称和权限代码不能为空", nil)
		return
	}

	parentId, _ := strconv.Atoi(parentIdStr)
	sortOrderInt, _ := strconv.Atoi(sortOrder)

	// 创建权限
	permission := &models.Permission{
		Code:        permissionCode,
		Name:        permissionName,
		Description: description,
		ParentId:    int64(parentId),
		Sort:        sortOrderInt,
		Status:      1,
	}
	err := models.CreatePermission(permission)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "创建权限失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "创建权限", "创建权限: "+permissionName)

	result := map[string]interface{}{
		"permissionId": permission.ID,
	}

	utils.SuccessResponse(&c.Controller, "创建成功", result)
}

// UpdatePermission 更新权限
func (c *PermissionController) UpdatePermission() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取权限ID
	permissionIdStr := c.Ctx.Input.Param(":id")
	if permissionIdStr == "" {
		utils.ErrorResponse(&c.Controller, 1002, "权限ID不能为空", nil)
		return
	}

	permissionId, err := strconv.Atoi(permissionIdStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "权限ID格式错误", nil)
		return
	}

	// 获取参数
	permissionName := c.GetString("permissionName")
	permissionCode := c.GetString("permissionCode")
	parentIdStr := c.GetString("parentId", "0")
	sortOrder := c.GetString("sortOrder", "0")
	description := c.GetString("description")

	if permissionName == "" || permissionCode == "" {
		utils.ErrorResponse(&c.Controller, 1002, "权限名称和权限代码不能为空", nil)
		return
	}

	parentId, _ := strconv.Atoi(parentIdStr)
	sortOrderInt, _ := strconv.Atoi(sortOrder)

	// 更新权限
	permission := &models.Permission{
		BaseModel:   models.BaseModel{ID: int64(permissionId)},
		Code:        permissionCode,
		Name:        permissionName,
		Description: description,
		ParentId:    int64(parentId),
		Sort:        sortOrderInt,
		Status:      1,
	}
	err = models.UpdatePermission(permission)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新权限失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "更新权限", "更新权限: "+permissionName)

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// DeletePermission 删除权限
func (c *PermissionController) DeletePermission() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取权限ID
	permissionIdStr := c.Ctx.Input.Param(":id")
	if permissionIdStr == "" {
		utils.ErrorResponse(&c.Controller, 1002, "权限ID不能为空", nil)
		return
	}

	permissionId, err := strconv.Atoi(permissionIdStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "权限ID格式错误", nil)
		return
	}

	// 检查权限是否有子权限
	if models.HasChildPermissions(int64(permissionId)) {
		utils.ErrorResponse(&c.Controller, 1003, "权限存在子权限，无法删除", nil)
		return
	}

	// 检查权限是否被角色使用
	if models.IsPermissionInUse(int64(permissionId)) {
		utils.ErrorResponse(&c.Controller, 1003, "权限正在使用中，无法删除", nil)
		return
	}

	// 删除权限
	err = models.DeletePermission(int64(permissionId))
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "删除权限失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "删除权限", "删除权限ID: "+permissionIdStr)

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}
