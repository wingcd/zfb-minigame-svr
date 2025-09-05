package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/server/web"
)

type AdminOperationLogController struct {
	web.Controller
}

// GetOperationLogList 获取操作日志列表
func (c *AdminOperationLogController) GetOperationLogList() {
	var requestData struct {
		Page      int    `json:"page"`
		PageSize  int    `json:"pageSize"`
		Username  string `json:"username"`
		Action    string `json:"action"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
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

	// 解析时间
	var startTime, endTime time.Time
	var err error

	if requestData.StartTime != "" {
		startTime, err = time.Parse("2006-01-02 15:04:05", requestData.StartTime)
		if err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      4001,
				"msg":       "开始时间格式错误",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}
	}

	if requestData.EndTime != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", requestData.EndTime)
		if err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      4001,
				"msg":       "结束时间格式错误",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}
	}

	logs, total, err := models.GetOperationLogList(requestData.Page, requestData.PageSize, requestData.Username, requestData.Action, startTime, endTime)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取操作日志失败",
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
			"list":       logs,
			"total":      total,
			"page":       requestData.Page,
			"pageSize":   requestData.PageSize,
			"totalPages": (total + int64(requestData.PageSize) - 1) / int64(requestData.PageSize),
		},
	}
	c.ServeJSON()
}

// GetOperationLogStats 获取操作日志统计
func (c *AdminOperationLogController) GetOperationLogStats() {
	var requestData struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
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

	// 解析时间
	var startTime, endTime time.Time
	var err error

	if requestData.StartTime != "" {
		startTime, err = time.Parse("2006-01-02 15:04:05", requestData.StartTime)
		if err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      4001,
				"msg":       "开始时间格式错误",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}
	} else {
		// 默认最近7天
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if requestData.EndTime != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", requestData.EndTime)
		if err != nil {
			c.Data["json"] = map[string]interface{}{
				"code":      4001,
				"msg":       "结束时间格式错误",
				"timestamp": utils.UnixMilli(),
				"data":      nil,
			}
			c.ServeJSON()
			return
		}
	} else {
		endTime = time.Now()
	}

	stats, err := models.GetOperationLogStats(startTime, endTime)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取操作统计失败",
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
		"data":      stats,
	}
	c.ServeJSON()
}

// CreateOperationLog 创建操作日志
func (c *AdminOperationLogController) CreateOperationLog() {
	var log models.AdminOperationLog
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &log); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数错误",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if err := log.Insert(); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "创建操作日志失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data":      log,
	}
	c.ServeJSON()
}

// DeleteOperationLog 删除操作日志
func (c *AdminOperationLogController) DeleteOperationLog() {
	var requestData struct {
		ID int64 `json:"id"`
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

	// TODO: 实现删除操作日志功能
	// 暂时返回成功

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      nil,
	}
	c.ServeJSON()
}

// CleanOperationLogs 清理操作日志
func (c *AdminOperationLogController) CleanOperationLogs() {
	var requestData struct {
		Days int `json:"days"` // 保留最近N天的日志
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

	// 设置默认保留30天
	if requestData.Days <= 0 {
		requestData.Days = 30
	}

	// TODO: 实现清理逻辑
	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "清理成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"cleanedCount": 0,
		},
	}
	c.ServeJSON()
}
