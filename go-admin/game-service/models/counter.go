package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// Counter 计数器模型 - 对应数据库设计的counter_[appid]表（简化结构，对齐JS功能）
type Counter struct {
	Id         int64     `orm:"auto" json:"id"`
	CounterKey string    `orm:"size(100);column(counter_key)" json:"counterKey"`
	Location   string    `orm:"size(100);default(default);column(location)" json:"location"`
	Value      int64     `orm:"default(0);column(value)" json:"value"`
	CreatedAt  time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt  time.Time `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// CounterConfig 计数器配置结构（从admin-service获取）
type CounterConfig struct {
	ID            int64     `json:"id"`
	AppId         string    `json:"appId"`
	CounterKey    string    `json:"counter_key"`
	ResetType     string    `json:"resetType"`
	ResetValue    int       `json:"resetValue"`
	NextResetTime time.Time `json:"nextResetTime"`
	Description   string    `json:"description"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// GetTableName 获取动态表名
func (c *Counter) GetTableName(appId string) string {
	return utils.GetCounterTableName(appId)
}

func GetCounterModel(appId string) (*Counter, string, error) {
	tableName := utils.GetCounterTableName(appId)
	counterModel := &Counter{}
	if err := utils.EnsureCounterTableRegistered(tableName, appId, counterModel); err != nil {
		return nil, "", err
	}
	return counterModel, tableName, nil
}

func GetCounterValues(appId, counterKey string) (map[string]int64, error) {
	// 1. 检查计数器配置是否存在
	_, err := getCounterConfig(appId, counterKey)
	if err != nil {
		return nil, err
	}

	tableName := utils.GetCounterTableName(appId)

	o := orm.NewOrm()
	var result []orm.Params
	selectSQL := fmt.Sprintf(`SELECT location, value FROM %s WHERE counter_key = ?`, tableName)
	_, err = o.Raw(selectSQL, counterKey).Values(&result)
	if err != nil {
		return nil, err
	}

	values := make(map[string]int64)
	for _, row := range result {
		location := row["location"].(string)
		// 数据库返回的是string，需要转换为int64
		valueStr := row["value"].(string)
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse counter value: %v", err)
		}
		values[location] = value
	}
	return values, nil
}

// GetCounterValue 获取计数器指定位置的值（对齐JS getCounter功能）
func GetCounterValue(appId, counterKey, location string) (int64, error) {
	if location == "" {
		location = "default"
	}

	// 1. 检查计数器配置是否存在
	config, err := getCounterConfig(appId, counterKey)
	if err != nil {
		return 0, err
	}

	// 2. 检查并处理重置逻辑
	err = checkAndResetCounter(appId, counterKey, location, config)
	if err != nil {
		return 0, err
	}

	// 3. 从数据库获取值
	counter, tableName, err := GetCounterModel(appId)
	if err != nil {
		return 0, err
	}

	o := orm.NewOrm()
	err = o.QueryTable(tableName).
		Filter("counter_key", counterKey).
		Filter("location", location).
		One(counter)

	if err == orm.ErrNoRows {
		return 0, nil // 计数器不存在，返回0
	} else if err != nil {
		return 0, err
	}

	return counter.Value, nil
}

// IncrementCounterValue 增加计数器值（对齐JS incrementCounter功能）
func IncrementCounterValue(appId, counterKey, location string, increment int64) (int64, error) {
	if location == "" {
		location = "default"
	}
	if increment <= 0 {
		increment = 1
	}

	// 1. 检查计数器配置是否存在
	config, err := getCounterConfig(appId, counterKey)
	if err != nil {
		return 0, err
	}

	// 2. 检查点位是否存在（通过config中的locations检查）
	err = checkLocationExists(appId, counterKey, location)
	if err != nil {
		return 0, err
	}

	// 3. 检查并处理重置逻辑
	err = checkAndResetCounter(appId, counterKey, location, config)
	if err != nil {
		return 0, err
	}

	// 4. 使用 UPSERT 更新计数器值
	_, tableName, err := GetCounterModel(appId)
	if err != nil {
		return 0, err
	}

	o := orm.NewOrm()

	// 使用 ON DUPLICATE KEY UPDATE 进行 upsert 操作
	sql := fmt.Sprintf(`
		INSERT INTO %s (counter_key, location, value, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			value = value + VALUES(value),
			updated_at = NOW()
	`, tableName)

	_, err = o.Raw(sql, counterKey, location, increment).Exec()
	if err != nil {
		return 0, err
	}

	// 5. 获取更新后的值
	var result []orm.Params
	selectSQL := fmt.Sprintf(`SELECT value FROM %s WHERE location = ?`, tableName)
	_, err = o.Raw(selectSQL, location).Values(&result)
	if err != nil {
		return 0, err
	}

	if len(result) > 0 {
		if valueStr, ok := result[0]["value"].(string); ok {
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse counter value: %v", err)
			}
			return value, nil
		}
	}

	return 0, fmt.Errorf("failed to get updated value")
}

// DecrementCounterValue 减少计数器值（保持现有接口兼容性）
func DecrementCounterValue(appId, counterKey, location string, decrement int64) (int64, error) {
	if location == "" {
		location = "default"
	}
	if decrement <= 0 {
		decrement = 1
	}

	// 1. 检查计数器配置是否存在
	config, err := getCounterConfig(appId, counterKey)
	if err != nil {
		return 0, err
	}

	// 2. 检查并处理重置逻辑
	err = checkAndResetCounter(appId, counterKey, location, config)
	if err != nil {
		return 0, err
	}

	// 3. 减少计数器值（不允许小于0）
	_, tableName, err := GetCounterModel(appId)
	if err != nil {
		return 0, err
	}

	o := orm.NewOrm()

	// 使用 UPDATE 语句减少值，确保不小于0
	sql := fmt.Sprintf(`
		UPDATE %s 
		SET value = GREATEST(0, value - ?), updated_at = NOW()
		WHERE counter_key = ? AND location = ?
	`, tableName)

	_, err = o.Raw(sql, decrement, counterKey, location).Exec()
	if err != nil {
		return 0, err
	}

	// 获取更新后的值
	var result []orm.Params
	selectSQL := fmt.Sprintf(`SELECT value FROM %s WHERE counter_key = ? AND location = ?`, tableName)
	_, err = o.Raw(selectSQL, counterKey, location).Values(&result)
	if err != nil {
		return 0, err
	}

	if len(result) > 0 {
		if value, ok := result[0]["value"].(int64); ok {
			return value, nil
		}
	}

	return 0, nil
}

// SetCounterValue 设置计数器值（对齐JS setCounter功能）
func SetCounterValue(appId, counterKey, location string, value int64) (int64, error) {
	if location == "" {
		location = "default"
	}

	// 1. 检查计数器配置是否存在
	config, err := getCounterConfig(appId, counterKey)
	if err != nil {
		return 0, err
	}

	// 2. 检查点位是否存在
	err = checkLocationExists(appId, counterKey, location)
	if err != nil {
		return 0, err
	}

	// 3. 检查并处理重置逻辑
	err = checkAndResetCounter(appId, counterKey, location, config)
	if err != nil {
		return 0, err
	}

	// 4. 使用 UPSERT 设置计数器值
	_, tableName, err := GetCounterModel(appId)
	if err != nil {
		return 0, err
	}

	o := orm.NewOrm()

	// 使用 ON DUPLICATE KEY UPDATE 进行 upsert 操作
	sql := fmt.Sprintf(`
		INSERT INTO %s (counter_key, location, value, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			value = VALUES(value),
			updated_at = NOW()
	`, tableName)

	_, err = o.Raw(sql, counterKey, location, value).Exec()
	if err != nil {
		return 0, err
	}

	return value, nil
}

// ResetCounterValue 重置计数器值（对齐JS resetCounter功能）
func ResetCounterValue(appId, counterKey, location string) (int64, error) {
	if location == "" {
		location = "default"
	}

	// 1. 检查计数器配置是否存在
	_, err := getCounterConfig(appId, counterKey)
	if err != nil {
		return 0, err
	}

	// 2. 检查点位是否存在
	err = checkLocationExists(appId, counterKey, location)
	if err != nil {
		return 0, err
	}

	// 3. 重置计数器值为0
	return SetCounterValue(appId, counterKey, location, 0)
}

// SetCounter 设置计数器值
func SetCounter(appId, counterKey, playerId string, value int64) error {
	counter, tableName, err := GetCounterModel(appId)
	if err != nil {
		return err
	}

	o := orm.NewOrm()

	err = o.QueryTable(tableName).
		Filter("counter_key", counterKey).
		Filter("location", playerId).
		One(counter)

	if err == orm.ErrNoRows {
		// 创建新计数器（使用新的字段名）
		counter.CounterKey = counterKey
		counter.Location = playerId
		counter.Value = value
		_, err = o.Insert(counter)
		return err
	} else if err != nil {
		return err
	}

	// 更新计数器值
	counter.Value = value
	_, err = o.Update(counter, "value", "updated_at")
	return err
}

// ResetCounter 重置计数器
func ResetCounter(appId, counterKey, playerId string) error {
	counter, tableName, err := GetCounterModel(appId)
	if err != nil {
		return err
	}

	o := orm.NewOrm()

	err = o.QueryTable(tableName).
		Filter("counter_key", counterKey).
		Filter("location", playerId).
		One(counter)

	if err == orm.ErrNoRows {
		return nil // 计数器不存在
	} else if err != nil {
		return err
	}

	// 重置计数器
	counter.Value = 0
	_, err = o.Update(counter, "value", "updated_at")
	return err
}

// GetCounterList 获取计数器列表（管理后台使用）
func GetCounterList(appId string, page, pageSize int, counterKey string) ([]Counter, int64, error) {
	_, tableName, err := GetCounterModel(appId)
	if err != nil {
		return nil, 0, err
	}

	o := orm.NewOrm()

	qs := o.QueryTable(tableName)
	if counterKey != "" {
		qs = qs.Filter("counter_key", counterKey)
	}

	total, _ := qs.Count()

	var results []Counter
	offset := (page - 1) * pageSize
	_, err = qs.OrderBy("-count", "counter_key").Limit(pageSize, offset).All(&results)

	return results, total, err
}

// 全局计数器函数（不需要用户ID）

// GetGlobalCounter 获取全局计数器值（已废弃，使用GetCounterValue替代）
func GetGlobalCounter(appId, counterKey string) (int64, error) {
	return GetCounterValue(appId, counterKey, "default")
}

// IncrementGlobalCounter 增加全局计数器（已废弃，使用IncrementCounterValue替代）
func IncrementGlobalCounter(appId, counterKey string, increment int64) (int64, error) {
	return IncrementCounterValue(appId, counterKey, "default", increment)
}

// DecrementGlobalCounter 减少全局计数器（已废弃，使用DecrementCounterValue替代）
func DecrementGlobalCounter(appId, counterKey string, decrement int64) (int64, error) {
	return DecrementCounterValue(appId, counterKey, "default", decrement)
}

// SetGlobalCounter 设置全局计数器值（已废弃，使用SetCounterValue替代）
func SetGlobalCounter(appId, counterKey string, value int64) error {
	_, err := SetCounterValue(appId, counterKey, "default", value)
	return err
}

// ResetGlobalCounter 重置全局计数器（已废弃，使用ResetCounterValue替代）
func ResetGlobalCounter(appId, counterKey string) error {
	_, err := ResetCounterValue(appId, counterKey, "default")
	return err
}

// GetAllGlobalCounters 获取所有全局计数器
func GetAllGlobalCounters(appId string) ([]Counter, error) {
	_, tableName, err := GetCounterModel(appId)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()

	var results []Counter
	_, err = o.QueryTable(tableName).Filter("location", "default").OrderBy("counter_key").All(&results)

	return results, err
}

// getCounterConfig 从admin-service获取计数器配置
func getCounterConfig(appId, counterKey string) (*CounterConfig, error) {
	// 这里应该调用admin-service的API获取配置
	// 暂时返回一个默认配置，实际使用时需要实现HTTP调用
	adminServiceURL := utils.GetAdminServiceURL()
	if adminServiceURL == "" {
		return &CounterConfig{
			AppId:      appId,
			CounterKey: counterKey,
			ResetType:  "permanent",
			IsActive:   true,
		}, nil
	}

	// 实现HTTP调用获取配置的逻辑
	url := fmt.Sprintf("%s/counter/getCounterConfig", adminServiceURL)

	requestBody := map[string]interface{}{
		"appId": appId,
		"key":   counterKey,
	}

	_, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		logs.Error("Failed to get counter config from admin-service: %v", err)
		// 返回默认配置
		return &CounterConfig{
			AppId:      appId,
			CounterKey: counterKey,
			ResetType:  "permanent",
			IsActive:   true,
		}, nil
	}
	defer resp.Body.Close()

	var result struct {
		Code int            `json:"code"`
		Data *CounterConfig `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil || result.Code != 0 {
		// 返回默认配置
		return &CounterConfig{
			AppId:      appId,
			CounterKey: counterKey,
			ResetType:  "permanent",
			IsActive:   true,
		}, nil
	}

	return result.Data, nil
}

// checkLocationExists 检查点位是否存在
func checkLocationExists(appId, counterKey, location string) error {
	// 这里可以通过admin-service检查点位是否在配置中存在
	// 暂时返回nil，表示所有点位都允许
	return nil
}

// checkAndResetCounter 检查并重置计数器
func checkAndResetCounter(appId, counterKey, location string, config *CounterConfig) error {
	if config.ResetType == "permanent" || config.NextResetTime.IsZero() {
		return nil
	}

	now := time.Now()
	if now.After(config.NextResetTime) {
		logs.Info("计数器需要重置: appId=%s, key=%s, location=%s, resetTime=%v",
			appId, counterKey, location, config.NextResetTime)

		// 重置计数器值
		err := resetCounterValue(appId, counterKey, location)
		if err != nil {
			return fmt.Errorf("重置计数器失败: %v", err)
		}

		// 计算下次重置时间并更新配置（通过admin-service）
		nextResetTime := calculateCounterNextResetTime(config.ResetType, config.ResetValue)
		if nextResetTime != nil {
			err = updateCounterResetTime(appId, counterKey, *nextResetTime)
			if err != nil {
				logs.Error("更新计数器重置时间失败: %v", err)
			} else {
				logs.Info("计数器重置完成，下次重置时间: %v", *nextResetTime)
			}
		}
	}

	return nil
}

// resetCounterValue 重置计数器值
func resetCounterValue(appId, counterKey, location string) error {
	_, tableName, err := GetCounterModel(appId)
	if err != nil {
		return err
	}

	o := orm.NewOrm()
	sql := fmt.Sprintf(`UPDATE %s SET value = 0, updated_at = NOW() WHERE counter_key = ? AND location = ?`, tableName)
	_, err = o.Raw(sql, counterKey, location).Exec()
	return err
}

// calculateCounterNextResetTime 计算计数器下次重置时间
func calculateCounterNextResetTime(resetType string, resetValue int) *time.Time {
	now := time.Now()
	var nextReset time.Time

	switch resetType {
	case "daily":
		nextReset = now.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	case "weekly":
		// 下周一0点
		weekday := now.Weekday()
		daysToMonday := (7 - int(weekday) + 1) % 7
		if daysToMonday == 0 {
			daysToMonday = 7
		}
		nextReset = now.AddDate(0, 0, daysToMonday).Truncate(24 * time.Hour)
	case "monthly":
		// 下月1号0点
		year, month, _ := now.Date()
		nextReset = time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location())
	case "custom":
		if resetValue > 0 {
			nextReset = now.Add(time.Duration(resetValue) * time.Hour)
		} else {
			return nil
		}
	default:
		return nil
	}

	return &nextReset
}

// updateCounterResetTime 更新计数器重置时间（通过admin-service）
func updateCounterResetTime(appId, counterKey string, resetTime time.Time) error {
	// 这里应该调用admin-service的API更新重置时间
	adminServiceURL := utils.GetAdminServiceURL()
	if adminServiceURL == "" {
		return nil // 如果没有admin-service，忽略更新
	}

	// 实现HTTP调用更新重置时间的逻辑
	url := fmt.Sprintf("%s/counter/updateCounterResetTime", adminServiceURL)

	requestBody := map[string]interface{}{
		"appId":     appId,
		"key":       counterKey,
		"resetTime": resetTime.Format("2006-01-02 15:04:05"),
	}

	_, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
