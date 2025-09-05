package models

import (
	"errors"

	"github.com/beego/beego/v2/client/orm"
)

// AdminRole 管理员角色模型
type AdminRole struct {
	BaseModel
	RoleCode    string `orm:"unique;size(50)" json:"role_code"` // 对齐云函数的roleCode
	RoleName    string `orm:"size(50)" json:"role_name"`        // 对齐云函数的roleName
	Name        string `orm:"unique;size(50)" json:"name"`      // 保持兼容性
	Description string `orm:"size(255)" json:"description"`
	Permissions string `orm:"type(text)" json:"permissions"`
	Status      int    `orm:"default(1)" json:"status"`
}

func (r *AdminRole) TableName() string {
	return "admin_roles"
}

// Insert 插入角色
func (r *AdminRole) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(r)
	return err
}

// Update 更新角色
func (r *AdminRole) Update(fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(r, fields...)
	return err
}

// GetById 根据ID获取角色
func (r *AdminRole) GetById(id int64) error {
	o := orm.NewOrm()
	r.Id = id
	return o.Read(r)
}

// GetByName 根据名称获取角色
func (r *AdminRole) GetByName(name string) error {
	o := orm.NewOrm()
	return o.QueryTable(r.TableName()).Filter("name", name).One(r)
}

// GetByRoleCode 根据角色代码获取角色
func (r *AdminRole) GetByRoleCode(roleCode string) error {
	o := orm.NewOrm()
	return o.QueryTable(r.TableName()).Filter("role_code", roleCode).One(r)
}

// GetList 获取角色列表
func GetRoleList(page, pageSize int, name string) ([]AdminRole, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("admin_roles").Filter("status", 1)

	if name != "" {
		qs = qs.Filter("name__icontains", name)
	}

	total, _ := qs.Count()

	var roles []AdminRole
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-id").Limit(pageSize, offset).All(&roles)

	return roles, total, err
}

// Delete 删除角色（软删除）
func DeleteRole(id int64) error {
	o := orm.NewOrm()

	// 检查是否有用户使用此角色
	count, err := o.QueryTable("admin_user_roles").Filter("role_id", id).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("role is in use by users") // 使用自定义错误
	}

	_, err = o.QueryTable("admin_roles").Filter("id", id).Delete()
	return err
}

// GetRoles 获取角色列表（别名）
func GetRoles(page, pageSize int, name string) ([]AdminRole, int64, error) {
	return GetRoleList(page, pageSize, name)
}

// GetRoleWithPermissions 获取角色及其权限
func GetRoleWithPermissions(id int64) (*AdminRole, error) {
	o := orm.NewOrm()
	role := &AdminRole{BaseModel: BaseModel{Id: id}}
	err := o.Read(role)
	return role, err
}

// CreateRole 创建角色
func CreateRole(role *AdminRole) error {
	return role.Insert()
}

// UpdateRole 更新角色
func UpdateRole(role *AdminRole) error {
	return role.Update()
}

// IsRoleInUse 检查角色是否被使用
func IsRoleInUse(roleId int64) bool {
	o := orm.NewOrm()
	count, _ := o.QueryTable("admin_users").Filter("role_id", roleId).Count()
	return count > 0
}

// GetAllPermissions 获取所有权限
func GetAllPermissions() ([]map[string]interface{}, error) {
	// 返回静态权限列表，实际项目中可能从数据库或配置文件获取
	permissions := []map[string]interface{}{
		{"code": "admin_manage", "name": "管理员管理", "group": "系统管理"},
		{"code": "role_manage", "name": "角色管理", "group": "系统管理"},
		{"code": "app_manage", "name": "应用管理", "group": "应用管理"},
		{"code": "user_manage", "name": "用户管理", "group": "用户管理"},
		{"code": "leaderboard_manage", "name": "排行榜管理", "group": "游戏管理"},
		{"code": "mail_manage", "name": "邮件管理", "group": "消息管理"},
		{"code": "stats_view", "name": "统计查看", "group": "数据统计"},
		{"code": "system_config", "name": "系统配置", "group": "系统管理"},
		{"code": "counter_manage", "name": "计数器管理", "group": "游戏管理"},
	}
	return permissions, nil
}

// GetPermissionTree 获取权限树结构
func GetPermissionTree() ([]map[string]interface{}, error) {
	// 返回权限树结构，按分组组织
	permissionTree := []map[string]interface{}{
		{
			"group": "系统管理",
			"permissions": []map[string]interface{}{
				{"code": "admin_manage", "name": "管理员管理"},
				{"code": "role_manage", "name": "角色管理"},
				{"code": "system_config", "name": "系统配置"},
			},
		},
		{
			"group": "应用管理",
			"permissions": []map[string]interface{}{
				{"code": "app_manage", "name": "应用管理"},
			},
		},
		{
			"group": "用户管理",
			"permissions": []map[string]interface{}{
				{"code": "user_manage", "name": "用户管理"},
			},
		},
		{
			"group": "游戏管理",
			"permissions": []map[string]interface{}{
				{"code": "leaderboard_manage", "name": "排行榜管理"},
				{"code": "counter_manage", "name": "计数器管理"},
			},
		},
		{
			"group": "消息管理",
			"permissions": []map[string]interface{}{
				{"code": "mail_manage", "name": "邮件管理"},
			},
		},
		{
			"group": "数据统计",
			"permissions": []map[string]interface{}{
				{"code": "stats_view", "name": "统计查看"},
			},
		},
	}
	return permissionTree, nil
}

func init() {
	orm.RegisterModel(new(AdminRole))
}
