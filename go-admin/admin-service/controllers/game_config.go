package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

type GameConfigController struct {
	web.Controller
}

// GetGameConfigList 获取游戏配置列表
func (c *GameConfigController) GetGameConfigList() {
	var requestData struct {
		AppId      string `json:"appId"`
		Page       int    `json:"page"`
		PageSize   int    `json:"pageSize"`
		ConfigType string `json:"configType"`
		Version    string `json:"version"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 设置默认值
	if requestData.Page <= 0 {
		requestData.Page = 1
	}
	if requestData.PageSize <= 0 {
		requestData.PageSize = 20
	}

	configs, total, err := models.GetGameConfigList(requestData.AppId, requestData.Page, requestData.PageSize, requestData.ConfigType, requestData.Version)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取游戏配置列表失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"list":       configs,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// CreateGameConfig 创建游戏配置
func (c *GameConfigController) CreateGameConfig() {
	var requestData struct {
		AppId       string      `json:"appId"`
		ConfigKey   string      `json:"configKey"`
		ConfigValue interface{} `json:"configValue"`
		ConfigType  string      `json:"configType"` // 修正字段名
		Version     string      `json:"version"`
		Description string      `json:"description"`
		IsActive    bool        `json:"isActive"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		logs.Error("JSON解析错误:", err)
		logs.Error("请求数据:", string(c.Ctx.Input.RequestBody))
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	logs.Info("接收到的请求数据:", requestData)

	// 参数验证
	if requestData.AppId == "" || requestData.ConfigKey == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "缺少必要参数",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 将 ConfigValue 转换为字符串
	configValueStr := ""
	if requestData.ConfigValue != nil {
		switch v := requestData.ConfigValue.(type) {
		case string:
			configValueStr = v
		default:
			if jsonBytes, err := json.Marshal(requestData.ConfigValue); err == nil {
				configValueStr = string(jsonBytes)
			}
		}
	}

	config := &models.GameConfig{
		AppID:       requestData.AppId,
		ConfigKey:   requestData.ConfigKey,
		ConfigValue: configValueStr,
		ConfigType:  requestData.ConfigType,
		Version:     requestData.Version,
		Description: requestData.Description,
		IsActive:    requestData.IsActive,
	}

	if err := models.CreateGameConfig(config); err != nil {
		if err.Error() == "配置已存在" {
			c.Data["json"] = map[string]interface{}{
				"code":      4002,
				"msg":       err.Error(),
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
		} else {
			c.Data["json"] = map[string]interface{}{
				"code":      5001,
				"msg":       "创建游戏配置失败",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data":      config,
	}
	c.ServeJSON()
}

// UpdateGameConfig 更新游戏配置
func (c *GameConfigController) UpdateGameConfig() {
	var requestData models.UpdateGameConfigRequest

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		logs.Error("UpdateGameConfig 解析请求参数失败: %v", err)
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.UpdateGameConfigByRequest(&requestData); err != nil {
		logs.Error("UpdateGameConfig 更新失败: %v", err)
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       fmt.Sprintf("更新游戏配置失败: %v", err),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// DeleteGameConfig 删除游戏配置
func (c *GameConfigController) DeleteGameConfig() {
	var requestData struct {
		AppId     string `json:"appId"`
		ConfigKey string `json:"configKey"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.DeleteGameConfigByKey(requestData.AppId, requestData.ConfigKey); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除游戏配置失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// GetGameConfig 获取指定游戏配置
func (c *GameConfigController) GetGameConfig() {
	var requestData struct {
		AppId     string `json:"appId"`
		ConfigKey string `json:"configKey"`
		Version   string `json:"version"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	config, err := models.GetGameConfig(requestData.AppId, requestData.ConfigKey, requestData.Version)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "配置不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      config,
	}
	c.ServeJSON()
}

// GetAllGameConfigs 获取所有游戏配置
func (c *GameConfigController) GetAllGameConfigs() {
	var requestData struct {
		AppId     string `json:"appId"`
		ConfigKey string `json:"configKey"`
		Page      int    `json:"page"`
		PageSize  int    `json:"pageSize"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 设置默认值
	if requestData.Page <= 0 {
		requestData.Page = 1
	}
	if requestData.PageSize <= 0 {
		requestData.PageSize = 20
	}

	configs, total, err := models.GetAllGameConfigs(requestData.Page, requestData.PageSize, requestData.AppId, requestData.ConfigKey)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取游戏配置失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"list":       configs,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// GetGameConfigById 根据ID获取游戏配置
func (c *GameConfigController) GetGameConfigById() {
	var requestData struct {
		ID    int64  `json:"id"`
		AppId string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	config, err := models.GetGameConfigById(requestData.ID, requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "配置不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      config,
	}
	c.ServeJSON()
}

// GetGameConfigByKey 根据AppId和Key获取游戏配置
func (c *GameConfigController) GetGameConfigByKey() {
	var requestData struct {
		AppId     string `json:"appId"`
		ConfigKey string `json:"configKey"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	config, err := models.GetGameConfigByKey(requestData.AppId, requestData.ConfigKey)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "配置不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      config,
	}
	c.ServeJSON()
}

// GetGameConfigsByAppId 根据AppId获取所有配置
func (c *GameConfigController) GetGameConfigsByAppId() {
	var requestData struct {
		AppId string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	configs, err := models.GetGameConfigsByAppId(requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取配置失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      configs,
	}
	c.ServeJSON()
}

// GetPublicGameConfigs 获取公开游戏配置
func (c *GameConfigController) GetPublicGameConfigs() {
	var requestData struct {
		AppId string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	configs, err := models.GetPublicGameConfigs(requestData.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取公开配置失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      configs,
	}
	c.ServeJSON()
}

// AddGameConfig 添加游戏配置
func (c *GameConfigController) AddGameConfig() {
	var config models.GameConfig
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.AddGameConfig(&config); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "添加配置失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "添加成功",
		"timestamp": utils.UnixMilli(),
		"data":      config,
	}
	c.ServeJSON()
}

// BatchUpdateGameConfigs 批量更新游戏配置
func (c *GameConfigController) BatchUpdateGameConfigs() {
	var requestData struct {
		AppId   string            `json:"appId"`
		Configs map[string]string `json:"configs"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.BatchUpdateGameConfigs(requestData.AppId, requestData.Configs); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "批量更新失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// DeleteGameConfigsByAppId 删除应用所有配置
func (c *GameConfigController) DeleteGameConfigsByAppId() {
	var requestData struct {
		AppId string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.DeleteGameConfigsByAppId(requestData.AppId); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除配置失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}
