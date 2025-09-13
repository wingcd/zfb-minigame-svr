package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// YallaConfig Yalla配置模型
type YallaConfig struct {
	ID          int64     `orm:"column(id);auto;pk" json:"id"`
	AppID       string    `orm:"column(app_id);size(100);unique" json:"appId"`
	AppGameID   string    `orm:"column(app_game_id);size(100)" json:"appGameId"`
	SecretKey   string    `orm:"column(secret_key);size(255)" json:"secretKey"`
	BaseURL     string    `orm:"column(base_url);size(255)" json:"baseUrl"`
	PushURL     string    `orm:"column(push_url);size(255)" json:"pushUrl"`                        // 推送域名
	Environment string    `orm:"column(environment);size(50);default(sandbox)" json:"environment"` // sandbox, production
	Timeout     int       `orm:"column(timeout);default(30)" json:"timeout"`                       // 超时时间（秒）
	RetryCount  int       `orm:"column(retry_count);default(3)" json:"retryCount"`                 // 重试次数
	Description string    `orm:"column(description);size(500);null" json:"description"`            // 配置描述
	IsActive    bool      `orm:"column(is_active);default(true)" json:"isActive"`                  // 是否启用
	CreateTime  time.Time `orm:"column(created_at);auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime  time.Time `orm:"column(updated_at);auto_now;type(datetime)" json:"updateTime"`
}

func init() {
	orm.RegisterModel(new(YallaConfig))
}

// TableName 设置表名
func (u *YallaConfig) TableName() string {
	return "yalla_config"
}

// GetYallaConfigList 获取Yalla配置列表
func GetYallaConfigList(conditions map[string]interface{}, page, pageSize int) ([]*YallaConfig, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(YallaConfig))

	// 应用查询条件
	for key, value := range conditions {
		qs = qs.Filter(key, value)
	}

	// 分页和排序
	offset := (page - 1) * pageSize
	var configs []*YallaConfig
	_, err := qs.OrderBy("-created_at").Limit(pageSize, offset).All(&configs)

	if err != nil {
		logs.Error("获取Yalla配置列表失败: %v", err)
		return nil, err
	}

	return configs, nil
}

// GetYallaConfigCount 获取Yalla配置总数
func GetYallaConfigCount(conditions map[string]interface{}) (int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(YallaConfig))

	// 应用查询条件
	for key, value := range conditions {
		qs = qs.Filter(key, value)
	}

	count, err := qs.Count()
	if err != nil {
		logs.Error("获取Yalla配置总数失败: %v", err)
		return 0, err
	}

	return count, nil
}

// CreateYallaConfig 创建Yalla配置
func CreateYallaConfig(config *YallaConfig) error {
	o := orm.NewOrm()
	_, err := o.Insert(config)
	if err != nil {
		logs.Error("创建Yalla配置失败: %v", err)
		return err
	}
	return nil
}

// GetYallaConfigByAppID 根据AppID获取Yalla配置
func GetYallaConfigByAppID(appID string) (*YallaConfig, error) {
	o := orm.NewOrm()
	config := &YallaConfig{}

	err := o.QueryTable(new(YallaConfig)).Filter("app_id", appID).One(config)
	if err != nil {
		if err == orm.ErrNoRows {
			logs.Info("Yalla配置不存在: AppID=%s", appID)
		} else {
			logs.Error("获取Yalla配置失败: AppID=%s, Error=%v", appID, err)
		}
		return nil, err
	}

	return config, nil
}

// UpdateYallaConfig 更新Yalla配置
func UpdateYallaConfig(config *YallaConfig) error {
	o := orm.NewOrm()
	_, err := o.Update(config)
	if err != nil {
		logs.Error("更新Yalla配置失败: AppID=%s, Error=%v", config.AppID, err)
		return err
	}
	return nil
}

// DeleteYallaConfig 删除Yalla配置
func DeleteYallaConfig(appID string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable(new(YallaConfig)).Filter("app_id", appID).Delete()
	if err != nil {
		logs.Error("删除Yalla配置失败: AppID=%s, Error=%v", appID, err)
		return err
	}
	return nil
}

// CheckYallaConfigExists 检查Yalla配置是否存在
func CheckYallaConfigExists(appID string) (bool, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable(new(YallaConfig)).Filter("app_id", appID).Count()
	if err != nil {
		logs.Error("检查Yalla配置是否存在失败: AppID=%s, Error=%v", appID, err)
		return false, err
	}
	return count > 0, nil
}

// GetActiveYallaConfig 获取启用的Yalla配置
func GetActiveYallaConfig(appID string) (*YallaConfig, error) {
	o := orm.NewOrm()
	config := &YallaConfig{}

	err := o.QueryTable(new(YallaConfig)).Filter("app_id", appID).Filter("is_active", true).One(config)
	if err != nil {
		if err == orm.ErrNoRows {
			logs.Info("启用的Yalla配置不存在: AppID=%s", appID)
		} else {
			logs.Error("获取启用的Yalla配置失败: AppID=%s, Error=%v", appID, err)
		}
		return nil, err
	}

	return config, nil
}

// GetAllActiveYallaConfigs 获取所有启用的Yalla配置
func GetAllActiveYallaConfigs() ([]*YallaConfig, error) {
	o := orm.NewOrm()
	var configs []*YallaConfig

	_, err := o.QueryTable(new(YallaConfig)).Filter("is_active", true).OrderBy("app_id").All(&configs)
	if err != nil {
		logs.Error("获取所有启用的Yalla配置失败: %v", err)
		return nil, err
	}

	return configs, nil
}
