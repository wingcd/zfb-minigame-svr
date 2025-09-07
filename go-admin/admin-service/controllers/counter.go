package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
)

// CounterController 计数器控制器（对齐云函数接口）
type CounterController struct {
	web.Controller
}

// CreateCounterRequest 创建计数器请求结构
type CreateCounterRequest struct {
	AppId       string `json:"appId"`
	Key         string `json:"key"`
	ResetType   string `json:"resetType"`
	ResetValue  int    `json:"resetValue"`
	Description string `json:"description"`
}

// CreateCounter 创建计数器（对齐云函数createCounter接口）
func (c *CounterController) CreateCounter() {
	var req CreateCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 参数校验
	if req.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[appId]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.Key == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[key]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 设置默认值
	if req.ResetType == "" {
		req.ResetType = "permanent"
	}

	// 验证重置类型
	validResetTypes := []string{"daily", "weekly", "monthly", "custom", "permanent"}
	validType := false
	for _, resetType := range validResetTypes {
		if req.ResetType == resetType {
			validType = true
			break
		}
	}
	if !validType {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "无效的重置类型",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 创建计数器配置
	counter := &models.CounterConfig{
		AppId:       req.AppId,
		CounterKey:  req.Key,
		ResetType:   req.ResetType,
		ResetValue:  req.ResetValue,
		Description: req.Description,
		IsActive:    true,
	}

	if err := models.CreateCounterConfig(counter); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "创建计数器失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 返回成功结果
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"id":          counter.Id,
			"appId":       req.AppId,
			"key":         req.Key,
			"resetType":   req.ResetType,
			"resetValue":  req.ResetValue,
			"description": req.Description,
			"createdAt":   counter.CreatedAt,
		},
	}
	c.ServeJSON()
}

// GetCounter 获取计数器当前值（返回所有点位）
func (c *CounterController) GetCounter() {
	var req GetCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 参数校验
	if req.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[appId]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.Key == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[key]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取计数器配置
	counterConfig, err := models.GetCounterConfig(req.AppId, req.Key)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       fmt.Sprintf("计数器[%s]不存在，请先在管理后台创建", req.Key),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取所有点位数据
	locations, err := models.GetCounterAllLocations(req.AppId, req.Key)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查是否需要重置
	now := time.Now()
	shouldReset := false
	timeToReset := int64(0)
	currentResetTime := ""

	if !counterConfig.NextResetTime.IsZero() {
		currentResetTime = counterConfig.NextResetTime.Format("2006-01-02 15:04:05")
		timeToReset = counterConfig.NextResetTime.Sub(now).Milliseconds()

		if now.After(counterConfig.NextResetTime) {
			shouldReset = true

			// 重新计算下次重置时间
			nextResetTime := calculateNextResetTime(counterConfig.ResetType, counterConfig.ResetValue)
			if !nextResetTime.IsZero() {
				// 更新配置中的重置时间
				err = models.UpdateCounterConfig(req.AppId, req.Key, map[string]interface{}{
					"next_resetTime": nextResetTime,
				})
				if err == nil {
					currentResetTime = nextResetTime.Format("2006-01-02 15:04:05")
					timeToReset = nextResetTime.Sub(now).Milliseconds()
				}

				// 重置所有点位的值
				for locationKey := range locations {
					models.UpdateCounterValue(req.AppId, req.Key, locationKey, 0)
				}

				// 重新获取重置后的数据
				locations, _ = models.GetCounterAllLocations(req.AppId, req.Key)
			}
		}
	}

	// 构建返回数据
	resultLocations := make(map[string]interface{})
	for locationKey, locationData := range locations {
		if shouldReset {
			resultLocations[locationKey] = map[string]interface{}{
				"value": 0,
			}
		} else {
			resultLocations[locationKey] = locationData
		}
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "success",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"key":         counterConfig.CounterKey,
			"locations":   resultLocations,
			"resetType":   counterConfig.ResetType,
			"resetValue":  counterConfig.ResetValue,
			"resetTime":   currentResetTime,
			"timeToReset": timeToReset,
			"description": counterConfig.Description,
		},
	}
	c.ServeJSON()
}

// GetCounterRequest 获取单个计数器请求结构
type GetCounterRequest struct {
	AppId string `json:"appId"`
	Key   string `json:"key"`
}

// GetCounterListRequest 获取计数器列表请求结构
type GetCounterListRequest struct {
	AppId      string `json:"appId"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Key        string `json:"key"`        // 计数器key筛选（模糊搜索）
	ResetType  string `json:"resetType"`  // 重置类型筛选
	GroupByKey bool   `json:"groupByKey"` // 是否按key分组
}

// GetCounterList 获取计数器列表（对齐云函数getCounterList接口）
func (c *CounterController) GetCounterList() {
	var req GetCounterListRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 参数校验
	if req.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[appId]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 获取计数器列表（支持筛选）
	counters, total, err := models.GetCounterConfigListWithFilter(req.AppId, req.Page, req.PageSize, req.Key, req.ResetType)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取计数器列表失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.GroupByKey {
		// 分组模式：返回按key分组的数据
		var counterList []map[string]interface{}
		for _, counter := range counters {
			// 获取该计数器的所有点位数据
			locations, _ := models.GetCounterAllLocations(req.AppId, counter.CounterKey)

			// 计算总值和点位数量
			totalValue := int64(0)
			locationCount := len(locations)
			locationsArray := make([]map[string]interface{}, 0)

			for locationKey, locationData := range locations {
				if valueMap, ok := locationData.(map[string]interface{}); ok {
					if value, ok := valueMap["value"]; ok {
						if valueInt, ok := value.(int64); ok {
							totalValue += valueInt
							locationsArray = append(locationsArray, map[string]interface{}{
								"location": locationKey,
								"value":    valueInt,
							})
						}
					}
				}
			}

			resetTime := ""
			if !counter.NextResetTime.IsZero() {
				resetTime = counter.NextResetTime.Format("2006-01-02 15:04:05")
			}

			counterList = append(counterList, map[string]interface{}{
				"_id":           counter.Id,
				"key":           counter.CounterKey,
				"locations":     locationsArray,
				"locationCount": locationCount,
				"totalValue":    totalValue,
				"resetType":     counter.ResetType,
				"resetValue":    counter.ResetValue,
				"resetTime":     resetTime,
				"description":   counter.Description,
				"gmtCreate":     counter.CreatedAt,
				"gmtModify":     counter.UpdatedAt,
			})
		}

		c.Data["json"] = map[string]interface{}{
			"code":      0,
			"msg":       "success",
			"timestamp": utils.UnixMilli(),
			"data": map[string]interface{}{
				"list":     counterList,
				"total":    total,
				"page":     req.Page,
				"pageSize": req.PageSize,
			},
		}
	} else {
		// 列表模式：返回扁平化的数据
		flatList := make([]map[string]interface{}, 0)

		for _, counter := range counters {
			// 获取该计数器的所有点位数据
			locations, _ := models.GetCounterAllLocations(req.AppId, counter.CounterKey)

			resetTime := ""
			if !counter.NextResetTime.IsZero() {
				resetTime = counter.NextResetTime.Format("2006-01-02 15:04:05")
			}

			for locationKey, locationData := range locations {
				value := int64(0)
				if valueMap, ok := locationData.(map[string]interface{}); ok {
					if v, ok := valueMap["value"]; ok {
						if valueInt, ok := v.(int64); ok {
							value = valueInt
						}
					}
				}

				flatList = append(flatList, map[string]interface{}{
					"_id":         counter.Id,
					"key":         counter.CounterKey,
					"location":    locationKey,
					"value":       value,
					"resetType":   counter.ResetType,
					"resetValue":  counter.ResetValue,
					"resetTime":   resetTime,
					"description": counter.Description,
					"gmtCreate":   counter.CreatedAt,
					"gmtModify":   counter.UpdatedAt,
				})
			}
		}

		// 计算总数和分页
		flatTotal := int64(len(flatList))
		skip := (req.Page - 1) * req.PageSize
		end := skip + req.PageSize
		if end > len(flatList) {
			end = len(flatList)
		}
		if skip > len(flatList) {
			skip = len(flatList)
		}

		paginatedList := flatList[skip:end]

		c.Data["json"] = map[string]interface{}{
			"code":      0,
			"msg":       "success",
			"timestamp": utils.UnixMilli(),
			"data": map[string]interface{}{
				"list":     paginatedList,
				"total":    flatTotal,
				"page":     req.Page,
				"pageSize": req.PageSize,
			},
		}
	}
	c.ServeJSON()
}

// UpdateCounterRequest 更新计数器请求结构
type UpdateCounterRequest struct {
	AppId       string                            `json:"appId"`
	Key         string                            `json:"key"`
	Location    string                            `json:"location"`
	ResetType   string                            `json:"resetType"`
	ResetValue  int                               `json:"resetValue"`
	Description string                            `json:"description"`
	Value       int64                             `json:"value"`
	Locations   map[string]map[string]interface{} `json:"locations"`
}

// UpdateCounter 更新计数器配置（对齐云函数updateCounter接口）
func (c *CounterController) UpdateCounter() {
	var req UpdateCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 参数校验
	if req.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[appId]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.Key == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数[key]错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 更新计数器配置
	err := models.UpdateCounterConfig(req.AppId, req.Key, map[string]interface{}{
		"reset_type":  req.ResetType,
		"reset_value": req.ResetValue,
		"description": req.Description,
	})
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新计数器失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 如果指定了location和value，更新对应点位的值
	if req.Location != "" {
		err = models.UpdateCounterValue(req.AppId, req.Key, req.Location, req.Value)
		if err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      5001,
				"msg":       "更新计数器值失败: " + err.Error(),
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}
	}

	// 批量更新点位配置
	if req.Locations != nil {
		for location, locationData := range req.Locations {
			if value, ok := locationData["value"]; ok {
				if valueInt, ok := value.(float64); ok {
					err = models.UpdateCounterValue(req.AppId, req.Key, location, int64(valueInt))
					if err != nil {
						// 记录错误但不中断处理
						continue
					}
				}
			}
		}
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "更新成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// DeleteCounterRequest 删除计数器请求结构
type DeleteCounterRequest struct {
	AppId string `json:"appId"`
	Key   string `json:"key"`
}

// DeleteCounter 删除计数器（对齐云函数deleteCounter接口）
func (c *CounterController) DeleteCounter() {
	var req DeleteCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 参数校验
	if req.AppId == "" || req.Key == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 删除计数器配置和数据
	err := models.DeleteCounterConfig(req.AppId, req.Key)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除计数器失败: " + err.Error(),
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

// GetAllCounterStats 获取所有计数器统计信息
func (c *CounterController) GetAllCounterStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	// 获取所有计数器配置
	counters, _, err := models.GetCounterConfigList(appId, 1, 1000)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取计数器列表失败: "+err.Error(), nil)
		return
	}

	// 统计信息
	stats := make([]map[string]interface{}, 0)
	for _, counter := range counters {
		// 获取该计数器的所有点位数据
		locations, _ := models.GetCounterAllLocations(appId, counter.CounterKey)

		// 计算总值
		totalValue := int64(0)
		for _, locationData := range locations {
			if valueMap, ok := locationData.(map[string]interface{}); ok {
				if value, ok := valueMap["value"]; ok {
					if valueInt, ok := value.(int64); ok {
						totalValue += valueInt
					}
				}
			}
		}

		stats = append(stats, map[string]interface{}{
			"key":           counter.CounterKey,
			"description":   counter.Description,
			"totalValue":    totalValue,
			"locationCount": len(locations),
			"resetType":     counter.ResetType,
		})
	}

	utils.SuccessResponse(&c.Controller, "获取成功", map[string]interface{}{
		"stats": stats,
		"total": len(stats),
	})
}

// GetCounterConfig 获取计数器配置
func (c *CounterController) GetCounterConfig() {
	var req GetCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	config, err := models.GetCounterConfig(req.AppId, req.Key)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "计数器配置不存在",
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

// GetCounterValue 获取计数器指定位置的值
func (c *CounterController) GetCounterValue() {
	var req struct {
		AppId    string `json:"appId"`
		Key      string `json:"key"`
		Location string `json:"location"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	value, err := models.GetCounterValue(req.AppId, req.Key, req.Location)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取计数器值失败",
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
			"value": value,
		},
	}
	c.ServeJSON()
}

// UpdateCounterValue 更新计数器值
func (c *CounterController) UpdateCounterValue() {
	var req struct {
		AppId    string `json:"appId"`
		Key      string `json:"key"`
		Location string `json:"location"`
		Value    int64  `json:"value"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := models.UpdateCounterValue(req.AppId, req.Key, req.Location, req.Value); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "更新计数器值失败",
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
