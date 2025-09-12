package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type MailController struct {
	web.Controller
}

// InitMailSystem 初始化邮件系统
func (c *MailController) InitMailSystem() {
	var requestData struct {
		AppId string `json:"appId"`
		Force bool   `json:"force"`
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

	if requestData.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "appId不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 邮件表已在应用创建时创建，这里只需要验证表是否存在
	mail := &models.MailSystem{}
	tableName := mail.GetTableName(requestData.AppId)

	// 检查邮件表是否存在
	o := orm.NewOrm()
	var count int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&count)
	if err != nil || count == 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "邮件表不存在，请先创建应用",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "邮件系统初始化完成",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"createdTables": 1,
			"warning":       "邮件系统已成功初始化！",
		},
	}
	c.ServeJSON()
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
		CreateBy   int64  `json:"createBy"`
		AppId      string `json:"appId"`
		Type       string `json:"type"`
		TargetType string `json:"targetType"`
		UserId     string `json:"userId"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Rewards    string `json:"rewards"`
		ExpireDays int    `json:"expireDays"`
		MailType   int    `json:"mailType"` // 0: 个人邮件, 1: 系统广播邮件
		Status     string `json:"status"`
		ExpireTime string `json:"expireTime"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		log.Println("参数错误", err)
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

	var rewardsString = requestData.Rewards

	// 使用新的邮件系统
	if requestData.MailType == 1 || requestData.UserId == "" {
		// 系统广播邮件
		// 调试信息
		fmt.Printf("DEBUG: 收到创建邮件请求，状态参数: %s\n", requestData.Status)

		systemMail := &models.MailSystem{
			AppId:      requestData.AppId,
			Title:      requestData.Title,
			Content:    requestData.Content,
			Rewards:    rewardsString,
			Type:       requestData.Type,
			TargetType: requestData.TargetType,
			Status:     requestData.Status,
			CreatedBy:  "admin", // 默认创建者为admin
		}

		// 设置创建者
		// 查询用户
		user, err := models.GetAdminUserById(requestData.CreateBy)
		if err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      5001,
				"msg":       "查询用户失败",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}
		systemMail.CreatedBy = user.Username

		if requestData.ExpireDays > 0 {
			expireTime := time.Now().AddDate(0, 0, requestData.ExpireDays)
			systemMail.ExpireTime = &expireTime
		}

		if err := models.CreateSystemMail(systemMail); err != nil {
			fmt.Printf("DEBUG: CreateSystemMail failed: %v\n", err)
			c.Data["json"] = map[string]interface{}{
				"code":      5001,
				"msg":       "创建系统邮件失败",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}

		c.Data["json"] = map[string]interface{}{
			"code":      0,
			"msg":       "系统邮件创建并发布成功",
			"timestamp": utils.UnixMilli(),
			"data":      systemMail,
		}
	} else {
		// 个人邮件 - 使用SendPersonalMail
		if err := models.SendPersonalMail(requestData.AppId, requestData.UserId, requestData.Title, requestData.Content, requestData.Rewards); err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      5001,
				"msg":       "发送个人邮件失败",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}

		c.Data["json"] = map[string]interface{}{
			"code":      0,
			"msg":       "个人邮件发送成功",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
	}
	c.ServeJSON()
}

// UpdateMail 更新邮件（支持动态邮件系统）
func (c *MailController) UpdateMail() {
	var requestData struct {
		ID          int64  `json:"id"` // 添加ID字段
		AppId       string `json:"appId"`
		Type        string `json:"type"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		Rewards     string `json:"rewards"`
		Status      string `json:"status"`
		TargetType  string `json:"targetType"`
		ExpireDays  int    `json:"expireDays"`
		PublishTime int64  `json:"publishTime"`
		ExpireTime  int64  `json:"expireTime"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		log.Println("参数错误", err)
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
	if requestData.AppId == "" || requestData.ID == 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数：appId 和 id",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 构造更新的邮件对象
	mail := &models.MailSystem{
		ID:    requestData.ID, // 设置邮件ID
		AppId: requestData.AppId,
	}

	var oldMail *models.MailSystem
	oldMail, err := models.GetMailById(requestData.AppId, requestData.ID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code": 5001,
			"msg":  "获取邮件失败",
		}
		c.ServeJSON()
		return
	}

	// 设置发布时间
	if requestData.PublishTime != 0 {
		// 将字符串转换为时间, 时间戳转换为时间
		publishTime := time.Unix(requestData.PublishTime, 0)
		mail.SendTime = &publishTime
	} else {
		mail.SendTime = oldMail.SendTime
	}

	// 设置过期时间
	if requestData.ExpireTime != 0 {
		expireTime := time.Unix(requestData.ExpireTime, 0)
		mail.ExpireTime = &expireTime
	} else {
		// 设置过期时间
		if requestData.ExpireDays > 0 {
			expireTime := time.Now().AddDate(0, 0, requestData.ExpireDays)
			mail.ExpireTime = &expireTime
		} else {
			mail.ExpireTime = oldMail.ExpireTime
		}
	}

	if requestData.Title != "" {
		mail.Title = requestData.Title
	} else {
		mail.Title = oldMail.Title
	}
	if requestData.Content != "" {
		mail.Content = requestData.Content
	} else {
		mail.Content = oldMail.Content
	}
	if requestData.Rewards != "" {
		mail.Rewards = requestData.Rewards
	} else {
		mail.Rewards = oldMail.Rewards
	}
	if requestData.Status != "" {
		mail.Status = requestData.Status
	} else {
		mail.Status = oldMail.Status
	}
	if requestData.TargetType != "" {
		mail.TargetType = requestData.TargetType
	} else {
		mail.TargetType = oldMail.TargetType
	}
	if requestData.Type != "" {
		mail.Type = requestData.Type
	} else {
		mail.Type = oldMail.Type
	}

	if err := models.UpdateSystemMail(mail); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新邮件失败: " + err.Error(),
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

// DeleteMail 删除邮件（支持动态邮件系统）
func (c *MailController) DeleteMail() {
	var requestData struct {
		AppId string `json:"appId"`
		ID    int64  `json:"id"`
		Type  string `json:"type"` // "system" 或 "personal"
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
	if requestData.AppId == "" || requestData.ID == 0 {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数：appId 和 id",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 根据邮件类型选择删除方法
	var err error
	if requestData.Type == "personal" {
		// 删除个人邮件（只删除关联记录）
		err = models.DeletePersonalMail(requestData.AppId, requestData.ID)
	} else {
		// 删除系统邮件（删除邮件内容和所有关联记录）
		err = models.DeleteSystemMail(requestData.AppId, requestData.ID)
		if err == nil {
			// 同时删除所有相关的个人邮件关联记录
			_ = models.DeletePersonalMail(requestData.AppId, requestData.ID)
		}
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除邮件失败: " + err.Error(),
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

	// 使用新的邮件系统创建并发布邮件
	systemMail := &models.MailSystem{
		AppId:      requestData.AppId,
		MailId:     fmt.Sprintf("mail_%d", time.Now().Unix()),
		Title:      requestData.Title,
		Content:    requestData.Content,
		Rewards:    requestData.Rewards,
		Type:       "system",
		TargetType: "all",
		Status:     "draft",
		CreatedBy:  "admin", // 默认创建者为admin
	}

	// 设置过期时间
	if requestData.ExpireDays > 0 {
		expireTime := time.Now().AddDate(0, 0, requestData.ExpireDays)
		systemMail.ExpireTime = &expireTime
	}

	// 创建系统邮件
	if err := models.CreateSystemMail(systemMail); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "创建系统邮件失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 发布邮件
	if err := models.PublishSystemMail(systemMail.ID, requestData.AppId); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5002,
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
		fmt.Println("参数错误", err)
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
		fmt.Println("获取邮件统计失败", err)
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

	mails, total, err := models.GetPlayerMailList(requestData.AppId, requestData.UserId, requestData.Page, requestData.PageSize)
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

	if err := models.SendBroadcastMail(requestData.AppId, requestData.Title, requestData.Content, requestData.Attachments, 7); err != nil {
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

// PublishExisting 发布已存在的邮件
func (c *MailController) PublishExisting() {
	var requestData struct {
		Id    int64  `json:"id"`
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

	fmt.Printf("DEBUG: 发布邮件请求 - ID: %d, AppId: %s\n", requestData.Id, requestData.AppId)

	// 查找邮件
	mail, err := models.GetMailById(requestData.AppId, requestData.Id)
	if err != nil {
		fmt.Printf("DEBUG: 查找邮件失败: %v\n", err)
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "邮件不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查邮件是否已经发布
	if mail.Status == "sent" {
		c.Data["json"] = map[string]interface{}{
			"code":      4002,
			"msg":       "邮件已经发布",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	fmt.Printf("DEBUG: 准备发布邮件 - 当前状态: %s\n", mail.Status)

	// 发布邮件给所有用户
	if err := models.PublishSystemMail(int64(requestData.Id), requestData.AppId); err != nil {
		fmt.Printf("DEBUG: PublishSystemMail failed: %v\n", err)
		c.Data["json"] = map[string]interface{}{
			"code":      5002,
			"msg":       "发布邮件失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	fmt.Printf("DEBUG: 邮件发布成功 - ID: %d\n", requestData.Id)

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "邮件发布成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}
