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

// GetMailList 获取邮件列表
func (c *MailController) GetMailList() {
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

	utils.SuccessResponse(c.Ctx, "删除成功", nil)
}
