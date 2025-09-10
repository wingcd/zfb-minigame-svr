package controllers

import (
	"encoding/json"
	"game-service/models"
	"game-service/utils"

	"github.com/beego/beego/v2/server/web"
)

type MailController struct {
	web.Controller
}

// ReadMailRequest 读取邮件请求
type ReadMailRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	MailId    int64  `json:"mailId"`
}

// ClaimRewardsRequest 领取奖励请求
type ClaimRewardsRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	MailId    int64  `json:"mailId"`
}

// DeleteMailRequest 删除邮件请求
type DeleteMailRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	MailId    int64  `json:"mailId"`
}

// GetUserMailsRequest 获取用户邮件请求
type GetUserMailsRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
}

// UpdateMailStatusRequest 更新邮件状态请求
type UpdateMailStatusRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	MailId    int64  `json:"mailId"`
	Status    string `json:"status"`
}

// parseRequest 解析请求参数
func (c *MailController) parseRequest(req interface{}) error {
	return json.Unmarshal(c.Ctx.Input.RequestBody, req)
}

// clearNewMailCache 清除用户新邮件缓存
func clearNewMailCache(appId, userId string) {
	if models.RedisClient != nil {
		newMailKey := "new_mail:" + appId + ":" + userId
		models.RedisClient.Del(models.RedisClient.Context(), newMailKey)
	}
}

// ReadMail 读取邮件
func (c *MailController) ReadMail() {
	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req ReadMailRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	if req.MailId == 0 {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	mailId := req.MailId
	userId := req.PlayerId

	// 读取邮件
	err := models.ReadMail(appId, userId, mailId)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "读取邮件失败: "+err.Error(), nil)
		return
	}

	// 清除新邮件缓存
	clearNewMailCache(appId, userId)

	utils.SuccessResponse(c.Ctx, "读取成功", nil)
}

// ClaimRewards 领取奖励
func (c *MailController) ClaimRewards() {
	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req ClaimRewardsRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	if req.MailId == 0 {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	mailId := req.MailId
	userId := req.PlayerId

	// 领取奖励
	rewards, err := models.ClaimRewards(appId, userId, mailId)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "领取奖励失败: "+err.Error(), nil)
		return
	}

	// 清除新邮件缓存
	clearNewMailCache(appId, userId)

	result := map[string]interface{}{
		"rewards": rewards,
	}

	utils.SuccessResponse(c.Ctx, "领取成功", result)
}

// DeleteMail 删除邮件
func (c *MailController) DeleteMail() {
	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req DeleteMailRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	if req.MailId == 0 {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	mailId := req.MailId
	userId := req.PlayerId

	// 删除邮件
	err := models.DeleteMail(appId, userId, mailId)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "删除邮件失败: "+err.Error(), nil)
		return
	}

	// 清除新邮件缓存
	clearNewMailCache(appId, userId)

	utils.SuccessResponse(c.Ctx, "删除成功", nil)
}

// ===== zy-sdk对齐接口 =====

// GetUserMails 获取用户邮件（zy-sdk接口）
func (c *MailController) GetUserMails() {
	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req GetUserMailsRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 设置默认值
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	userId := req.PlayerId

	// 获取邮件列表
	mails, total, err := models.GetMailList(appId, userId, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取邮件列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"mails":    mails,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}

// UpdateMailStatus 更新邮件状态（zy-sdk接口）
func (c *MailController) UpdateMailStatus() {
	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req UpdateMailStatusRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	if req.MailId == 0 {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	if req.Status == "" {
		utils.ErrorResponse(c.Ctx, 1002, "status参数不能为空", nil)
		return
	}

	mailId := req.MailId
	status := req.Status
	userId := req.PlayerId

	// 根据状态执行相应操作
	switch status {
	case "read":
		// 标记为已读
		err := models.ReadMail(appId, userId, mailId)
		if err != nil {
			utils.ErrorResponse(c.Ctx, 1003, "标记邮件已读失败: "+err.Error(), nil)
			return
		}
	case "claimed":
		// 领取奖励
		_, err := models.ClaimRewards(appId, userId, mailId)
		if err != nil {
			utils.ErrorResponse(c.Ctx, 1003, "领取奖励失败: "+err.Error(), nil)
			return
		}
	case "deleted":
		// 删除邮件
		err := models.DeleteMail(appId, userId, mailId)
		if err != nil {
			utils.ErrorResponse(c.Ctx, 1003, "删除邮件失败: "+err.Error(), nil)
			return
		}
	default:
		utils.ErrorResponse(c.Ctx, 1002, "无效的状态参数", nil)
		return
	}

	// 清除新邮件缓存
	clearNewMailCache(appId, userId)

	utils.SuccessResponse(c.Ctx, "状态更新成功", nil)
}
