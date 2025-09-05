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

// GetAllUsers 获取所有用户
func (c *UserController) GetAllUsers() {
	var requestData struct {
		AppId    string `json:"appId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Keyword  string `json:"keyword"`
		Status   string `json:"status"`
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

	users, total, err := models.GetUserList(requestData.Page, requestData.PageSize, requestData.Keyword, requestData.Status, requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取用户列表失败",
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
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// BanUser 封禁用户
func (c *UserController) BanUser() {
	var requestData struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
		Reason   string `json:"reason"`
		Duration int    `json:"duration"` // 封禁时长（小时），0表示永久
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

	if requestData.AppId == "" || requestData.PlayerId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取管理员ID（假设从JWT中获取）
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	if err := models.BanUser(requestData.AppId, requestData.PlayerId, claims.UserID, "temporary", requestData.Reason, requestData.Duration); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "封禁用户失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "封禁成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// UnbanUser 解封用户
func (c *UserController) UnbanUser() {
	var requestData struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
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

	// 获取管理员ID
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	if err := models.UnbanUser(requestData.AppId, requestData.PlayerId, claims.UserID, "管理员解封"); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "解封用户失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "解封成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser() {
	var requestData struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
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

	if err := models.DeleteUser(requestData.AppId, requestData.PlayerId); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除用户失败",
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

// GetUserDetail 获取用户详情
func (c *UserController) GetUserDetail() {
	var requestData struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
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

	user, err := models.GetUserDetail(requestData.AppId, requestData.PlayerId)
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

// SetUserDetail 设置用户详情
func (c *UserController) SetUserDetail() {
	var requestData struct {
		AppId    string                 `json:"appId"`
		PlayerId string                 `json:"playerId"`
		UserData map[string]interface{} `json:"userData"`
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

	// 将UserData转换为JSON字符串
	userDataJSON, err := json.Marshal(requestData.UserData)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "用户数据格式错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.SetUserDetail(requestData.AppId, requestData.PlayerId, string(userDataJSON)); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "设置用户详情失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "设置成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// GetUserStats 获取用户统计
func (c *UserController) GetUserStats() {
	var requestData struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
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

	stats, err := models.GetUserStats(requestData.AppId, requestData.PlayerId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取用户统计失败",
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
		"data":      stats,
	}
	c.ServeJSON()
}
