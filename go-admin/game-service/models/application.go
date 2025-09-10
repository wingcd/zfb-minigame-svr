package models

import (
	"errors"

	"github.com/beego/beego/v2/client/orm"
)

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

func (a *Application) TableName() string {
	return "apps"
}

// GetByAppId 根据AppId获取应用
func (a *Application) GetByAppId(appId string) error {
	o := orm.NewOrm()
	// 修复状态过滤，改为字符串类型的"active"
	err := o.QueryTable(a.TableName()).Filter("app_id", appId).Filter("status", "active").One(a)
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.New("application not found")
		}
		return err
	}
	return nil
}
