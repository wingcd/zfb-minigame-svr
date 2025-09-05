package models

import (
	"errors"

	"github.com/beego/beego/v2/client/orm"
)

type Application struct {
	Id          int64  `orm:"auto" json:"id"`
	AppId       string `orm:"unique;size(50)" json:"appId"`
	AppName     string `orm:"size(100)" json:"appName"`
	AppSecret   string `orm:"size(100)" json:"appSecret"`
	Description string `orm:"type(text)" json:"description"`
	Status      int    `orm:"default(1)" json:"status"`
	CreatedAt   string `orm:"auto_now_add;type(datetime)" json:"createdAt"`
	UpdatedAt   string `orm:"auto_now;type(datetime)" json:"updatedAt"`
}

func (a *Application) TableName() string {
	return "applications"
}

// GetByAppId 根据AppId获取应用
func (a *Application) GetByAppId() error {
	o := orm.NewOrm()
	err := o.QueryTable(a.TableName()).Filter("app_id", a.AppId).Filter("status", 1).One(a)
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.New("application not found")
		}
		return err
	}
	return nil
}
