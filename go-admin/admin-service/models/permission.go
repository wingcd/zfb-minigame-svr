package models

import (
	"encoding/json"
	"errors"

	"github.com/beego/beego/v2/client/orm"
)

// Permission 权限模型
type Permission struct {
	BaseModel
	Code        string `orm:"unique;size(50)" json:"code"`
	Name        string `orm:"size(100)" json:"name"`
	Description string `orm:"size(255)" json:"description"`
	ParentId    int64  `orm:"default(0)" json:"parent_id"`
	Sort        int    `orm:"default(0)" json:"sort"`
	Status      int    `orm:"default(1)" json:"status"`
}

func (p *Permission) TableName() string {
	return "permissions"
}

// CreatePermission 创建权限
func CreatePermission(permission *Permission) error {
	o := orm.NewOrm()
	_, err := o.Insert(permission)
	return err
}

// UpdatePermission 更新权限
func UpdatePermission(permission *Permission) error {
	o := orm.NewOrm()
	_, err := o.Update(permission)
	return err
}

// DeletePermission 删除权限
func DeletePermission(id int64) error {
	o := orm.NewOrm()

	// 检查是否有子权限
	if HasChildPermissions(id) {
		return errors.New("该权限下存在子权限，不能删除")
	}

	// 检查是否被使用
	if IsPermissionInUse(id) {
		return errors.New("该权限正在使用中，不能删除")
	}

	_, err := o.QueryTable("permissions").Filter("id", id).Delete()
	return err
}

// HasChildPermissions 检查是否有子权限
func HasChildPermissions(parentId int64) bool {
	o := orm.NewOrm()
	count, _ := o.QueryTable("permissions").Filter("parent_id", parentId).Count()
	return count > 0
}

// IsPermissionInUse 检查权限是否被使用
func IsPermissionInUse(permissionId int64) bool {
	o := orm.NewOrm()

	// 检查角色权限关联
	var roles []AdminRole
	o.QueryTable("admin_roles").All(&roles)

	for _, role := range roles {
		if role.Permissions != "" {
			var permissions []string
			if err := json.Unmarshal([]byte(role.Permissions), &permissions); err == nil {
				for _, perm := range permissions {
					if perm == string(permissionId) {
						return true
					}
				}
			}
		}
	}

	return false
}

// GetPermissionList 获取权限列表
func GetPermissionList(page, pageSize int, parentId int64) ([]Permission, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("permissions").Filter("status", 1)

	if parentId >= 0 {
		qs = qs.Filter("parent_id", parentId)
	}

	total, _ := qs.Count()

	var permissions []Permission
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("sort", "id").Limit(pageSize, offset).All(&permissions)

	return permissions, total, err
}

func init() {
	orm.RegisterModel(new(Permission))
}
