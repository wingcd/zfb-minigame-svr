package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"fmt"
	"strings"
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

	// 创建邮件表
	err := createMailTable(requestData.AppId, requestData.Force)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "邮件系统初始化失败: " + err.Error(),
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
		AppId      string   `json:"appId"`
		UserId     string   `json:"userId"`
		Title      string   `json:"title"`
		Content    string   `json:"content"`
		Rewards    []string `json:"rewards"`
		ExpireDays int      `json:"expireDays"`
		MailType   int      `json:"mailType"` // 0: 个人邮件, 1: 系统广播邮件
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		fmt.Println("参数错误", err)
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

	// 使用新的邮件系统 - expireAt变量已移除

	// 使用新的邮件系统
	if requestData.MailType == 1 || requestData.UserId == "" {
		// 系统广播邮件
		systemMail := &models.MailSystem{
			AppId:      requestData.AppId,
			MailId:     fmt.Sprintf("mail_%d", time.Now().Unix()),
			Title:      requestData.Title,
			Content:    requestData.Content,
			Rewards:    strings.Join(requestData.Rewards, ","),
			Type:       "system",
			TargetType: "all",
			Status:     "draft",
		}

		if requestData.ExpireDays > 0 {
			systemMail.ExpireTime = time.Now().AddDate(0, 0, requestData.ExpireDays)
		}

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

		// 发布邮件给所有用户
		if err := models.PublishSystemMail(systemMail.MailId, requestData.AppId); err != nil {
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
			"msg":       "系统邮件创建并发布成功",
			"timestamp": utils.UnixMilli(),
			"data":      systemMail,
		}
	} else {
		// 个人邮件 - 使用SendPersonalMail
		rewards := strings.Join(requestData.Rewards, ",")
		if err := models.SendPersonalMail(requestData.AppId, requestData.UserId, requestData.Title, requestData.Content, rewards); err != nil {
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

// createMailTable 创建邮件表
func createMailTable(appId string, force bool) error {
	o := orm.NewOrm()

	mail := &models.Mail{}
	tableName := mail.GetTableName(appId)

	// 检查表是否已存在
	var count int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&count)
	if err != nil {
		return fmt.Errorf("检查表是否存在时出错: %v", err)
	}

	// 如果表已存在且不是强制模式
	if count > 0 && !force {
		// 检查表中是否有数据
		var dataCount int64
		err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).QueryRow(&dataCount)
		if err == nil && dataCount > 0 {
			return fmt.Errorf("邮件系统已初始化，如需重新初始化请设置 force=true")
		}
	}

	// 如果是强制模式，先删除现有表
	if force && count > 0 {
		_, err = o.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)).Exec()
		if err != nil {
			return fmt.Errorf("删除现有表失败: %v", err)
		}
	}

	// 创建邮件表
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			app_id VARCHAR(100) NOT NULL,
			user_id VARCHAR(100) NOT NULL,
			title VARCHAR(200) NOT NULL,
			content TEXT,
			rewards TEXT,
			status INT DEFAULT 0 COMMENT '0:未读 1:已读 2:已领取',
			expire_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_app_id (app_id),
			INDEX idx_user_id (user_id),
			INDEX idx_status (status),
			INDEX idx_expire_at (expire_at),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件表'
	`, tableName)

	_, err = o.Raw(createTableSQL).Exec()
	if err != nil {
		return fmt.Errorf("创建邮件表失败: %v", err)
	}

	return nil
}
