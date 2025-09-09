package controllers

import (
	"game-service/models"
	"game-service/utils"

	"github.com/beego/beego/v2/server/web"
)

type UserDataController struct {
	web.Controller
}

// SaveData 保存用户数据
func (c *UserDataController) SaveData() {
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取数据
	data := c.GetString("data")
	if data == "" {
		utils.ErrorResponse(c.Ctx, 1002, "data参数不能为空", nil)
		return
	}

	// 保存数据
	err = models.SaveUserDataWithKey(appId, userId, "default", data)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "保存数据失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "保存成功", nil)
}

// GetData 获取用户数据
func (c *UserDataController) GetData() {
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取数据
	data, err := models.GetUserDataWithKey(appId, userId, "default")
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取数据失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"data": data,
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}

// DeleteData 删除用户数据
func (c *UserDataController) DeleteData() {
	// 验证签名
	appId, userId, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 删除数据
	err = models.DeleteUserDataWithKey(appId, userId, "default")
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "删除数据失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "删除成功", nil)
}
