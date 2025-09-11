package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// CounterConfig 计数器配置模型
type CounterConfig struct {
	BaseModel
	AppId         string    `orm:"size(100);column(app_id)" json:"appId"`
	CounterKey    string    `orm:"size(100);column(counter_key)" json:"counter_key"`
	ResetType     string    `orm:"size(20);default(permanent);column(reset_type)" json:"resetType"` // daily, weekly, monthly, custom, permanent
	ResetValue    int       `orm:"null;column(reset_value)" json:"resetValue"`                      // 自定义重置时间(小时)
	NextResetTime time.Time `orm:"type(datetime);null;column(next_reset_time)" json:"nextResetTime"`
	Description   string    `orm:"type(text);null;column(description)" json:"description"`
	IsActive      bool      `orm:"default(true);column(is_active)" json:"isActive"`
}

// TableName 指定表名
func (c *CounterConfig) TableName() string {
	return "counter_config"
}

// CounterData 计数器数据模型（动态表）
type CounterData struct {
	Id         int64  `orm:"auto" json:"id"`
	CounterKey string `orm:"size(100);column(counter_key)" json:"counter_key"`
	Location   string `orm:"size(100);default(default);column(location)" json:"location"`
	Value      int64  `orm:"default(0)" json:"value"`
	created_at string `orm:"auto_now_add;type(datetime);column(created_at)" json:"created_at"`
	updated_at string `orm:"auto_now;type(datetime);column(updated_at)" json:"updated_at"`
}

// GetTableName 获取动态表名
func (c *CounterData) GetTableName(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return fmt.Sprintf("counter_%s", cleanAppId)
}

// CreateCounterConfig 创建计数器配置
func CreateCounterConfig(config *CounterConfig) error {
	o := orm.NewOrm()

	// 检查是否已存在
	exist := &CounterConfig{}
	err := o.QueryTable("counter_config").
		Filter("appId", config.AppId).
		Filter("counter_key", config.CounterKey).
		One(exist)

	if err == nil {
		return fmt.Errorf("计数器[%s]已存在", config.CounterKey)
	}

	// 计算下次重置时间
	if config.ResetType != "permanent" {
		config.NextResetTime = calculateNextResetTime(config.ResetType, config.ResetValue)
	}

	_, err = o.Insert(config)
	if err != nil {
		return err
	}

	// 创建动态计数器表
	return createCounterTable(config.AppId)
}

// GetCounterConfig 获取单个计数器配置
func GetCounterConfig(appId, key string) (*CounterConfig, error) {
	o := orm.NewOrm()
	config := &CounterConfig{}
	err := o.QueryTable("counter_config").
		Filter("app_id", appId).
		Filter("counter_key", key).
		Filter("is_active", true).
		One(config)
	return config, err
}

// GetCounterConfigList 获取计数器配置列表（包含软删除的记录）
func GetCounterConfigList(appId string, page, pageSize int) ([]*CounterConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("counter_config").Filter("app_id", appId)

	total, _ := qs.Count()

	var configs []*CounterConfig
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-id").Limit(pageSize, offset).All(&configs)

	return configs, total, err
}

// GetCounterConfigListWithFilter 获取计数器配置列表（支持筛选，包含软删除的记录）
func GetCounterConfigListWithFilter(appId string, page, pageSize int, key, resetType string) ([]*CounterConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("counter_config").Filter("app_id", appId)

	// 添加key筛选（模糊搜索）
	if key != "" {
		qs = qs.Filter("counter_key__icontains", key)
	}

	// 添加resetType筛选
	if resetType != "" {
		qs = qs.Filter("reset_type", resetType)
	}

	total, _ := qs.Count()

	var configs []*CounterConfig
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-updated_at").Limit(pageSize, offset).All(&configs)

	return configs, total, err
}

// UpdateCounterConfig 更新计数器配置
func UpdateCounterConfig(appId, key string, fields map[string]interface{}) error {
	o := orm.NewOrm()

	// 添加更新时间
	fields["updated_at"] = time.Now()

	_, err := o.QueryTable("counter_config").
		Filter("app_id", appId).
		Filter("counter_key", key).
		Update(fields)

	return err
}

// RestoreCounterConfig 恢复软删除的计数器配置
func RestoreCounterConfig(appId, key string) error {
	o := orm.NewOrm()

	fmt.Printf("开始恢复计数器配置: AppId=%s, Key=%s\n", appId, key)

	// 检查计数器是否存在且为软删除状态
	config := &CounterConfig{}
	err := o.QueryTable("counter_config").
		Filter("app_id", appId).
		Filter("counter_key", key).
		Filter("is_active", false).
		One(config)

	if err != nil {
		fmt.Printf("查询软删除配置失败: %v\n", err)
		return fmt.Errorf("未找到可恢复的计数器配置")
	}

	// 恢复配置
	configResult, err := o.QueryTable("counter_config").
		Filter("app_id", appId).
		Filter("counter_key", key).
		Update(orm.Params{
			"is_active":  true,
			"updated_at": time.Now(),
		})

	if err != nil {
		fmt.Printf("恢复配置失败: %v\n", err)
		return err
	}

	fmt.Printf("恢复配置成功，影响行数: %d\n", configResult)
	return nil
}

// DeleteCounterConfig 删除计数器配置（第一次软删除，第二次硬删除）
func DeleteCounterConfig(appId, key string) error {
	o := orm.NewOrm()

	fmt.Printf("开始删除计数器配置: AppId=%s, Key=%s\n", appId, key)

	// 先查询当前配置状态
	config := &CounterConfig{}
	err := o.QueryTable("counter_config").
		Filter("app_id", appId).
		Filter("counter_key", key).
		One(config)

	if err != nil {
		fmt.Printf("查询配置失败: %v\n", err)
		return fmt.Errorf("计数器配置不存在")
	}

	if config.IsActive {
		// 第一次删除：软删除
		fmt.Printf("执行软删除: AppId=%s, Key=%s\n", appId, key)
		configResult, err := o.QueryTable("counter_config").
			Filter("app_id", appId).
			Filter("counter_key", key).
			Update(orm.Params{
				"is_active":  false,
				"updated_at": time.Now(),
			})

		if err != nil {
			fmt.Printf("软删除配置失败: %v\n", err)
			return err
		}

		fmt.Printf("软删除配置成功，影响行数: %d\n", configResult)
		return nil
	} else {
		// 第二次删除：硬删除（删除配置和所有相关数据）
		fmt.Printf("执行硬删除: AppId=%s, Key=%s\n", appId, key)

		// 删除配置记录
		configResult, err := o.QueryTable("counter_config").
			Filter("app_id", appId).
			Filter("counter_key", key).
			Delete()

		if err != nil {
			fmt.Printf("硬删除配置失败: %v\n", err)
			return err
		}

		fmt.Printf("硬删除配置成功，影响行数: %d\n", configResult)

		// 删除计数器数据表中对应的数据
		counterData := &CounterData{}
		tableName := counterData.GetTableName(appId)

		sql := fmt.Sprintf("DELETE FROM %s WHERE counter_key = ?", tableName)
		fmt.Printf("执行删除数据SQL: %s, 参数: key=%s\n", sql, key)

		dataResult, err := o.Raw(sql, key).Exec()
		if err != nil {
			fmt.Printf("删除数据失败: %v\n", err)
			return err
		}

		rowsAffected, _ := dataResult.RowsAffected()
		fmt.Printf("删除数据成功，影响行数: %d\n", rowsAffected)

		return nil
	}
}

// DeleteCounterLocation 删除计数器特定点位
func DeleteCounterLocation(appId, key, location string) error {
	o := orm.NewOrm()

	// 删除计数器数据表中对应的特定点位数据
	counterData := &CounterData{}
	tableName := counterData.GetTableName(appId)

	sql := fmt.Sprintf("DELETE FROM %s WHERE counter_key = ? AND location = ?", tableName)
	fmt.Printf("执行删除点位SQL: %s, 参数: key=%s, location=%s\n", sql, key, location)

	result, err := o.Raw(sql, key, location).Exec()
	if err != nil {
		fmt.Printf("删除点位失败: %v\n", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("删除点位成功，影响行数: %d\n", rowsAffected)

	return nil
}

// UpdateCounterValue 更新计数器值
func UpdateCounterValue(appId, key, location string, value int64) error {
	o := orm.NewOrm()

	counterData := &CounterData{}
	tableName := counterData.GetTableName(appId)

	// 检查记录是否存在
	var existingId int64
	checkSQL := fmt.Sprintf("SELECT id FROM %s WHERE counter_key = ? AND location = ?", tableName)
	err := o.Raw(checkSQL, key, location).QueryRow(&existingId)

	if err == orm.ErrNoRows {
		// 插入新记录
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (counter_key, location, value, created_at, updated_at) 
			VALUES (?, ?, ?, NOW(), NOW())
		`, tableName)
		_, err = o.Raw(insertSQL, key, location, value).Exec()
	} else if err == nil {
		// 更新现有记录
		updateSQL := fmt.Sprintf(`
			UPDATE %s SET value = ?, updated_at = NOW() 
			WHERE counter_key = ? AND location = ?
		`, tableName)
		_, err = o.Raw(updateSQL, value, key, location).Exec()
	}

	return err
}

// GetCounterValue 获取计数器值
func GetCounterValue(appId, key, location string) (int64, error) {
	o := orm.NewOrm()

	counterData := &CounterData{}
	tableName := counterData.GetTableName(appId)

	var value int64
	querySQL := fmt.Sprintf("SELECT value FROM %s WHERE counter_key = ? AND location = ?", tableName)
	err := o.Raw(querySQL, key, location).QueryRow(&value)

	if err == orm.ErrNoRows {
		return 0, nil // 不存在则返回0
	}

	return value, err
}

// GetCounterAllLocations 获取计数器所有点位数据
func GetCounterAllLocations(appId, key string) (map[string]interface{}, error) {
	o := orm.NewOrm()

	counterData := &CounterData{}
	tableName := counterData.GetTableName(appId)

	var results []orm.Params
	querySQL := fmt.Sprintf("SELECT location, value FROM %s WHERE counter_key = ?", tableName)
	_, err := o.Raw(querySQL, key).Values(&results)

	if err != nil {
		return nil, err
	}

	locations := make(map[string]interface{})
	for _, result := range results {
		location := result["location"].(string)
		value := result["value"]
		locations[location] = map[string]interface{}{
			"value": value,
		}
	}

	return locations, nil
}

// createCounterTable 创建计数器数据表
func createCounterTable(appId string) error {
	o := orm.NewOrm()

	counterData := &CounterData{}
	tableName := counterData.GetTableName(appId)

	// 检查表是否存在
	checkSQL := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	var exists string
	err := o.Raw(checkSQL).QueryRow(&exists)

	if err == orm.ErrNoRows {
		// 表不存在，创建表
		createSQL := fmt.Sprintf(`
			CREATE TABLE %s (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				counter_key VARCHAR(100) NOT NULL,
				location VARCHAR(100) DEFAULT 'default',
				value BIGINT DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				UNIQUE KEY uk_key_location (counter_key, location),
				INDEX idx_counter_key (counter_key)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`, tableName)

		_, err = o.Raw(createSQL).Exec()
		return err
	}

	return nil
}

// calculateNextResetTime 计算下次重置时间
func calculateNextResetTime(resetType string, resetValue int) time.Time {
	now := time.Now()

	switch resetType {
	case "daily":
		return now.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	case "weekly":
		// 下周一0点
		weekday := now.Weekday()
		daysToMonday := (7 - int(weekday) + 1) % 7
		if daysToMonday == 0 {
			daysToMonday = 7
		}
		return now.AddDate(0, 0, daysToMonday).Truncate(24 * time.Hour)
	case "monthly":
		// 下月1号0点
		year, month, _ := now.Date()
		return time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location())
	case "custom":
		if resetValue > 0 {
			return now.Add(time.Duration(resetValue) * time.Hour)
		}
		return time.Time{} // 无效的自定义时间
	default:
		return time.Time{} // permanent类型不设置重置时间
	}
}

// GetCounterCount 获取计数器数量统计
func GetCounterCount(appId string) (int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("counter_%s", appId)
	var count int64
	err := o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).QueryRow(&count)
	return count, err
}

func init() {
	orm.RegisterModel(new(CounterConfig))
}
