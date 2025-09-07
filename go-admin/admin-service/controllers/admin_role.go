package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
)

type AdminRoleController struct {
	web.Controller
}

// GetRoleList 获取角色列表
func (c *AdminRoleController) GetRoleList() {
	var requestData struct {
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		RoleName string `json:"roleName"`
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

	// 设置默认值
	if requestData.Page <= 0 {
		requestData.Page = 1
	}
	if requestData.PageSize <= 0 {
		requestData.PageSize = 20
	}

	roles, total, err := models.GetRoleList(requestData.Page, requestData.PageSize, requestData.RoleName)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取角色列表失败",
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
			"list":       roles,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// CreateRole 创建角色
func (c *AdminRoleController) CreateRole() {
	var role models.AdminRole
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &role); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 参数验证
	if role.RoleName == "" || role.RoleCode == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := role.Insert(); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "创建角色失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data":      role,
	}
	c.ServeJSON()
}

// UpdateRole 更新角色
func (c *AdminRoleController) UpdateRole() {
	var role models.AdminRole
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &role); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 只更新允许更新的字段，不包括ID和CreatedAt
	if err := role.Update("role_code", "role_name", "description", "permissions", "status"); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新角色失败",
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
		"data":      role,
	}
	c.ServeJSON()
}

// DeleteRole 删除角色
func (c *AdminRoleController) DeleteRole() {
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

	if err := models.DeleteRole(requestData.ID); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除角色失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// GetRole 获取单个角色详情
func (c *AdminRoleController) GetRole() {
	var req struct {
		ID int64 `json:"id"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.ID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的角色ID",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取角色详情
	role := &models.AdminRole{}
	err := role.GetById(req.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "角色不存在",
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
			"id":          role.ID,
			"roleName":    role.RoleName,
			"roleCode":    role.RoleCode,
			"description": role.Description,
			"permissions": role.Permissions,
		},
	}
	c.ServeJSON()
}

// GetAllRoles 获取所有角色列表（对齐云函数getAllRoles接口）
func (c *AdminRoleController) GetAllRoles() {
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
		// 解析权限JSON
		permissions := []string{}
		if role.Permissions != "" {
			if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
				permissions = []string{}
			}
		}

		roleList = append(roleList, map[string]interface{}{
			"roleCode":    role.RoleCode,
			"roleName":    role.RoleName,
			"description": role.Description,
			"permissions": permissions,
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
