package login

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
)

// BaseLoginController 基础登录控制器
type BaseLoginController struct {
	web.Controller
}

// LoginData 登录响应数据结构
type LoginData struct {
	Token    string `json:"token"`
	PlayerId string `json:"playerId"`
	IsNew    bool   `json:"isNew"`
	OpenId   string `json:"openId,omitempty"`
	UnionId  string `json:"unionId,omitempty"`
	Data     string `json:"data"`
}

// CommonResponse 通用响应结构
type CommonResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

// parseRequest 解析请求参数
func (c *BaseLoginController) parseRequest(req interface{}) error {
	return json.Unmarshal(c.Ctx.Input.RequestBody, req)
}

// validateBasicParams 验证基础参数
func (c *BaseLoginController) validateBasicParams(appId, code string) *CommonResponse {
	if appId == "" {
		return &CommonResponse{
			Success: false,
			Code:    4001,
			Msg:     "appId不能为空",
		}
	}

	if code == "" {
		return &CommonResponse{
			Success: false,
			Code:    4001,
			Msg:     "code不能为空",
		}
	}

	return nil
}

// createSuccessResponse 创建成功响应
func (c *BaseLoginController) createSuccessResponse(data interface{}) CommonResponse {
	return CommonResponse{
		Success: true,
		Code:    0,
		Msg:     "success",
		Data:    data,
	}
}

// createErrorResponse 创建错误响应
func (c *BaseLoginController) createErrorResponse(code int, msg string) CommonResponse {
	return CommonResponse{
		Success: false,
		Code:    code,
		Msg:     msg,
	}
}

// sendResponse 发送响应
func (c *BaseLoginController) sendResponse(response CommonResponse) {
	c.Data["json"] = response
	c.ServeJSON()
}

// generateToken 生成token
func generateToken(appId, playerId string) string {
	// 生成简单的token（时间戳 + 哈希）
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%s:%s:%d", appId, playerId, timestamp)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	return fmt.Sprintf("%d:%s", timestamp, hash)
}

// validateSignature 验证签名（预留）
func (c *BaseLoginController) validateSignature(params map[string]interface{}, secret string) bool {
	// TODO: 实现签名验证逻辑
	return true
}
