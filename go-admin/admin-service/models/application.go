package models

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

// Application 应用模型
type Application struct {
	BaseModel
	AppId       string `orm:"unique;size(50)" json:"app_id"`
	AppName     string `orm:"size(100)" json:"app_name"`
	AppSecret   string `orm:"size(100)" json:"app_secret"`
	Description string `orm:"type(text)" json:"description"`
	Status      int    `orm:"default(1)" json:"status"`
}

func (a *Application) TableName() string {
	return "applications"
}

// Insert 插入应用并创建相关数据表
func (a *Application) Insert() error {
	o := orm.NewOrm()

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
  player_id varchar(100) NOT NULL COMMENT '玩家ID',
  data longtext COMMENT '用户数据（JSON格式）',
  create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_player_id (player_id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户数据表_%s'`, cleanAppId, cleanAppId)

	// 排行榜表现在按需创建，不在应用创建时预先创建

	// 创建计数器表
	counterSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS counter_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  counter_name varchar(100) NOT NULL COMMENT '计数器名称',
  user_id varchar(100) DEFAULT NULL COMMENT '用户ID（全局计数器为空）',
  count bigint(20) NOT NULL DEFAULT 0 COMMENT '计数值',
  reset_time datetime DEFAULT NULL COMMENT '重置时间',
  reset_interval int(11) DEFAULT NULL COMMENT '重置间隔（秒）',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_counter_user (counter_name, user_id),
  KEY idx_counter_name (counter_name),
  KEY idx_reset_time (reset_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计数器数据表_%s'`, cleanAppId, cleanAppId)

	// 创建邮件表
	mailSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS mail_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  user_id varchar(100) NOT NULL COMMENT '收件人用户ID',
  title varchar(200) NOT NULL COMMENT '邮件标题',
  content text COMMENT '邮件内容',
  rewards text COMMENT '奖励物品（JSON格式）',
  status tinyint(1) NOT NULL DEFAULT 0 COMMENT '状态 0:未读 1:已读 2:已领取',
  expire_at datetime DEFAULT NULL COMMENT '过期时间',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  KEY idx_status (status),
  KEY idx_expire_at (expire_at),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件数据表_%s'`, cleanAppId, cleanAppId)

	// 创建游戏配置表
	gameConfigSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS game_config_%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  config_key varchar(100) NOT NULL COMMENT '配置键',
  config_value longtext COMMENT '配置值（JSON格式）',
  version varchar(50) DEFAULT NULL COMMENT '版本号',
  description varchar(255) DEFAULT NULL COMMENT '配置描述',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_config_key (config_key),
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
	return o.QueryTable(a.TableName()).Filter("app_id", appId).One(a)
}

// GetList 获取应用列表
func GetApplicationList(page, pageSize int, keyword string) ([]Application, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("applications")

	if keyword != "" {
		qs = qs.Filter("app_name__icontains", keyword).
			Filter("app_id__icontains", keyword)
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
	app.Status = 0
	_, err = o.Update(app, "status", "updated_at")
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
	count, err := o.QueryTable("applications").Filter("status", 1).Count()
	return count, err
}

// GetApplicationsByStatus 获取指定状态的应用数量
func GetApplicationsByStatus(status int) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("applications").Filter("status", status).Count()
	return count, err
}

func init() {
	orm.RegisterModel(new(Application))
}
