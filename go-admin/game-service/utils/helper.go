package utils

import (
	"crypto/md5"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/config"
)

func ValidateSignature(request *http.Request) (string, string) {
	appId := request.Header.Get("appId")
	playerId := request.Header.Get("playerId")
	return appId, playerId
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
	cryptoRand.Read(bytes)
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

// GeneratePlayerId 生成数字格式的playerId（简化版本，不检查数据库重复）
func GeneratePlayerId() string {
	// 使用当前时间戳的后6位
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	timestampSuffix := timestamp[len(timestamp)-6:]

	// 生成3位随机数
	randomNum, _ := cryptoRand.Int(cryptoRand.Reader, big.NewInt(1000))
	randomStr := fmt.Sprintf("%03d", randomNum)

	// 组合：6 + 时间戳后6位 + 3位随机数
	return fmt.Sprintf("6%s%s", timestampSuffix, randomStr)
}

// GenerateOpenID 生成OpenID（基于微信授权码和应用ID）
func GenerateOpenID(code, appId string) string {
	data := fmt.Sprintf("wx_%s_%s", code, appId)
	return MD5(data)
}

// GenerateSessionToken 生成会话token
func GenerateSessionToken(userId, appId string) string {
	randomStr := GenerateRandomString(16)
	data := fmt.Sprintf("session_%s_%s_%s", userId, appId, randomStr)
	return MD5(data)
}

// GenerateUUID 生成简单的UUID（实际上是MD5哈希）
func GenerateUUID() string {
	randomStr := GenerateRandomString(32)
	return MD5(fmt.Sprintf("uuid_%d_%s", time.Now().UnixNano(), randomStr))
}
