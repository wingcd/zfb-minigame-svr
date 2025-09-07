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
	AppId         string    `orm:"size(100)" json:"appId"`
	CounterKey    string    `orm:"size(100)" json:"counterKey"`
	ResetType     string    `orm:"size(20);default(permanent)" json:"resetType"` // daily, weekly, monthly, custom, permanent
	ResetValue    int       `orm:"null" json:"resetValue"`                       // 自定义重置时间(小时)
	NextResetTime time.Time `orm:"type(datetime);null" json:"nextResetTime"`
	Description   string    `orm:"type(text);null" json:"description"`
	IsActive      bool      `orm:"default(true)" json:"isActive"`
}

// TableName 指定表名
func (c *CounterConfig) TableName() string {
	return "counter_config"
}

// CounterData 计数器数据模型（动态表）
type CounterData struct {
	Id         int64     `orm:"auto" json:"id"`
	CounterKey string    `orm:"size(100)" json:"counterKey"`
	Location   string    `orm:"size(100);default(default)" json:"location"`
	Value      int64     `orm:"default(0)" json:"value"`
	ResetTime  time.Time `orm:"type(datetime);null" json:"resetTime"`
	CreatedAt  string    `orm:"auto_now_add;type(datetime)" json:"createdAt"`
	UpdatedAt  string    `orm:"auto_now;type(datetime)" json:"updatedAt"`
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
		Filter("counterKey", config.CounterKey).
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
		Filter("appId", appId).
		Filter("counterKey", key).
		Filter("isActive", true).
		One(config)
	return config, err
}

// GetCounterConfigList 获取计数器配置列表
func GetCounterConfigList(appId string, page, pageSize int) ([]*CounterConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("counter_config").Filter("appId", appId).Filter("isActive", true)

	total, _ := qs.Count()

	var configs []*CounterConfig
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-id").Limit(pageSize, offset).All(&configs)

	return configs, total, err
}

// GetCounterConfigListWithFilter 获取计数器配置列表（支持筛选）
func GetCounterConfigListWithFilter(appId string, page, pageSize int, key, resetType string) ([]*CounterConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("counter_config").Filter("appId", appId).Filter("isActive", true)

	// 添加key筛选（模糊搜索）
	if key != "" {
		qs = qs.Filter("counterKey__icontains", key)
	}

	// 添加resetType筛选
	if resetType != "" {
		qs = qs.Filter("reset_type", resetType)
	}

	total, _ := qs.Count()

	var configs []*CounterConfig
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-updatedAt").Limit(pageSize, offset).All(&configs)

	return configs, total, err
}

// UpdateCounterConfig 更新计数器配置
func UpdateCounterConfig(appId, key string, fields map[string]interface{}) error {
	o := orm.NewOrm()

	// 添加更新时间
	fields["updatedAt"] = time.Now()

	_, err := o.QueryTable("counter_config").
		Filter("appId", appId).
		Filter("counterKey", key).
		Update(fields)

	return err
}

// DeleteCounterConfig 删除计数器配置
func DeleteCounterConfig(appId, key string) error {
	o := orm.NewOrm()

	// 软删除配置
	_, err := o.QueryTable("counter_config").
		Filter("appId", appId).
		Filter("counterKey", key).
		Update(orm.Params{
			"isActive":  false,
			"updatedAt": time.Now(),
		})

	if err != nil {
		return err
	}

	// 删除计数器数据表中对应的数据
	counterData := &CounterData{}
	tableName := counterData.GetTableName(appId)

	_, err = o.Raw(fmt.Sprintf("DELETE FROM %s WHERE counterKey = ?", tableName), key).Exec()
	return err
}

// UpdateCounterValue 更新计数器值
func UpdateCounterValue(appId, key, location string, value int64) error {
	o := orm.NewOrm()

	counterData := &CounterData{}
	tableName := counterData.GetTableName(appId)

	// 检查记录是否存在
	var existingId int64
	checkSQL := fmt.Sprintf("SELECT id FROM %s WHERE counterKey = ? AND location = ?", tableName)
	err := o.Raw(checkSQL, key, location).QueryRow(&existingId)

	if err == orm.ErrNoRows {
		// 插入新记录
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (counterKey, location, value, createdAt, updatedAt) 
			VALUES (?, ?, ?, NOW(), NOW())
		`, tableName)
		_, err = o.Raw(insertSQL, key, location, value).Exec()
	} else if err == nil {
		// 更新现有记录
		updateSQL := fmt.Sprintf(`
			UPDATE %s SET value = ?, updatedAt = NOW() 
			WHERE counterKey = ? AND location = ?
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
	querySQL := fmt.Sprintf("SELECT value FROM %s WHERE counterKey = ? AND location = ?", tableName)
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
	querySQL := fmt.Sprintf("SELECT location, value FROM %s WHERE counterKey = ?", tableName)
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
				counterKey VARCHAR(100) NOT NULL,
				location VARCHAR(100) DEFAULT 'default',
				value BIGINT DEFAULT 0,
				resetTime DATETIME NULL,
				createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
				updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				UNIQUE KEY uk_key_location (counterKey, location),
				INDEX idx_counterKey (counterKey)
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
