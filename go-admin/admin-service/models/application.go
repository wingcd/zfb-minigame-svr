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
	AppId         string `orm:"unique;size(50);column(appId)" json:"appId"`                // 应用ID（唯一）
	AppName       string `orm:"size(100);column(appName)" json:"appName"`                  // 应用名称
	Description   string `orm:"type(text);column(description)" json:"description"`         // 应用描述
	ChannelAppId  string `orm:"size(100);column(channelAppId)" json:"channelAppId"`        // 渠道应用ID
	ChannelAppKey string `orm:"size(100);column(channelAppKey)" json:"channelAppKey"`      // 渠道应用密钥
	AppSecret     string `orm:"size(100);column(appSecret)" json:"appSecret"`              // 应用密钥
	Category      string `orm:"size(50);default('game');column(category)" json:"category"` // 应用分类: game/tool/social
	Platform      string `orm:"size(50);column(platform)" json:"platform"`                 // 平台: alipay/wechat/baidu
	Status        string `orm:"size(20);default('active');column(status)" json:"status"`   // 状态: active/inactive/pending
	Version       string `orm:"size(50);column(version)" json:"version"`                   // 当前版本
	MinVersion    string `orm:"size(50);column(minVersion)" json:"minVersion"`             // 最低支持版本
	Settings      string `orm:"type(text);column(settings)" json:"settings"`               // 应用设置(JSON格式)
	UserCount     int64  `orm:"default(0);column(userCount)" json:"userCount"`             // 用户数量
	ScoreCount    int64  `orm:"default(0);column(scoreCount)" json:"scoreCount"`           // 分数记录数
	DailyActive   int64  `orm:"default(0);column(dailyActive)" json:"dailyActive"`         // 日活跃用户
	MonthlyActive int64  `orm:"default(0);column(monthlyActive)" json:"monthlyActive"`     // 月活跃用户
	CreatedBy     string `orm:"size(50);column(createdBy)" json:"createdBy"`               // 创建者
}

func (a *Application) TableName() string {
	return "apps"
}

// Insert 插入应用并创建相关数据表
func (a *Application) Insert() error {
	o := orm.NewOrm()

	// 生成AppId和AppSecret（如果没有设置）
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

	// 创建相关数据表
	err = a.createAppTables()
	if err != nil {
		return fmt.Errorf("create app tables failed: %v", err)
	}

	// 提交事务
	return tx.Commit()
}

// createAppTables 为应用创建相关数据表
func (a *Application) createAppTables() error {
	o := orm.NewOrm()

	// 清理应用ID，确保表名安全
	cleanAppId := strings.ReplaceAll(a.AppId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")

	// 创建用户数据表（与用户管理模块匹配）
	userDataSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS user_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  playerId varchar(100) NOT NULL COMMENT '玩家ID',
  data longtext COMMENT '用户数据（JSON格式）',
  createdAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_playerId (playerId),
  KEY idx_updatedAt (updatedAt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户数据表_%s'`, cleanAppId, cleanAppId)

	// 排行榜表现在按需创建，不在应用创建时预先创建

	// 创建计数器表
	counterSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS counter_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  counterName varchar(100) NOT NULL COMMENT '计数器名称',
  playerId varchar(100) DEFAULT NULL COMMENT '用户ID（全局计数器为空）',
  count bigint(20) NOT NULL DEFAULT 0 COMMENT '计数值',
  resetTime datetime DEFAULT NULL COMMENT '重置时间',
  resetInterval int(11) DEFAULT NULL COMMENT '重置间隔（秒）',
  createdAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_counter_user (counterName, playerId),
  KEY idx_counterName (counterName),
  KEY idx_resetTime (resetTime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计数器数据表_%s'`, cleanAppId, cleanAppId)

	// 创建邮件表
	mailSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS mail_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  playerId varchar(100) NOT NULL COMMENT '收件人用户ID',
  title varchar(200) NOT NULL COMMENT '邮件标题',
  content text COMMENT '邮件内容',
  rewards text COMMENT '奖励物品（JSON格式）',
  status tinyint(1) NOT NULL DEFAULT 0 COMMENT '状态 0:未读 1:已读 2:已领取',
  expireTime datetime DEFAULT NULL COMMENT '过期时间',
  createdAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_playerId (playerId),
  KEY idx_status (status),
  KEY idx_expireTime (expireAt),
  KEY idx_createdAt (createdAt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件数据表_%s'`, cleanAppId, cleanAppId)

	// 创建游戏配置表
	gameConfigSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS game_config_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  configKey varchar(100) NOT NULL COMMENT '配置键',
  configValue longtext COMMENT '配置值（JSON格式）',
  version varchar(50) DEFAULT NULL COMMENT '版本号',
  description varchar(255) DEFAULT NULL COMMENT '配置描述',
  createdAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_configKey (configKey),
  KEY idx_version (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏配置表_%s'`, cleanAppId, cleanAppId)

	// 执行创建表的SQL
	sqls := []string{userDataSQL, counterSQL, mailSQL, gameConfigSQL}
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
	a.Id = id
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

// Delete 删除应用（软删除）
func DeleteApplication(id int64) error {
	o := orm.NewOrm()

	// 获取应用信息
	app := &Application{}
	app.Id = id
	err := o.Read(app)
	if err != nil {
		return err
	}

	// 这里可以选择是否删除相关数据表
	// 为了安全起见，我们只是标记应用为禁用状态
	app.Status = "inactive" // inactive = 禁用, active = 启用
	_, err = o.Update(app, "status", "updatedAt")
	return err
}

// dropAppTables 删除应用相关数据表（危险操作，谨慎使用）
func (a *Application) dropAppTables() error {
	o := orm.NewOrm()

	cleanAppId := strings.ReplaceAll(a.AppId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")

	tables := []string{
		fmt.Sprintf("user_data_%s", cleanAppId),
		fmt.Sprintf("leaderboard_%s", cleanAppId),
		fmt.Sprintf("counter_%s", cleanAppId),
		fmt.Sprintf("mail_%s", cleanAppId),
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
