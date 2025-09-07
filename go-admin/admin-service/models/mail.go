package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Mail 邮件模型
type Mail struct {
	ID        int64     `orm:"pk;auto" json:"id"`
	AppId     string    `orm:"size(100)" json:"appId"`
	UserId    string    `orm:"size(100)" json:"userId"`
	Title     string    `orm:"size(200)" json:"title"`
	Content   string    `orm:"type(text)" json:"content"`
	Rewards   string    `orm:"type(text)" json:"rewards"`
	Status    int       `orm:"default(0)" json:"status"` // 0:未读 1:已读 2:已领取
	ExpireAt  time.Time `orm:"type(datetime);null" json:"expireAt"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"createdAt"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"updatedAt"`
}

// MailConfig 邮件配置模型
type MailConfig struct {
	BaseModel
	AppId         string `orm:"size(100)" json:"appId"`
	MailType      string `orm:"size(50)" json:"mailType"` // personal, broadcast
	Title         string `orm:"size(200)" json:"title"`
	Content       string `orm:"type(text)" json:"content"`
	Rewards       string `orm:"type(text)" json:"rewards"`
	ExpireDays    int    `orm:"default(7)" json:"expireDays"`
	Status        int    `orm:"default(1)" json:"status"`        // 1:启用 0:禁用
	SendCondition string `orm:"type(text)" json:"sendCondition"` // 发送条件JSON
}

// GetMailCount 获取邮件数量统计
func GetMailCount(appId string) (int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("mail_%s", appId)
	var count int64
	err := o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).QueryRow(&count)
	return count, err
}

func init() {
	orm.RegisterModel(new(Mail))
	orm.RegisterModel(new(MailConfig))
}

// TableName 获取表名
func (m *Mail) TableName() string {
	return "mail"
}

// GetTableName 获取动态表名
func (m *Mail) GetTableName(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return fmt.Sprintf("mail_%s", cleanAppId)
}

// TableName 获取表名
func (mc *MailConfig) TableName() string {
	return "mail_config"
}

// GetMailList 获取邮件列表
func GetMailList(appId string, page, pageSize int, userId string) ([]map[string]interface{}, int64, error) {
	return GetAllMailList(appId, page, pageSize)
}

// CreateMail 创建邮件
func CreateMail(mail *Mail) error {
	o := orm.NewOrm()
	_, err := o.Insert(mail)
	return err
}

// UpdateMail 更新邮件
func UpdateMail(mail *Mail) error {
	o := orm.NewOrm()
	_, err := o.Update(mail)
	return err
}

// DeleteMail 删除邮件
func DeleteMail(id int64) error {
	o := orm.NewOrm()
	mail := &Mail{ID: id}
	_, err := o.Delete(mail)
	return err
}

// PublishMail 发布邮件
func PublishMail(appId, title, content, rewards string, expireDays int) error {
	// 这里实现发布邮件的逻辑
	// 可以是广播邮件或者特定条件的邮件
	return SendBroadcastMail(appId, title, content, rewards)
}

// GetMailStats 获取邮件统计
func GetMailStats(appId string) (map[string]interface{}, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("mail_%s", appId)

	// 检查表是否存在
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	tableName = fmt.Sprintf("mail_%s", cleanAppId)

	stats := make(map[string]interface{})

	// 总邮件数
	var total int64
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := o.Raw(sql).QueryRow(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// 未读邮件数
	var unread int64
	sql = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status = 0", tableName)
	err = o.Raw(sql).QueryRow(&unread)
	if err != nil {
		return nil, err
	}
	stats["unread"] = unread

	// 已读邮件数
	var read int64
	sql = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status = 1", tableName)
	err = o.Raw(sql).QueryRow(&read)
	if err != nil {
		return nil, err
	}
	stats["read"] = read

	// 已领取邮件数
	var claimed int64
	sql = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status = 2", tableName)
	err = o.Raw(sql).QueryRow(&claimed)
	if err != nil {
		return nil, err
	}
	stats["claimed"] = claimed

	return stats, nil
}

// GetUserMails 获取用户邮件
func GetUserMails(appId, userId string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	tableName := fmt.Sprintf("mail_%s", cleanAppId)

	// 获取用户邮件
	sql := fmt.Sprintf("SELECT * FROM %s WHERE playerId = ? ORDER BY createdAt DESC LIMIT ? OFFSET ?", tableName)
	params := []interface{}{userId, pageSize, (page - 1) * pageSize}

	var results []orm.Params
	_, err := o.Raw(sql, params...).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 计算总数
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE playerId = ?", tableName)
	var total int64
	err = o.Raw(countSql, userId).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 转换结果
	var list []map[string]interface{}
	for _, result := range results {
		item := make(map[string]interface{})
		for k, v := range result {
			item[k] = v
		}
		list = append(list, item)
	}

	return list, total, nil
}

// CreateMailConfig 创建邮件配置
func CreateMailConfig(config *MailConfig) error {
	o := orm.NewOrm()
	_, err := o.Insert(config)
	return err
}

// UpdateMailConfig 更新邮件配置
func UpdateMailConfig(config *MailConfig) error {
	o := orm.NewOrm()
	_, err := o.Update(config)
	return err
}

// DeleteMailConfig 删除邮件配置
func DeleteMailConfig(id int64) error {
	o := orm.NewOrm()
	config := &MailConfig{}
	config.Id = id
	_, err := o.Delete(config)
	return err
}

// GetMailConfigList 获取邮件配置列表
func GetMailConfigList(appId string, page, pageSize int) ([]*MailConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("mail_config")

	if appId != "" {
		qs = qs.Filter("appId", appId)
	}

	total, _ := qs.Count()

	var configs []*MailConfig
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&configs)

	return configs, total, err
}
