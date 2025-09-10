package controllers

import (
	"encoding/json"
	"game-service/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"
)

// HealthController 健康检查控制器
type HealthController struct {
	web.Controller
}

// Health 健康检查接口
func (c *HealthController) Health() {
	// 检查数据库连接
	dbStatus := "ok"
	if models.RedisClient == nil {
		dbStatus = "error"
	}

	// 检查Redis连接
	redisStatus := "ok"
	if models.RedisClient != nil {
		_, err := models.RedisClient.Ping(models.RedisClient.Context()).Result()
		if err != nil {
			redisStatus = "error"
		}
	} else {
		redisStatus = "error"
	}

	health := map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"status":  "ok",
			"service": "game-service",
			"database": map[string]string{
				"mysql": dbStatus,
				"redis": redisStatus,
			},
		},
		"timestamp": time.Now().UnixNano() / 1e6,
	}

	c.Data["json"] = health
	c.ServeJSON()
}

// HeartbeatRequest 心跳请求结构
type HeartbeatRequest struct {
	AppId    string `json:"appId"`    // 应用ID
	PlayerId string `json:"playerId"` // 玩家ID
}

// Heartbeat 心跳接口
func (c *HealthController) Heartbeat() {
	var req HeartbeatRequest

	// 解析JSON请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	// 验证必需参数
	if req.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":    400,
			"message": "appId is required",
		}
		c.ServeJSON()
		return
	}

	if req.PlayerId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":    400,
			"message": "playerId is required",
		}
		c.ServeJSON()
		return
	}

	// 验证playerId格式（应该是数字）
	_, err := strconv.ParseInt(req.PlayerId, 10, 64)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":    400,
			"message": "invalid playerId format",
		}
		c.ServeJSON()
		return
	}

	// 检查是否有新邮件 - 首先从Redis缓存检查，如果没有则查询数据库
	hasNewMail := false

	if models.RedisClient != nil {
		// 检查用户是否有新邮件标记（Redis缓存）
		newMailKey := "new_mail:" + req.AppId + ":" + req.PlayerId
		result, err := models.RedisClient.Get(models.RedisClient.Context(), newMailKey).Result()
		if err == nil && result == "1" {
			hasNewMail = true
		} else {
			// Redis中没有缓存，查询数据库
			hasNewMailDB, err := models.HasNewMail(req.AppId, req.PlayerId)
			if err == nil {
				hasNewMail = hasNewMailDB
				// 将结果缓存到Redis（5分钟过期）
				if hasNewMail {
					models.RedisClient.Set(models.RedisClient.Context(), newMailKey, "1", 5*60*time.Second)
				} else {
					models.RedisClient.Set(models.RedisClient.Context(), newMailKey, "0", 5*60*time.Second)
				}
			}
		}
	} else {
		// 没有Redis，直接查询数据库
		hasNewMailDB, err := models.HasNewMail(req.AppId, req.PlayerId)
		if err == nil {
			hasNewMail = hasNewMailDB
		}
	}

	response := map[string]interface{}{
		"code":      0,
		"message":   "success",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"data": map[string]interface{}{
			"hasNewMail": hasNewMail,
		}}

	c.Data["json"] = response
	c.ServeJSON()
}
