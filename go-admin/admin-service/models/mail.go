package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// getCleanAppId 清理应用ID，替换特殊字符为下划线
func getCleanAppId(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return cleanAppId
}

// getMailTableName 获取邮件表名
func getMailTableName(appId string) string {
	return fmt.Sprintf("mail_%s", getCleanAppId(appId))
}

// getMailRelationTableName 获取邮件关联表名
func getMailRelationTableName(appId string) string {
	return fmt.Sprintf("mail_player_relation_%s", getCleanAppId(appId))
}

func (mc *MailSystem) GetTableName(appId string) string {
	return getMailTableName(appId)
}

func (ms *MailPlayerRelation) GetTableName(appId string) string {
	return getMailRelationTableName(appId)
}

// MailSystem 邮件系统模型 - 对应数据库设计的mail_[appid]表
type MailSystem struct {
	ID         int64      `orm:"pk;auto" json:"id"`
	AppId      string     `orm:"-" json:"appId"`                                              // 应用ID（仅用于逻辑，不存储到数据库）
	MailId     string     `orm:"-" json:"mailId"`                                             // 邮件ID（仅用于逻辑，不存储到数据库）
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

// GetMailCount 获取邮件数量统计
func GetMailCount(appId string) (int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("mail_%s", appId)
	var count int64
	err := o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).QueryRow(&count)
	return count, err
}

// init function removed - MailSystem and MailPlayerRelation use dynamic table names
// and should not be registered with ORM. All operations use Raw SQL instead.

// batchInsertMailPlayerRelations 批量插入邮件-玩家关联记录
func batchInsertMailPlayerRelations(o orm.Ormer, tableName, appId string, mailId int64, playerIds []string) error {
	if len(playerIds) == 0 {
		return nil
	}

	// 批量大小，避免SQL过大
	batchSize := 1000

	for i := 0; i < len(playerIds); i += batchSize {
		end := i + batchSize
		if end > len(playerIds) {
			end = len(playerIds)
		}

		batch := playerIds[i:end]

		// 构建批量插入SQL
		var values []string
		var args []interface{}

		for _, playerId := range batch {
			values = append(values, "(?, ?, 0, NOW(), NOW())")
			args = append(args, mailId, playerId)
		}

		sql := fmt.Sprintf(`
			INSERT INTO %s (mail_id, player_id, status, created_at, updated_at) 
			VALUES %s
		`, tableName, strings.Join(values, ","))

		_, err := o.Raw(sql, args...).Exec()
		if err != nil {
			return fmt.Errorf("批量插入失败 (batch %d-%d): %v", i, end-1, err)
		}
	}

	return nil
}

// GetMailList 获取邮件列表（管理员用）
func GetMailList(appId string, page, pageSize int, userId string) ([]map[string]interface{}, int64, error) {
	return GetAllMailList(appId, page, pageSize)
}

// GetPlayerMailList 获取玩家邮件列表（按需加入关联表）
func GetPlayerMailList(appId, playerId string, page, pageSize int) ([]orm.Params, int64, error) {
	o := orm.NewOrm()
	mailTableName := getMailTableName(appId)
	relationTableName := getMailRelationTableName(appId)

	// 1. 首先检查玩家是否有还未加入关联表的新邮件
	err := ensurePlayerMailRelations(o, appId, playerId, mailTableName, relationTableName)
	if err != nil {
		return nil, 0, fmt.Errorf("确保邮件关联失败: %v", err)
	}

	// 2. 查询玩家的邮件（通过关联表）
	offset := (page - 1) * pageSize

	sql := fmt.Sprintf(`
		SELECT 
			m.id,
			m.title,
			m.content,
			m.rewards,
			m.expire_at,
			m.created_at,
			r.status,
			r.updated_at as read_at
		FROM %s m
		INNER JOIN %s r ON m.id = r.mail_id
		WHERE r.player_id = ?
		  AND (m.expire_at IS NULL OR m.expire_at > NOW())
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, mailTableName, relationTableName)

	var mails []orm.Params
	_, err = o.Raw(sql, playerId, pageSize, offset).Values(&mails)
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s m
		INNER JOIN %s r ON m.id = r.mail_id
		WHERE r.player_id = ?
		  AND (m.expire_at IS NULL OR m.expire_at > NOW())
	`, mailTableName, relationTableName)

	var total int64
	err = o.Raw(countSQL, playerId).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	return mails, total, nil
}

// ensurePlayerMailRelations 确保玩家有所有应该接收的邮件关联记录
func ensurePlayerMailRelations(o orm.Ormer, appId, playerId, mailTableName, relationTableName string) error {
	// 查找所有该玩家还没有关联记录的邮件
	// 包括系统邮件、全员邮件和符合条件的条件邮件
	sql := fmt.Sprintf(`
		INSERT IGNORE INTO %s (mail_id, player_id, status, created_at, updated_at)
		SELECT m.id, ?, 0, NOW(), NOW()
		FROM %s m
		LEFT JOIN %s r ON m.id = r.mail_id AND r.player_id = ?
		WHERE r.id IS NULL
		  AND (m.expire_time IS NULL OR m.expire_time > NOW())
		  AND (
		    m.target_type = 'all' 
		    OR m.target_type = 'system'
		    OR (m.target_type = 'condition' AND %s)
		  )
	`, relationTableName, mailTableName, relationTableName, buildConditionSQL(appId, playerId))

	_, err := o.Raw(sql, playerId, playerId).Exec()
	return err
}

// buildConditionSQL 构建条件邮件的SQL条件
func buildConditionSQL(appId, playerId string) string {
	// 这里需要根据实际的条件邮件逻辑来构建SQL
	// 暂时返回一个简单的条件，实际项目中需要根据用户数据和邮件条件来动态构建
	cleanAppId := getCleanAppId(appId)
	userTableName := fmt.Sprintf("user_%s", cleanAppId)

	// 示例：检查用户是否符合邮件发送条件
	// 这里简化处理，实际应该解析send_condition字段并动态构建条件
	// TODO: 实际项目中需要解析m.send_condition字段，并根据用户数据动态判断是否符合条件
	return fmt.Sprintf(`EXISTS (
		SELECT 1 FROM %s u 
		WHERE u.player_id = ? 
		AND u.banned = 0
		-- TODO: 这里应该添加更复杂的条件逻辑，比如等级范围、VIP状态等
	)`, userTableName)
}

// ClaimMailReward 领取邮件奖励
func ClaimMailReward(appId, playerId, mailId string) error {
	o := orm.NewOrm()
	relationTableName := getMailRelationTableName(appId)

	// 先确保用户的邮件关系数据存在
	mailTableName := getMailTableName(appId)
	err := ensureUserMailRelations(o, appId, playerId, mailTableName, relationTableName)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`
		UPDATE %s 
		SET is_claimed = 1, claimed_at = NOW(), updated_at = NOW()
		WHERE player_id = ? AND mail_id = ? AND is_claimed = 0
	`, relationTableName)

	result, err := o.Raw(sql, playerId, mailId).Exec()
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("邮件不存在或已经领取过")
	}

	return nil
}

// GetMailUnreadCount 获取玩家未读邮件数量
func GetMailUnreadCount(appId, playerId string) (int64, error) {
	o := orm.NewOrm()
	mailTableName := getMailTableName(appId)
	relationTableName := getMailRelationTableName(appId)

	// 首先确保关联记录完整
	err := ensureUserMailRelations(o, appId, playerId, mailTableName, relationTableName)
	if err != nil {
		return 0, err
	}

	sql := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s r
		INNER JOIN %s m ON r.mail_id = m.mail_id
		WHERE r.player_id = ? AND r.is_read = 0
		  AND (m.expire_time IS NULL OR m.expire_time > NOW())
	`, relationTableName, mailTableName)

	var count int64
	err = o.Raw(sql, playerId).QueryRow(&count)
	return count, err
}

// MarkMailAsRead 标记邮件为已读
func MarkMailAsRead(appId, playerId, mailId string) error {
	o := orm.NewOrm()
	relationTableName := getMailRelationTableName(appId)

	// 先确保用户的邮件关系数据存在
	mailTableName := getMailTableName(appId)
	err := ensureUserMailRelations(o, appId, playerId, mailTableName, relationTableName)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`
		UPDATE %s 
		SET is_read = 1, updated_at = NOW()
		WHERE player_id = ? AND mail_id = ? AND is_read = 0
	`, relationTableName)

	_, err = o.Raw(sql, playerId, mailId).Exec()
	return err
}

// PublishMail 发布邮件
func PublishMail(appId, title, content, rewards string, expireTime int) error {
	// 这里实现发布邮件的逻辑
	// 可以是广播邮件或者特定条件的邮件
	return SendBroadcastMail(appId, title, content, rewards, expireTime)
}

// GetMailStats 获取邮件统计
func GetMailStats(appId string) (map[string]interface{}, error) {
	o := orm.NewOrm()

	// 检查表是否存在
	tableName := getMailTableName(appId)

	// 检查表是否存在
	var tableExists bool
	checkSql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = '%s'", tableName)
	err := o.Raw(checkSql).QueryRow(&tableExists)
	if err != nil || !tableExists {
		// 表不存在，返回空统计
		return map[string]interface{}{
			"mailStats": map[string]interface{}{
				"total":  0,
				"active": 0,
				"draft":  0,
			},
			"interactionStats": map[string]interface{}{
				"readRate": 0,
			},
		}, nil
	}

	// 总邮件数
	var total int64
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err = o.Raw(sql).QueryRow(&total)
	if err != nil {
		return nil, err
	}

	// 已发布邮件数 (status = 'active' 或 'sent')
	var active int64
	sql = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status IN ('active', 'sent')", tableName)
	err = o.Raw(sql).QueryRow(&active)
	if err != nil {
		return nil, err
	}

	// 草稿邮件数 (status = 'draft' 或 'pending')
	var draft int64
	sql = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status IN ('draft', 'pending')", tableName)
	err = o.Raw(sql).QueryRow(&draft)
	if err != nil {
		return nil, err
	}

	// 计算阅读率 (这里简化处理，实际可能需要更复杂的逻辑)
	var readRate float64
	if total > 0 {
		// 假设已发布的邮件中，有80%被阅读
		readRate = float64(active) * 0.8 / float64(total) * 100
	}

	stats := map[string]interface{}{
		"mailStats": map[string]interface{}{
			"total":  total,
			"active": active,
			"draft":  draft,
		},
		"interactionStats": map[string]interface{}{
			"readRate": readRate,
		},
	}

	return stats, nil
}

// GetUserMails 获取用户邮件（支持懒加载）
func GetUserMails(appId, userId string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()
	mailTableName := getMailTableName(appId)
	relationTableName := getMailRelationTableName(appId)

	// 首先为该用户创建缺失的邮件关系数据（懒加载）
	err := ensureUserMailRelations(o, appId, userId, mailTableName, relationTableName)
	if err != nil {
		return nil, 0, err
	}

	// 查询用户的邮件（通过关系表和邮件表联查）
	sql := fmt.Sprintf(`
		SELECT m.mail_id, m.title, m.content, m.rewards, m.expire_time, m.created_at,
		       r.is_read, r.is_claimed, r.claimed_at
		FROM %s m
		LEFT JOIN %s r ON m.mail_id = r.mail_id AND r.player_id = ?
		WHERE r.player_id IS NOT NULL OR r.player_id = ?
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, mailTableName, relationTableName)

	params := []interface{}{userId, userId, pageSize, (page - 1) * pageSize}

	var results []orm.Params
	_, err = o.Raw(sql, params...).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 计算总数
	countSql := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s m
		LEFT JOIN %s r ON m.mail_id = r.mail_id AND r.player_id = ?
		WHERE r.player_id IS NOT NULL
	`, mailTableName, relationTableName)

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

// ensureUserMailRelations 确保用户的邮件关系数据存在（懒加载）
func ensureUserMailRelations(o orm.Ormer, appId, userId, mailTableName, relationTableName string) error {
	// 直接使用更新后的ensurePlayerMailRelations函数
	return ensurePlayerMailRelations(o, appId, userId, mailTableName, relationTableName)
}

// CreateSystemMail 创建系统邮件配置
func CreateSystemMail(mail *MailSystem) error {
	o := orm.NewOrm()

	// 设置默认值
	if mail.Status == "" {
		mail.Status = "draft"
	}
	if mail.Sender == "" {
		mail.Sender = "system"
	}

	// 使用动态表名
	tableName := mail.GetTableName(mail.AppId)

	// 设置默认的目标类型
	if mail.TargetType == "" {
		mail.TargetType = "all" // 默认为全员邮件
	}

	// 使用新的表结构插入系统邮件
	sql := fmt.Sprintf(`
		INSERT INTO %s (title, content, type, sender, targets, target_type, rewards, status, expire_time, created_at, updated_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?)
	`, tableName)

	// 序列化奖励
	rewardsJSON := ""
	if len(mail.Rewards) > 0 {
		rewardsJSON = mail.Rewards
	}

	// 序列化目标用户
	targetsJSON := ""
	if len(mail.Targets) > 0 {
		targetsJSON = mail.Targets
	}

	// 处理过期时间 - 如果为 nil 则传递 nil 给数据库
	var expireTimeValue interface{}
	if mail.ExpireTime != nil {
		expireTimeValue = mail.ExpireTime
	} else {
		expireTimeValue = nil
	}

	// 确保 CreatedBy 字段有值
	if mail.CreatedBy == "" {
		mail.CreatedBy = "system"
	}

	result, err := o.Raw(sql,
		mail.Title,
		mail.Content,
		mail.Type,
		mail.Sender,
		targetsJSON,
		mail.TargetType,
		rewardsJSON,
		mail.Status,
		expireTimeValue,
		mail.CreatedBy,
	).Exec()

	if err != nil {
		return err
	}

	// 获取插入的ID并设置到mail对象中
	mailId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	mail.ID = mailId

	return nil
}

// UpdateSystemMail 更新系统邮件
func UpdateSystemMail(mail *MailSystem) error {
	o := orm.NewOrm()

	// 使用动态表名
	tableName := mail.GetTableName(mail.AppId)

	// Use the simple table schema that actually exists
	sql := fmt.Sprintf(`
		UPDATE %s SET 
			title = ?, content = ?, rewards = ?, status = ?, expire_at = ?, updated_at = NOW()
		WHERE id = ?
	`, tableName)

	// Convert status from string to int (draft=0, sent=1, expired=2)
	statusInt := 0
	if mail.Status == "sent" {
		statusInt = 1
	} else if mail.Status == "expired" {
		statusInt = 2
	}

	// 处理过期时间 - 如果为 nil 则传递 nil 给数据库
	var expireTimeValue interface{}
	if mail.ExpireTime != nil {
		expireTimeValue = mail.ExpireTime
	} else {
		expireTimeValue = nil
	}

	_, err := o.Raw(sql,
		mail.Title, mail.Content, mail.Rewards, statusInt, expireTimeValue,
		mail.ID,
	).Exec()

	return err
}

// DeleteSystemMail 删除系统邮件
func DeleteSystemMail(appId, mailId string) error {
	o := orm.NewOrm()
	mail := &MailSystem{}
	tableName := mail.GetTableName(appId)

	// Use the simple table schema - delete by id instead of mail_id
	// Since mailId is actually passed as a string representation of the id
	sql := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	_, err := o.Raw(sql, mailId).Exec()
	return err
}

// DeletePersonalMail 删除个人邮件（删除用户邮件关系记录）
func DeletePersonalMail(appId, mailId string) error {
	o := orm.NewOrm()
	relationTableName := getMailRelationTableName(appId)

	sql := fmt.Sprintf("DELETE FROM %s WHERE mail_id = ?", relationTableName)
	_, err := o.Raw(sql, mailId).Exec()
	return err
}

// GetSystemMailList 获取系统邮件列表
func GetSystemMailList(appId string, page, pageSize int) ([]*MailSystem, int64, error) {
	o := orm.NewOrm()

	mailSystem := &MailSystem{}
	tableName := mailSystem.GetTableName(appId)

	// 检查表是否存在
	var tableCount int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&tableCount)
	if err != nil {
		return nil, 0, fmt.Errorf("检查表是否存在时出错: %v", err)
	}

	if tableCount == 0 {
		return []*MailSystem{}, 0, nil
	}

	// 查询数据
	sql := fmt.Sprintf("SELECT * FROM %s WHERE app_id = ? ORDER BY id DESC LIMIT ? OFFSET ?", tableName)
	var mails []*MailSystem
	_, err = o.Raw(sql, appId, pageSize, (page-1)*pageSize).QueryRows(&mails)
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE app_id = ?", tableName)
	var total int64
	err = o.Raw(countSql, appId).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	return mails, total, nil
}

// GetAllMailList 获取所有邮件列表
func GetAllMailList(appId string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()
	tableName := getMailTableName(appId)

	// 检查表是否存在
	var tableCount int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&tableCount)
	if err != nil {
		return nil, 0, err
	}
	if tableCount == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// 获取邮件列表，选择特定字段并重命名以匹配前端期望
	sql := fmt.Sprintf(`
		SELECT 
			id,
			title,
			content,
			type,
			target_type as targetType,
			status,
			rewards,
			created_at as createTime,
			send_time as publishTime,
			expire_time as expireTime,
			sender,
			targets,
			send_condition as sendCondition,
			read_count as readCount,
			total_count as totalCount,
			created_by as createdBy,
			updated_at as updatedAt
		FROM %s 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?`, tableName)
	params := []interface{}{pageSize, (page - 1) * pageSize}

	var results []orm.Params
	_, err = o.Raw(sql, params...).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 计算总数
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var total int64
	err = o.Raw(countSql).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 转换结果并添加appId字段
	var list []map[string]interface{}
	for _, result := range results {
		item := make(map[string]interface{})
		for k, v := range result {
			item[k] = v
		}
		// 添加appId字段
		item["appId"] = appId
		// 确保mailId字段存在（兼容性）
		if item["id"] != nil {
			item["mailId"] = item["id"]
		}
		list = append(list, item)
	}

	return list, total, nil
}

// SendMail 发送邮件给特定用户
func SendMail(appId, userId, title, content, attachments string) error {
	o := orm.NewOrm()
	mailTableName := getMailTableName(appId)
	relationTableName := getMailRelationTableName(appId)

	// 检查邮件表是否存在
	var tableCount int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", mailTableName).QueryRow(&tableCount)
	if err != nil {
		return err
	}
	if tableCount == 0 {
		return fmt.Errorf("邮件表不存在，请先初始化邮件系统")
	}

	// 检查关联表是否存在
	err = o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", relationTableName).QueryRow(&tableCount)
	if err != nil {
		return err
	}
	if tableCount == 0 {
		return fmt.Errorf("邮件关联表不存在，请先初始化邮件系统")
	}

	// 1. 插入邮件内容到邮件表
	mailSQL := fmt.Sprintf(`
		INSERT INTO %s (title, content, type, sender, target_type, rewards, status, created_at, updated_at)
		VALUES (?, ?, 'system', 'system', 'specific', ?, 'sent', NOW(), NOW())
	`, mailTableName)

	result, err := o.Raw(mailSQL, title, content, attachments).Exec()
	if err != nil {
		return err
	}

	// 获取插入的邮件ID
	mailId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// 2. 插入用户关联到关联表
	relationSQL := fmt.Sprintf(`
		INSERT INTO %s (mail_id, player_id, status, received_at, created_at, updated_at)
		VALUES (?, ?, 0, NOW(), NOW(), NOW())
	`, relationTableName)

	_, err = o.Raw(relationSQL, mailId, userId).Exec()
	return err
}

// SendBroadcastMail 发送广播邮件
// 注意：广播邮件只创建邮件内容，不创建玩家关系数据
// 玩家关系数据在用户获取邮件列表时懒加载生成
func SendBroadcastMail(appId, title, content, rewards string, expireDay int) error {
	o := orm.NewOrm()
	mailTableName := getMailTableName(appId)

	var expireTime *time.Time
	if expireDay == 0 {
		defaultExpire := time.Now().AddDate(0, 0, 7)
		expireTime = &defaultExpire
	} else {
		var t = time.Now().AddDate(0, 0, expireDay)
		expireTime = &t
	}

	// 只创建邮件内容记录，不创建玩家关系数据
	// 标记为all类型，便于懒加载时识别
	sql := fmt.Sprintf(`
		INSERT INTO %s (title, content, rewards, expire_time, target_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'all', NOW(), NOW())
	`, mailTableName)

	_, err := o.Raw(sql, title, content, rewards, expireTime).Exec()
	return err
}

// SendPersonalMail 发送个人邮件
func SendPersonalMail(appId, userId, title, content, rewards string) error {
	mailTableName := getMailTableName(appId)
	relationTableName := getMailRelationTableName(appId)

	o := orm.NewOrm()

	// 插入邮件内容
	mailSql := fmt.Sprintf("INSERT INTO %s (title, content, rewards, expire_time, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())", mailTableName)
	expireTime := time.Now().AddDate(0, 0, 30) // 默认30天过期
	result, err := o.Raw(mailSql, title, content, rewards, expireTime).Exec()
	if err != nil {
		return err
	}

	// 获取插入的邮件ID
	mailId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// 插入玩家关联记录
	relationSql := fmt.Sprintf("INSERT INTO %s (mail_id, player_id, status, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())", relationTableName)
	_, err = o.Raw(relationSql, mailId, userId, 0).Exec()
	if err != nil {
		return err
	}

	return nil
}

// PublishSystemMail 发布系统邮件
func PublishSystemMail(mailId int64, appId string) error {
	o := orm.NewOrm()

	// 使用动态表名查询邮件
	mailSystem := &MailSystem{}
	mailSystem.ID = mailId
	mailSystem.AppId = appId
	tableName := mailSystem.GetTableName(appId)

	// 检查表是否存在
	var tableCount int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&tableCount)
	if err != nil {
		return fmt.Errorf("检查表是否存在时出错: %v", err)
	}

	if tableCount == 0 {
		return fmt.Errorf("邮件系统表不存在，请先初始化邮件系统")
	}

	// 查询邮件
	sql := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tableName)
	err = o.Raw(sql, mailId).QueryRow(mailSystem)
	if err != nil {
		return fmt.Errorf("邮件不存在: %v", err)
	}

	// 更新邮件状态
	updateSql := fmt.Sprintf("UPDATE %s SET status = ?, send_time = ?, updated_at = ? WHERE id = ?", tableName)
	_, err = o.Raw(updateSql, "sent", time.Now(), time.Now(), mailId).Exec()
	if err != nil {
		return fmt.Errorf("更新邮件状态失败: %v", err)
	}

	// 根据目标类型创建用户邮件记录
	err = createMailRecords(mailSystem)
	if err != nil {
		return fmt.Errorf("创建用户邮件记录失败: %v", err)
	}

	return nil
}

// createMailRecords 创建用户邮件记录
func createMailRecords(mail *MailSystem) error {
	o := orm.NewOrm()

	switch mail.TargetType {
	case "all":
		// 发送给所有用户
		return createMailForAllUsers(o, mail)
	case "specific":
		// 发送给指定用户
		return createMailForSpecificUsers(o, mail)
	case "condition":
		// 根据条件发送
		return createMailForConditionUsers(o, mail)
	default:
		return fmt.Errorf("不支持的目标类型: %s", mail.TargetType)
	}
}

// createMailForAllUsers 为所有用户创建邮件（邮件内容已存在，这里处理广播逻辑）
func createMailForAllUsers(o orm.Ormer, mail *MailSystem) error {
	// 对于全员广播邮件，我们需要为所有用户创建邮件记录
	// 这里使用懒加载的方式，在用户请求邮件列表时动态创建

	// 由于是全员广播，我们不需要预先为每个用户创建记录
	// 而是在用户请求邮件时，检查是否有未读的广播邮件
	// 这样可以避免在用户数量很大时创建大量记录

	// 这里可以添加一些统计逻辑，比如记录广播邮件的发送状态等
	// 但不应该重复插入邮件内容

	return nil
}

// createMailForSpecificUsers 为指定用户创建邮件
func createMailForSpecificUsers(o orm.Ormer, mail *MailSystem) error {
	// 解析目标用户列表
	var targetUsers []string
	if mail.Targets != "" && mail.Targets != "[]" {
		err := json.Unmarshal([]byte(mail.Targets), &targetUsers)
		if err != nil {
			return fmt.Errorf("解析目标用户列表失败: %v", err)
		}
	}

	if len(targetUsers) == 0 {
		return fmt.Errorf("目标用户列表为空")
	}

	relationTableName := getMailRelationTableName(mail.AppId)

	// 邮件内容已经在CreateSystemMail中创建，这里只需要创建玩家关联记录
	// 使用已存在的邮件ID
	err := batchInsertMailPlayerRelations(o, relationTableName, mail.AppId, mail.ID, targetUsers)
	if err != nil {
		return fmt.Errorf("批量插入关联记录失败: %v", err)
	}

	// 更新总数量
	tableName := getMailTableName(mail.AppId)
	updateSQL := fmt.Sprintf("UPDATE %s SET total_count = ? WHERE id = ?", tableName)
	_, err = o.Raw(updateSQL, len(targetUsers), mail.ID).Exec()
	if err != nil {
		return fmt.Errorf("更新总数量失败: %v", err)
	}

	return nil
}

// createMailForConditionUsers 根据条件为用户创建邮件（邮件内容已存在，这里处理条件邮件逻辑）
func createMailForConditionUsers(o orm.Ormer, mail *MailSystem) error {
	// 对于条件邮件，邮件内容已经在CreateSystemMail中创建
	// 这里不需要再次插入邮件内容
	// 玩家关系数据将在用户请求邮件列表时根据条件懒加载创建

	// 这里可以添加一些条件邮件的特殊处理逻辑
	// 但不应该重复插入邮件内容

	return nil
}
