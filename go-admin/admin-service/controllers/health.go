package controllers

import (
	"admin-service/models"
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
		"service": "admin-service",
		"database": map[string]string{
			"mysql": dbStatus,
			"redis": redisStatus,
		},
		"timestamp": time.Now().Unix(),
	}

	c.Data["json"] = health
	c.ServeJSON()
}
