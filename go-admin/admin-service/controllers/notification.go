package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"strconv"

	"github.com/beego/beego/v2/server/web"
)

type NotificationController struct {
	web.Controller
}

// GetNotifications 获取通知列表
func (c *NotificationController) GetNotifications() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")
	notificationType := c.GetString("type", "")
	status := c.GetString("status", "")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	notifications, total, err := models.GetNotifications(page, pageSize, notificationType, status)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取通知列表失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", map[string]interface{}{
		"list":     notifications,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetNotification 获取单个通知
func (c *NotificationController) GetNotification() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "无效的通知ID", nil)
		return
	}

	notification, err := models.GetNotification(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1004, "通知不存在", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", notification)
}

// CreateNotification 创建通知
func (c *NotificationController) CreateNotification() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	var notification models.Notification
	if err := utils.ParseJSON(&c.Controller, &notification); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "参数解析失败", nil)
		return
	}

	// 验证必填参数
	if !utils.ValidateRequired(&c.Controller, map[string]interface{}{
		"title":   notification.Title,
		"content": notification.Content,
	}) {
		return
	}

	// 设置创建者
	notification.CreatedBy = claims.UserID

	if err := models.CreateNotification(&notification); err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "创建通知失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	models.LogAdminOperation(claims.UserID, claims.Username, "CREATE", "NOTIFICATION", map[string]interface{}{
		"notificationId":    notification.Id,
		"notificationTitle": notification.Title,
	})

	utils.SuccessResponse(&c.Controller, "创建成功", notification)
}

// UpdateNotification 更新通知
func (c *NotificationController) UpdateNotification() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "无效的通知ID", nil)
		return
	}

	// 检查通知是否存在
	existingNotification, err := models.GetNotification(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1004, "通知不存在", nil)
		return
	}

	var notification models.Notification
	if err := utils.ParseJSON(&c.Controller, &notification); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "参数解析失败", nil)
		return
	}

	// 设置ID和保留创建者
	notification.Id = id
	notification.CreatedBy = existingNotification.CreatedBy

	if err := models.UpdateNotification(&notification); err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新通知失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	models.LogAdminOperation(claims.UserID, claims.Username, "UPDATE", "NOTIFICATION", map[string]interface{}{
		"notificationId":    notification.Id,
		"notificationTitle": notification.Title,
	})

	utils.SuccessResponse(&c.Controller, "更新成功", notification)
}

// DeleteNotification 删除通知
func (c *NotificationController) DeleteNotification() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "无效的通知ID", nil)
		return
	}

	// 检查通知是否存在
	notification, err := models.GetNotification(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1004, "通知不存在", nil)
		return
	}

	if err := models.DeleteNotification(id); err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "删除通知失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	models.LogAdminOperation(claims.UserID, claims.Username, "DELETE", "NOTIFICATION", map[string]interface{}{
		"notificationId":    id,
		"notificationTitle": notification.Title,
	})

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}

// SendNotification 发送通知
func (c *NotificationController) SendNotification() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "无效的通知ID", nil)
		return
	}

	// 获取通知
	notification, err := models.GetNotification(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1004, "通知不存在", nil)
		return
	}

	// TODO: 实现发送通知逻辑
	// 这里可以实现推送到消息队列、发送邮件、短信等

	// 记录操作日志
	models.LogAdminOperation(claims.UserID, claims.Username, "SEND", "NOTIFICATION", map[string]interface{}{
		"notificationId":    id,
		"notificationTitle": notification.Title,
	})

	utils.SuccessResponse(&c.Controller, "发送成功", nil)
}

// GetNotificationTemplates 获取通知模板列表
func (c *NotificationController) GetNotificationTemplates() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	templates, total, err := models.GetNotificationTemplates(page, pageSize)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取模板列表失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", map[string]interface{}{
		"list":     templates,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CreateNotificationTemplate 创建通知模板
func (c *NotificationController) CreateNotificationTemplate() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	var template models.NotificationTemplate
	if err := utils.ParseJSON(&c.Controller, &template); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "参数解析失败", nil)
		return
	}

	// 验证必填参数
	if !utils.ValidateRequired(&c.Controller, map[string]interface{}{
		"name":    template.Name,
		"title":   template.Title,
		"content": template.Content,
	}) {
		return
	}

	if err := models.CreateNotificationTemplate(&template); err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "创建模板失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	models.LogAdminOperation(claims.UserID, claims.Username, "CREATE", "NOTIFICATION_TEMPLATE", map[string]interface{}{
		"templateId":   template.Id,
		"templateName": template.Name,
	})

	utils.SuccessResponse(&c.Controller, "创建成功", template)
}

// GetNotificationLogs 获取通知日志
func (c *NotificationController) GetNotificationLogs() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")
	userIdStr := c.GetString("userId", "0")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	logs, total, err := models.GetNotificationLogs(page, pageSize, userId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取通知日志失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", map[string]interface{}{
		"list":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetNotificationStats 获取通知统计
func (c *NotificationController) GetNotificationStats() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	stats, err := models.GetNotificationStats()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取统计失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", stats)
}

// MarkAsRead 标记通知为已读
func (c *NotificationController) MarkAsRead() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "无效的通知ID", nil)
		return
	}

	if err := models.MarkAsRead(id, claims.UserID); err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "标记已读失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "标记成功", nil)
}
