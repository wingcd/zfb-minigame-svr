package models

import (
	"admin-service/utils"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

// Application 应用模型
type Application struct {
	BaseModel
	AppId         string `orm:"unique;size(50);column(app_id)" json:"appId"`               // 应用ID（唯一）
	AppName       string `orm:"size(100);column(app_name)" json:"appName"`                 // 应用名称
	Description   string `orm:"type(text);column(description)" json:"description"`         // 应用描述
	ChannelAppId  string `orm:"size(100);column(channel_app_id)" json:"channelAppId"`      // 渠道应用ID
	ChannelAppKey string `orm:"size(100);column(channel_app_key)" json:"channelAppKey"`    // 渠道应用密钥
	Category      string `orm:"size(50);default('game');column(category)" json:"category"` // 应用分类: game/tool/social
	Platform      string `orm:"size(50);column(platform)" json:"platform"`                 // 平台: alipay/wechat/baidu
	Status        string `orm:"size(20);default('active');column(status)" json:"status"`   // 状态: active/inactive/pending
	Version       string `orm:"size(50);column(version)" json:"version"`                   // 当前版本
	MinVersion    string `orm:"size(50);column(min_version)" json:"minVersion"`            // 最低支持版本
	Settings      string `orm:"type(text);column(settings)" json:"settings"`               // 应用设置(JSON格式)
	UserCount     int64  `orm:"default(0);column(user_count)" json:"userCount"`            // 用户数量
	ScoreCount    int64  `orm:"default(0);column(score_count)" json:"scoreCount"`          // 分数记录数
	DailyActive   int64  `orm:"default(0);column(daily_active)" json:"dailyActive"`        // 日活跃用户
	MonthlyActive int64  `orm:"default(0);column(monthly_active)" json:"monthlyActive"`    // 月活跃用户
	CreatedBy     string `orm:"size(50);column(created_by)" json:"createdBy"`              // 创建者
}

func (a *Application) TableName() string {
	return "apps"
}

// Insert 插入应用并创建相关数据表
func (a *Application) Insert() error {
	o := orm.NewOrm()

	// 生成AppId（如果没有设置）
	if a.AppId == "" {
		a.AppId = utils.GenerateAppId()
	}

	// 开始事务
	tx, err := o.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 插入应用记录
	_, err = tx.Insert(a)
	if err != nil {
		return fmt.Errorf("insert application failed: %v", err)
	}

	// 创建相关数据表（在事务外执行，因为DDL操作不能在事务中）
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction failed: %v", err)
	}

	// 事务提交后再创建数据表
	err = a.createAppTables()
	if err != nil {
		// 如果创建表失败，需要删除已插入的应用记录
		o.Delete(a)
		return fmt.Errorf("create app tables failed: %v", err)
	}

	return nil
}

// createAppTables 为应用创建相关数据表
func (a *Application) createAppTables() error {
	o := orm.NewOrm()

	// 清理应用ID，确保表名安全
	cleanAppId := strings.ReplaceAll(a.AppId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")

	// 创建用户数据表（按照数据库设计.md的正确结构）
	userDataSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS user_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  open_id varchar(100) NOT NULL COMMENT '用户唯一标识',
  player_id varchar(100) NOT NULL COMMENT '玩家ID（唯一，自动生成）',
  token varchar(255) COMMENT '登录Token',
  nickname varchar(100) COMMENT '昵称',
  avatar varchar(500) COMMENT '头像URL',
  data longtext COMMENT '游戏数据（JSON格式）',
  level int(11) NOT NULL DEFAULT 1 COMMENT '等级',
  exp bigint(20) NOT NULL DEFAULT 0 COMMENT '经验值',
  coin bigint(20) NOT NULL DEFAULT 0 COMMENT '金币',
  diamond bigint(20) NOT NULL DEFAULT 0 COMMENT '钻石',
  vip_level int(11) NOT NULL DEFAULT 0 COMMENT 'VIP等级',
  banned tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否封禁',
  ban_reason varchar(500) COMMENT '封禁原因',
  ban_expire datetime COMMENT '封禁到期时间',
  login_count int(11) NOT NULL DEFAULT 0 COMMENT '登录次数',
  last_login_time datetime COMMENT '最后登录时间',
  last_login_ip varchar(50) COMMENT '最后登录IP',
  register_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_open_id (open_id),
  UNIQUE KEY uk_player_id (player_id),
  KEY idx_token (token),
  KEY idx_updated_at (updated_at),
  KEY idx_created_at (created_at),
  KEY idx_banned_ban_expire (banned, ban_expire)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户数据表_%s'`, cleanAppId, cleanAppId)

	// 创建排行榜统计表
	leaderboardStatsSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS leaderboard_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  type varchar(50) NOT NULL COMMENT '排行榜类型',
  player_id varchar(100) NOT NULL COMMENT '玩家ID',
  score bigint(20) NOT NULL DEFAULT 0 COMMENT '分数',
  extra_data text COMMENT '额外数据（JSON格式）',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_leaderboard_user (type, player_id),
  KEY idx_leaderboard_score (type, score DESC),
  KEY idx_leaderboard_type (type),
  KEY idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排行榜数据表_%s'`, cleanAppId, cleanAppId)

	// 创建计数器表
	counterSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS counter_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  counter_name varchar(100) NOT NULL COMMENT '计数器名称',
  player_id varchar(100) DEFAULT NULL COMMENT '用户ID（全局计数器为空）',
  count bigint(20) NOT NULL DEFAULT 0 COMMENT '计数值',
  reset_time datetime DEFAULT NULL COMMENT '重置时间',
  reset_interval int(11) DEFAULT NULL COMMENT '重置间隔（秒）',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_counter_user (counter_name, player_id),
  KEY idx_counter_name (counter_name),
  KEY idx_reset_time (reset_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计数器数据表_%s'`, cleanAppId, cleanAppId)

	// 创建邮件表（存储邮件内容）
	mailSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS mail_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  title varchar(200) NOT NULL COMMENT '邮件标题',
  content text COMMENT '邮件内容',
  type varchar(50) NOT NULL DEFAULT 'system' COMMENT '邮件类型: system/activity/reward',
  sender varchar(100) NOT NULL DEFAULT 'system' COMMENT '发送者',
  targets text COMMENT '目标用户（JSON数组，all表示全体）',
  target_type varchar(50) NOT NULL DEFAULT 'all' COMMENT '目标类型: all/specific/condition',
  send_condition text COMMENT '发送条件（JSON）',
  rewards text COMMENT '奖励列表（JSON数组）',
  status varchar(50) NOT NULL DEFAULT 'draft' COMMENT '状态: draft/sent/expired',
  send_time datetime DEFAULT NULL COMMENT '发送时间',
  expire_time datetime DEFAULT NULL COMMENT '过期时间',
  read_count int NOT NULL DEFAULT 0 COMMENT '已读数量',
  total_count int NOT NULL DEFAULT 0 COMMENT '总发送数量',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  created_by varchar(100) DEFAULT NULL COMMENT '创建者',
  PRIMARY KEY (id),
  KEY idx_type (type),
  KEY idx_status (status),
  KEY idx_target_type (target_type),
  KEY idx_expire_time (expire_time),
  KEY idx_send_time (send_time),
  KEY idx_create_time (created_at),
  KEY idx_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件数据表_%s'`, cleanAppId, cleanAppId)

	// 创建邮件-玩家关联表（存储邮件发送状态）
	mailPlayerRelationSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS mail_player_relation_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  mail_id bigint(20) NOT NULL COMMENT '邮件ID（关联mail_%s表）',
  player_id varchar(100) NOT NULL COMMENT '玩家ID',
  status tinyint(1) NOT NULL DEFAULT 0 COMMENT '状态 0:未读 1:已读 2:已领取',
  received_at datetime DEFAULT NULL COMMENT '接收时间',
  read_at datetime DEFAULT NULL COMMENT '阅读时间',
  claim_at datetime DEFAULT NULL COMMENT '领取时间',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mail_player (mail_id, player_id),
  KEY idx_mail_id (mail_id),
  KEY idx_player_id (player_id),
  KEY idx_status (status),
  KEY idx_received_at (received_at),
  KEY idx_create_time (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件玩家关联表_%s'`, cleanAppId, cleanAppId, cleanAppId)

	// 创建游戏配置表
	gameConfigSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS game_config_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  config_key varchar(100) NOT NULL COMMENT '配置键',
  config_type varchar(50) DEFAULT NULL COMMENT '配置类型',
  config_value longtext COMMENT '配置值（JSON格式）',
  version varchar(50) DEFAULT NULL COMMENT '版本号',
  description varchar(255) DEFAULT NULL COMMENT '配置描述',
  is_active tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否激活',
  priority int(11) NOT NULL DEFAULT 1 COMMENT '优先级',
  tags varchar(255) DEFAULT NULL COMMENT '标签',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_by varchar(50) DEFAULT NULL COMMENT '创建者',
  PRIMARY KEY (id),
  UNIQUE KEY uk_config_key (config_key),
  KEY idx_version (version),
  KEY idx_is_active (is_active),
  KEY idx_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏配置表_%s'`, cleanAppId, cleanAppId)

	// 执行创建表的SQL
	sqls := []string{userDataSQL, leaderboardStatsSQL, counterSQL, mailSQL, mailPlayerRelationSQL, gameConfigSQL}
	for _, sql := range sqls {
		_, err := o.Raw(sql).Exec()
		if err != nil {
			return fmt.Errorf("create table failed: %v, sql: %s", err, sql)
		}
	}

	return nil
}

// Update 更新应用
func (a *Application) Update(fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(a, fields...)
	return err
}

// GetById 根据ID获取应用
func (a *Application) GetById(id int64) error {
	o := orm.NewOrm()
	a.ID = id
	return o.Read(a)
}

// GetByAppId 根据AppId获取应用
func (a *Application) GetByAppId(appId string) error {
	o := orm.NewOrm()
	return o.QueryTable(a.TableName()).Filter("appId", appId).One(a)
}

// GetList 获取应用列表
func GetApplicationList(page, pageSize int, keyword string) ([]Application, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("apps")

	if keyword != "" {
		qs = qs.Filter("app_name__icontains", keyword).
			Filter("appId__icontains", keyword)
	}

	total, _ := qs.Count()

	var apps []Application
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-id").Limit(pageSize, offset).All(&apps)

	return apps, total, err
}

// DeleteApplication 删除应用（软删除，只标记为禁用状态）
func DeleteApplication(id int64) error {
	o := orm.NewOrm()

	// 获取应用信息
	app := &Application{}
	app.ID = id
	err := o.Read(app)
	if err != nil {
		return err
	}

	// 软删除：只标记应用为禁用状态
	app.Status = "inactive" // inactive = 禁用, active = 启用
	_, err = o.Update(app, "status", "updatedAt")
	return err
}

// HardDeleteApplication 硬删除应用（彻底删除应用和相关数据表，仅超级管理员可用）
func HardDeleteApplication(id int64) error {
	o := orm.NewOrm()

	// 获取应用信息
	app := &Application{}
	app.ID = id
	err := o.Read(app)
	if err != nil {
		return err
	}

	// 开始事务
	tx, err := o.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 删除应用相关数据表
	err = app.dropAppTables()
	if err != nil {
		return fmt.Errorf("drop app tables failed: %v", err)
	}

	// 删除应用记录
	_, err = tx.Delete(app)
	if err != nil {
		return fmt.Errorf("delete application failed: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction failed: %v", err)
	}

	return nil
}

// dropAppTables 删除应用相关数据表（危险操作，谨慎使用）
func (a *Application) dropAppTables() error {
	o := orm.NewOrm()

	cleanAppId := strings.ReplaceAll(a.AppId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")

	tables := []string{
		fmt.Sprintf("user_%s", cleanAppId),
		fmt.Sprintf("leaderboard_%s", cleanAppId),
		fmt.Sprintf("counter_%s", cleanAppId),
		fmt.Sprintf("mail_%s", cleanAppId),
		fmt.Sprintf("mail_player_relation_%s", cleanAppId),
		fmt.Sprintf("game_config_%s", cleanAppId),
	}

	for _, table := range tables {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		_, err := o.Raw(sql).Exec()
		if err != nil {
			return fmt.Errorf("drop table %s failed: %v", table, err)
		}
	}

	return nil
}

// GetTotalApplications 获取应用总数
func GetTotalApplications() (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("apps").Filter("status", "active").Count() // active = 启用
	return count, err
}

// GetApplicationsByStatus 获取指定状态的应用数量
func GetApplicationsByStatus(status string) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("apps").Filter("status", status).Count()
	return count, err
}

func init() {
	orm.RegisterModel(new(Application))
}
