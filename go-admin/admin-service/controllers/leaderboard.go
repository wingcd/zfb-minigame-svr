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

// GetAllLeaderboards 获取所有排行榜（对齐云函数getLeaderboards接口）
func (c *LeaderboardController) GetAllLeaderboards() {
	var req struct {
		AppId           string `json:"appId"`
		Page            int    `json:"page"`
		PageSize        int    `json:"pageSize"`
		LeaderboardName string `json:"leaderboardName"`
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

	leaderboards, total, err := models.GetLeaderboardList(req.AppId, req.Page, req.PageSize, req.LeaderboardName)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取排行榜列表失败: " + err.Error(),
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
			"page":       req.Page,
			"pageSize":   req.PageSize,
			"totalPages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		},
	}
	c.ServeJSON()
}

// CreateLeaderboard 创建排行榜（对齐云函数createLeaderboard接口）
func (c *LeaderboardController) CreateLeaderboard() {
	var req struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		ScoreType       string `json:"scoreType"`
		MaxRank         int    `json:"maxRank"`
		Category        string `json:"category"`
		ResetType       string `json:"resetType"`
		ResetValue      int    `json:"resetValue"`
		Enabled         bool   `json:"enabled"`
		UpdateStrategy  int    `json:"updateStrategy"`
		Sort            int    `json:"sort"`
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

	// 参数验证
	if req.AppId == "" || req.LeaderboardType == "" || req.Name == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	leaderboard := &models.LeaderboardConfig{
		AppId:           req.AppId,
		LeaderboardType: req.LeaderboardType,
		Name:            req.Name,
		Description:     req.Description,
		ScoreType:       req.ScoreType,
		MaxRank:         req.MaxRank,
		Category:        req.Category,
		ResetType:       req.ResetType,
		ResetValue:      req.ResetValue,
		Enabled:         req.Enabled,
		UpdateStrategy:  req.UpdateStrategy,
		Sort:            req.Sort,
	}

	if err := models.CreateLeaderboardConfig(leaderboard); err != nil {
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
				"msg":       "创建排行榜失败: " + err.Error(),
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "CREATE", "LEADERBOARD", map[string]interface{}{
		"appId":           req.AppId,
		"leaderboardName": req.Name,
		"type":            req.LeaderboardType,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data":      leaderboard,
	}
	c.ServeJSON()
}

// UpdateLeaderboard 更新排行榜配置（对齐云函数updateLeaderboard接口）
func (c *LeaderboardController) UpdateLeaderboard() {
	var req struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		ScoreType       string `json:"scoreType"`
		MaxRank         int    `json:"maxRank"`
		Category        string `json:"category"`
		ResetType       string `json:"resetType"`
		ResetValue      int    `json:"resetValue"`
		Enabled         bool   `json:"enabled"`
		UpdateStrategy  int    `json:"updateStrategy"`
		Sort            int    `json:"sort"`
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

	fields := make(map[string]interface{})
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.ScoreType != "" {
		fields["score_type"] = req.ScoreType
	}
	if req.MaxRank > 0 {
		fields["max_rank"] = req.MaxRank
	}
	if req.Category != "" {
		fields["category"] = req.Category
	}
	if req.ResetType != "" {
		fields["reset_type"] = req.ResetType
	}
	if req.ResetValue > 0 {
		fields["reset_value"] = req.ResetValue
	}
	fields["enabled"] = req.Enabled
	if req.UpdateStrategy >= 0 {
		fields["update_strategy"] = req.UpdateStrategy
	}
	if req.Sort >= 0 {
		fields["sort"] = req.Sort
	}

	if err := models.UpdateLeaderboard(req.AppId, req.LeaderboardType, fields); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新排行榜失败: " + err.Error(),
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
		"data":      map[string]interface{}{},
	}
	c.ServeJSON()
}

// DeleteLeaderboard 删除排行榜（对齐云函数deleteLeaderboard接口）
func (c *LeaderboardController) DeleteLeaderboard() {
	var req struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
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

	if err := models.DeleteLeaderboard(req.AppId, req.LeaderboardType); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除排行榜失败: " + err.Error(),
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
		"data":      map[string]interface{}{},
	}
	c.ServeJSON()
}

// GetLeaderboardData 获取排行榜数据
func (c *LeaderboardController) GetLeaderboardData() {
	var requestData struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
		Page            int    `json:"page"`
		PageSize        int    `json:"pageSize"`
		StartRank       int    `json:"startRank"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	// 设置默认值
	if requestData.Page <= 0 {
		requestData.Page = 1
	}
	if requestData.PageSize <= 0 {
		requestData.PageSize = 20
	}

	data, total, err := models.GetLeaderboardData(requestData.AppId, requestData.LeaderboardType, requestData.Page, requestData.PageSize)
	if err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeServerError, "获取排行榜数据失败", nil)
		return
	}

	result := map[string]interface{}{
		"list":       data,
		"total":      total,
		"page":       requestData.Page,
		"pageSize":   requestData.PageSize,
		"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
	}

	utils.SuccessResponse(&c.Controller, "success", result)
}

// UpdateLeaderboardScore 更新排行榜分数
func (c *LeaderboardController) UpdateLeaderboardScore() {
	var requestData struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
		PlayerId        string `json:"playerId"`
		Score           int64  `json:"score"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	if err := models.UpdateLeaderboardScore(requestData.AppId, requestData.LeaderboardType, requestData.PlayerId, requestData.Score); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeServerError, "更新分数失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", nil)
}

// DeleteLeaderboardScore 删除排行榜分数
func (c *LeaderboardController) DeleteLeaderboardScore() {
	var requestData struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
		PlayerId        string `json:"playerId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	if err := models.DeleteLeaderboardScore(requestData.AppId, requestData.LeaderboardType, requestData.PlayerId); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeServerError, "删除分数失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", nil)
}

// CommitLeaderboardScore 提交排行榜分数
func (c *LeaderboardController) CommitLeaderboardScore() {
	var requestData struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
		PlayerId        string `json:"playerId"`
		Score           int64  `json:"score"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	if err := models.CommitLeaderboardScore(requestData.AppId, requestData.LeaderboardType, requestData.PlayerId, requestData.Score); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeServerError, "提交分数失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", nil)
}

// QueryLeaderboardScore 查询排行榜分数
func (c *LeaderboardController) QueryLeaderboardScore() {
	var requestData struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
		PlayerId        string `json:"playerId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	score, rank, err := models.QueryLeaderboardScore(requestData.AppId, requestData.LeaderboardType, requestData.PlayerId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeServerError, "查询分数失败", nil)
		return
	}

	data := map[string]interface{}{
		"score": score,
		"rank":  rank,
	}

	utils.SuccessResponse(&c.Controller, "success", data)
}

// FixLeaderboardUserInfo 修复排行榜用户信息
func (c *LeaderboardController) FixLeaderboardUserInfo() {
	var requestData struct {
		AppId           string `json:"appId"`
		LeaderboardType string `json:"leaderboardType"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeBadRequest, "参数错误", nil)
		return
	}

	count, err := models.FixLeaderboardUserInfo(requestData.AppId, requestData.LeaderboardType)
	if err != nil {
		utils.ErrorResponse(&c.Controller, utils.CodeServerError, "修复用户信息失败", nil)
		return
	}

	data := map[string]interface{}{
		"fixedCount": count,
	}

	utils.SuccessResponse(&c.Controller, "success", data)
}
