package controllers

import (
	"game-service/models"
	"game-service/utils"

	"strconv"

	"github.com/beego/beego/v2/server/web"
)

type CounterController struct {
	web.Controller
}

// GetCounter 获取计数器
func (c *CounterController) GetCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 获取参数
	counterName := c.GetString("counterName")
	if counterName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "counterName参数不能为空", nil)
		return
	}

	// 获取全局计数器
	value, err := models.GetGlobalCounter(appId, counterName)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取计数器失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"counterName": counterName,
		"value":       value,
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}

// IncrementCounter 增加计数器
func (c *CounterController) IncrementCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 获取参数
	counterName := c.GetString("counterName")
	incrementStr := c.GetString("increment", "1")

	if counterName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "counterName参数不能为空", nil)
		return
	}

	increment, err := strconv.ParseInt(incrementStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "increment参数格式错误", nil)
		return
	}

	// 增加全局计数器
	newValue, err := models.IncrementGlobalCounter(appId, counterName, increment)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "增加计数器失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"counterName": counterName,
		"value":       newValue,
	}

	utils.SuccessResponse(c.Ctx, "增加成功", result)
}

// DecrementCounter 减少计数器
func (c *CounterController) DecrementCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 获取参数
	counterName := c.GetString("counterName")
	decrementStr := c.GetString("decrement", "1")

	if counterName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "counterName参数不能为空", nil)
		return
	}

	decrement, err := strconv.ParseInt(decrementStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "decrement参数格式错误", nil)
		return
	}

	// 减少全局计数器
	newValue, err := models.DecrementGlobalCounter(appId, counterName, decrement)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "减少计数器失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"counterName": counterName,
		"value":       newValue,
	}

	utils.SuccessResponse(c.Ctx, "减少成功", result)
}

// SetCounter 设置计数器
func (c *CounterController) SetCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 获取参数
	counterName := c.GetString("counterName")
	valueStr := c.GetString("value")

	if counterName == "" || valueStr == "" {
		utils.ErrorResponse(c.Ctx, 1002, "counterName和value参数不能为空", nil)
		return
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "value参数格式错误", nil)
		return
	}

	// 设置全局计数器
	err = models.SetGlobalCounter(appId, counterName, value)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "设置计数器失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"counterName": counterName,
		"value":       value,
	}

	utils.SuccessResponse(c.Ctx, "设置成功", result)
}

// ResetCounter 重置计数器
func (c *CounterController) ResetCounter() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 获取参数
	counterName := c.GetString("counterName")
	if counterName == "" {
		utils.ErrorResponse(c.Ctx, 1002, "counterName参数不能为空", nil)
		return
	}

	// 重置全局计数器
	err := models.ResetGlobalCounter(appId, counterName)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "重置计数器失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"counterName": counterName,
		"value":       0,
	}

	utils.SuccessResponse(c.Ctx, "重置成功", result)
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
