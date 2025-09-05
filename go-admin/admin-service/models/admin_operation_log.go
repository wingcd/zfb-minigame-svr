package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// AdminOperationLog 管理员操作日志模型
type AdminOperationLog struct {
	BaseModel
	UserId    int64  `orm:"" json:"user_id"`
	Username  string `orm:"size(50)" json:"username"`
	Action    string `orm:"size(100)" json:"action"`
	Resource  string `orm:"size(100)" json:"resource"`
	Params    string `orm:"type(text)" json:"params"`
	IpAddress string `orm:"size(45)" json:"ip_address"`
	UserAgent string `orm:"size(500)" json:"user_agent"`
}

func (l *AdminOperationLog) TableName() string {
	return "admin_operation_logs"
}

// Insert 插入日志
func (l *AdminOperationLog) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(l)
	return err
}

// GetList 获取操作日志列表
func GetOperationLogList(page, pageSize int, username, action string, startTime, endTime time.Time) ([]AdminOperationLog, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("admin_operation_logs")

	// 条件过滤
	if username != "" {
		qs = qs.Filter("username__icontains", username)
	}
	if action != "" {
		qs = qs.Filter("action__icontains", action)
	}
	if !startTime.IsZero() {
		qs = qs.Filter("created_at__gte", startTime)
	}
	if !endTime.IsZero() {
		qs = qs.Filter("created_at__lte", endTime)
	}

	// 获取总数
	total, _ := qs.Count()

	// 获取列表
	var logs []AdminOperationLog
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-created_at").Limit(pageSize, offset).All(&logs)

	return logs, total, err
}

// GetStats 获取统计数据
func GetOperationLogStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	o := orm.NewOrm()
	result := make(map[string]interface{})

	// 总操作次数
	totalCount, err := o.QueryTable("admin_operation_logs").
		Filter("created_at__gte", startTime).
		Filter("created_at__lte", endTime).
		Count()
	if err != nil {
		return nil, err
	}
	result["total_count"] = totalCount

	// 按模块统计
	var moduleStats []orm.Params
	_, err = o.Raw("SELECT action as module, COUNT(*) as count FROM admin_operation_logs WHERE created_at BETWEEN ? AND ? GROUP BY action ORDER BY count DESC", startTime, endTime).Values(&moduleStats)
	if err != nil {
		return nil, err
	}
	result["module_stats"] = moduleStats

	// 按管理员统计
	var adminStats []orm.Params
	_, err = o.Raw("SELECT username, COUNT(*) as count FROM admin_operation_logs WHERE created_at BETWEEN ? AND ? GROUP BY username ORDER BY count DESC LIMIT 10", startTime, endTime).Values(&adminStats)
	if err != nil {
		return nil, err
	}
	result["admin_stats"] = adminStats

	return result, nil
}

// GetOperationsByDate 获取指定日期的操作数量
func GetOperationsByDate(date string) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("admin_operation_logs").
		Filter("created_at__date", date).
		Count()
	return count, err
}

// GetOperationLogs 获取操作日志列表
func GetOperationLogs(page, pageSize int, adminId, action, startDate, endDate string) ([]*AdminOperationLog, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("admin_operation_logs")

	if adminId != "" && adminId != "0" {
		qs = qs.Filter("user_id", adminId)
	}
	if action != "" {
		qs = qs.Filter("action__icontains", action)
	}
	if startDate != "" {
		qs = qs.Filter("created_at__gte", startDate)
	}
	if endDate != "" {
		qs = qs.Filter("created_at__lte", endDate)
	}

	total, _ := qs.Count()

	var logs []*AdminOperationLog
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-created_at").Limit(pageSize, offset).All(&logs)

	return logs, total, err
}

func init() {
	orm.RegisterModel(new(AdminOperationLog))
}
