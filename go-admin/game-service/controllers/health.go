package controllers

import (
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
		"status":  "ok",
		"service": "game-service",
		"database": map[string]string{
			"mysql": dbStatus,
			"redis": redisStatus,
		},
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.Data["json"] = health
	c.ServeJSON()
}

// Heartbeat 心跳接口
func (c *HealthController) Heartbeat() {
	// 获取用户ID参数
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.Data["json"] = map[string]interface{}{
			"code":    400,
			"message": "user_id is required",
		}
		c.ServeJSON()
		return
	}

	_, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":    400,
			"message": "invalid user_id format",
		}
		c.ServeJSON()
		return
	}

	// 获取服务器当前时间戳（毫秒）
	serverTime := time.Now().UnixNano() / int64(time.Millisecond)

	// 检查是否有新邮件 - 首先从Redis缓存检查，如果没有则查询数据库
	hasNewMail := false
	appId := c.GetString("app_id", "default") // 获取appId，默认为"default"

	if models.RedisClient != nil {
		// 检查用户是否有新邮件标记（Redis缓存）
		newMailKey := "new_mail:" + appId + ":" + userIDStr
		result, err := models.RedisClient.Get(models.RedisClient.Context(), newMailKey).Result()
		if err == nil && result == "1" {
			hasNewMail = true
		} else {
			// Redis中没有缓存，查询数据库
			hasNewMailDB, err := models.HasNewMail(appId, userIDStr)
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
		hasNewMailDB, err := models.HasNewMail(appId, userIDStr)
		if err == nil {
			hasNewMail = hasNewMailDB
		}
	}

	response := map[string]interface{}{
		"code":         200,
		"message":      "success",
		"server_time":  serverTime,
		"has_new_mail": hasNewMail,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// Options 处理OPTIONS预检请求
func (c *HealthController) Options() {
	// 直接返回200状态码，CORS头部已在中间件中设置
	c.Ctx.Output.SetStatus(200)
}
