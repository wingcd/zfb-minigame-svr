package controllers

import (
	"game-service/models"
	"game-service/utils"
	"strconv"

	"github.com/beego/beego/v2/server/web"
)

type MailController struct {
	web.Controller
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
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	mailIdStr := c.GetString("mailId")
	if mailIdStr == "" {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	mailId, err := strconv.ParseInt(mailIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数格式错误", nil)
		return
	}

	// 读取邮件
	err = models.ReadMail(appId, userId, mailId)
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
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	mailIdStr := c.GetString("mailId")
	if mailIdStr == "" {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	mailId, err := strconv.ParseInt(mailIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数格式错误", nil)
		return
	}

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
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	mailIdStr := c.GetString("mailId")
	if mailIdStr == "" {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	mailId, err := strconv.ParseInt(mailIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数格式错误", nil)
		return
	}

	// 删除邮件
	err = models.DeleteMail(appId, userId, mailId)
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
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

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
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	mailIdStr := c.GetString("mailId")
	if mailIdStr == "" {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数不能为空", nil)
		return
	}

	status := c.GetString("status")
	if status == "" {
		utils.ErrorResponse(c.Ctx, 1002, "status参数不能为空", nil)
		return
	}

	mailId, err := strconv.ParseInt(mailIdStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "mailId参数格式错误", nil)
		return
	}

	// 根据状态执行相应操作
	switch status {
	case "read":
		// 标记为已读
		err = models.ReadMail(appId, userId, mailId)
		if err != nil {
			utils.ErrorResponse(c.Ctx, 1003, "标记邮件已读失败: "+err.Error(), nil)
			return
		}
	case "claimed":
		// 领取奖励
		_, err = models.ClaimRewards(appId, userId, mailId)
		if err != nil {
			utils.ErrorResponse(c.Ctx, 1003, "领取奖励失败: "+err.Error(), nil)
			return
		}
	case "deleted":
		// 删除邮件
		err = models.DeleteMail(appId, userId, mailId)
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
