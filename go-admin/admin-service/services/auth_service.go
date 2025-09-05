package services

import (
	"admin-service/models"
	"admin-service/utils"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/dgrijalva/jwt-go"
)

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Login 管理员登录
func (s *AuthService) Login(username, password string) (*models.AdminUser, string, error) {
	o := orm.NewOrm()

	// 查找用户
	user := &models.AdminUser{}
	err := o.QueryTable("admin_users").Filter("username", username).Filter("status", 1).One(user)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, "", fmt.Errorf("用户不存在或已被禁用")
		}
		return nil, "", fmt.Errorf("查询用户失败: %v", err)
	}

	// 验证密码
	hashedPassword := utils.HashPassword(password)
	if user.Password != hashedPassword {
		return nil, "", fmt.Errorf("用户名或密码错误")
	}

	// 生成JWT Token
	token, err := s.GenerateToken(user.Id, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %v", err)
	}

	// 更新登录时间和IP
	user.LastLoginAt = time.Now()
	_, err = o.Update(user, "last_login_at")
	if err != nil {
		// 登录时间更新失败不影响登录流程，只记录日志
		fmt.Printf("更新用户登录时间失败: %v\n", err)
	}

	return user, token, nil
}

// GenerateToken 生成JWT Token
func (s *AuthService) GenerateToken(userId int64, username string) (string, error) {
	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userId,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
		"iat":      time.Now().Unix(),
	})

	// 签名token
	jwtSecret := utils.GetJWTSecret()
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT Token
func (s *AuthService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	// 解析token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(utils.GetJWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	// 检查token是否有效
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 检查是否过期
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				return nil, fmt.Errorf("token已过期")
			}
		}
		return &claims, nil
	}

	return nil, fmt.Errorf("无效的token")
}

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
	user.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	_, err = o.Update(user, "password", "updated_at")
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
	admin.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
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
