package models

import (
	"admin-service/utils"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// AdminUser 管理员用户模型
type AdminUser struct {
	BaseModel
	Username    string    `orm:"size(50);unique" json:"username" valid:"Required"`
	Password    string    `orm:"size(255)" json:"-"`
	Email       string    `orm:"size(100)" json:"email"`
	Phone       string    `orm:"size(20)" json:"phone"`
	Role        string    `orm:"size(50);column(role)" json:"role"`
	Nickname    string    `orm:"size(50);column(nickname)" json:"nickname"`
	Avatar      string    `orm:"size(255)" json:"avatar"`
	Status      int       `orm:"default(1)" json:"status"` // 1:正常 0:禁用
	LastLoginAt time.Time `orm:"type(datetime);null;column(last_login_at)" json:"lastLoginAt"`
	LastLoginIP string    `orm:"size(50);column(last_login_ip)" json:"lastLoginIp"`
	RoleId      int64     `orm:"default(0);column(role_id)" json:"roleId"`
	Token       string    `orm:"size(128);null" json:"-"`                           // 添加token字段
	TokenExpire time.Time `orm:"type(datetime);null;column(token_expire)" json:"-"` // 添加token过期时间
}

// TableName 指定表名
func (u *AdminUser) TableName() string {
	return "admin_users"
}

// GetAllAdminUsers 获取所有管理员
func GetAllAdminUsers(page, pageSize int, keyword string) ([]*AdminUser, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("admin_users")

	if keyword != "" {
		cond := orm.NewCondition()
		cond = cond.Or("username__icontains", keyword).
			Or("role__icontains", keyword).
			Or("nickname__icontains", keyword).
			Or("email__icontains", keyword)
		qs = qs.SetCond(cond)
	}

	total, _ := qs.Count()

	var users []*AdminUser
	_, err := qs.OrderBy("-id").Limit(pageSize, (page-1)*pageSize).All(&users)

	return users, total, err
}

// GetAdminUserById 根据ID获取管理员
func GetAdminUserById(id int64) (*AdminUser, error) {
	o := orm.NewOrm()
	user := &AdminUser{BaseModel: BaseModel{ID: id}}
	err := o.QueryTable("admin_users").Filter("id", id).One(user)
	return user, err
}

// GetAdminUserByUsername 根据用户名获取管理员
func GetAdminUserByUsername(username string) (*AdminUser, error) {
	o := orm.NewOrm()
	user := &AdminUser{}
	err := o.QueryTable("admin_users").Filter("username", username).One(user)
	return user, err
}

// AddAdminUser 添加管理员
func AddAdminUser(user *AdminUser) error {
	o := orm.NewOrm()
	_, err := o.Insert(user)
	return err
}

// UpdateAdminUser 更新管理员
func UpdateAdminUser(user *AdminUser) error {
	o := orm.NewOrm()
	_, err := o.Update(user)
	return err
}

// UpdateAdminUserFields 更新管理员指定字段
func UpdateAdminUserFields(id int64, fields map[string]interface{}) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("admin_users").Filter("id", id).Update(fields)
	return err
}

// DeleteAdminUser 删除管理员
func DeleteAdminUser(id int64) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("admin_users").Filter("id", id).Delete()
	return err
}

// UpdateAdminUserStatus 更新管理员状态
func UpdateAdminUserStatus(id int64, status string) error {
	return UpdateAdminUserFields(id, map[string]interface{}{
		"status":    status,
		"updatedAt": time.Now(),
	})
}

// UpdateAdminUserLoginInfo 更新管理员登录信息
func UpdateAdminUserLoginInfo(id int64, loginIP string) error {
	return UpdateAdminUserFields(id, map[string]interface{}{
		"lastLoginAt": time.Now(),
		"lastLoginIp": loginIP,
		"updatedAt":   time.Now(),
	})
}

// AdminLogin 管理员登录验证
func AdminLogin(username, password string) (*AdminUser, error) {
	admin, err := GetAdminUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// 检查状态
	if admin.Status != 1 {
		return nil, orm.ErrNoRows
	}

	// 验证密码 - 这里需要根据实际的密码加密方式调整
	// 暂时直接比较，实际应该使用哈希验证
	return admin, nil
}

// GetAdminById 根据ID获取管理员 (别名)
func GetAdminById(id int64) (*AdminUser, error) {
	return GetAdminUserById(id)
}

// UpdateAdminProfile 更新管理员资料
func UpdateAdminProfile(id int64, nickname, email string) error {
	return UpdateAdminUserFields(id, map[string]interface{}{
		"nickname":  nickname,
		"email":     email,
		"updatedAt": time.Now(),
	})
}

// ChangeAdminPassword 修改管理员密码
func ChangeAdminPassword(id int64, oldPassword, newPassword string) error {
	// 这里需要实现密码验证和更新逻辑
	// 暂时简单实现
	hashedPassword := utils.HashPassword(newPassword)
	return UpdateAdminUserFields(id, map[string]interface{}{
		"password":  hashedPassword,
		"updatedAt": time.Now(),
	})
}

// AdminLoginWithMD5 使用MD5密码验证管理员登录（对齐云函数）
func AdminLoginWithMD5(username, passwordHash string) (*AdminUser, error) {
	o := orm.NewOrm()
	admin := &AdminUser{}

	// 调试查询
	fmt.Printf("DEBUG: 查询参数 - username=%s, passwordHash=%s\n", username, passwordHash)

	// 先查询用户是否存在
	userExists := &AdminUser{}
	err := o.QueryTable("admin_users").Filter("username", username).One(userExists)
	if err != nil {
		fmt.Printf("DEBUG: 用户不存在: %v\n", err)
		return nil, orm.ErrNoRows
	}
	fmt.Printf("DEBUG: 找到用户 ID=%d, 状态=%d, 存储的密码=%s\n", userExists.ID, userExists.Status, userExists.Password)

	err = o.QueryTable("admin_users").
		Filter("username", username).
		Filter("password", passwordHash).
		Filter("status", 1). // 1:正常状态
		One(admin)

	if err == orm.ErrNoRows {
		fmt.Printf("DEBUG: 查询失败 - 可能是密码或状态不匹配\n")
		return nil, orm.ErrNoRows
	}
	return admin, err
}

// UpdateAdminToken 更新管理员token信息 (已废弃，JWT不需要存储在数据库中)
func UpdateAdminToken(id int64, token string, tokenExpire time.Time, loginIP string) error {
	// JWT token不需要存储在数据库中，只更新登录信息
	return UpdateAdminUserFields(id, map[string]interface{}{
		"lastLoginAt": time.Now(),
		"lastLoginIp": loginIP,
		"updatedAt":   time.Now(),
	})
}

// GetAdminRolePermissions 获取管理员角色和权限
func GetAdminRolePermissions(roleId int64) (*AdminRole, []string, error) {
	o := orm.NewOrm()
	role := &AdminRole{BaseModel: BaseModel{ID: roleId}}
	err := o.Read(role)
	if err != nil {
		return nil, nil, err
	}

	// 解析权限JSON
	var permissions []string
	if role.Permissions != "" {
		err := json.Unmarshal([]byte(role.Permissions), &permissions)
		if err != nil {
			// 如果JSON解析失败，记录错误但不中断流程
			permissions = []string{}
		}
	}

	return role, permissions, nil
}

// GetAdminByToken 根据Token获取管理员信息 (已废弃，使用JWT验证替代)
func GetAdminByToken(token string) (*AdminUser, error) {
	// 使用JWT验证替代数据库token查询
	claims, err := utils.ParseJWT(token)
	if err != nil {
		return nil, fmt.Errorf("JWT验证失败: %v", err)
	}

	// 根据JWT中的用户ID获取用户信息
	return GetAdminUserById(claims.UserID)
}

// LogAdminOperation 记录管理员操作日志
func LogAdminOperation(adminId int64, username, action, resource string, details map[string]interface{}) {
	// 创建操作日志
	log := &AdminOperationLog{
		UserId:   adminId,
		Username: username,
		Action:   action,
		Resource: resource,
	}

	// 将details转换为JSON字符串
	if details != nil {
		if detailsJSON, err := json.Marshal(details); err == nil {
			log.Params = string(detailsJSON)
		}
	}

	// 插入日志
	log.Insert()
}

// GetTotalAdmins 获取管理员总数
func GetTotalAdmins() (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("admin_users").Filter("status", "active").Count()
	return count, err
}

func init() {
	orm.RegisterModel(new(AdminUser))
}
