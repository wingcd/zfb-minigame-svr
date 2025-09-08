package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
)

type StatsController struct {
	web.Controller
}

// GetDashboardStats 获取仪表板统计数据
func (c *StatsController) GetDashboardStats() {
	var requestData struct {
		AppId string `json:"appId"`
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

	stats, err := models.GetDashboardStats()
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取统计数据失败",
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

// GetTopApps 获取热门应用
func (c *StatsController) GetTopApps() {
	var requestData struct {
		Limit int `json:"limit"`
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

	if requestData.Limit <= 0 {
		requestData.Limit = 10
	}

	apps, err := models.GetTopApps(requestData.Limit)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取热门应用失败",
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
		"data":      apps,
	}
	c.ServeJSON()
}

// GetRecentActivity 获取最近活动
func (c *StatsController) GetRecentActivity() {
	var requestData struct {
		AppId string `json:"appId"`
		Limit int    `json:"limit"`
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

	if requestData.Limit <= 0 {
		requestData.Limit = 50
	}

	activities, err := models.GetRecentActivity(requestData.Limit)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取最近活动失败",
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
		"data":      activities,
	}
	c.ServeJSON()
}

// GetUserGrowth 获取用户增长统计
func (c *StatsController) GetUserGrowth() {
	var requestData struct {
		AppId string `json:"appId"`
		Days  int    `json:"days"`
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

	if requestData.Days <= 0 {
		requestData.Days = 30
	}

	growth, err := models.GetUserGrowth(requestData.AppId, requestData.Days)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取用户增长统计失败",
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
		"data":      growth,
	}
	c.ServeJSON()
}

// GetAppStats 获取应用统计
func (c *StatsController) GetAppStats() {
	var requestData struct {
		AppId string `json:"appId"`
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

	stats, err := models.GetAppStats(requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取应用统计失败",
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

// GetLeaderboardStats 获取排行榜统计
func (c *StatsController) GetLeaderboardStats() {
	var requestData struct {
		AppId string `json:"appId"`
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

	stats, err := models.GetLeaderboardStats(requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取排行榜统计失败",
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

// GetPlatformDistribution 获取平台分布统计
func (c *StatsController) GetPlatformDistribution() {
	var requestData struct {
		AppId string `json:"appId"`
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

	distribution, err := models.GetPlatformDistribution()
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取平台分布统计失败",
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
		"data":      distribution,
	}
	c.ServeJSON()
}
