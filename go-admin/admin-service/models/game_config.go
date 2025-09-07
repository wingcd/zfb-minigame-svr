package models

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// GameConfig 游戏配置模型
type GameConfig struct {
	BaseModel
	AppID       string `orm:"size(32)" json:"appId" valid:"Required"`
	ConfigKey   string `orm:"size(100)" json:"configKey" valid:"Required"`
	ConfigValue string `orm:"type(text)" json:"configValue"`
	Version     string `orm:"size(50)" json:"version"`
	Description string `orm:"size(255)" json:"description"`
	ConfigType  string `orm:"size(50);default(string)" json:"configType"`
	IsActive    bool   `orm:"default(true)" json:"isActive"`
	Priority    int    `orm:"default(1)" json:"priority"`
	Tags        string `orm:"type(text)" json:"tags"` // JSON array stored as string
	CreatedBy   string `orm:"size(50)" json:"createdBy"`
}

// TableName 指定表名
func GameConfigTableName(appId string) string {
	// 清理 appId 确保表名安全
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return fmt.Sprintf("game_config_%s", cleanAppId)
}

// TableName 为 GameConfig 实现 TableName 接口
func (d *GameConfig) TableName() string {
	// 这里需要在运行时设置，通过全局变量或者其他方式
	return "game_config_default"
}

// GetAllGameConfigs 获取所有游戏配置
func GetAllGameConfigs(page, pageSize int, appId, configKey string) ([]*GameConfig, int64, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	// 构建查询条件
	var whereConditions []string
	var params []interface{}

	if configKey != "" {
		whereConditions = append(whereConditions, "config_key LIKE ?")
		params = append(params, "%"+configKey+"%")
	}

	var whereClause string
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 获取总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, whereClause)
	var total int64
	err := o.Raw(countSQL, params...).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s %s ORDER BY id DESC LIMIT ? OFFSET ?", tableName, whereClause)
	params = append(params, pageSize, offset)

	var configs []*GameConfig
	_, err = o.Raw(querySQL, params...).QueryRows(&configs)
	if err != nil {
		return nil, 0, err
	}

	// 设置 AppID 字段
	for _, config := range configs {
		config.AppID = appId
	}

	return configs, total, nil
}

// GetGameConfigById 根据ID获取游戏配置
func GetGameConfigById(id int64, appId string) (*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	config := &GameConfig{}
	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE id = ?", tableName)
	err := o.Raw(querySQL, id).QueryRow(config)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	config.AppID = appId
	return config, nil
}

// GetGameConfigByKey 根据AppId和Key获取游戏配置
func GetGameConfigByKey(appId, configKey string) (*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	config := &GameConfig{}
	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE config_key = ?", tableName)
	err := o.Raw(querySQL, configKey).QueryRow(config)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	config.AppID = appId
	return config, nil
}

// GetGameConfigsByAppId 根据AppId获取所有配置
func GetGameConfigsByAppId(appId string) ([]*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	var configs []*GameConfig
	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s ORDER BY id DESC", tableName)
	_, err := o.Raw(querySQL).QueryRows(&configs)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	for _, config := range configs {
		config.AppID = appId
	}

	return configs, nil
}

// GetPublicGameConfigs 获取公开的游戏配置
func GetPublicGameConfigs(appId string) ([]*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	var configs []*GameConfig
	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE is_active = true ORDER BY priority DESC, id DESC", tableName)
	_, err := o.Raw(querySQL).QueryRows(&configs)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	for _, config := range configs {
		config.AppID = appId
	}

	return configs, nil
}

// AddGameConfig 添加游戏配置
func AddGameConfig(config *GameConfig) error {

	o := orm.NewOrm()
	tableName := GameConfigTableName(config.AppID)

	// 设置默认值
	if config.ConfigType == "" {
		config.ConfigType = "string"
	}
	if config.Priority == 0 {
		config.Priority = 1
	}

	// 使用 Raw SQL 插入到指定表
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?)
	`, tableName)

	_, err := o.Raw(insertSQL,
		config.ConfigKey,
		config.ConfigValue,
		config.Version,
		config.Description,
		config.ConfigType,
		config.IsActive,
		config.Priority,
		config.CreatedBy,
	).Exec()

	return err
}

// UpdateGameConfig 更新游戏配置
func UpdateGameConfig(config *GameConfig) error {
	o := orm.NewOrm()
	tableName := GameConfigTableName(config.AppID)

	updateSQL := fmt.Sprintf(`
		UPDATE %s SET 
		config_value = ?, version = ?, description = ?, 
		config_type = ?, is_active = ?, priority = ?,
		updated_at = NOW() 
		WHERE id = ?
	`, tableName)

	_, err := o.Raw(updateSQL,
		config.ConfigValue,
		config.Version,
		config.Description,
		config.ConfigType,
		config.IsActive,
		config.Priority,
		config.ID,
	).Exec()

	return err
}

// DeleteGameConfig 删除游戏配置
func DeleteGameConfig(id int64, appId string) error {

	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	// 使用 Raw SQL 删除指定表中的记录
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	_, err := o.Raw(deleteSQL, id).Exec()

	return err
}

// BatchUpdateGameConfigs 批量更新游戏配置
func BatchUpdateGameConfigs(appId string, configs map[string]string) error {

	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	// 开启事务
	tx, err := o.Begin()
	if err != nil {
		return err
	}

	for key, value := range configs {
		// 检查配置是否存在
		checkSQL := fmt.Sprintf("SELECT id FROM %s WHERE config_key = ?", tableName)
		var id int64
		err := tx.Raw(checkSQL, key).QueryRow(&id)

		if err == orm.ErrNoRows {
			// 不存在则创建
			insertSQL := fmt.Sprintf(`
				INSERT INTO %s (config_key, config_value, config_type, is_active, priority, created_at, updated_at)
				VALUES (?, ?, 'string', true, 1, NOW(), NOW())
			`, tableName)
			_, err = tx.Raw(insertSQL, key, value).Exec()
		} else if err == nil {
			// 存在则更新
			updateSQL := fmt.Sprintf("UPDATE %s SET config_value = ?, updated_at = NOW() WHERE config_key = ?", tableName)
			_, err = tx.Raw(updateSQL, value, key).Exec()
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
	tableName := GameConfigTableName(appId)

	// 使用 Raw SQL 删除所有记录
	deleteSQL := fmt.Sprintf("DELETE FROM %s", tableName)
	_, err := o.Raw(deleteSQL).Exec()

	return err
}

// GetGameConfigList 获取游戏配置列表 (控制器调用的函数)
func GetGameConfigList(appId string, page, pageSize int, configType, version string) ([]*GameConfig, int64, error) {

	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	// 构建查询条件
	var whereConditions []string
	var params []interface{}

	if configType != "" {
		whereConditions = append(whereConditions, "config_type = ?")
		params = append(params, configType)
	}

	var whereClause string
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 获取总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, whereClause)
	var total int64
	err := o.Raw(countSQL, params...).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s %s ORDER BY id DESC LIMIT ? OFFSET ?", tableName, whereClause)
	params = append(params, pageSize, offset)

	var configs []*GameConfig
	_, err = o.Raw(querySQL, params...).QueryRows(&configs)
	if err != nil {
		return nil, 0, err
	}

	// 设置 AppID 字段
	for _, config := range configs {
		config.AppID = appId
	}

	return configs, total, nil
}

// CreateGameConfig 创建游戏配置
func CreateGameConfig(config *GameConfig) error {

	o := orm.NewOrm()
	tableName := GameConfigTableName(config.AppID)

	// 检查配置是否已存在
	checkSQL := fmt.Sprintf("SELECT id FROM %s WHERE config_key = ?", tableName)
	var existingId int64
	err := o.Raw(checkSQL, config.ConfigKey).QueryRow(&existingId)
	if err == nil {
		return fmt.Errorf("配置已存在")
	} else if err != orm.ErrNoRows {
		return err
	}

	// 设置默认值
	if config.ConfigType == "" {
		config.ConfigType = "string"
	}
	if config.Priority == 0 {
		config.Priority = 1
	}

	// 使用 Raw SQL 插入
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (config_key, config_value, config_type, description, is_active, priority, version, created_at, updated_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?)
	`, tableName)

	_, err = o.Raw(insertSQL,
		config.ConfigKey,
		config.ConfigValue,
		config.ConfigType,
		config.Description,
		config.IsActive,
		config.Priority,
		config.Version,
		config.CreatedBy,
	).Exec()

	return err
}

type UpdateGameConfigRequest struct {
	AppId       string      `json:"appId"`
	ID          int64       `json:"id"`
	ConfigKey   string      `json:"configKey"`
	IsActive    bool        `json:"isActive"`
	ConfigValue interface{} `json:"configValue"`
	ConfigType  string      `json:"configType"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	Priority    int         `json:"priority"`
}

// UpdateGameConfigByKey 根据AppId和Key更新游戏配置
func UpdateGameConfigByRequest(requestData *UpdateGameConfigRequest) error {

	o := orm.NewOrm()
	tableName := GameConfigTableName(requestData.AppId)

	logs.Info("UpdateGameConfig 开始更新: AppId=%s, ConfigKey=%s, Updates=%+v", requestData.AppId, requestData.ConfigKey, requestData)

	// 构建更新语句
	var setParts []string
	var params []interface{}

	if requestData.ConfigValue != nil {
		setParts = append(setParts, "config_value = ?")
		params = append(params, requestData.ConfigValue)
	}
	if requestData.Version != "" {
		setParts = append(setParts, "version = ?")
		params = append(params, requestData.Version)
	}
	if requestData.Description != "" {
		setParts = append(setParts, "description = ?")
		params = append(params, requestData.Description)
	}
	if requestData.ConfigType != "" {
		setParts = append(setParts, "config_type = ?")
		params = append(params, requestData.ConfigType)
	}
	setParts = append(setParts, "is_active = ?")
	params = append(params, requestData.IsActive)

	setParts = append(setParts, "priority = ?")
	params = append(params, requestData.Priority)

	if len(setParts) == 0 {
		logs.Error("没有要更新的字段，原始更新数据: %+v", requestData)
		return fmt.Errorf("没有要更新的字段")
	}

	// 添加更新时间
	setParts = append(setParts, "updated_at = NOW()")

	updateSQL := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, strings.Join(setParts, ", "))
	params = append(params, requestData.ID)
	logs.Info("执行 SQL: %s, 参数: %+v", updateSQL, params)

	result, err := o.Raw(updateSQL, params...).Exec()
	if err != nil {
		logs.Error("SQL 执行失败: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logs.Info("SQL 执行成功，影响行数: %d", rowsAffected)

	if rowsAffected == 0 {
		logs.Warning("没有找到匹配的配置记录: AppId=%s, ID=%d", requestData.AppId, requestData.ID)
		return fmt.Errorf("没有找到匹配的配置记录")
	}

	return err
}

// DeleteGameConfigByKey 根据AppId和Key删除游戏配置
func DeleteGameConfigByKey(appId, configKey string) error {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE config_key = ?", tableName)
	_, err := o.Raw(deleteSQL, configKey).Exec()
	return err
}

// GetGameConfig 根据AppId和Key获取游戏配置（支持版本）
func GetGameConfig(appId, configKey, version string) (*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	var querySQL string
	var params []interface{}

	if version != "" {
		// 优先查找指定版本的配置
		querySQL = fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE config_key = ? AND version = ? AND is_active = true ORDER BY priority DESC LIMIT 1", tableName)
		params = []interface{}{configKey, version}
	} else {
		// 查找全局配置（无版本限制）
		querySQL = fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE config_key = ? AND is_active = true ORDER BY priority DESC LIMIT 1", tableName)
		params = []interface{}{configKey}
	}

	config := &GameConfig{}
	err := o.Raw(querySQL, params...).QueryRow(config)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	config.AppID = appId
	return config, nil
}

// GetConfigCount 获取配置数量统计
func GetConfigCount(appId string) (int64, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE is_active = true", tableName)
	var count int64
	err := o.Raw(countSQL).QueryRow(&count)
	return count, err
}

// GetConfigsByType 根据配置类型获取配置列表
func GetConfigsByType(appId, configType string) ([]*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE config_type = ? AND is_active = true ORDER BY priority DESC", tableName)

	var configs []*GameConfig
	_, err := o.Raw(querySQL, configType).QueryRows(&configs)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	for _, config := range configs {
		config.AppID = appId
	}
	return configs, nil
}

// GetConfigsByTag 根据标签获取配置列表
func GetConfigsByTag(appId, tag string) ([]*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	// 使用 LIKE 查询包含指定标签的配置
	querySQL := fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, tags, created_at, updated_at, created_by FROM %s WHERE tags LIKE ? AND is_active = true ORDER BY priority DESC", tableName)

	var configs []*GameConfig
	tagPattern := fmt.Sprintf("%%\"%s\"%%", tag) // 查找包含该标签的 JSON 数组
	_, err := o.Raw(querySQL, tagPattern).QueryRows(&configs)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	for _, config := range configs {
		config.AppID = appId
	}
	return configs, nil
}

// GetConfigsByVersion 根据版本获取配置列表
func GetConfigsByVersion(appId, version string) ([]*GameConfig, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	var querySQL string
	var params []interface{}

	if version == "" {
		// 获取全局配置（无版本限制）
		querySQL = fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE (version = '' OR version IS NULL) AND is_active = true ORDER BY priority DESC", tableName)
		params = []interface{}{}
	} else {
		// 获取指定版本的配置
		querySQL = fmt.Sprintf("SELECT id, config_key, config_value, version, description, config_type, is_active, priority, created_at, updated_at, created_by FROM %s WHERE version = ? AND is_active = true ORDER BY priority DESC", tableName)
		params = []interface{}{version}
	}

	var configs []*GameConfig
	_, err := o.Raw(querySQL, params...).QueryRows(&configs)
	if err != nil {
		return nil, err
	}

	// 设置 AppID 字段
	for _, config := range configs {
		config.AppID = appId
	}
	return configs, nil
}

// UpdateConfigStatus 批量更新配置状态
func UpdateConfigStatus(appId string, configKeys []string, isActive bool) error {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	if len(configKeys) == 0 {
		return nil
	}

	// 构建 IN 查询
	placeholders := make([]string, len(configKeys))
	params := make([]interface{}, len(configKeys)+1)
	for i := range placeholders {
		placeholders[i] = "?"
		params[i] = configKeys[i]
	}
	params[len(configKeys)] = isActive

	updateSQL := fmt.Sprintf("UPDATE %s SET is_active = ? WHERE config_key IN (%s)", tableName, strings.Join(placeholders, ","))
	_, err := o.Raw(updateSQL, params...).Exec()
	return err
}

// GetActiveConfigKeys 获取所有激活的配置键名
func GetActiveConfigKeys(appId string) ([]string, error) {
	o := orm.NewOrm()
	tableName := GameConfigTableName(appId)

	querySQL := fmt.Sprintf("SELECT DISTINCT config_key FROM %s WHERE is_active = true ORDER BY config_key", tableName)

	var keys []string
	_, err := o.Raw(querySQL).QueryRows(&keys)
	return keys, err
}

func init() {
	orm.RegisterModel(new(GameConfig))
}
