package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Notification 通知模型
type Notification struct {
	BaseModel
	Title       string    `orm:"size(200)" json:"title"`
	Content     string    `orm:"type(text)" json:"content"`
	Type        string    `orm:"size(50)" json:"type"`          // info, warning, error, success
	Status      int       `orm:"default(1)" json:"status"`      // 1:启用 0:禁用
	Priority    int       `orm:"default(1)" json:"priority"`    // 1:低 2:中 3:高
	TargetUsers string    `orm:"type(text)" json:"targetUsers"` // 目标用户JSON
	SendTime    time.Time `orm:"type(datetime);null" json:"sendTime"`
	ExpireTime  time.Time `orm:"type(datetime);null" json:"expireTime"`
	CreatedBy   int64     `json:"createdBy"`
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	BaseModel
	Name        string `orm:"size(100)" json:"name"`
	Title       string `orm:"size(200)" json:"title"`
	Content     string `orm:"type(text)" json:"content"`
	Type        string `orm:"size(50)" json:"type"`
	Variables   string `orm:"type(text)" json:"variables"` // 变量定义JSON
	Description string `orm:"size(500)" json:"description"`
	Status      int    `orm:"default(1)" json:"status"`
}

// NotificationLog 通知日志
type NotificationLog struct {
	BaseModel
	NotificationId int64     `json:"notificationId"`
	UserId         int64     `json:"userId"`
	Status         int       `orm:"default(0)" json:"status"` // 0:未读 1:已读
	ReadTime       time.Time `orm:"type(datetime);null" json:"readTime"`
}

func init() {
	orm.RegisterModel(new(Notification))
	orm.RegisterModel(new(NotificationTemplate))
	orm.RegisterModel(new(NotificationLog))
}

// TableName 获取表名
func (n *Notification) TableName() string {
	return "notifications"
}

// TableName 获取表名
func (nt *NotificationTemplate) TableName() string {
	return "notification_templates"
}

// TableName 获取表名
func (nl *NotificationLog) TableName() string {
	return "notification_logs"
}

// GetNotifications 获取通知列表
func GetNotifications(page, pageSize int, notificationType, status string) ([]*Notification, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("notifications")

	if notificationType != "" {
		qs = qs.Filter("type", notificationType)
	}

	if status != "" {
		qs = qs.Filter("status", status)
	}

	total, _ := qs.Count()

	var notifications []*Notification
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&notifications)

	return notifications, total, err
}

// GetNotification 获取单个通知
func GetNotification(id int64) (*Notification, error) {
	o := orm.NewOrm()
	notification := &Notification{BaseModel: BaseModel{ID: id}}
	err := o.Read(notification)
	return notification, err
}

// CreateNotification 创建通知
func CreateNotification(notification *Notification) error {
	o := orm.NewOrm()
	_, err := o.Insert(notification)
	return err
}

// UpdateNotification 更新通知
func UpdateNotification(notification *Notification) error {
	o := orm.NewOrm()
	_, err := o.Update(notification)
	return err
}

// DeleteNotification 删除通知
func DeleteNotification(id int64) error {
	o := orm.NewOrm()
	notification := &Notification{BaseModel: BaseModel{ID: id}}
	_, err := o.Delete(notification)
	return err
}

// GetNotificationTemplates 获取通知模板列表
func GetNotificationTemplates(page, pageSize int) ([]*NotificationTemplate, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("notification_templates")

	total, _ := qs.Count()

	var templates []*NotificationTemplate
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&templates)

	return templates, total, err
}

// CreateNotificationTemplate 创建通知模板
func CreateNotificationTemplate(template *NotificationTemplate) error {
	o := orm.NewOrm()
	_, err := o.Insert(template)
	return err
}

// GetNotificationLogs 获取通知日志
func GetNotificationLogs(page, pageSize int, userId int64) ([]*NotificationLog, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("notification_logs")

	if userId > 0 {
		qs = qs.Filter("playerId", userId)
	}

	total, _ := qs.Count()

	var logs []*NotificationLog
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&logs)

	return logs, total, err
}

// GetNotificationStats 获取通知统计
func GetNotificationStats() (map[string]interface{}, error) {
	o := orm.NewOrm()

	stats := make(map[string]interface{})

	// 总通知数
	total, err := o.QueryTable("notifications").Count()
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// 活跃通知数
	active, err := o.QueryTable("notifications").Filter("status", 1).Count()
	if err != nil {
		return nil, err
	}
	stats["active"] = active

	// 未读通知数
	unread, err := o.QueryTable("notification_logs").Filter("status", 0).Count()
	if err != nil {
		return nil, err
	}
	stats["unread"] = unread

	return stats, nil
}

// MarkAsRead 标记通知为已读
func MarkAsRead(notificationId, userId int64) error {
	o := orm.NewOrm()

	// 查找通知日志
	log := &NotificationLog{}
	err := o.QueryTable("notification_logs").
		Filter("notification_id", notificationId).
		Filter("playerId", userId).
		One(log)

	if err != nil {
		// 如果不存在日志，创建一个
		log = &NotificationLog{
			NotificationId: notificationId,
			UserId:         userId,
			Status:         1,
			ReadTime:       time.Now(),
		}
		_, err = o.Insert(log)
		return err
	}

	// 更新状态
	log.Status = 1
	log.ReadTime = time.Now()
	_, err = o.Update(log, "status", "read_time")
	return err
}
