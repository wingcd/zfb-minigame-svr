package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// YallaConfig Yalla SDK配置
type YallaConfig struct {
	ID         int       `orm:"auto;pk" json:"id"`
	AppID      string    `orm:"size(50)" json:"app_id" description:"应用ID"`
	SecretKey  string    `orm:"size(200)" json:"secret_key" description:"秘钥"`
	BaseURL    string    `orm:"size(200)" json:"base_url" description:"API基础URL"`
	PushURL    string    `orm:"size(200)" json:"push_url" description:"推送域名URL"`
	Timeout    int       `orm:"default(30)" json:"timeout" description:"请求超时时间(秒)"`
	RetryCount int       `orm:"default(3)" json:"retry_count" description:"重试次数"`
	EnableLog  bool      `orm:"default(true)" json:"enable_log" description:"是否开启日志"`
	Status     int       `orm:"default(1)" json:"status" description:"状态 1启用 0禁用"`
	Remark     string    `orm:"size(500)" json:"remark" description:"备注"`
	CreatedAt  time.Time `orm:"auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt  time.Time `orm:"auto_now;type(datetime)" json:"updated_at"`
}

// YallaCallLog Yalla API调用日志
type YallaCallLog struct {
	ID           int       `orm:"auto;pk" json:"id"`
	AppID        string    `orm:"size(50)" json:"app_id" description:"应用ID"`
	UserID       string    `orm:"size(100)" json:"user_id" description:"用户ID"`
	Method       string    `orm:"size(20)" json:"method" description:"请求方法"`
	Endpoint     string    `orm:"size(200)" json:"endpoint" description:"接口端点"`
	RequestData  string    `orm:"type(text)" json:"request_data" description:"请求数据"`
	ResponseData string    `orm:"type(text)" json:"response_data" description:"响应数据"`
	StatusCode   int       `json:"status_code" description:"HTTP状态码"`
	Duration     int64     `json:"duration" description:"请求耗时(毫秒)"`
	Success      bool      `orm:"default(false)" json:"success" description:"是否成功"`
	ErrorMsg     string    `orm:"size(500)" json:"error_msg" description:"错误信息"`
	CreatedAt    time.Time `orm:"auto_now_add;type(datetime)" json:"created_at"`
}

// YallaUserBinding Yalla用户绑定
type YallaUserBinding struct {
	ID          int       `orm:"auto;pk" json:"id"`
	AppID       string    `orm:"size(50)" json:"app_id" description:"应用ID"`
	GameUserID  string    `orm:"size(100)" json:"game_user_id" description:"游戏用户ID"`
	YallaUserID string    `orm:"size(100)" json:"yalla_user_id" description:"Yalla用户ID"`
	YallaToken  string    `orm:"size(500)" json:"yalla_token" description:"Yalla用户令牌"`
	ExpiresAt   time.Time `json:"expires_at" description:"令牌过期时间"`
	Status      int       `orm:"default(1)" json:"status" description:"状态 1有效 0无效"`
	BindAt      time.Time `orm:"auto_now_add;type(datetime)" json:"bind_at" description:"绑定时间"`
	UpdatedAt   time.Time `orm:"auto_now;type(datetime)" json:"updated_at"`
}

func init() {
	// 注册模型
	orm.RegisterModel(new(YallaConfig))
	orm.RegisterModel(new(YallaCallLog))
	orm.RegisterModel(new(YallaUserBinding))
}

// TableName 设置表名
func (m *YallaConfig) TableName() string {
	return "yalla_config"
}

func (m *YallaCallLog) TableName() string {
	return "yalla_call_logs"
}

func (m *YallaUserBinding) TableName() string {
	return "yalla_user_bindings"
}
