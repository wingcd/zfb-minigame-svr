package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/config"
)

func ValidateSignature(request *http.Request) (string, string, error) {
	appId := request.Header.Get("App-Id")
	userId := request.Header.Get("User-Id")
	sign := request.Header.Get("Sign")
	_, err := VerifySign(appId, userId, sign)
	if err != nil {
		return appId, userId, err
	}
	return appId, userId, err
}

// VerifySign 验证签名
func VerifySign(appId, userId, sign string) (bool, error) {
	if sign == "" {
		return false, errors.New("sign is empty")
	}

	// 直接查询数据库获取应用密钥
	o := orm.NewOrm()
	var appSecret string
	err := o.Raw("SELECT app_secret FROM application WHERE app_id = ?", appId).QueryRow(&appSecret)
	if err != nil {
		return false, errors.New("application not found")
	}

	// 构建签名字符串
	params := make(map[string]string)
	params["appId"] = appId
	if userId != "" {
		params["userId"] = userId
	}
	params["appSecret"] = appSecret

	// 按key排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var signStr strings.Builder
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
		signStr.WriteString("&")
	}

	// 去掉最后的&
	signString := strings.TrimSuffix(signStr.String(), "&")

	// 计算MD5
	expectedSign := fmt.Sprintf("%x", md5.Sum([]byte(signString)))

	ret := strings.EqualFold(sign, expectedSign)
	if !ret {
		return false, errors.New("sign is invalid")
	}
	return true, nil
}

// GenerateSign 生成签名（测试用）
func GenerateSign(appId, userId, appSecret string) string {
	params := make(map[string]string)
	params["appId"] = appId
	if userId != "" {
		params["userId"] = userId
	}
	params["appSecret"] = appSecret

	// 按key排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var signStr strings.Builder
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
		signStr.WriteString("&")
	}

	// 去掉最后的&
	signString := strings.TrimSuffix(signStr.String(), "&")

	// 计算MD5
	return fmt.Sprintf("%x", md5.Sum([]byte(signString)))
}

var (
	apiSecret string
	md5Salt   string
)

func init() {
	appconf, _ := config.NewConfig("ini", "conf/app.conf")
	apiSecret = getConfigString(appconf, "api_secret", "default_api_secret")
	md5Salt = getConfigString(appconf, "md5_salt", "default_md5_salt")
}

// getConfigString 获取配置字符串，支持默认值
func getConfigString(conf config.Configer, key, defaultValue string) string {
	if conf != nil {
		if value, _ := conf.String(key); value != "" {
			return value
		}
	}
	return defaultValue
}

// MD5 MD5加密
func MD5(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// MD5WithSalt MD5加盐加密
func MD5WithSalt(data, salt string) string {
	return MD5(data + salt)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
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

// GetAPISecret 获取API密钥
func GetAPISecret() string {
	if apiSecret == "" {
		return "default_api_secret"
	}
	return apiSecret
}

// GetMD5Salt 获取MD5盐值
func GetMD5Salt() string {
	if md5Salt == "" {
		return "default_md5_salt"
	}
	return md5Salt
}

// HashPassword 密码哈希（使用MD5加盐）
func HashPassword(password string) string {
	return MD5WithSalt(password, GetMD5Salt())
}

// GenerateUserID 生成用户ID（基于用户名和应用ID）
func GenerateUserID(username, appId string) string {
	data := fmt.Sprintf("user_%s_%s", username, appId)
	return MD5(data)
}

// GenerateOpenID 生成OpenID（基于微信授权码和应用ID）
func GenerateOpenID(code, appId string) string {
	data := fmt.Sprintf("wx_%s_%s", code, appId)
	return MD5(data)
}

// GenerateUserIDFromOpenID 基于OpenID生成用户ID
func GenerateUserIDFromOpenID(openId, appId string) string {
	data := fmt.Sprintf("wxuser_%s_%s", openId, appId)
	return MD5(data)
}

// GenerateSessionToken 生成会话token
func GenerateSessionToken(userId, appId string) string {
	randomStr := GenerateRandomString(16)
	data := fmt.Sprintf("session_%s_%s_%s", userId, appId, randomStr)
	return MD5(data)
}
