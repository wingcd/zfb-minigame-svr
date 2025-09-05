package models

import (
	"fmt"

	"github.com/beego/beego/v2/client/orm"
)

// GameConfig 游戏配置模型
type GameConfig struct {
	BaseModel
	AppId       string `orm:"size(32)" json:"app_id" valid:"Required"`
	ConfigKey   string `orm:"size(100)" json:"config_key" valid:"Required"`
	ConfigValue string `orm:"type(text)" json:"config_value"`
	ConfigType  string `orm:"size(20)" json:"config_type"` // string, number, boolean, json
	Description string `orm:"size(255)" json:"description"`
	IsPublic    int    `orm:"default(1)" json:"is_public"` // 1:公开 0:私有
	Status      int    `orm:"default(1)" json:"status"`    // 1:启用 0:禁用
}

// TableName 指定表名
func (c *GameConfig) TableName() string {
	return "game_configs"
}

// GetAllGameConfigs 获取所有游戏配置
func GetAllGameConfigs(page, pageSize int, appId, configKey string) ([]*GameConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("game_configs")

	if appId != "" {
		qs = qs.Filter("app_id", appId)
	}

	if configKey != "" {
		qs = qs.Filter("config_key__icontains", configKey)
	}

	total, _ := qs.Count()

	var configs []*GameConfig
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&configs)

	return configs, total, err
}

// GetGameConfigById 根据ID获取游戏配置
func GetGameConfigById(id int64) (*GameConfig, error) {
	o := orm.NewOrm()
	config := &GameConfig{BaseModel: BaseModel{Id: id}}
	err := o.QueryTable("game_configs").Filter("id", id).One(config)
	return config, err
}

// GetGameConfigByKey 根据AppId和Key获取游戏配置
func GetGameConfigByKey(appId, configKey string) (*GameConfig, error) {
	o := orm.NewOrm()
	config := &GameConfig{}
	err := o.QueryTable("game_configs").Filter("app_id", appId).Filter("config_key", configKey).One(config)
	return config, err
}

// GetGameConfigsByAppId 根据AppId获取所有配置
func GetGameConfigsByAppId(appId string) ([]*GameConfig, error) {
	o := orm.NewOrm()
	var configs []*GameConfig
	_, err := o.QueryTable("game_configs").Filter("app_id", appId).Filter("status", 1).All(&configs)
	return configs, err
}

// GetPublicGameConfigs 获取公开的游戏配置
func GetPublicGameConfigs(appId string) ([]*GameConfig, error) {
	o := orm.NewOrm()
	var configs []*GameConfig
	_, err := o.QueryTable("game_configs").Filter("app_id", appId).Filter("is_public", 1).Filter("status", 1).All(&configs)
	return configs, err
}

// AddGameConfig 添加游戏配置
func AddGameConfig(config *GameConfig) error {
	o := orm.NewOrm()
	_, err := o.Insert(config)
	return err
}

// UpdateGameConfig 更新游戏配置
func UpdateGameConfig(config *GameConfig) error {
	o := orm.NewOrm()
	_, err := o.Update(config)
	return err
}

// DeleteGameConfig 删除游戏配置
func DeleteGameConfig(id int64) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("game_configs").Filter("id", id).Delete()
	return err
}

// BatchUpdateGameConfigs 批量更新游戏配置
func BatchUpdateGameConfigs(appId string, configs map[string]string) error {
	o := orm.NewOrm()

	// 开启事务
	tx, err := o.Begin()
	if err != nil {
		return err
	}

	for key, value := range configs {
		// 查找现有配置
		config := &GameConfig{}
		err := o.QueryTable("game_configs").Filter("app_id", appId).Filter("config_key", key).One(config)

		if err == orm.ErrNoRows {
			// 不存在则创建
			newConfig := &GameConfig{
				AppId:       appId,
				ConfigKey:   key,
				ConfigValue: value,
				ConfigType:  "string",
				Status:      1,
				IsPublic:    1,
			}
			_, err = tx.Insert(newConfig)
		} else if err == nil {
			// 存在则更新
			config.ConfigValue = value
			_, err = tx.Update(config, "config_value", "updated_at")
		}

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// DeleteGameConfigsByAppId 删除应用的所有配置
func DeleteGameConfigsByAppId(appId string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("game_configs").Filter("app_id", appId).Delete()
	return err
}

// GetGameConfigList 获取游戏配置列表 (控制器调用的函数)
func GetGameConfigList(appId string, page, pageSize int, configType, version string) ([]*GameConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("game_configs")

	if appId != "" {
		qs = qs.Filter("app_id", appId)
	}

	if configType != "" {
		qs = qs.Filter("config_type", configType)
	}

	if version != "" {
		// 如果有版本过滤需求，可以在这里添加版本相关的逻辑
		// 目前的GameConfig模型中没有version字段，所以先忽略
	}

	total, _ := qs.Count()

	var configs []*GameConfig
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&configs)

	return configs, total, err
}

// CreateGameConfig 创建游戏配置
func CreateGameConfig(config *GameConfig) error {
	// 检查配置是否已存在
	existing := &GameConfig{}
	o := orm.NewOrm()
	err := o.QueryTable("game_configs").Filter("app_id", config.AppId).Filter("config_key", config.ConfigKey).One(existing)

	if err == nil {
		return fmt.Errorf("配置已存在")
	} else if err != orm.ErrNoRows {
		return err
	}

	// 设置默认值
	if config.ConfigType == "" {
		config.ConfigType = "string"
	}
	if config.Status == 0 {
		config.Status = 1
	}
	if config.IsPublic == 0 {
		config.IsPublic = 1
	}

	_, err = o.Insert(config)
	return err
}

// UpdateGameConfigByKey 根据AppId和Key更新游戏配置
func UpdateGameConfigByKey(appId, configKey string, updates map[string]interface{}) error {
	o := orm.NewOrm()

	// 先查找配置
	config := &GameConfig{}
	err := o.QueryTable("game_configs").Filter("app_id", appId).Filter("config_key", configKey).One(config)
	if err != nil {
		return err
	}

	// 更新字段
	if value, exists := updates["configValue"]; exists {
		if strValue, ok := value.(string); ok {
			config.ConfigValue = strValue
		}
	}
	if description, exists := updates["description"]; exists {
		if strDescription, ok := description.(string); ok {
			config.Description = strDescription
		}
	}
	if configType, exists := updates["configType"]; exists {
		if strType, ok := configType.(string); ok {
			config.ConfigType = strType
		}
	}

	_, err = o.Update(config)
	return err
}

// DeleteGameConfigByKey 根据AppId和Key删除游戏配置
func DeleteGameConfigByKey(appId, configKey string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("game_configs").Filter("app_id", appId).Filter("config_key", configKey).Delete()
	return err
}

// GetGameConfig 根据AppId和Key获取游戏配置（支持版本）
func GetGameConfig(appId, configKey, version string) (*GameConfig, error) {
	o := orm.NewOrm()
	config := &GameConfig{}
	qs := o.QueryTable("game_configs").Filter("app_id", appId).Filter("config_key", configKey)

	// 如果有版本要求，可以在这里添加版本过滤逻辑
	// 目前模型中没有version字段，所以忽略version参数

	err := qs.One(config)
	return config, err
}

// GetConfigCount 获取配置数量统计
func GetConfigCount(appId string) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("game_configs").Filter("app_id", appId).Count()
	return count, err
}

func init() {
	orm.RegisterModel(new(GameConfig))
}
