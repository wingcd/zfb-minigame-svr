package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// YallaCrypto Yalla加密工具
type YallaCrypto struct {
	SecretKey string
}

// NewYallaCrypto 创建加密工具实例
func NewYallaCrypto(secretKey string) *YallaCrypto {
	return &YallaCrypto{
		SecretKey: secretKey,
	}
}

// GenerateSign 生成签名
func (c *YallaCrypto) GenerateSign(params map[string]interface{}, timestamp int64) string {
	// 添加时间戳
	params["timestamp"] = timestamp

	// 参数排序
	var keys []string
	for k := range params {
		if k != "sign" { // 排除sign参数
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接参数字符串
	var paramStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			paramStr.WriteString("&")
		}
		paramStr.WriteString(fmt.Sprintf("%s=%v", k, params[k]))
	}

	// 添加密钥
	signStr := paramStr.String() + "&key=" + c.SecretKey

	// MD5加密
	hash := md5.Sum([]byte(signStr))
	return fmt.Sprintf("%x", hash)
}

// VerifySign 验证签名
func (c *YallaCrypto) VerifySign(params map[string]interface{}, sign string, timestamp int64) bool {
	// 检查时间戳是否在有效期内（5分钟）
	now := time.Now().Unix()
	if now-timestamp > 300 || timestamp-now > 300 {
		return false
	}

	expectedSign := c.GenerateSign(params, timestamp)
	return strings.EqualFold(expectedSign, sign)
}

// GenerateToken 生成访问令牌
func (c *YallaCrypto) GenerateToken(userID string, timestamp int64) string {
	data := fmt.Sprintf("%s_%d_%s", userID, timestamp, c.SecretKey)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// VerifyToken 验证访问令牌
func (c *YallaCrypto) VerifyToken(userID string, token string, timestamp int64) bool {
	expectedToken := c.GenerateToken(userID, timestamp)
	return expectedToken == token
}

// GenerateRequestID 生成请求ID
func (c *YallaCrypto) GenerateRequestID() string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%d_%s", timestamp, c.SecretKey)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}

// ParamsToMap 将结构体转换为map（用于签名计算）
func ParamsToMap(params interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 这里可以使用反射来转换结构体到map
	// 为了简化，这里提供一个基础实现
	switch v := params.(type) {
	case map[string]interface{}:
		return v
	case map[string]string:
		for k, val := range v {
			result[k] = val
		}
	}

	return result
}

// ConvertToString 将interface{}转换为字符串
func ConvertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}
