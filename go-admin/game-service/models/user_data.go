package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// User 用户模型 - 按照数据库设计.md的正确结构
type User struct {
	ID            int64     `orm:"pk;auto" json:"id"`
	AppId         string    `orm:"-" json:"appId"`                                                        // 应用ID（仅用于逻辑，不存储到数据库）
	OpenId        string    `orm:"size(100);unique;column(open_id)" json:"openId"`                        // 用户唯一标识
	PlayerId      string    `orm:"size(100);unique;column(player_id)" json:"playerId"`                    // 玩家ID（唯一，自动生成）
	Token         string    `orm:"size(255)" json:"token"`                                                // 登录Token
	Nickname      string    `orm:"size(100)" json:"nickname"`                                             // 昵称
	Avatar        string    `orm:"size(500)" json:"avatar"`                                               // 头像URL
	Data          string    `orm:"type(longtext)" json:"data"`                                            // 游戏数据（JSON格式）
	Level         int       `orm:"default(1)" json:"level"`                                               // 等级
	Exp           int64     `orm:"default(0)" json:"exp"`                                                 // 经验值
	Coin          int64     `orm:"default(0)" json:"coin"`                                                // 金币
	Diamond       int64     `orm:"default(0)" json:"diamond"`                                             // 钻石
	VipLevel      int       `orm:"default(0);column(vip_level)" json:"vipLevel"`                          // VIP等级
	Banned        bool      `orm:"default(false)" json:"banned"`                                          // 是否封禁
	BanReason     string    `orm:"size(500);column(ban_reason)" json:"banReason"`                         // 封禁原因
	BanExpire     time.Time `orm:"null;type(datetime);column(ban_expire)" json:"banExpire"`               // 封禁到期时间
	LoginCount    int       `orm:"default(0);column(login_count)" json:"loginCount"`                      // 登录次数
	LastLoginTime time.Time `orm:"null;type(datetime);column(last_login_time)" json:"lastLoginTime"`      // 最后登录时间
	LastLoginIp   string    `orm:"size(50);column(last_login_ip)" json:"lastLoginIp"`                     // 最后登录IP
	RegisterTime  time.Time `orm:"auto_now_add;type(datetime);column(register_time)" json:"registerTime"` // 注册时间
	UpdatedAt     time.Time `orm:"auto_now;type(datetime);column(updated_at)" json:"updated_at"`          // 修改时间
	CreatedAt     time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"created_at"`      // 创建时间
}

// 为了兼容旧代码，保留 UserDataEntry 结构但已废弃
// Deprecated: 请使用 User 结构
type UserDataEntry = User

// GetTableName 获取动态表名
func (u *User) GetTableName(appId string) string {
	return utils.GetUserTableName(appId)
}

// TableName 返回表名（实现TableName接口）
func (u *User) TableName() string {
	if u.AppId != "" {
		return u.GetTableName(u.AppId)
	}
	return "user_default" // 默认表名
}

// GetUserByOpenId 根据openId获取用户
func GetUserByOpenId(appId, openId string) (*User, error) {
	o := orm.NewOrm()
	user := &User{AppId: appId}

	tableName := utils.GetUserTableName(appId)

	err := o.Raw("SELECT * FROM "+tableName+" WHERE open_id = ?", openId).QueryRow(user)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil // 用户不存在
		}
		logs.Error("获取用户失败: %v", err)
		return nil, err
	}

	return user, nil
}

// GetUserByPlayerId 根据playerId获取用户
func GetUserByPlayerId(appId, playerId string) (*User, error) {
	o := orm.NewOrm()
	user := &User{AppId: appId}

	tableName := utils.GetUserTableName(appId)

	err := o.Raw("SELECT * FROM "+tableName+" WHERE player_id = ?", playerId).QueryRow(user)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil // 用户不存在
		}
		logs.Error("获取用户失败: %v", err)
		return nil, err
	}

	return user, nil
}

// CreateUser 创建新用户
func CreateUser(appId string, user *User) error {
	o := orm.NewOrm()

	tableName := utils.GetUserTableName(appId)
	user.AppId = appId

	// 生成唯一的playerId
	if user.PlayerId == "" {
		user.PlayerId = utils.GeneratePlayerId()
	}

	// 设置默认值
	if user.Level == 0 {
		user.Level = 1
	}

	// 构建插入SQL
	sql := fmt.Sprintf(`
		INSERT INTO %s (open_id, player_id, token, nickname, avatar, data, level, exp, coin, diamond, vip_level, banned, ban_reason, ban_expire, login_count, last_login_time, last_login_ip, register_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)

	_, err := o.Raw(sql,
		user.OpenId, user.PlayerId, user.Token, user.Nickname, user.Avatar, user.Data,
		user.Level, user.Exp, user.Coin, user.Diamond, user.VipLevel, user.Banned,
		user.BanReason, user.BanExpire, user.LoginCount, user.LastLoginTime, user.LastLoginIp, user.RegisterTime,
	).Exec()

	if err != nil {
		logs.Error("创建用户失败: %v", err)
		return err
	}

	logs.Info("用户创建成功: openId=%s, playerId=%s", user.OpenId, user.PlayerId)
	return nil
}

// UpdateUser 更新用户信息
func UpdateUser(appId string, user *User) error {
	o := orm.NewOrm()

	tableName := utils.GetUserTableName(appId)

	sql := fmt.Sprintf(`
		UPDATE %s SET 
			token = ?, nickname = ?, avatar = ?, data = ?, level = ?, exp = ?, coin = ?, diamond = ?, 
			vip_level = ?, banned = ?, ban_reason = ?, ban_expire = ?, login_count = ?, last_login_time = ?, 
			last_login_ip = ?, updated_at = NOW()
		WHERE player_id = ?
	`, tableName)

	_, err := o.Raw(sql,
		user.Token, user.Nickname, user.Avatar, user.Data, user.Level, user.Exp, user.Coin, user.Diamond,
		user.VipLevel, user.Banned, user.BanReason, user.BanExpire, user.LoginCount, user.LastLoginTime,
		user.LastLoginIp, user.PlayerId,
	).Exec()

	if err != nil {
		logs.Error("更新用户失败: %v", err)
		return err
	}

	return nil
}

// SaveUserData 保存用户游戏数据
func SaveUserData(appId, playerId string, data string) error {
	o := orm.NewOrm()
	tableName := utils.GetUserTableName(appId)

	sql := fmt.Sprintf("UPDATE %s SET data = ?, updated_at = NOW() WHERE player_id = ?", tableName)
	_, err := o.Raw(sql, data, playerId).Exec()

	if err != nil {
		logs.Error("保存用户数据失败: %v", err)
		return err
	}

	return nil
}

// GetUserData 获取用户游戏数据
func GetUserData(appId, playerId string) (map[string]interface{}, error) {
	user, err := GetUserByPlayerId(appId, playerId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	if user.Data == "" {
		return make(map[string]interface{}), nil
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(user.Data), &data)
	if err != nil {
		logs.Error("反序列化用户数据失败: %v", err)
		return nil, err
	}

	return data, nil
}

// UpdateLoginInfo 更新登录信息
func UpdateLoginInfo(appId, playerId, loginIp string) error {
	o := orm.NewOrm()
	tableName := utils.GetUserTableName(appId)

	sql := fmt.Sprintf(`
		UPDATE %s SET 
			login_count = login_count + 1, 
			last_login_time = NOW(), 
			last_login_ip = ?, 
			updated_at = NOW()
		WHERE player_id = ?
	`, tableName)

	_, err := o.Raw(sql, loginIp, playerId).Exec()
	if err != nil {
		logs.Error("更新登录信息失败: %v", err)
		return err
	}

	return nil
}

// GetUserInfo 获取用户基本信息
func GetUserInfo(appId, playerId string) (map[string]interface{}, error) {
	user, err := GetUserByPlayerId(appId, playerId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	userInfo := map[string]interface{}{
		"playerId":      user.PlayerId,
		"openId":        user.OpenId,
		"nickname":      user.Nickname,
		"avatar":        user.Avatar,
		"level":         user.Level,
		"exp":           user.Exp,
		"coin":          user.Coin,
		"diamond":       user.Diamond,
		"vipLevel":      user.VipLevel,
		"banned":        user.Banned,
		"banReason":     user.BanReason,
		"banExpire":     user.BanExpire,
		"loginCount":    user.LoginCount,
		"lastLoginTime": user.LastLoginTime,
		"lastLoginIp":   user.LastLoginIp,
		"registerTime":  user.RegisterTime,
		"data":          user.Data,
	}

	return userInfo, nil
}

// UpdateUserInfo 更新用户基本信息（昵称、头像）
func UpdateUserInfo(appId, playerId, nickname, avatar string) error {
	o := orm.NewOrm()
	tableName := utils.GetUserTableName(appId)

	// 构建动态更新SQL
	setParts := []string{}
	params := []interface{}{}

	if nickname != "" {
		setParts = append(setParts, "nickname = ?")
		params = append(params, nickname)
	}

	if avatar != "" {
		setParts = append(setParts, "avatar = ?")
		params = append(params, avatar)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	// 添加修改时间
	setParts = append(setParts, "updated_at = NOW()")

	// 构建SQL
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE player_id = ?", tableName, setClause)
	params = append(params, playerId)

	_, err := o.Raw(sql, params...).Exec()
	if err != nil {
		logs.Error("更新用户基本信息失败: %v", err)
		return err
	}

	return nil
}

// ClearUserGameData 清空用户的游戏数据（保留用户基本信息）
func ClearUserGameData(appId, playerId string) error {
	o := orm.NewOrm()
	tableName := utils.GetUserTableName(appId)

	sql := fmt.Sprintf("UPDATE %s SET data = '{}', updated_at = NOW() WHERE player_id = ?", tableName)
	_, err := o.Raw(sql, playerId).Exec()

	if err != nil {
		logs.Error("清空用户游戏数据失败: %v", err)
		return err
	}

	return nil
}

// 在redis中保存用户token
func SaveUserStatusToRedis(appId, playerId, token string) error {
	RedisClient.Set(context.Background(), fmt.Sprintf("user_token_%s_%s", appId, playerId), token, 0)
	return nil
}

// 在redis中获取用户token
func GetUserStatusFromRedis(appId, playerId string) (string, error) {
	token, err := RedisClient.Get(context.Background(), fmt.Sprintf("user_token_%s_%s", appId, playerId)).Result()
	return token, err
}
