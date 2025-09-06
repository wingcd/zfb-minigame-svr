package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret    string
	passwordSalt string
	apiSecret    string
)

// FindConfigFile 查找配置文件
func FindConfigFile() string {
	configPaths := []string{
		"conf/app.conf",       // 从项目根目录运行
		"../conf/app.conf",    // 从tests目录运行
		"../../conf/app.conf", // 从更深层目录运行
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "conf/app.conf" // 默认路径
}

func init() {
	configPath := FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)

	// 如果找不到配置文件，使用默认值
	if err == nil && appconf != nil {
		jwtSecret, _ = appconf.String("jwt_secret")
		passwordSalt, _ = appconf.String("password_salt")
		apiSecret, _ = appconf.String("api_secret")
	}
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	RoleID   int64  `json:"role_id"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// AdminInfo 管理员信息
type AdminInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RoleID   int64  `json:"roleId"`
}

// GenerateJWT 生成JWT令牌
func GenerateJWT(userID int64, username string, roleID int64) (string, error) {
	expireTime := time.Now().Add(24 * time.Hour)

	// 根据roleID设置role字符串
	role := "user"
	if roleID == 1 {
		role = "super_admin"
	} else if roleID == 2 {
		role = "admin"
	}

	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "admin-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ParseJWT 解析JWT令牌
func ParseJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashPassword 密码加密
func HashPassword(password string) string {
	// 使用MD5加密（为了兼容性）
	return MD5HashWithSalt(password, passwordSalt)
}

// HashPasswordBcrypt 使用bcrypt加密密码
func HashPasswordBcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password+passwordSalt), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+passwordSalt))
	return err == nil
}

// MD5Hash MD5加密
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// MD5HashWithSalt MD5加盐加密
func MD5HashWithSalt(text, salt string) string {
	return MD5Hash(text + salt)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// GenerateAppSecret 生成应用密钥
func GenerateAppSecret() string {
	return GenerateRandomString(32)
}

// GenerateAppId 生成应用ID
func GenerateAppId() string {
	// 生成格式为 app_xxxxxxxx 的应用ID
	return "app_" + GenerateRandomString(8)
}

// ValidateAPISign 验证API签名
func ValidateAPISign(params map[string]string, timestamp, sign string) bool {
	// 构建签名字符串
	signStr := ""
	for key, value := range params {
		if key != "sign" && key != "timestamp" {
			signStr += key + "=" + value + "&"
		}
	}
	signStr += "timestamp=" + timestamp + "&"
	signStr += "secret=" + apiSecret

	// 计算MD5
	expectedSign := MD5Hash(signStr)

	return expectedSign == sign
}

// GenerateAPISign 生成API签名
func GenerateAPISign(params map[string]string, timestamp string) string {
	// 构建签名字符串
	signStr := ""
	for key, value := range params {
		if key != "sign" && key != "timestamp" {
			signStr += key + "=" + value + "&"
		}
	}
	signStr += "timestamp=" + timestamp + "&"
	signStr += "secret=" + apiSecret

	return MD5Hash(signStr)
}

// ValidateJWT 验证JWT令牌并返回claims
func ValidateJWT(ctx *context.Context) *JWTClaims {
	// 从Header中获取Token
	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		ctx.Output.Header("Content-Type", "application/json")
		ctx.Output.SetStatus(200)
		response := NewErrorResponse(CodeUnauthorized, "未登录")
		ctx.Output.JSON(response, false, false)
		return nil
	}

	// 检查Bearer前缀
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.Output.Header("Content-Type", "application/json")
		ctx.Output.SetStatus(200)
		response := NewErrorResponse(CodeUnauthorized, "Token格式错误")
		ctx.Output.JSON(response, false, false)
		return nil
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// 解析Token
	claims, err := ParseJWT(tokenString)
	if err != nil {
		ctx.Output.Header("Content-Type", "application/json")
		ctx.Output.SetStatus(200)
		response := NewErrorResponse(CodeUnauthorized, "Token无效: "+err.Error())
		ctx.Output.JSON(response, false, false)
		return nil
	}

	// 检查Token是否过期
	if claims.ExpiresAt < time.Now().Unix() {
		ctx.Output.Header("Content-Type", "application/json")
		ctx.Output.SetStatus(200)
		response := NewErrorResponse(CodeUnauthorized, "Token已过期")
		ctx.Output.JSON(response, false, false)
		return nil
	}

	return claims
}

// GetJWTUserInfo 获取JWT用户信息
func GetJWTUserInfo(ctx *context.Context) *AdminInfo {
	claims := ValidateJWT(ctx)
	if claims == nil {
		return nil
	}

	return &AdminInfo{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		RoleID:   claims.RoleID,
	}
}

// GetJWTSecret 获取JWT密钥
func GetJWTSecret() string {
	if jwtSecret == "" {
		return "default_jwt_secret_key"
	}
	return jwtSecret
}

// SetJWTSecret 设置JWT密钥（用于测试）
func SetJWTSecret(secret string) {
	jwtSecret = secret
}

// LogOperation 记录操作日志
func LogOperation(adminId int64, action, description string) {
	// 这里可以实现日志记录逻辑
	// 暂时留空或简单打印
}

// GetPasswordSalt 获取密码盐值
func GetPasswordSalt() string {
	if passwordSalt == "" {
		return "default_password_salt"
	}
	return passwordSalt
}

// GetAPISecret 获取API密钥
func GetAPISecret() string {
	if apiSecret == "" {
		return "default_api_secret_key"
	}
	return apiSecret
}

// CleanAppId 清理应用ID中的特殊字符
func CleanAppId(appId string) string {
	// 替换特殊字符为下划线
	cleaned := strings.ReplaceAll(appId, "-", "_")
	cleaned = strings.ReplaceAll(cleaned, ".", "_")
	cleaned = strings.ReplaceAll(cleaned, " ", "_")
	return cleaned
}

// BuildSignString 构建签名字符串
func BuildSignString(params map[string]interface{}, timestamp int64) string {
	// 按参数名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}

	var signParts []string
	for _, k := range keys {
		v := params[k]
		if v != nil {
			signParts = append(signParts, fmt.Sprintf("%s=%v", k, v))
		}
	}

	// 构建最终签名字符串
	signStr := strings.Join(signParts, "&")
	if signStr != "" {
		signStr += "&"
	}
	signStr += fmt.Sprintf("timestamp=%d&key=%s", timestamp, GetAPISecret())

	return signStr
}

// BuildPlaceholders 构建SQL占位符
func BuildPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}

	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}
