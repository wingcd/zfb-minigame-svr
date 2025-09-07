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

// MailSystem 邮件系统模型 - 对应数据库设计的mail_system表
type MailSystem struct {
	ID         int64     `orm:"pk;auto" json:"id"`
	AppId      string    `orm:"size(100);column(app_id)" json:"appId"`
	MailId     string    `orm:"size(100);column(mail_id)" json:"mailId"`                   // 邮件ID（唯一）
	Title      string    `orm:"size(200)" json:"title"`                                    // 邮件标题
	Content    string    `orm:"type(text)" json:"content"`                                 // 邮件内容
	Type       string    `orm:"size(50)" json:"type"`                                      // 邮件类型: system/activity/reward
	Sender     string    `orm:"size(100)" json:"sender"`                                   // 发送者
	Targets    string    `orm:"type(text)" json:"targets"`                                 // 目标用户（JSON数组，all表示全体）
	TargetType string    `orm:"size(50)" json:"targetType"`                                // 目标类型: all/specific/condition
	Condition  string    `orm:"type(text)" json:"condition"`                               // 发送条件（JSON）
	Rewards    string    `orm:"type(text)" json:"rewards"`                                 // 奖励列表（JSON数组）
	Status     string    `orm:"size(50);default(draft)" json:"status"`                     // 状态: draft/sent/expired
	SendTime   time.Time `orm:"type(datetime);null;column(send_time)" json:"sendTime"`     // 发送时间
	ExpireTime time.Time `orm:"type(datetime);null;column(expire_time)" json:"expireTime"` // 过期时间
	ReadCount  int       `orm:"default(0);column(read_count)" json:"readCount"`            // 已读数量
	TotalCount int       `orm:"default(0);column(total_count)" json:"totalCount"`          // 总发送数量
	CreateTime time.Time `orm:"auto_now_add;type(datetime);column(create_time)" json:"createTime"`
	UpdateTime time.Time `orm:"auto_now;type(datetime);column(update_time)" json:"updateTime"`
	CreatedBy  string    `orm:"size(100);column(created_by)" json:"createdBy"` // 创建者
}

// MailPlayerRelation 邮件-玩家关联表模型（动态表名: mail_player_relation_[appid]）
type MailPlayerRelation struct {
	ID        int64     `orm:"pk;auto" json:"id"`
	AppId     string    `orm:"size(100);column(app_id)" json:"appId"`
	MailId    int64     `orm:"column(mail_id)" json:"mailId"`               // 对应mail_[appid]表的id
	PlayerId  string    `orm:"size(100);column(player_id)" json:"playerId"` // 玩家ID
	Status    int       `orm:"default(0)" json:"status"`                    // 0:未读 1:已读 2:已领取
	CreatedAt time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
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
	orm.RegisterModel(new(MailSystem))
	orm.RegisterModel(new(MailPlayerRelation))
}

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
			values = append(values, "(?, ?, ?, 0, NOW(), NOW())")
			args = append(args, appId, mailId, playerId)
		}

		sql := fmt.Sprintf(`
			INSERT INTO %s (app_id, mail_id, player_id, status, created_at, updated_at) 
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
		WHERE r.app_id = ? AND r.player_id = ?
		  AND (m.expire_at IS NULL OR m.expire_at > NOW())
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, mailTableName, relationTableName)

	var mails []orm.Params
	_, err = o.Raw(sql, appId, playerId, pageSize, offset).Values(&mails)
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s m
		INNER JOIN %s r ON m.id = r.mail_id
		WHERE r.app_id = ? AND r.player_id = ?
		  AND (m.expire_at IS NULL OR m.expire_at > NOW())
	`, mailTableName, relationTableName)

	var total int64
	err = o.Raw(countSQL, appId, playerId).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	return mails, total, nil
}

// ensurePlayerMailRelations 确保玩家有所有应该接收的邮件关联记录
func ensurePlayerMailRelations(o orm.Ormer, appId, playerId, mailTableName, relationTableName string) error {
	// 查找所有该玩家还没有关联记录的邮件（这些可能是新发布的全员邮件）
	sql := fmt.Sprintf(`
		INSERT IGNORE INTO %s (app_id, mail_id, player_id, status, created_at, updated_at)
		SELECT ?, m.id, ?, 0, NOW(), NOW()
		FROM %s m
		LEFT JOIN %s r ON m.id = r.mail_id AND r.player_id = ?
		WHERE m.app_id = ? 
		  AND r.id IS NULL
		  AND (m.expire_at IS NULL OR m.expire_at > NOW())
	`, relationTableName, mailTableName, relationTableName)

	_, err := o.Raw(sql, appId, playerId, playerId, appId).Exec()
	return err
}

// MarkMailAsRead 标记邮件为已读
func MarkMailAsRead(appId, playerId string, mailId int64) error {
	o := orm.NewOrm()
	relationTableName := getMailRelationTableName(appId)

	sql := fmt.Sprintf(`
		UPDATE %s 
		SET status = 1, updated_at = NOW()
		WHERE app_id = ? AND player_id = ? AND mail_id = ? AND status = 0
	`, relationTableName)

	_, err := o.Raw(sql, appId, playerId, mailId).Exec()
	return err
}

// ClaimMailReward 领取邮件奖励
func ClaimMailReward(appId, playerId string, mailId int64) error {
	o := orm.NewOrm()
	relationTableName := getMailRelationTableName(appId)

	sql := fmt.Sprintf(`
		UPDATE %s 
		SET status = 2, updated_at = NOW()
		WHERE app_id = ? AND player_id = ? AND mail_id = ? AND status IN (0, 1)
	`, relationTableName)

	result, err := o.Raw(sql, appId, playerId, mailId).Exec()
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
	err := ensurePlayerMailRelations(o, appId, playerId, mailTableName, relationTableName)
	if err != nil {
		return 0, err
	}

	sql := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s r
		INNER JOIN %s m ON r.mail_id = m.id
		WHERE r.app_id = ? AND r.player_id = ? AND r.status = 0
		  AND (m.expire_at IS NULL OR m.expire_at > NOW())
	`, relationTableName, mailTableName)

	var count int64
	err = o.Raw(sql, appId, playerId).QueryRow(&count)
	return count, err
}

// CreateMail 创建邮件
func CreateMail(mail *MailSystem) error {
	o := orm.NewOrm()
	_, err := o.Insert(mail)
	return err
}

// UpdateMail 更新邮件
func UpdateMail(mail *MailSystem) error {
	o := orm.NewOrm()
	_, err := o.Update(mail)
	return err
}

// DeleteMail 删除邮件
func DeleteMail(id int64) error {
	o := orm.NewOrm()
	mail := &MailSystem{ID: id}
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

	// 检查表是否存在
	tableName := getMailTableName(appId)

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
	tableName := getMailTableName(appId)

	// 获取用户邮件
	sql := fmt.Sprintf("SELECT * FROM %s WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", tableName)
	params := []interface{}{userId, pageSize, (page - 1) * pageSize}

	var results []orm.Params
	_, err := o.Raw(sql, params...).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 计算总数
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id = ?", tableName)
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

// CreateSystemMail 创建系统邮件配置
func CreateSystemMail(mail *MailSystem) error {
	o := orm.NewOrm()

	// 生成唯一的邮件ID
	if mail.MailId == "" {
		mail.MailId = generateMailId()
	}

	// 设置默认值
	if mail.Status == "" {
		mail.Status = "draft"
	}
	if mail.Sender == "" {
		mail.Sender = "system"
	}

	_, err := o.Insert(mail)
	return err
}

// UpdateSystemMail 更新系统邮件
func UpdateSystemMail(mail *MailSystem) error {
	o := orm.NewOrm()
	_, err := o.Update(mail)
	return err
}

// DeleteSystemMail 删除系统邮件
func DeleteSystemMail(id int64) error {
	o := orm.NewOrm()
	mail := &MailSystem{ID: id}
	_, err := o.Delete(mail)
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

	// 获取邮件列表
	sql := fmt.Sprintf("SELECT * FROM %s ORDER BY created_at DESC LIMIT ? OFFSET ?", tableName)
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

// SendMail 发送邮件给特定用户
func SendMail(appId, userId, title, content, attachments string) error {
	o := orm.NewOrm()
	tableName := getMailTableName(appId)

	// 检查表是否存在
	var tableCount int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&tableCount)
	if err != nil {
		return err
	}
	if tableCount == 0 {
		return fmt.Errorf("邮件表不存在，请先初始化邮件系统")
	}

	// 插入邮件
	sql := fmt.Sprintf(`
		INSERT INTO %s (app_id, user_id, title, content, rewards, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, NOW(), NOW())
	`, tableName)

	_, err = o.Raw(sql, appId, userId, title, content, attachments).Exec()
	return err
}

// SendBroadcastMail 发送广播邮件
func SendBroadcastMail(appId, title, content, rewards string) error {
	o := orm.NewOrm()
	tableName := getMailTableName(appId)

	// 检查表是否存在
	var tableCount int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&tableCount)
	if err != nil {
		return err
	}
	if tableCount == 0 {
		return fmt.Errorf("邮件表不存在，请先初始化邮件系统")
	}

	// 广播邮件使用空的user_id表示给所有用户
	sql := fmt.Sprintf(`
		INSERT INTO %s (app_id, user_id, title, content, rewards, status, created_at, updated_at)
		VALUES (?, '', ?, ?, ?, 0, NOW(), NOW())
	`, tableName)

	_, err = o.Raw(sql, appId, title, content, rewards).Exec()
	return err
}

// SendPersonalMail 发送个人邮件
func SendPersonalMail(appId, userId, title, content, rewards string) error {
	mailTableName := getMailTableName(appId)
	relationTableName := getMailRelationTableName(appId)

	o := orm.NewOrm()

	// 生成邮件ID
	mailId := generateMailId()

	// 插入邮件内容
	mailSql := fmt.Sprintf("INSERT INTO %s (mail_id, title, content, rewards, expire_time, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", mailTableName)
	expireTime := time.Now().AddDate(0, 0, 30) // 默认30天过期
	_, err := o.Raw(mailSql, mailId, title, content, rewards, expireTime, time.Now(), time.Now()).Exec()
	if err != nil {
		return err
	}

	// 插入玩家关联记录
	relationSql := fmt.Sprintf("INSERT INTO %s (mail_id, player_id, is_read, is_claimed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", relationTableName)
	_, err = o.Raw(relationSql, mailId, userId, 0, 0, time.Now(), time.Now()).Exec()
	if err != nil {
		return err
	}

	return nil
}

// PublishSystemMail 发布系统邮件
func PublishSystemMail(mailId string, appId string) error {
	o := orm.NewOrm()

	// 使用动态表名查询邮件
	mailSystem := &MailSystem{}
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
	sql := fmt.Sprintf("SELECT * FROM %s WHERE mail_id = ? AND app_id = ?", tableName)
	err = o.Raw(sql, mailId, appId).QueryRow(mailSystem)
	if err != nil {
		return fmt.Errorf("邮件不存在: %v", err)
	}

	// 更新邮件状态
	updateSql := fmt.Sprintf("UPDATE %s SET status = ?, send_time = ?, update_time = ? WHERE mail_id = ? AND app_id = ?", tableName)
	_, err = o.Raw(updateSql, "sent", time.Now(), time.Now(), mailId, appId).Exec()
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

// createMailForAllUsers 为所有用户创建邮件
func createMailForAllUsers(o orm.Ormer, mail *MailSystem) error {
	cleanAppId := getCleanAppId(mail.AppId)
	userTableName := fmt.Sprintf("user_%s", cleanAppId)
	mailTableName := getMailTableName(mail.AppId)
	relationTableName := getMailRelationTableName(mail.AppId)

	// 1. 首先在邮件表中插入邮件内容
	mailInsertSQL := fmt.Sprintf(`
		INSERT INTO %s (app_id, title, content, rewards, expire_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`, mailTableName)

	result, err := o.Raw(mailInsertSQL, mail.AppId, mail.Title, mail.Content, mail.Rewards, mail.ExpireTime).Exec()
	if err != nil {
		return fmt.Errorf("插入邮件内容失败: %v", err)
	}

	mailId, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取邮件ID失败: %v", err)
	}

	// 2. 查询所有未封禁用户
	var playerIds []string
	_, err = o.Raw(fmt.Sprintf("SELECT player_id FROM %s WHERE banned = 0", userTableName)).QueryRows(&playerIds)
	if err != nil {
		return err
	}

	if len(playerIds) == 0 {
		return fmt.Errorf("没有可发送邮件的用户")
	}

	// 3. 批量插入邮件-玩家关联记录
	err = batchInsertMailPlayerRelations(o, relationTableName, mail.AppId, mailId, playerIds)
	if err != nil {
		return fmt.Errorf("批量插入关联记录失败: %v", err)
	}

	// 4. 更新总数量
	mail.TotalCount = len(playerIds)
	_, err = o.Update(mail, "total_count")

	return err
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

	mailTableName := getMailTableName(mail.AppId)
	relationTableName := getMailRelationTableName(mail.AppId)

	// 1. 首先在邮件表中插入邮件内容
	mailInsertSQL := fmt.Sprintf(`
		INSERT INTO %s (app_id, title, content, rewards, expire_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`, mailTableName)

	result, err := o.Raw(mailInsertSQL, mail.AppId, mail.Title, mail.Content, mail.Rewards, mail.ExpireTime).Exec()
	if err != nil {
		return fmt.Errorf("插入邮件内容失败: %v", err)
	}

	mailId, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取邮件ID失败: %v", err)
	}

	// 2. 批量插入邮件-玩家关联记录
	err = batchInsertMailPlayerRelations(o, relationTableName, mail.AppId, mailId, targetUsers)
	if err != nil {
		return fmt.Errorf("批量插入关联记录失败: %v", err)
	}

	// 3. 更新总数量
	mail.TotalCount = len(targetUsers)
	_, err = o.Update(mail, "total_count")

	return err
}

// createMailForConditionUsers 根据条件为用户创建邮件
func createMailForConditionUsers(o orm.Ormer, mail *MailSystem) error {
	// 解析条件
	var condition map[string]interface{}
	if mail.Condition != "" {
		err := json.Unmarshal([]byte(mail.Condition), &condition)
		if err != nil {
			return fmt.Errorf("解析发送条件失败: %v", err)
		}
	}

	// 构建查询条件
	cleanAppId := getCleanAppId(mail.AppId)
	userTableName := fmt.Sprintf("user_%s", cleanAppId)
	mailTableName := getMailTableName(mail.AppId)
	relationTableName := getMailRelationTableName(mail.AppId)
	whereClause := "banned = 0"

	// 处理等级范围条件
	if minLevel, ok := condition["minLevel"]; ok {
		whereClause += fmt.Sprintf(" AND level >= %v", minLevel)
	}
	if maxLevel, ok := condition["maxLevel"]; ok {
		whereClause += fmt.Sprintf(" AND level <= %v", maxLevel)
	}

	// 查询符合条件的用户
	var playerIds []string
	sql := fmt.Sprintf("SELECT player_id FROM %s WHERE %s", userTableName, whereClause)
	_, err := o.Raw(sql).QueryRows(&playerIds)
	if err != nil {
		return err
	}

	if len(playerIds) == 0 {
		return fmt.Errorf("没有符合条件的用户")
	}

	// 1. 首先在邮件表中插入邮件内容
	mailInsertSQL := fmt.Sprintf(`
		INSERT INTO %s (app_id, title, content, rewards, expire_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`, mailTableName)

	result, err := o.Raw(mailInsertSQL, mail.AppId, mail.Title, mail.Content, mail.Rewards, mail.ExpireTime).Exec()
	if err != nil {
		return fmt.Errorf("插入邮件内容失败: %v", err)
	}

	mailId, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取邮件ID失败: %v", err)
	}

	// 2. 批量插入邮件-玩家关联记录
	err = batchInsertMailPlayerRelations(o, relationTableName, mail.AppId, mailId, playerIds)
	if err != nil {
		return fmt.Errorf("批量插入关联记录失败: %v", err)
	}

	// 3. 更新总数量
	mail.TotalCount = len(playerIds)
	_, err = o.Update(mail, "total_count")

	return err
}

// generateMailId 生成邮件ID
func generateMailId() string {
	return fmt.Sprintf("mail_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}
