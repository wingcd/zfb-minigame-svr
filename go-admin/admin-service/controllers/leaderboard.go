package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
)

type LeaderboardController struct {
	web.Controller
}

// GetAllLeaderboards 获取所有排行榜
func (c *LeaderboardController) GetAllLeaderboards() {
	var requestData struct {
		AppId           string `json:"appId"`
		Page            int    `json:"page"`
		PageSize        int    `json:"pageSize"`
		LeaderboardName string `json:"leaderboardName"`
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
		requestData.PageSize = 10
	}

	leaderboards, total, err := models.GetLeaderboardList(requestData.AppId, requestData.Page, requestData.PageSize, requestData.LeaderboardName)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取排行榜列表失败",
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
			"list":       leaderboards,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// CreateLeaderboard 创建排行榜
func (c *LeaderboardController) CreateLeaderboard() {
	var requestData struct {
		AppId       string `json:"appId"`
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ResetType   string `json:"resetType"`
		MaxEntries  int    `json:"maxEntries"`
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

	// 参数验证
	if requestData.AppId == "" || requestData.Type == "" || requestData.Name == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	leaderboard := &models.Leaderboard{
		AppId:       requestData.AppId,
		Type:        requestData.Type,
		Name:        requestData.Name,
		Description: requestData.Description,
		ResetType:   requestData.ResetType,
		MaxEntries:  requestData.MaxEntries,
	}

	if err := models.CreateLeaderboard(leaderboard); err != nil {
		if err.Error() == "排行榜已存在" {
			c.Data["json"] = map[string]interface{}{
				"code":      4002,
				"msg":       err.Error(),
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
		} else {
			c.Data["json"] = map[string]interface{}{
				"code":      5001,
				"msg":       "创建排行榜失败",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data":      leaderboard,
	}
	c.ServeJSON()
}

// UpdateLeaderboard 更新排行榜配置
func (c *LeaderboardController) UpdateLeaderboard() {
	var requestData struct {
		AppId       string `json:"appId"`
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ResetType   string `json:"resetType"`
		MaxEntries  int    `json:"maxEntries"`
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

	fields := make(map[string]interface{})
	if requestData.Name != "" {
		fields["name"] = requestData.Name
	}
	if requestData.Description != "" {
		fields["description"] = requestData.Description
	}
	if requestData.ResetType != "" {
		fields["reset_type"] = requestData.ResetType
	}
	if requestData.MaxEntries > 0 {
		fields["max_entries"] = requestData.MaxEntries
	}

	if err := models.UpdateLeaderboard(requestData.AppId, requestData.Type, fields); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新排行榜失败",
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

// DeleteLeaderboard 删除排行榜
func (c *LeaderboardController) DeleteLeaderboard() {
	var requestData struct {
		AppId string `json:"appId"`
		Type  string `json:"type"`
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

	if err := models.DeleteLeaderboard(requestData.AppId, requestData.Type); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除排行榜失败",
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

// GetLeaderboardData 获取排行榜数据
func (c *LeaderboardController) GetLeaderboardData() {
	var requestData struct {
		AppId     string `json:"appId"`
		Type      string `json:"type"`
		Page      int    `json:"page"`
		PageSize  int    `json:"pageSize"`
		StartRank int    `json:"startRank"`
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

	data, total, err := models.GetLeaderboardData(requestData.AppId, requestData.Type, requestData.Page, requestData.PageSize)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取排行榜数据失败",
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
			"list":       data,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// UpdateLeaderboardScore 更新排行榜分数
func (c *LeaderboardController) UpdateLeaderboardScore() {
	var requestData struct {
		AppId    string `json:"appId"`
		Type     string `json:"type"`
		PlayerId string `json:"playerId"`
		Score    int64  `json:"score"`
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

	if err := models.UpdateLeaderboardScore(requestData.AppId, requestData.Type, requestData.PlayerId, requestData.Score); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新分数失败",
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

// DeleteLeaderboardScore 删除排行榜分数
func (c *LeaderboardController) DeleteLeaderboardScore() {
	var requestData struct {
		AppId    string `json:"appId"`
		Type     string `json:"type"`
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

	if err := models.DeleteLeaderboardScore(requestData.AppId, requestData.Type, requestData.PlayerId); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除分数失败",
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

// CommitLeaderboardScore 提交排行榜分数
func (c *LeaderboardController) CommitLeaderboardScore() {
	var requestData struct {
		AppId    string `json:"appId"`
		Type     string `json:"type"`
		PlayerId string `json:"playerId"`
		Score    int64  `json:"score"`
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

	if err := models.CommitLeaderboardScore(requestData.AppId, requestData.Type, requestData.PlayerId, requestData.Score); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "提交分数失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "提交成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// QueryLeaderboardScore 查询排行榜分数
func (c *LeaderboardController) QueryLeaderboardScore() {
	var requestData struct {
		AppId    string `json:"appId"`
		Type     string `json:"type"`
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

	score, rank, err := models.QueryLeaderboardScore(requestData.AppId, requestData.Type, requestData.PlayerId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "查询分数失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "查询成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"score": score,
			"rank":  rank,
		},
	}
	c.ServeJSON()
}

// FixLeaderboardUserInfo 修复排行榜用户信息
func (c *LeaderboardController) FixLeaderboardUserInfo() {
	var requestData struct {
		AppId string `json:"appId"`
		Type  string `json:"type"`
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

	count, err := models.FixLeaderboardUserInfo(requestData.AppId, requestData.Type)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "修复用户信息失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "修复成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"fixedCount": count,
		},
	}
	c.ServeJSON()
}
