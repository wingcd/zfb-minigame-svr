package models

import (
	"fmt"
	"time"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
)

// Mail 邮件模型 - 兼容旧版本API
type Mail struct {
	Id        int64     `orm:"auto" json:"id"`
	UserId    string    `orm:"size(100)" json:"user_id"`
	Title     string    `orm:"size(200)" json:"title"`
	Content   string    `orm:"type(text)" json:"content"`
	Rewards   string    `orm:"type(text)" json:"rewards"`
	Status    int       `orm:"default(0)" json:"status"` // 0:未读 1:已读 2:已领取
	ExpireAt  time.Time `orm:"null" json:"expire_at"`
	CreatedAt string    `orm:"auto_now_add;type(datetime)" json:"create_time"`
	UpdatedAt string    `orm:"auto_now;type(datetime)" json:"update_time"`
}

// MailSystem 邮件系统模型 - 对应数据库设计的mail_[appid]表
type MailSystem struct {
	ID         int64      `orm:"pk;auto" json:"id"`
	AppId      string     `orm:"-" json:"appId"`                                              // 应用ID（仅用于逻辑，不存储到数据库）
	Title      string     `orm:"size(200)" json:"title"`                                      // 邮件标题
	Content    string     `orm:"type(text)" json:"content"`                                   // 邮件内容
	Type       string     `orm:"size(50);default(system)" json:"type"`                        // 邮件类型: system/activity/reward
	Sender     string     `orm:"size(100);default(system)" json:"sender"`                     // 发送者
	Targets    string     `orm:"type(text)" json:"targets"`                                   // 目标用户（JSON数组，all表示全体）
	TargetType string     `orm:"size(50);default(all);column(target_type)" json:"targetType"` // 目标类型: all/specific/condition
	Condition  string     `orm:"type(text);column(send_condition)" json:"condition"`          // 发送条件（JSON）
	Rewards    string     `orm:"type(text)" json:"rewards"`                                   // 奖励列表（JSON数组）
	Status     string     `orm:"size(50);default(draft)" json:"status"`                       // 状态: draft/sent/expired
	SendTime   *time.Time `orm:"type(datetime);null;column(send_time)" json:"sendTime"`       // 发送时间
	ExpireTime *time.Time `orm:"type(datetime);null;column(expire_time)" json:"expireTime"`   // 过期时间
	ReadCount  int        `orm:"default(0);column(read_count)" json:"readCount"`              // 已读数量
	TotalCount int        `orm:"default(0);column(total_count)" json:"totalCount"`            // 总发送数量
	CreatedAt  time.Time  `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt  time.Time  `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
	CreatedBy  string     `orm:"size(100);column(created_by)" json:"createdBy"` // 创建者
}

// MailPlayerRelation 邮件-玩家关联表模型（动态表名: mail_player_relation_[appid]）
type MailPlayerRelation struct {
	ID         int64      `orm:"pk;auto" json:"id"`
	MailId     int64      `orm:"column(mail_id)" json:"mailId"`                             // 对应mail_[appid]表的id
	PlayerId   string     `orm:"size(100);column(player_id)" json:"playerId"`               // 玩家ID
	Status     int        `orm:"default(0)" json:"status"`                                  // 0:未读 1:已读 2:已领取
	ReceivedAt *time.Time `orm:"type(datetime);null;column(received_at)" json:"receivedAt"` // 接收时间
	ReadAt     *time.Time `orm:"type(datetime);null;column(read_at)" json:"readAt"`         // 阅读时间
	ClaimAt    *time.Time `orm:"type(datetime);null;column(claim_at)" json:"claimAt"`       // 领取时间
	CreatedAt  time.Time  `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt  time.Time  `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// GetTableName 获取动态表名
func (m *Mail) GetTableName(appId string) string {
	return utils.GetMailTableName(appId)
}

func (ms *MailSystem) GetTableName(appId string) string {
	return utils.GetMailTableName(appId)
}

func (mr *MailPlayerRelation) GetTableName(appId string) string {
	return utils.GetMailRelationTableName(appId)
}

// GetMailList 获取用户邮件列表（通过关联表）
func GetMailList(appId, userId string, page, pageSize int) ([]Mail, int64, error) {
	o := orm.NewOrm()
	mailTableName := utils.GetMailTableName(appId)
	relationTableName := utils.GetMailRelationTableName(appId)

	// 首先确保用户的邮件关联记录完整
	err := ensurePlayerMailRelations(o, appId, userId, mailTableName, relationTableName)
	if err != nil {
		return nil, 0, fmt.Errorf("确保邮件关联失败: %v", err)
	}

	// 查询用户的邮件（通过关联表）
	offset := (page - 1) * pageSize
	sql := fmt.Sprintf(`
		SELECT 
			m.id,
			m.title,
			m.content,
			m.rewards,
			m.expire_time as expire_at,
			m.created_at as create_time,
			r.status,
			r.updated_at as update_time
		FROM %s m
		INNER JOIN %s r ON m.id = r.mail_id
		WHERE r.player_id = ?
		  AND (m.expire_time IS NULL OR m.expire_time > NOW())
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, mailTableName, relationTableName)

	var results []orm.Params
	_, err = o.Raw(sql, userId, pageSize, offset).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 转换为Mail结构
	var mails []Mail
	for _, result := range results {
		mail := Mail{
			UserId: userId,
		}

		if id, ok := result["id"]; ok {
			if idInt, ok := id.(int64); ok {
				mail.Id = idInt
			}
		}
		if title, ok := result["title"].(string); ok {
			mail.Title = title
		}
		if content, ok := result["content"].(string); ok {
			mail.Content = content
		}
		if rewards, ok := result["rewards"].(string); ok {
			mail.Rewards = rewards
		}
		if status, ok := result["status"]; ok {
			if statusInt, ok := status.(int64); ok {
				mail.Status = int(statusInt)
			}
		}
		if createTime, ok := result["create_time"].(time.Time); ok {
			mail.CreatedAt = createTime.Format("2006-01-02 15:04:05")
		}
		if updateTime, ok := result["update_time"].(time.Time); ok {
			mail.UpdatedAt = updateTime.Format("2006-01-02 15:04:05")
		}
		if expireAt, ok := result["expire_at"].(time.Time); ok {
			mail.ExpireAt = expireAt
		}

		mails = append(mails, mail)
	}

	// 查询总数
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s m
		INNER JOIN %s r ON m.id = r.mail_id
		WHERE r.player_id = ?
		  AND (m.expire_time IS NULL OR m.expire_time > NOW())
	`, mailTableName, relationTableName)

	var total int64
	err = o.Raw(countSQL, userId).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	return mails, total, nil
}

// ReadMail 读取邮件
func ReadMail(appId, userId string, mailId int64) error {
	o := orm.NewOrm()
	relationTableName := utils.GetMailRelationTableName(appId)

	// 标记邮件为已读
	sql := fmt.Sprintf(`
		UPDATE %s 
		SET status = 1, read_at = NOW(), updated_at = NOW()
		WHERE mail_id = ? AND player_id = ? AND status = 0
	`, relationTableName)

	result, err := o.Raw(sql, mailId, userId).Exec()
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("邮件不存在或已读取")
	}

	return nil
}

// ClaimRewards 领取邮件奖励
func ClaimRewards(appId, userId string, mailId int64) (string, error) {
	o := orm.NewOrm()
	mailTableName := utils.GetMailTableName(appId)
	relationTableName := utils.GetMailRelationTableName(appId)

	// 首先获取邮件奖励信息和检查状态
	sql := fmt.Sprintf(`
		SELECT m.rewards, r.status, m.expire_time
		FROM %s m
		INNER JOIN %s r ON m.id = r.mail_id
		WHERE r.mail_id = ? AND r.player_id = ?
	`, mailTableName, relationTableName)

	var result []orm.Params
	_, err := o.Raw(sql, mailId, userId).Values(&result)
	if err != nil {
		return "", fmt.Errorf("查询邮件失败: %v", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("邮件不存在")
	}

	mailData := result[0]

	// 检查状态
	if status, ok := mailData["status"].(int64); ok && status == 2 {
		return "", fmt.Errorf("奖励已领取")
	}

	// 检查过期时间
	if expireTime, ok := mailData["expire_time"].(time.Time); ok && !expireTime.IsZero() && time.Now().After(expireTime) {
		return "", fmt.Errorf("邮件已过期")
	}

	// 获取奖励内容
	rewards := ""
	if rewardsData, ok := mailData["rewards"].(string); ok {
		rewards = rewardsData
	}

	// 标记为已领取
	updateSQL := fmt.Sprintf(`
		UPDATE %s 
		SET status = 2, claim_at = NOW(), updated_at = NOW()
		WHERE mail_id = ? AND player_id = ?
	`, relationTableName)

	_, err = o.Raw(updateSQL, mailId, userId).Exec()
	if err != nil {
		return "", fmt.Errorf("更新领取状态失败: %v", err)
	}

	return rewards, nil
}

// DeleteMail 删除邮件（删除关联关系）
func DeleteMail(appId, userId string, mailId int64) error {
	o := orm.NewOrm()
	relationTableName := utils.GetMailRelationTableName(appId)

	// 删除邮件关联关系
	sql := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE mail_id = ? AND player_id = ?
	`, relationTableName)

	_, err := o.Raw(sql, mailId, userId).Exec()
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
	_, err := qs.OrderBy("-create_time").Limit(pageSize, offset).All(&mails)

	return mails, total, err
}

// CleanExpiredMails 清理过期邮件
func CleanExpiredMails(appId string) error {
	o := orm.NewOrm()
	relationTableName := utils.GetMailRelationTableName(appId)
	mailTableName := utils.GetMailTableName(appId)

	// 删除过期邮件的关联关系
	sql := fmt.Sprintf(`
		DELETE r FROM %s r
		INNER JOIN %s m ON r.mail_id = m.id
		WHERE m.expire_time IS NOT NULL AND m.expire_time < NOW()
	`, relationTableName, mailTableName)

	_, err := o.Raw(sql).Exec()
	if err != nil {
		return err
	}

	// 删除过期邮件
	deleteMailSQL := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE expire_time IS NOT NULL AND expire_time < NOW()
	`, mailTableName)

	_, err = o.Raw(deleteMailSQL).Exec()
	return err
}

// ensurePlayerMailRelations 确保玩家有所有应该接收的邮件关联记录
func ensurePlayerMailRelations(o orm.Ormer, appId, playerId, mailTableName, relationTableName string) error {
	// 查找玩家缺失的邮件关联
	sql := fmt.Sprintf(`
		INSERT INTO %s (mail_id, player_id, status, received_at, created_at, updated_at)
		SELECT 
			m.id,
			?,
			0,
			NOW(),
			NOW(),
			NOW()
		FROM %s m
		LEFT JOIN %s r ON m.id = r.mail_id AND r.player_id = ?
		WHERE r.id IS NULL
		  AND m.status = 'sent'
		  AND (m.expire_time IS NULL OR m.expire_time > NOW())
		  AND (
			m.target_type = 'all' 
			OR (m.target_type = 'specific' AND JSON_CONTAINS(m.targets, JSON_QUOTE(?)))
		  )
	`, relationTableName, mailTableName, relationTableName)

	_, err := o.Raw(sql, playerId, playerId, playerId).Exec()
	return err
}

// HasNewMail 检查用户是否有新邮件（状态为0的邮件）
func HasNewMail(appId, userId string) (bool, error) {
	o := orm.NewOrm()
	mailTableName := utils.GetMailTableName(appId)
	relationTableName := utils.GetMailRelationTableName(appId)

	// 首先确保用户的邮件关联记录完整
	err := ensurePlayerMailRelations(o, appId, userId, mailTableName, relationTableName)
	if err != nil {
		return false, fmt.Errorf("确保邮件关联失败: %v", err)
	}

	// 查询是否有未读邮件
	sql := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s m
		INNER JOIN %s r ON m.id = r.mail_id
		WHERE r.player_id = ?
		  AND r.status = 0
		  AND (m.expire_time IS NULL OR m.expire_time > NOW())
	`, mailTableName, relationTableName)

	var count int64
	err = o.Raw(sql, userId).QueryRow(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
