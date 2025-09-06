package services

import (
	"admin-service/models"
	"admin-service/utils"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{}
}

// 登录逻辑已移至 models 层，AuthService 保留核心认证功能

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userId int64) (*models.AdminUser, error) {
	o := orm.NewOrm()

	user := &models.AdminUser{}
	err := o.QueryTable("admin_users").Filter("id", userId).Filter("status", 1).One(user)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("用户不存在或已被禁用")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return user, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userId int64, oldPassword, newPassword string) error {
	o := orm.NewOrm()

	// 获取用户信息
	user := &models.AdminUser{}
	err := o.QueryTable("admin_users").Filter("id", userId).One(user)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	hashedOldPassword := utils.HashPassword(oldPassword)
	if user.Password != hashedOldPassword {
		return fmt.Errorf("原密码错误")
	}

	// 更新新密码
	user.Password = utils.HashPassword(newPassword)
	user.UpdatedAt = time.Now()

	_, err = o.Update(user, "password", "updateTime")
	if err != nil {
		return fmt.Errorf("修改密码失败: %v", err)
	}

	return nil
}

// CreateInitialAdmin 创建初始管理员
func (s *AuthService) CreateInitialAdmin() error {
	o := orm.NewOrm()

	// 检查是否已有管理员
	count, err := o.QueryTable("admin_users").Count()
	if err != nil {
		return fmt.Errorf("检查管理员数量失败: %v", err)
	}

	if count > 0 {
		return nil // 已有管理员，不需要创建
	}

	// 创建默认管理员
	admin := &models.AdminUser{
		Username: "admin",
		Password: utils.HashPassword("admin123"),
		RealName: "超级管理员",
		Status:   1,
	}
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = admin.CreatedAt

	_, err = o.Insert(admin)
	if err != nil {
		return fmt.Errorf("创建初始管理员失败: %v", err)
	}

	return nil
}

// GenerateAPISign 生成API签名
func (s *AuthService) GenerateAPISign(params map[string]interface{}, timestamp int64) string {
	// 按参数名排序并构建签名字符串
	signStr := utils.BuildSignString(params, timestamp)

	// MD5签名
	h := md5.New()
	h.Write([]byte(signStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ValidateAPISign 验证API签名
func (s *AuthService) ValidateAPISign(params map[string]interface{}, timestamp int64, sign string) error {
	// 检查时间戳是否在有效期内（5分钟）
	if time.Now().Unix()-timestamp > 300 {
		return fmt.Errorf("请求时间戳过期")
	}

	// 生成期望的签名
	expectedSign := s.GenerateAPISign(params, timestamp)

	// 比较签名
	if sign != expectedSign {
		return fmt.Errorf("签名验证失败")
	}

	return nil
}
