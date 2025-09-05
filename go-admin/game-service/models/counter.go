package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Counter 计数器模型
type Counter struct {
	Id            int64     `orm:"auto" json:"id"`
	CounterName   string    `orm:"size(100)" json:"counter_name"`
	UserId        string    `orm:"size(100);null" json:"user_id"`
	Count         int64     `orm:"default(0)" json:"count"`
	ResetTime     time.Time `orm:"null" json:"reset_time"`
	ResetInterval int       `orm:"null" json:"reset_interval"`
	CreatedAt     string    `orm:"auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt     string    `orm:"auto_now;type(datetime)" json:"updated_at"`
}

// GetTableName 获取动态表名
func (c *Counter) GetTableName(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return fmt.Sprintf("counter_%s", cleanAppId)
}

// GetCounter 获取计数器值
func GetCounter(appId, counterName, userId string) (int64, error) {
	o := orm.NewOrm()

	counter := &Counter{}
	tableName := counter.GetTableName(appId)

	err := o.QueryTable(tableName).
		Filter("counter_name", counterName).
		Filter("user_id", userId).
		One(counter)

	if err == orm.ErrNoRows {
		return 0, nil // 计数器不存在，返回0
	} else if err != nil {
		return 0, err
	}

	// 检查是否需要重置
	if !counter.ResetTime.IsZero() && time.Now().After(counter.ResetTime) {
		err = resetCounterIfNeeded(counter, tableName)
		if err != nil {
			return 0, err
		}
	}

	return counter.Count, nil
}

// IncrementCounter 增加计数器
func IncrementCounter(appId, counterName, userId string, increment int64) (int64, error) {
	o := orm.NewOrm()

	counter := &Counter{}
	tableName := counter.GetTableName(appId)

	err := o.QueryTable(tableName).
		Filter("counter_name", counterName).
		Filter("user_id", userId).
		One(counter)

	if err == orm.ErrNoRows {
		// 创建新计数器
		counter.CounterName = counterName
		counter.UserId = userId
		counter.Count = increment
		_, err = o.Insert(counter)
		return increment, err
	} else if err != nil {
		return 0, err
	}

	// 检查是否需要重置
	if !counter.ResetTime.IsZero() && time.Now().After(counter.ResetTime) {
		err = resetCounterIfNeeded(counter, tableName)
		if err != nil {
			return 0, err
		}
	}

	// 增加计数
	counter.Count += increment
	_, err = o.Update(counter, "count", "updated_at")
	return counter.Count, err
}

// DecrementCounter 减少计数器
func DecrementCounter(appId, counterName, userId string, decrement int64) (int64, error) {
	o := orm.NewOrm()

	counter := &Counter{}
	tableName := counter.GetTableName(appId)

	err := o.QueryTable(tableName).
		Filter("counter_name", counterName).
		Filter("user_id", userId).
		One(counter)

	if err == orm.ErrNoRows {
		return 0, nil // 计数器不存在，返回0
	} else if err != nil {
		return 0, err
	}

	// 检查是否需要重置
	if !counter.ResetTime.IsZero() && time.Now().After(counter.ResetTime) {
		err = resetCounterIfNeeded(counter, tableName)
		if err != nil {
			return 0, err
		}
	}

	// 减少计数（不允许小于0）
	counter.Count -= decrement
	if counter.Count < 0 {
		counter.Count = 0
	}
	_, err = o.Update(counter, "count", "updated_at")
	return counter.Count, err
}

// SetCounter 设置计数器值
func SetCounter(appId, counterName, userId string, value int64) error {
	o := orm.NewOrm()

	counter := &Counter{}
	tableName := counter.GetTableName(appId)

	err := o.QueryTable(tableName).
		Filter("counter_name", counterName).
		Filter("user_id", userId).
		One(counter)

	if err == orm.ErrNoRows {
		// 创建新计数器
		counter.CounterName = counterName
		counter.UserId = userId
		counter.Count = value
		_, err = o.Insert(counter)
		return err
	} else if err != nil {
		return err
	}

	// 更新计数器值
	counter.Count = value
	_, err = o.Update(counter, "count", "updated_at")
	return err
}

// ResetCounter 重置计数器
func ResetCounter(appId, counterName, userId string) error {
	o := orm.NewOrm()

	counter := &Counter{}
	tableName := counter.GetTableName(appId)

	err := o.QueryTable(tableName).
		Filter("counter_name", counterName).
		Filter("user_id", userId).
		One(counter)

	if err == orm.ErrNoRows {
		return nil // 计数器不存在
	} else if err != nil {
		return err
	}

	// 重置计数器
	counter.Count = 0
	_, err = o.Update(counter, "count", "updated_at")
	return err
}

// resetCounterIfNeeded 检查并重置计数器
func resetCounterIfNeeded(counter *Counter, tableName string) error {
	if counter.ResetInterval <= 0 {
		return nil
	}

	o := orm.NewOrm()

	// 计算下一次重置时间
	nextResetTime := counter.ResetTime.Add(time.Duration(counter.ResetInterval) * time.Second)
	for time.Now().After(nextResetTime) {
		nextResetTime = nextResetTime.Add(time.Duration(counter.ResetInterval) * time.Second)
	}

	// 重置计数器
	counter.Count = 0
	counter.ResetTime = nextResetTime
	_, err := o.Update(counter, "count", "reset_time", "updated_at")
	return err
}

// GetCounterList 获取计数器列表（管理后台使用）
func GetCounterList(appId string, page, pageSize int, counterName string) ([]Counter, int64, error) {
	o := orm.NewOrm()

	counter := &Counter{}
	tableName := counter.GetTableName(appId)

	qs := o.QueryTable(tableName)
	if counterName != "" {
		qs = qs.Filter("counter_name", counterName)
	}

	total, _ := qs.Count()

	var results []Counter
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-count", "counter_name").Limit(pageSize, offset).All(&results)

	return results, total, err
}

// 全局计数器函数（不需要用户ID）

// GetGlobalCounter 获取全局计数器值
func GetGlobalCounter(appId, counterName string) (int64, error) {
	return GetCounter(appId, counterName, "")
}

// IncrementGlobalCounter 增加全局计数器
func IncrementGlobalCounter(appId, counterName string, increment int64) (int64, error) {
	return IncrementCounter(appId, counterName, "", increment)
}

// DecrementGlobalCounter 减少全局计数器
func DecrementGlobalCounter(appId, counterName string, decrement int64) (int64, error) {
	return DecrementCounter(appId, counterName, "", decrement)
}

// SetGlobalCounter 设置全局计数器值
func SetGlobalCounter(appId, counterName string, value int64) error {
	return SetCounter(appId, counterName, "", value)
}

// ResetGlobalCounter 重置全局计数器
func ResetGlobalCounter(appId, counterName string) error {
	return ResetCounter(appId, counterName, "")
}

// GetAllGlobalCounters 获取所有全局计数器
func GetAllGlobalCounters(appId string) ([]Counter, error) {
	o := orm.NewOrm()

	counter := &Counter{}
	tableName := counter.GetTableName(appId)

	var results []Counter
	_, err := o.QueryTable(tableName).Filter("user_id", "").OrderBy("counter_name").All(&results)

	return results, err
}
