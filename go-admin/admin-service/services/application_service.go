package services

import (
	"admin-service/models"
	"admin-service/utils"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// ApplicationService 应用服务
type ApplicationService struct{}

// NewApplicationService 创建应用服务实例
func NewApplicationService() *ApplicationService {
	return &ApplicationService{}
}

// CreateApplication 创建应用
func (s *ApplicationService) CreateApplication(appId, appName, description string) (*models.Application, error) {
	o := orm.NewOrm()

	// 检查应用ID是否已存在
	exists := o.QueryTable("apps").Filter("appId", appId).Exist()
	if exists {
		return nil, fmt.Errorf("应用ID已存在")
	}

	// 创建应用记录
	app := &models.Application{
		AppId:       appId,
		AppName:     appName,
		Description: description,
		Status:      "active", // active = 启用
	}
	app.CreatedAt = time.Now()
	app.UpdatedAt = app.CreatedAt

	// 插入数据库
	_, err := o.Insert(app)
	if err != nil {
		return nil, fmt.Errorf("创建应用失败: %v", err)
	}

	// 创建应用对应的动态表
	err = s.CreateAppTables(appId)
	if err != nil {
		// 如果创建表失败，删除应用记录
		o.Delete(app)
		return nil, fmt.Errorf("创建应用数据表失败: %v", err)
	}

	return app, nil
}

// GetApplicationList 获取应用列表
func (s *ApplicationService) GetApplicationList(page, pageSize int, keyword string) (*models.PageData, error) {
	o := orm.NewOrm()

	qs := o.QueryTable("apps")

	// 搜索条件
	if keyword != "" {
		cond := orm.NewCondition()
		cond = cond.Or("appName__icontains", keyword).
			Or("appId__icontains", keyword)
		qs = qs.SetCond(cond)
	}

	// 统计总数
	total, err := qs.Count()
	if err != nil {
		return nil, fmt.Errorf("统计应用数量失败: %v", err)
	}

	// 分页查询
	var apps []models.Application
	offset := (page - 1) * pageSize
	_, err = qs.OrderBy("-id").Limit(pageSize, offset).All(&apps)
	if err != nil {
		return nil, fmt.Errorf("查询应用列表失败: %v", err)
	}

	return &models.PageData{
		List:     apps,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetApplication 获取应用详情
func (s *ApplicationService) GetApplication(appId string) (*models.Application, error) {
	o := orm.NewOrm()

	app := &models.Application{}
	err := o.QueryTable("apps").Filter("appId", appId).One(app)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("应用不存在")
		}
		return nil, fmt.Errorf("查询应用失败: %v", err)
	}

	return app, nil
}

// UpdateApplication 更新应用信息
func (s *ApplicationService) UpdateApplication(appId, appName, description string, status string) error {
	o := orm.NewOrm()

	// 查找应用
	app := &models.Application{}
	err := o.QueryTable("apps").Filter("appId", appId).One(app)
	if err != nil {
		if err == orm.ErrNoRows {
			return fmt.Errorf("应用不存在")
		}
		return fmt.Errorf("查询应用失败: %v", err)
	}

	// 更新信息
	app.AppName = appName
	app.Description = description
	app.Status = status
	app.UpdatedAt = time.Now()

	_, err = o.Update(app, "appName", "description", "status", "updateTime")
	if err != nil {
		return fmt.Errorf("更新应用失败: %v", err)
	}

	return nil
}

// DeleteApplication 删除应用
func (s *ApplicationService) DeleteApplication(appId string) error {
	o := orm.NewOrm()

	// 开启事务
	tx, err := o.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 删除应用记录
	_, err = tx.QueryTable("apps").Filter("appId", appId).Delete()
	if err != nil {
		return fmt.Errorf("删除应用记录失败: %v", err)
	}

	// 删除应用对应的动态表
	err = s.DropAppTables(appId)
	if err != nil {
		return fmt.Errorf("删除应用数据表失败: %v", err)
	}

	return nil
}

// ResetAppSecret 重置应用密钥
func (s *ApplicationService) ResetAppSecret(appId string) (string, error) {
	o := orm.NewOrm()

	// 查找应用
	app := &models.Application{}
	err := o.QueryTable("apps").Filter("appId", appId).One(app)
	if err != nil {
		if err == orm.ErrNoRows {
			return "", fmt.Errorf("应用不存在")
		}
		return "", fmt.Errorf("查询应用失败: %v", err)
	}

	// 生成新密钥
	newSecret := utils.GenerateRandomString(32)
	app.AppSecret = newSecret
	app.UpdatedAt = time.Now()

	_, err = o.Update(app, "appSecret", "updateTime")
	if err != nil {
		return "", fmt.Errorf("重置应用密钥失败: %v", err)
	}

	return newSecret, nil
}

// CreateAppTables 创建应用对应的动态表
func (s *ApplicationService) CreateAppTables(appId string) error {
	o := orm.NewOrm()

	// 清理appId中的特殊字符
	cleanAppId := utils.CleanAppId(appId)

	// 创建用户数据表
	userDataSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS user_data_%s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			userId VARCHAR(100) NOT NULL,
			data LONGTEXT,
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_id (userId),
			KEY idx_update_time (updateTime)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户数据表'
	`, cleanAppId)

	// 排行榜表现在按需创建，不在应用创建时预先创建

	// 创建计数器表
	counterSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS counter_%s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			counterName VARCHAR(100) NOT NULL,
			userId VARCHAR(100) DEFAULT NULL,
			count BIGINT DEFAULT 0,
			resetTime DATETIME DEFAULT NULL,
			resetInterval INT DEFAULT NULL,
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_counter_user (counterName, userId),
			KEY idx_counter_name (counterName),
			KEY idx_reset_time (resetTime)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计数器数据表'
	`, cleanAppId)

	// 创建邮件表
	mailSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS mail_%s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			userId VARCHAR(100) NOT NULL,
			title VARCHAR(200) NOT NULL,
			content TEXT,
			rewards TEXT,
			status TINYINT DEFAULT 0 COMMENT '0:未读 1:已读 2:已领取',
			expireAt DATETIME DEFAULT NULL,
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			KEY idx_user_id (userId),
			KEY idx_status (status),
			KEY idx_expire_at (expireAt),
			KEY idx_create_time (createTime)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件数据表'
	`, cleanAppId)

	// 创建游戏配置表
	configSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS game_config_%s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			configKey VARCHAR(100) NOT NULL,
			configValue LONGTEXT,
			version VARCHAR(50) DEFAULT NULL,
			description VARCHAR(255) DEFAULT NULL,
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_config_key (configKey),
			KEY idx_version (version)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏配置表'
	`, cleanAppId)

	// 执行SQL语句
	sqls := []string{userDataSQL, counterSQL, mailSQL, configSQL}
	for _, sql := range sqls {
		_, err := o.Raw(sql).Exec()
		if err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}

	return nil
}

// DropAppTables 删除应用对应的动态表
func (s *ApplicationService) DropAppTables(appId string) error {
	o := orm.NewOrm()

	// 清理appId中的特殊字符
	cleanAppId := utils.CleanAppId(appId)

	// 删除表的SQL语句
	tables := []string{
		fmt.Sprintf("user_data_%s", cleanAppId),
		fmt.Sprintf("leaderboard_%s", cleanAppId), // 仍需删除可能存在的排行榜表
		fmt.Sprintf("counter_%s", cleanAppId),
		fmt.Sprintf("mail_%s", cleanAppId),
		fmt.Sprintf("game_config_%s", cleanAppId),
	}

	for _, table := range tables {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		_, err := o.Raw(sql).Exec()
		if err != nil {
			return fmt.Errorf("删除表 %s 失败: %v", table, err)
		}
	}

	return nil
}

// GetApplicationStats 获取应用统计信息
func (s *ApplicationService) GetApplicationStats(appId string) (map[string]interface{}, error) {
	o := orm.NewOrm()

	// 检查应用是否存在
	exists := o.QueryTable("apps").Filter("appId", appId).Exist()
	if !exists {
		return nil, fmt.Errorf("应用不存在")
	}

	cleanAppId := utils.CleanAppId(appId)
	stats := make(map[string]interface{})

	// 统计用户数量
	userCountSQL := fmt.Sprintf("SELECT COUNT(*) as count FROM user_data_%s", cleanAppId)
	var userCount int64
	err := o.Raw(userCountSQL).QueryRow(&userCount)
	if err == nil {
		stats["user_count"] = userCount
	} else {
		stats["user_count"] = 0
	}

	// 统计排行榜记录数
	leaderboardCountSQL := fmt.Sprintf("SELECT COUNT(*) as count FROM leaderboard_%s", cleanAppId)
	var leaderboardCount int64
	err = o.Raw(leaderboardCountSQL).QueryRow(&leaderboardCount)
	if err == nil {
		stats["leaderboard_count"] = leaderboardCount
	} else {
		stats["leaderboard_count"] = 0
	}

	// 统计计数器数量
	counterCountSQL := fmt.Sprintf("SELECT COUNT(*) as count FROM counter_%s", cleanAppId)
	var counterCount int64
	err = o.Raw(counterCountSQL).QueryRow(&counterCount)
	if err == nil {
		stats["counter_count"] = counterCount
	} else {
		stats["counter_count"] = 0
	}

	// 统计邮件数量
	mailCountSQL := fmt.Sprintf("SELECT COUNT(*) as count FROM mail_%s", cleanAppId)
	var mailCount int64
	err = o.Raw(mailCountSQL).QueryRow(&mailCount)
	if err == nil {
		stats["mail_count"] = mailCount
	} else {
		stats["mail_count"] = 0
	}

	// 统计配置数量
	configCountSQL := fmt.Sprintf("SELECT COUNT(*) as count FROM game_config_%s", cleanAppId)
	var configCount int64
	err = o.Raw(configCountSQL).QueryRow(&configCount)
	if err == nil {
		stats["config_count"] = configCount
	} else {
		stats["config_count"] = 0
	}

	return stats, nil
}
