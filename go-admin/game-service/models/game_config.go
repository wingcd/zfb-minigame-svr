package models

import (
	"fmt"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
)

// GameConfig 游戏配置模型
type GameConfig struct {
	Id          int64  `orm:"auto" json:"id"`
	ConfigKey   string `orm:"size(100);unique" json:"config_key"`
	ConfigValue string `orm:"type(longtext)" json:"config_value"`
	Version     string `orm:"size(50)" json:"version"`
	Description string `orm:"size(255)" json:"description"`
	CreatedAt   string `orm:"auto_now_add;type(datetime)" json:"create_time"`
	UpdatedAt   string `orm:"auto_now;type(datetime)" json:"update_time"`
}

// GetTableName 获取动态表名
func (g *GameConfig) GetTableName(appId string) string {
	return fmt.Sprintf("game_config_%s", utils.CleanAppId(appId))
}

// GetConfig 获取配置
func GetConfig(appId, configKey string) (string, error) {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	err := o.QueryTable(tableName).Filter("config_key", configKey).One(config)
	if err == orm.ErrNoRows {
		return "", nil // 配置不存在，返回空字符串
	} else if err != nil {
		return "", err
	}

	return config.ConfigValue, nil
}

// SetConfig 设置配置
func SetConfig(appId, configKey, configValue, version, description string) error {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	// 检查配置是否存在
	err := o.QueryTable(tableName).Filter("config_key", configKey).One(config)
	if err == orm.ErrNoRows {
		// 创建新配置
		config.ConfigKey = configKey
		config.ConfigValue = configValue
		config.Version = version
		config.Description = description
		_, err = o.Insert(config)
	} else if err == nil {
		// 更新配置
		config.ConfigValue = configValue
		if version != "" {
			config.Version = version
		}
		if description != "" {
			config.Description = description
		}
		_, err = o.Update(config, "config_value", "version", "description", "update_time")
	}

	return err
}

// GetConfigsByVersion 获取指定版本的所有配置
func GetConfigsByVersion(appId, version string) (map[string]string, error) {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	var configs []GameConfig
	qs := o.QueryTable(tableName)
	if version != "" {
		qs = qs.Filter("version", version)
	}

	_, err := qs.All(&configs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.ConfigKey] = cfg.ConfigValue
	}

	return result, nil
}

// GetAllConfigs 获取所有配置
func GetAllConfigs(appId string) (map[string]string, error) {
	return GetConfigsByVersion(appId, "")
}

// DeleteConfig 删除配置
func DeleteConfig(appId, configKey string) error {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	_, err := o.QueryTable(tableName).Filter("config_key", configKey).Delete()
	return err
}

// GetConfigList 获取配置列表（管理后台使用）
func GetConfigList(appId string, page, pageSize int, keyword string) ([]GameConfig, int64, error) {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	qs := o.QueryTable(tableName)
	if keyword != "" {
		qs = qs.Filter("config_key__icontains", keyword).
			Filter("description__icontains", keyword)
	}

	total, _ := qs.Count()

	var configs []GameConfig
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("config_key").Limit(pageSize, offset).All(&configs)

	return configs, total, err
}

// GetConfigDetails 获取配置详情（管理后台使用）
func GetConfigDetails(appId, configKey string) (*GameConfig, error) {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	err := o.QueryTable(tableName).Filter("config_key", configKey).One(config)
	return config, err
}

// UpdateConfigDescription 更新配置描述
func UpdateConfigDescription(appId, configKey, description string) error {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	err := o.QueryTable(tableName).Filter("config_key", configKey).One(config)
	if err != nil {
		return err
	}

	config.Description = description
	_, err = o.Update(config, "description", "update_time")
	return err
}

// GetVersionList 获取版本列表
func GetVersionList(appId string) ([]string, error) {
	o := orm.NewOrm()

	config := &GameConfig{}
	tableName := config.GetTableName(appId)

	var versions []string
	_, err := o.Raw(fmt.Sprintf("SELECT DISTINCT version FROM %s WHERE version IS NOT NULL AND version != '' ORDER BY version", tableName)).QueryRows(&versions)

	return versions, err
}
