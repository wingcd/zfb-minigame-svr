package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/server/web"
)

type MailController struct {
	web.Controller
}

// GetAllMails 获取所有邮件
func (c *MailController) GetAllMails() {
	var requestData struct {
		AppId    string `json:"appId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		UserId   string `json:"userId"`
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

	mails, total, err := models.GetMailList(requestData.AppId, requestData.Page, requestData.PageSize, requestData.UserId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取邮件列表失败",
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
			"list":       mails,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// CreateMail 创建邮件
func (c *MailController) CreateMail() {
	var requestData struct {
		AppId      string `json:"appId"`
		UserId     string `json:"userId"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Rewards    string `json:"rewards"`
		ExpireDays int    `json:"expireDays"`
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
	if requestData.AppId == "" || requestData.Title == "" || requestData.Content == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 设置过期时间
	var expireAt time.Time
	if requestData.ExpireDays > 0 {
		expireAt = time.Now().AddDate(0, 0, requestData.ExpireDays)
	}

	mail := &models.Mail{
		AppId:    requestData.AppId,
		UserId:   requestData.UserId,
		Title:    requestData.Title,
		Content:  requestData.Content,
		Rewards:  requestData.Rewards,
		Status:   0, // 未读
		ExpireAt: expireAt,
	}

	if err := models.CreateMail(mail); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "创建邮件失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data":      mail,
	}
	c.ServeJSON()
}

// UpdateMail 更新邮件
func (c *MailController) UpdateMail() {
	var requestData struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Rewards string `json:"rewards"`
		Status  int    `json:"status"`
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

	mail := &models.Mail{
		ID:      requestData.ID,
		Title:   requestData.Title,
		Content: requestData.Content,
		Rewards: requestData.Rewards,
		Status:  requestData.Status,
	}

	if err := models.UpdateMail(mail); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新邮件失败",
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
		"data":      mail,
	}
	c.ServeJSON()
}

// DeleteMail 删除邮件
func (c *MailController) DeleteMail() {
	var requestData struct {
		ID int64 `json:"id"`
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

	if err := models.DeleteMail(requestData.ID); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除邮件失败",
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

// PublishMail 发布邮件
func (c *MailController) PublishMail() {
	var requestData struct {
		AppId      string `json:"appId"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Rewards    string `json:"rewards"`
		ExpireDays int    `json:"expireDays"`
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

	// 设置默认过期天数
	if requestData.ExpireDays <= 0 {
		requestData.ExpireDays = 7
	}

	if err := models.PublishMail(requestData.AppId, requestData.Title, requestData.Content, requestData.Rewards, requestData.ExpireDays); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "发布邮件失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "发布成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// GetMailStats 获取邮件统计
func (c *MailController) GetMailStats() {
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

	stats, err := models.GetMailStats(requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取邮件统计失败",
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

// GetUserMails 获取用户邮件
func (c *MailController) GetUserMails() {
	var requestData struct {
		AppId    string `json:"appId"`
		UserId   string `json:"userId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
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

	mails, total, err := models.GetUserMails(requestData.AppId, requestData.UserId, requestData.Page, requestData.PageSize)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取用户邮件失败",
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
			"list":       mails,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// SendMail 发送邮件给特定用户
func (c *MailController) SendMail() {
	var requestData struct {
		AppId       string `json:"appId"`
		UserId      string `json:"userId"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		Attachments string `json:"attachments"`
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

	if err := models.SendMail(requestData.AppId, requestData.UserId, requestData.Title, requestData.Content, requestData.Attachments); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "发送邮件失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "发送成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// SendBroadcastMail 发送广播邮件
func (c *MailController) SendBroadcastMail() {
	var requestData struct {
		AppId       string `json:"appId"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		Attachments string `json:"attachments"`
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

	if err := models.SendBroadcastMail(requestData.AppId, requestData.Title, requestData.Content, requestData.Attachments); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "发送广播邮件失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "发送成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}
