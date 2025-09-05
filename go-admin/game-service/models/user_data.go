package models

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

// UserData 用户数据模型
type UserData struct {
	Id        int64  `orm:"auto" json:"id"`
	UserId    string `orm:"size(100);unique" json:"user_id"`
	Data      string `orm:"type(longtext)" json:"data"`
	CreatedAt string `orm:"auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt string `orm:"auto_now;type(datetime)" json:"updated_at"`
}

// GetTableName 获取动态表名
func (u *UserData) GetTableName(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return fmt.Sprintf("user_data_%s", cleanAppId)
}

// SaveData 保存用户数据
func SaveUserData(appId, userId, data string) error {
	o := orm.NewOrm()

	userData := &UserData{}
	tableName := userData.GetTableName(appId)

	// 检查用户数据是否存在
	err := o.QueryTable(tableName).Filter("user_id", userId).One(userData)
	if err == orm.ErrNoRows {
		// 新建用户数据
		userData.UserId = userId
		userData.Data = data
		_, err = o.Insert(userData)
	} else if err == nil {
		// 更新用户数据
		userData.Data = data
		_, err = o.Update(userData, "data", "updated_at")
	}

	return err
}

// GetData 获取用户数据
func GetUserData(appId, userId string) (string, error) {
	o := orm.NewOrm()

	userData := &UserData{}
	tableName := userData.GetTableName(appId)

	err := o.QueryTable(tableName).Filter("user_id", userId).One(userData)
	if err == orm.ErrNoRows {
		return "", nil // 返回空字符串表示无数据
	} else if err != nil {
		return "", err
	}

	return userData.Data, nil
}

// DeleteData 删除用户数据
func DeleteUserData(appId, userId string) error {
	o := orm.NewOrm()

	userData := &UserData{}
	tableName := userData.GetTableName(appId)

	_, err := o.QueryTable(tableName).Filter("user_id", userId).Delete()
	return err
}

// GetUserDataList 获取用户数据列表（管理后台使用）
func GetUserDataList(appId string, page, pageSize int) ([]UserData, int64, error) {
	o := orm.NewOrm()

	userData := &UserData{}
	tableName := userData.GetTableName(appId)

	qs := o.QueryTable(tableName)
	total, _ := qs.Count()

	var dataList []UserData
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-id").Limit(pageSize, offset).All(&dataList)

	return dataList, total, err
}

func init() {
	orm.RegisterModel(new(UserData))
}
