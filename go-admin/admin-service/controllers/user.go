package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
)

type UserController struct {
	web.Controller
}

// GetAllUsers 获取所有用户（对齐云函数getUserList接口）
func (c *UserController) GetAllUsers() {
	var req struct {
		AppId    string `json:"appId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Status   string `json:"status"`
		PlayerId string `json:"playerId"`
		OpenId   string `json:"openId"`
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

	users, total, err := models.GetUserList(req.Page, req.PageSize, req.Status, req.AppId, req.PlayerId, req.OpenId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取用户列表失败: " + err.Error(),
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

// BanUser 封禁用户（对齐云函数banUser接口）
func (c *UserController) BanUser() {
	var req struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
		Reason   string `json:"reason"`
		Duration int    `json:"duration"` // 封禁时长（小时），0表示永久
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

	if req.AppId == "" || req.PlayerId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.BanUser(req.AppId, req.PlayerId, 0, "temporary", req.Reason, req.Duration); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "封禁用户失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "BAN", "USER", map[string]interface{}{
		"appId":    req.AppId,
		"playerId": req.PlayerId,
		"reason":   req.Reason,
		"duration": req.Duration,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "封禁成功",
		"timestamp": utils.UnixMilli(),
		"data":      map[string]interface{}{},
	}
	c.ServeJSON()
}

// UnbanUser 解封用户（对齐云函数unbanUser接口）
func (c *UserController) UnbanUser() {
	var req struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
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

	if err := models.UnbanUser(req.AppId, req.PlayerId, 0, "管理员解封"); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "解封用户失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "UNBAN", "USER", map[string]interface{}{
		"appId":    req.AppId,
		"playerId": req.PlayerId,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "解封成功",
		"timestamp": utils.UnixMilli(),
		"data":      map[string]interface{}{},
	}
	c.ServeJSON()
}

// DeleteUser 删除用户（对齐云函数deleteUser接口）
func (c *UserController) DeleteUser() {
	var req struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
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

	if err := models.DeleteUser(req.AppId, req.PlayerId); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除用户失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "DELETE", "USER", map[string]interface{}{
		"appId":    req.AppId,
		"playerId": req.PlayerId,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      map[string]interface{}{},
	}
	c.ServeJSON()
}

// GetUserDetail 获取用户详情
func (c *UserController) GetUserDetail() {
	var requestData struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	user, err := models.GetUserDetail(requestData.AppId, requestData.PlayerId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeNotFound, "用户不存在", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", user)
}

// SetUserDetail 设置用户详情
func (c *UserController) SetUserDetail() {
	var requestData struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
		UserData string `json:"userData"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	if err := models.SetUserDetail(requestData.AppId, requestData.PlayerId, requestData.UserData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeServerError, "设置用户详情失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", nil)
}

// GetUserStats 获取用户统计（应用级别统计，对齐云函数 getUserStats）
func (c *UserController) GetUserStats() {
	var requestData struct {
		AppId string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
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
	if requestData.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[appId]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 验证应用是否存在
	app := &models.Application{}
	if err := app.GetByAppId(requestData.AppId); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "应用不存在或用户表不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取应用用户统计
	stats, err := models.GetAppUserStats(requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "success",
		"timestamp": utils.UnixMilli(),
		"data":      stats,
	}
	c.ServeJSON()
}
