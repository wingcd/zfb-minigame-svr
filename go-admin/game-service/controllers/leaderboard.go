package controllers

import (
	"game-service/models"
	"game-service/utils"
	"strconv"

	"github.com/beego/beego/v2/server/web"
)

type LeaderboardController struct {
	web.Controller
}

// SubmitScore 提交分数
func (c *LeaderboardController) SubmitScore() {
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	leaderboardName := c.GetString("leaderboardName")
	scoreStr := c.GetString("score")
	extraData := c.GetString("extraData")

	if leaderboardName == "" || scoreStr == "" {
		utils.ErrorResponse(c.Ctx, 1002, "leaderboardName和score参数不能为空", nil)
		return
	}

	// 转换分数
	score, err := strconv.ParseInt(scoreStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "score参数格式错误", nil)
		return
	}

	// 提交分数
	err = models.SubmitScore(appId, userId, leaderboardName, score, extraData)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "提交分数失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "提交成功", nil)
}

// GetLeaderboard 获取排行榜
func (c *LeaderboardController) GetLeaderboard() {
	// 验证签名
	appId, _, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	leaderboardName := c.GetString("leaderboardName")
	limitStr := c.GetString("limit", "10")

	if leaderboardName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "leaderboardName参数不能为空", nil)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	// 获取排行榜
	rankings, err := models.GetLeaderboard(appId, leaderboardName, limit)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取排行榜失败: "+err.Error(), nil)
		return
	}

	// 构建响应数据
	result := make([]map[string]interface{}, len(rankings))
	for i, ranking := range rankings {
		result[i] = map[string]interface{}{
			"rank":      i + 1,
			"userId":    ranking.UserId,
			"score":     ranking.Score,
			"extraData": ranking.ExtraData,
		}
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}

// GetUserRank 获取用户排名
func (c *LeaderboardController) GetUserRank() {
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	leaderboardName := c.GetString("leaderboardName")
	if leaderboardName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "leaderboardName参数不能为空", nil)
		return
	}

	// 获取用户排名
	rank, score, err := models.GetUserRank(appId, userId, leaderboardName)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取排名失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"rank":  rank,
		"score": score,
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}

// ResetLeaderboard 重置排行榜
func (c *LeaderboardController) ResetLeaderboard() {
	// 验证签名
	appId, _, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	leaderboardName := c.GetString("leaderboardName")
	if leaderboardName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "leaderboardName参数不能为空", nil)
		return
	}

	// 重置排行榜
	err = models.ResetLeaderboard(appId, leaderboardName)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "重置排行榜失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "重置成功", nil)
}

// ===== zy-sdk对齐接口 =====

// CommitScore 提交分数（zy-sdk接口）
func (c *LeaderboardController) CommitScore() {
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	leaderboardName := c.GetString("leaderboardName")
	scoreStr := c.GetString("score")
	extraData := c.GetString("extraData")

	if leaderboardName == "" || scoreStr == "" {
		utils.ErrorResponse(c.Ctx, 1002, "leaderboardName和score参数不能为空", nil)
		return
	}

	// 转换分数
	score, err := strconv.ParseInt(scoreStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "score参数格式错误", nil)
		return
	}

	// 提交分数
	err = models.SubmitScore(appId, userId, leaderboardName, score, extraData)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "提交分数失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "提交成功", nil)
}

// QueryTopRank 查询排行榜前几名（zy-sdk接口）
func (c *LeaderboardController) QueryTopRank() {
	// 验证签名
	appId, _, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	leaderboardName := c.GetString("leaderboardName")
	limitStr := c.GetString("limit", "10")

	if leaderboardName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "leaderboardName参数不能为空", nil)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	// 获取排行榜
	rankings, err := models.GetLeaderboard(appId, leaderboardName, limit)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取排行榜失败: "+err.Error(), nil)
		return
	}

	// 构建响应数据
	result := make([]map[string]interface{}, len(rankings))
	for i, ranking := range rankings {
		result[i] = map[string]interface{}{
			"rank":      i + 1,
			"userId":    ranking.UserId,
			"score":     ranking.Score,
			"extraData": ranking.ExtraData,
		}
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}
