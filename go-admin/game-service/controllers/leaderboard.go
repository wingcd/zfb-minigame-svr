package controllers

import (
	"encoding/json"
	"game-service/models"
	"game-service/utils"
	"strings"

	"github.com/beego/beego/v2/server/web"
)

type LeaderboardController struct {
	web.Controller
}

// SubmitScoreRequest 提交分数请求
type SubmitScoreRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	Type      string `json:"type"`
	Score     int64  `json:"score"`
	ExtraData string `json:"extraData"`
}

// GetLeaderboardRequest 获取排行榜请求
type GetLeaderboardRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	Type      string `json:"type"`
	Limit     int    `json:"limit"`
}

// GetUserRankRequest 获取用户排名请求
type GetUserRankRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	Type      string `json:"type"`
}

// ResetLeaderboardRequest 重置排行榜请求
type ResetLeaderboardRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	Type      string `json:"type"`
}

// parseRequest 解析请求参数
func (c *LeaderboardController) parseRequest(req interface{}) error {
	return json.Unmarshal(c.Ctx.Input.RequestBody, req)
}

// CommitScore 提交分数（zy-sdk接口）
func (c *LeaderboardController) CommitScore() {
	// 解析请求参数
	var req SubmitScoreRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 参数验证
	if req.AppId == "" {
		utils.ErrorResponse(c.Ctx, 1002, "appId参数不能为空", nil)
		return
	}
	if req.PlayerId == "" {
		utils.ErrorResponse(c.Ctx, 1002, "playerId参数不能为空", nil)
		return
	}
	if req.Type == "" {
		utils.ErrorResponse(c.Ctx, 1002, "type参数不能为空", nil)
		return
	}

	leaderboardName := req.Type
	score := req.Score
	extraData := req.ExtraData
	userId := req.PlayerId

	// 提交分数（包含用户验证、更新策略、重置检查等）
	err := models.SubmitScore(req.AppId, userId, leaderboardName, score, extraData)
	if err != nil {
		// 根据错误类型返回不同的错误码
		errorCode := 1003
		if strings.Contains(err.Error(), "用户不存在") {
			errorCode = 1004
		} else if strings.Contains(err.Error(), "更新策略异常") {
			errorCode = 1001
		}
		utils.ErrorResponse(c.Ctx, errorCode, "提交分数失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "提交成功", nil)
}

// QueryTopRank 查询排行榜前几名（zy-sdk接口）
func (c *LeaderboardController) QueryTopRank() {
	// 解析请求参数
	var req GetLeaderboardRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 参数验证
	if req.AppId == "" {
		utils.ErrorResponse(c.Ctx, 1002, "appId参数不能为空", nil)
		return
	}
	if req.Type == "" {
		utils.ErrorResponse(c.Ctx, 1002, "type参数不能为空", nil)
		return
	}

	leaderboardName := req.Type
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// 获取排行榜（包含重置检查和用户信息）
	rankings, err := models.GetLeaderboard(req.AppId, leaderboardName, limit)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取排行榜失败: "+err.Error(), nil)
		return
	}

	// 构建响应数据（按照JS版本的格式）
	resultList := make([]map[string]interface{}, len(rankings))
	for i, ranking := range rankings {
		item := map[string]interface{}{
			"playerId": ranking.UserId,
			"score":    ranking.Score,
			"userInfo": ranking.UserInfo,
		}
		if ranking.ExtraData != "" {
			item["extraData"] = ranking.ExtraData
		}
		resultList[i] = item
	}

	// 返回结果（按照JS版本格式）
	result := map[string]interface{}{
		"type":  leaderboardName,
		"count": len(resultList),
		"list":  resultList,
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}
