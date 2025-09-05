package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Mail 邮件模型
type Mail struct {
	Id        int64     `orm:"auto" json:"id"`
	UserId    string    `orm:"size(100)" json:"user_id"`
	Title     string    `orm:"size(200)" json:"title"`
	Content   string    `orm:"type(text)" json:"content"`
	Rewards   string    `orm:"type(text)" json:"rewards"`
	Status    int       `orm:"default(0)" json:"status"` // 0:未读 1:已读 2:已领取
	ExpireAt  time.Time `orm:"null" json:"expire_at"`
	CreatedAt string    `orm:"auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt string    `orm:"auto_now;type(datetime)" json:"updated_at"`
}

// GetTableName 获取动态表名
func (m *Mail) GetTableName(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return fmt.Sprintf("mail_%s", cleanAppId)
}

// GetMailList 获取用户邮件列表
func GetMailList(appId, userId string, page, pageSize int) ([]Mail, int64, error) {
	o := orm.NewOrm()

	mail := &Mail{}
	tableName := mail.GetTableName(appId)

	qs := o.QueryTable(tableName).Filter("user_id", userId)

	// 过滤掉已过期的邮件
	now := time.Now()
	qs = qs.Filter("expire_at__isnull", true).Filter("expire_at__gt", now)

	total, _ := qs.Count()

	var mails []Mail
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-created_at").Limit(pageSize, offset).All(&mails)

	return mails, total, err
}

// ReadMail 读取邮件
func ReadMail(appId, userId string, mailId int64) error {
	o := orm.NewOrm()

	mail := &Mail{}
	tableName := mail.GetTableName(appId)

	// 获取邮件
	err := o.QueryTable(tableName).Filter("id", mailId).Filter("user_id", userId).One(mail)
	if err != nil {
		return err
	}

	// 检查是否过期
	if !mail.ExpireAt.IsZero() && time.Now().After(mail.ExpireAt) {
		return fmt.Errorf("mail expired")
	}

	// 标记为已读
	if mail.Status == 0 {
		mail.Status = 1
		_, err = o.Update(mail, "status", "updated_at")
	}

	return err
}

// ClaimRewards 领取邮件奖励
func ClaimRewards(appId, userId string, mailId int64) (string, error) {
	o := orm.NewOrm()

	mail := &Mail{}
	tableName := mail.GetTableName(appId)

	// 获取邮件
	err := o.QueryTable(tableName).Filter("id", mailId).Filter("user_id", userId).One(mail)
	if err != nil {
		return "", err
	}

	// 检查是否过期
	if !mail.ExpireAt.IsZero() && time.Now().After(mail.ExpireAt) {
		return "", fmt.Errorf("mail expired")
	}

	// 检查是否已领取
	if mail.Status == 2 {
		return "", fmt.Errorf("rewards already claimed")
	}

	// 标记为已领取
	mail.Status = 2
	_, err = o.Update(mail, "status", "updated_at")
	if err != nil {
		return "", err
	}

	return mail.Rewards, nil
}

// DeleteMail 删除邮件
func DeleteMail(appId, userId string, mailId int64) error {
	o := orm.NewOrm()

	mail := &Mail{}
	tableName := mail.GetTableName(appId)

	_, err := o.QueryTable(tableName).Filter("id", mailId).Filter("user_id", userId).Delete()
	return err
}

// SendMail 发送邮件（管理后台使用）
func SendMail(appId, userId, title, content, rewards string, expireHours int) error {
	o := orm.NewOrm()

	mail := &Mail{}

	mail.UserId = userId
	mail.Title = title
	mail.Content = content
	mail.Rewards = rewards
	mail.Status = 0

	if expireHours > 0 {
		mail.ExpireAt = time.Now().Add(time.Duration(expireHours) * time.Hour)
	}

	_, err := o.Insert(mail)
	return err
}

// SendBroadcastMail 发送广播邮件（管理后台使用）
func SendBroadcastMail(appId, title, content, rewards string, expireHours int, userIds []string) error {
	o := orm.NewOrm()

	var expireTime time.Time
	if expireHours > 0 {
		expireTime = time.Now().Add(time.Duration(expireHours) * time.Hour)
	}

	// 批量插入邮件
	for _, userId := range userIds {
		mailRecord := &Mail{
			UserId:  userId,
			Title:   title,
			Content: content,
			Rewards: rewards,
			Status:  0,
		}

		if !expireTime.IsZero() {
			mailRecord.ExpireAt = expireTime
		}

		_, err := o.Insert(mailRecord)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMailDetails 获取邮件详情（管理后台使用）
func GetMailDetails(appId string, mailId int64) (*Mail, error) {
	o := orm.NewOrm()

	mail := &Mail{}
	tableName := mail.GetTableName(appId)

	err := o.QueryTable(tableName).Filter("id", mailId).One(mail)
	return mail, err
}

// GetAllMailList 获取所有邮件列表（管理后台使用）
func GetAllMailList(appId string, page, pageSize int, userId string) ([]Mail, int64, error) {
	o := orm.NewOrm()

	mail := &Mail{}
	tableName := mail.GetTableName(appId)

	qs := o.QueryTable(tableName)
	if userId != "" {
		qs = qs.Filter("user_id", userId)
	}

	total, _ := qs.Count()

	var mails []Mail
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-created_at").Limit(pageSize, offset).All(&mails)

	return mails, total, err
}

// CleanExpiredMails 清理过期邮件
func CleanExpiredMails(appId string) error {
	o := orm.NewOrm()

	mail := &Mail{}
	tableName := mail.GetTableName(appId)

	now := time.Now()
	_, err := o.QueryTable(tableName).Filter("expire_at__lt", now).Delete()
	return err
}
