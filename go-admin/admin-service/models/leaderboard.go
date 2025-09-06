package models

import (
	"fmt"
	"time"

	"admin-service/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// LeaderboardConfig 排行榜配置结构（管理表）
type LeaderboardConfig struct {
	BaseModel
	AppId       string `orm:"size(100)" json:"appId"`
	Type        string `orm:"size(100)" json:"type"`
	Name        string `orm:"size(200)" json:"name"`
	Description string `orm:"type(text)" json:"description"`
	ResetType   string `orm:"size(50);default(never)" json:"resetType"` // never, daily, weekly, monthly
	MaxEntries  int    `orm:"default(1000)" json:"maxEntries"`
	ScoreType   string `orm:"size(20);default(higher_better)" json:"scoreType"` // higher_better, lower_better
	Status      int    `orm:"default(1)" json:"status"`                         // 1=启用, 0=禁用
}

// Leaderboard 排行榜配置结构（兼容性保持）
type Leaderboard struct {
	ID          int64     `orm:"pk;auto" json:"id"`
	AppId       string    `orm:"size(100)" json:"appId"`
	Type        string    `orm:"size(100)" json:"type"`
	Name        string    `orm:"size(200)" json:"name"`
	Description string    `orm:"type(text)" json:"description"`
	ResetType   string    `orm:"size(50)" json:"resetType"` // never, daily, weekly, monthly
	MaxEntries  int       `orm:"default(1000)" json:"maxEntries"`
	ScoreType   string    `orm:"size(20);default(higher_better)" json:"scoreType"` // higher_better, lower_better
	Status      int       `orm:"default(1)" json:"status"`                         // 1=启用, 0=禁用
	CreateTime  time.Time `orm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime  time.Time `orm:"auto_now;type(datetime)" json:"updateTime"`
}

// LeaderboardData 排行榜数据模型（动态表）
type LeaderboardData struct {
	Id              int64  `orm:"auto" json:"id"`
	LeaderboardName string `orm:"size(100)" json:"leaderboard_name"`
	UserId          string `orm:"size(100)" json:"user_id"`
	Score           int64  `orm:"default(0)" json:"score"`
	ExtraData       string `orm:"type(text)" json:"extra_data"`
	CreatedAt       string `orm:"auto_now_add;type(datetime)" json:"create_time"`
	UpdatedAt       string `orm:"auto_now;type(datetime)" json:"update_time"`
}

// LeaderboardScore 排行榜分数记录（兼容性保持）
type LeaderboardScore struct {
	ID            int64     `orm:"pk;auto" json:"id"`
	AppId         string    `orm:"size(100)" json:"appId"`
	LeaderboardId string    `orm:"size(100)" json:"leaderboardId"`
	PlayerId      string    `orm:"size(100)" json:"playerId"`
	OpenId        string    `orm:"size(100)" json:"openId"`
	Score         int64     `json:"score"`
	ExtraData     string    `orm:"type(text)" json:"extraData"`
	HasUserInfo   int       `orm:"default(0)" json:"hasUserInfo"` // 0=无用户信息, 1=有用户信息
	CreateTime    time.Time `orm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime    time.Time `orm:"auto_now;type(datetime)" json:"updateTime"`
}

// GetLeaderboardCount 获取排行榜数量统计
func GetLeaderboardCount(appId string) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("leaderboard_config").Filter("app_id", appId).Count()
	return count, err
}

// GetLeaderboardList 获取排行榜配置列表
func GetLeaderboardList(appId string, page, pageSize int, leaderboardName string) ([]*LeaderboardConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("leaderboard_config").Filter("app_id", appId)

	// 添加名称筛选
	if leaderboardName != "" {
		qs = qs.Filter("name__icontains", leaderboardName)
	}

	total, _ := qs.Count()

	var leaderboards []*LeaderboardConfig
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-id").Limit(pageSize, offset).All(&leaderboards)

	return leaderboards, total, err
}

// GetTableName 获取动态表名
func (l *LeaderboardData) GetTableName(appId string) string {
	cleanAppId := utils.CleanAppId(appId)
	return fmt.Sprintf("leaderboard_%s", cleanAppId)
}

// TableName 获取配置表名
func (l *LeaderboardConfig) TableName() string {
	return "leaderboard_config"
}

func init() {
	orm.RegisterModel(new(LeaderboardConfig))
	orm.RegisterModel(new(Leaderboard))
	orm.RegisterModel(new(LeaderboardScore))
}

// TableName 获取表名
func (l *Leaderboard) TableName() string {
	return "leaderboard_config"
}

// TableName 获取表名
func (ls *LeaderboardScore) TableName() string {
	return "leaderboard_score"
}

// CreateLeaderboardConfig 创建排行榜配置
func CreateLeaderboardConfig(config *LeaderboardConfig) error {
	o := orm.NewOrm()

	// 检查是否已存在
	exist := o.QueryTable("leaderboard_config").
		Filter("app_id", config.AppId).
		Filter("type", config.Type).
		Exist()

	if exist {
		return fmt.Errorf("排行榜已存在")
	}

	_, err := o.Insert(config)
	if err != nil {
		return err
	}

	// 创建动态排行榜表（如果不存在）
	return createLeaderboardTable(config.AppId)
}

// CreateLeaderboard 创建排行榜（兼容性保持）
func CreateLeaderboard(leaderboard *Leaderboard) error {
	o := orm.NewOrm()

	// 检查是否已存在
	exist := o.QueryTable("leaderboard_config").
		Filter("app_id", leaderboard.AppId).
		Filter("type", leaderboard.Type).
		Exist()

	if exist {
		return fmt.Errorf("排行榜已存在")
	}

	_, err := o.Insert(leaderboard)
	if err != nil {
		return err
	}

	// 创建动态排行榜表（如果不存在）
	return createLeaderboardTable(leaderboard.AppId)
}

// UpdateLeaderboard 更新排行榜配置
func UpdateLeaderboard(appId, leaderboardType string, fields map[string]interface{}) error {
	o := orm.NewOrm()

	// 添加更新时间
	fields["update_time"] = time.Now()

	qs := o.QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("type", leaderboardType)

	_, err := qs.Update(fields)
	return err
}

// DeleteLeaderboard 删除排行榜
func DeleteLeaderboard(appId, leaderboardType string) error {
	o := orm.NewOrm()

	// 删除排行榜配置
	_, err := o.QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("type", leaderboardType).
		Delete()

	if err != nil {
		return err
	}

	// 删除动态表中的排行榜数据
	leaderboardData := &LeaderboardData{}
	tableName := leaderboardData.GetTableName(appId)

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE leaderboard_name = ?", tableName)
	_, err = o.Raw(deleteSQL, leaderboardType).Exec()

	return err
}

// GetLeaderboardData 获取排行榜数据
func GetLeaderboardData(appId, leaderboardType string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()

	// 使用动态表
	leaderboardData := &LeaderboardData{}
	tableName := leaderboardData.GetTableName(appId)

	// 获取总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE leaderboard_name = ?", tableName)
	var total int64
	err := o.Raw(countSQL, leaderboardType).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	querySQL := fmt.Sprintf(`
		SELECT id, leaderboard_name, user_id, score, extra_data, create_time, update_time 
		FROM %s 
		WHERE leaderboard_name = ? 
		ORDER BY score DESC, create_time ASC 
		LIMIT ? OFFSET ?
	`, tableName)

	var results []orm.Params
	_, err = o.Raw(querySQL, leaderboardType, pageSize, offset).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 转换为map格式并添加排名
	var result []map[string]interface{}
	for i, row := range results {
		item := map[string]interface{}{
			"rank":       offset + i + 1,
			"playerId":   row["user_id"],
			"score":      row["score"],
			"extraData":  row["extra_data"],
			"createTime": row["create_time"],
			"updateTime": row["update_time"],
		}
		result = append(result, item)
	}

	return result, total, nil
}

// UpdateLeaderboardScore 更新排行榜分数
func UpdateLeaderboardScore(appId, leaderboardType, playerId string, score int64) error {
	o := orm.NewOrm()

	// 使用动态表
	leaderboardData := &LeaderboardData{}
	tableName := leaderboardData.GetTableName(appId)

	// 检查记录是否存在
	var existingId int64
	checkSQL := fmt.Sprintf("SELECT id FROM %s WHERE leaderboard_name = ? AND user_id = ?", tableName)
	err := o.Raw(checkSQL, leaderboardType, playerId).QueryRow(&existingId)

	if err == orm.ErrNoRows {
		// 插入新记录
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (leaderboard_name, user_id, score, create_time, update_time) 
			VALUES (?, ?, ?, NOW(), NOW())
		`, tableName)
		_, err = o.Raw(insertSQL, leaderboardType, playerId, score).Exec()
	} else if err == nil {
		// 更新现有记录
		updateSQL := fmt.Sprintf(`
			UPDATE %s SET score = ?, update_time = NOW() 
			WHERE leaderboard_name = ? AND user_id = ?
		`, tableName)
		_, err = o.Raw(updateSQL, score, leaderboardType, playerId).Exec()
	}

	return err
}

// DeleteLeaderboardScore 删除排行榜分数
func DeleteLeaderboardScore(appId, leaderboardType, playerId string) error {
	o := orm.NewOrm()

	// 使用动态表
	leaderboardData := &LeaderboardData{}
	tableName := leaderboardData.GetTableName(appId)

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE leaderboard_name = ? AND user_id = ?", tableName)
	_, err := o.Raw(deleteSQL, leaderboardType, playerId).Exec()

	return err
}

// CommitLeaderboardScore 提交排行榜分数
func CommitLeaderboardScore(appId, leaderboardType, playerId string, score int64) error {
	o := orm.NewOrm()

	// 检查排行榜是否存在且启用
	var leaderboard Leaderboard
	err := o.QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("type", leaderboardType).
		Filter("status", 1).
		One(&leaderboard)

	if err != nil {
		if err == orm.ErrNoRows {
			return fmt.Errorf("排行榜不存在或已禁用")
		}
		return err
	}

	// 使用动态表
	leaderboardData := &LeaderboardData{}
	tableName := leaderboardData.GetTableName(appId)

	// 查找现有记录
	var existingScore int64
	checkSQL := fmt.Sprintf("SELECT score FROM %s WHERE leaderboard_name = ? AND user_id = ?", tableName)
	err = o.Raw(checkSQL, leaderboardType, playerId).QueryRow(&existingScore)

	if err == orm.ErrNoRows {
		// 检查是否超过最大条目数
		if leaderboard.MaxEntries > 0 {
			countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE leaderboard_name = ?", tableName)
			var count int64
			o.Raw(countSQL, leaderboardType).QueryRow(&count)

			if count >= int64(leaderboard.MaxEntries) {
				// 删除最低分记录
				lowestSQL := fmt.Sprintf("SELECT score FROM %s WHERE leaderboard_name = ? ORDER BY score ASC, create_time DESC LIMIT 1", tableName)
				var lowestScore int64
				err = o.Raw(lowestSQL, leaderboardType).QueryRow(&lowestScore)

				if err == nil && score > lowestScore {
					deleteLowestSQL := fmt.Sprintf("DELETE FROM %s WHERE leaderboard_name = ? ORDER BY score ASC, create_time DESC LIMIT 1", tableName)
					o.Raw(deleteLowestSQL, leaderboardType).Exec()
				} else if err == nil {
					return fmt.Errorf("分数太低，无法进入排行榜")
				}
			}
		}

		// 创建新记录
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (leaderboard_name, user_id, score, create_time, update_time) 
			VALUES (?, ?, ?, NOW(), NOW())
		`, tableName)
		_, err = o.Raw(insertSQL, leaderboardType, playerId, score).Exec()
		return err
	} else if err != nil {
		return err
	}

	// 检查分数类型决定是否更新
	shouldUpdate := false
	if leaderboard.ScoreType == "higher_better" && score > existingScore {
		shouldUpdate = true
	} else if leaderboard.ScoreType == "lower_better" && score < existingScore {
		shouldUpdate = true
	}

	if shouldUpdate {
		updateSQL := fmt.Sprintf(`
			UPDATE %s SET score = ?, update_time = NOW() 
			WHERE leaderboard_name = ? AND user_id = ?
		`, tableName)
		_, err = o.Raw(updateSQL, score, leaderboardType, playerId).Exec()
	}

	return err
}

// QueryLeaderboardScore 查询排行榜分数
func QueryLeaderboardScore(appId, leaderboardType, playerId string) (int64, int, error) {
	o := orm.NewOrm()

	// 使用动态表
	leaderboardData := &LeaderboardData{}
	tableName := leaderboardData.GetTableName(appId)

	// 获取用户分数
	var userScore int64
	scoreSQL := fmt.Sprintf("SELECT score FROM %s WHERE leaderboard_name = ? AND user_id = ?", tableName)
	err := o.Raw(scoreSQL, leaderboardType, playerId).QueryRow(&userScore)

	if err != nil {
		if err == orm.ErrNoRows {
			return 0, 0, nil // 用户没有分数记录
		}
		return 0, 0, err
	}

	// 计算排名
	rankSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE leaderboard_name = ? AND score > ?", tableName)
	var rank int64
	err = o.Raw(rankSQL, leaderboardType, userScore).QueryRow(&rank)

	if err != nil {
		return userScore, 0, err
	}

	return userScore, int(rank) + 1, nil
}

// FixLeaderboardUserInfo 修复排行榜用户信息（暂时保留兼容性）
func FixLeaderboardUserInfo(appId, leaderboardType string) (int64, error) {
	logs.Info("排行榜已迁移到动态表，用户信息修复功能已不需要")
	return 0, nil
}

// createLeaderboardTable 创建排行榜数据表
func createLeaderboardTable(appId string) error {
	o := orm.NewOrm()

	leaderboardData := &LeaderboardData{}
	tableName := leaderboardData.GetTableName(appId)

	// 检查表是否存在
	checkSQL := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	var exists string
	err := o.Raw(checkSQL).QueryRow(&exists)

	if err == orm.ErrNoRows {
		// 表不存在，创建表
		createSQL := fmt.Sprintf(`
			CREATE TABLE %s (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				leaderboard_name VARCHAR(100) NOT NULL,
				user_id VARCHAR(100) NOT NULL,
				score BIGINT DEFAULT 0,
				extra_data TEXT,
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
				update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				UNIQUE KEY uk_leaderboard_user (leaderboard_name, user_id),
				KEY idx_leaderboard_score (leaderboard_name, score DESC),
				KEY idx_update_time (update_time)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`, tableName)

		_, err = o.Raw(createSQL).Exec()
		return err
	}

	return nil
}
