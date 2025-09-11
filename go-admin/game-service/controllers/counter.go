package controllers

import (
	"encoding/json"
	"game-service/models"
	"game-service/utils"

	"github.com/beego/beego/v2/server/web"
)

type CounterController struct {
	web.Controller
}

// JSON请求结构体（对齐JS接口）
type GetCounterRequest struct {
	Key string `json:"key"`
}

type IncrementCounterRequest struct {
	Key       string `json:"key"`
	Location  string `json:"location,omitempty"`
	Increment int64  `json:"increment,omitempty"`
}

type DecrementCounterRequest struct {
	Key       string `json:"key"`
	Location  string `json:"location,omitempty"`
	Decrement int64  `json:"decrement,omitempty"`
}

type SetCounterRequest struct {
	Key      string `json:"key"`
	Location string `json:"location,omitempty"`
	Value    int64  `json:"value"`
}

type ResetCounterRequest struct {
	Key      string `json:"key"`
	Location string `json:"location,omitempty"`
}

// GetCounter 获取计数器（对齐JS getCounter功能）
// 传入key，获取所有location数值
func (c *CounterController) GetCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析JSON请求体
	var req GetCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(c.Ctx, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 验证参数
	if req.Key == "" {
		utils.ErrorResponse(c.Ctx, 4001, "参数[key]错误", nil)
		return
	}

	// 获取计数器值
	value, err := models.GetCounterValues(appId, req.Key)
	if err != nil {
		if err.Error() == "计数器配置不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]不存在，请先在管理后台创建", nil)
		} else {
			utils.ErrorResponse(c.Ctx, 5001, err.Error(), nil)
		}
		return
	}

	result := map[string]interface{}{
		"key":       req.Key,
		"locations": value,
	}

	utils.SuccessResponse(c.Ctx, "success", result)
}

// IncrementCounter 增加计数器（对齐JS incrementCounter功能）
func (c *CounterController) IncrementCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析JSON请求体
	var req IncrementCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(c.Ctx, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 验证参数
	if req.Key == "" {
		utils.ErrorResponse(c.Ctx, 4001, "参数[key]错误", nil)
		return
	}

	// 设置默认值
	location := req.Location
	if location == "" {
		location = "default"
	}

	increment := req.Increment
	if increment <= 0 {
		increment = 1
	}

	// 增加计数器值
	newValue, err := models.IncrementCounterValue(appId, req.Key, location, increment)
	if err != nil {
		if err.Error() == "计数器配置不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]不存在，请先在管理后台创建", nil)
		} else if err.Error() == "点位不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]的点位["+location+"]不存在", nil)
		} else {
			utils.ErrorResponse(c.Ctx, 5001, err.Error(), nil)
		}
		return
	}

	result := map[string]interface{}{
		"key":          req.Key,
		"location":     location,
		"currentValue": newValue,
	}

	utils.SuccessResponse(c.Ctx, "success", result)
}

// DecrementCounter 减少计数器（对齐JS decrementCounter功能）
func (c *CounterController) DecrementCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析JSON请求体
	var req DecrementCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(c.Ctx, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 验证参数
	if req.Key == "" {
		utils.ErrorResponse(c.Ctx, 4001, "参数[key]错误", nil)
		return
	}

	// 设置默认值
	location := req.Location
	if location == "" {
		location = "default"
	}

	decrement := req.Decrement
	if decrement <= 0 {
		decrement = 1
	}

	// 减少计数器值
	newValue, err := models.DecrementCounterValue(appId, req.Key, location, decrement)
	if err != nil {
		if err.Error() == "计数器配置不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]不存在，请先在管理后台创建", nil)
		} else if err.Error() == "点位不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]的点位["+location+"]不存在", nil)
		} else {
			utils.ErrorResponse(c.Ctx, 5001, err.Error(), nil)
		}
		return
	}

	result := map[string]interface{}{
		"key":          req.Key,
		"location":     location,
		"currentValue": newValue,
	}

	utils.SuccessResponse(c.Ctx, "success", result)
}

// SetCounter 设置计数器（对齐JS setCounter功能）
func (c *CounterController) SetCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析JSON请求体
	var req SetCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(c.Ctx, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 验证参数
	if req.Key == "" {
		utils.ErrorResponse(c.Ctx, 4001, "参数[key]错误", nil)
		return
	}

	// 设置默认location
	location := req.Location
	if location == "" {
		location = "default"
	}

	// 设置计数器值
	newValue, err := models.SetCounterValue(appId, req.Key, location, req.Value)
	if err != nil {
		if err.Error() == "计数器配置不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]不存在，请先在管理后台创建", nil)
		} else if err.Error() == "点位不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]的点位["+location+"]不存在", nil)
		} else {
			utils.ErrorResponse(c.Ctx, 5001, err.Error(), nil)
		}
		return
	}

	result := map[string]interface{}{
		"key":          req.Key,
		"location":     location,
		"currentValue": newValue,
	}

	utils.SuccessResponse(c.Ctx, "success", result)
}

// ResetCounter 重置计数器（对齐JS resetCounter功能）
func (c *CounterController) ResetCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析JSON请求体
	var req ResetCounterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(c.Ctx, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 验证参数
	if req.Key == "" {
		utils.ErrorResponse(c.Ctx, 4001, "参数[key]错误", nil)
		return
	}

	// 设置默认location
	location := req.Location
	if location == "" {
		location = "default"
	}

	// 重置计数器值
	newValue, err := models.ResetCounterValue(appId, req.Key, location)
	if err != nil {
		if err.Error() == "计数器配置不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]不存在，请先在管理后台创建", nil)
		} else if err.Error() == "点位不存在" {
			utils.ErrorResponse(c.Ctx, 4004, "计数器["+req.Key+"]的点位["+location+"]不存在", nil)
		} else {
			utils.ErrorResponse(c.Ctx, 5001, err.Error(), nil)
		}
		return
	}

	result := map[string]interface{}{
		"key":          req.Key,
		"location":     location,
		"currentValue": newValue,
	}

	utils.SuccessResponse(c.Ctx, "success", result)
}

// GetAllCounters 获取所有计数器
func (c *CounterController) GetAllCounters() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 获取所有全局计数器
	counters, err := models.GetAllGlobalCounters(appId)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取计数器列表失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "获取成功", counters)
}
