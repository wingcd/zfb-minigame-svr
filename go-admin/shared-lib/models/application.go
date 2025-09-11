package models

// Application 应用模型
type Application struct {
	BaseModel
	AppId         string `orm:"unique;size(50);column(app_id)" json:"appId"`               // 应用ID（唯一）
	AppName       string `orm:"size(100);column(app_name)" json:"appName"`                 // 应用名称
	Description   string `orm:"type(text);column(description)" json:"description"`         // 应用描述
	ChannelAppId  string `orm:"size(100);column(channel_app_id)" json:"channelAppId"`      // 渠道应用ID
	ChannelAppKey string `orm:"size(100);column(channel_app_key)" json:"channelAppKey"`    // 渠道应用密钥
	Category      string `orm:"size(50);default('game');column(category)" json:"category"` // 应用分类: game/tool/social
	Platform      string `orm:"size(50);column(platform)" json:"platform"`                 // 平台: alipay/wechat/baidu
	Status        string `orm:"size(20);default('active');column(status)" json:"status"`   // 状态: active/inactive/pending
	Version       string `orm:"size(50);column(version)" json:"version"`                   // 当前版本
	MinVersion    string `orm:"size(50);column(min_version)" json:"minVersion"`            // 最低支持版本
	Settings      string `orm:"type(text);column(settings)" json:"settings"`               // 应用设置(JSON格式)
	UserCount     int64  `orm:"default(0);column(user_count)" json:"userCount"`            // 用户数量
	ScoreCount    int64  `orm:"default(0);column(score_count)" json:"scoreCount"`          // 分数记录数
	DailyActive   int64  `orm:"default(0);column(daily_active)" json:"dailyActive"`        // 日活跃用户
	MonthlyActive int64  `orm:"default(0);column(monthly_active)" json:"monthlyActive"`    // 月活跃用户
	CreatedBy     string `orm:"size(50);column(created_by)" json:"createdBy"`              // 创建者
}

// TableName 返回表名
func (a *Application) TableName() string {
	return "apps"
}

// Register 注册模型
func (a *Application) Register() error {
	return RegisterModel(new(Application))
}

// RegisterWithSuffix 带后缀注册模型
func (a *Application) RegisterWithSuffix(suffix string) error {
	return RegisterModelWithSuffix(new(Application), suffix)
}

// Init 初始化模型
func (a *Application) Init() error {
	return a.Register()
}

// IsActive 检查应用是否活跃
func (a *Application) IsActive() bool {
	return a.Status == "active"
}
