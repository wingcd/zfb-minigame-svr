package services

import (
	"admin-service/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/go-redis/redis/v8"
)

// LeaderboardRedisService Redis排行榜服务
type LeaderboardRedisService struct {
	client *redis.Client
	ctx    context.Context
}

// NewLeaderboardRedisService 创建Redis排行榜服务实例
func NewLeaderboardRedisService() *LeaderboardRedisService {
	return &LeaderboardRedisService{
		client: models.RedisClient,
		ctx:    context.Background(),
	}
}

// PlayerData 玩家数据结构（用于Redis存储）
type PlayerData struct {
	PlayerId  string                 `json:"playerId"`
	Score     int64                  `json:"score"`
	ExtraData map[string]interface{} `json:"extraData"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank      int         `json:"rank"`
	PlayerId  string      `json:"playerId"`
	Score     int64       `json:"score"`
	ExtraData interface{} `json:"extraData"`
	CreatedAt interface{} `json:"createdAt"`
	UpdatedAt interface{} `json:"updatedAt"`
	// 玩家详细信息字段（保持与MySQL版本一致）
	Token    interface{} `json:"token"`
	Nickname interface{} `json:"nickname"`
	Avatar   interface{} `json:"avatar"`
	Data     interface{} `json:"data"`
	Level    interface{} `json:"level"`
	Exp      interface{} `json:"exp"`
	Coin     interface{} `json:"coin"`
	Diamond  interface{} `json:"diamond"`
	VipLevel interface{} `json:"vipLevel"`
}

// getLeaderboardKey 获取排行榜Redis键名
func (s *LeaderboardRedisService) getLeaderboardKey(appId, leaderboardType string) string {
	return fmt.Sprintf("leaderboard:%s:%s", appId, leaderboardType)
}

// getPlayerDataKey 获取玩家数据Redis键名
func (s *LeaderboardRedisService) getPlayerDataKey(appId, leaderboardType, playerId string) string {
	return fmt.Sprintf("leaderboard_data:%s:%s:%s", appId, leaderboardType, playerId)
}

// getLeaderboardConfigKey 获取排行榜配置Redis键名
func (s *LeaderboardRedisService) getLeaderboardConfigKey(appId, leaderboardType string) string {
	return fmt.Sprintf("leaderboard_config:%s:%s", appId, leaderboardType)
}

// UpdateScore 更新玩家分数
func (s *LeaderboardRedisService) UpdateScore(appId, leaderboardType, playerId string, score int64, extraData map[string]interface{}) error {
	// 获取排行榜配置
	config, err := s.getLeaderboardConfig(appId, leaderboardType)
	if err != nil {
		return fmt.Errorf("获取排行榜配置失败: %v", err)
	}

	if !config.Enabled {
		return fmt.Errorf("排行榜已禁用")
	}

	leaderboardKey := s.getLeaderboardKey(appId, leaderboardType)
	playerDataKey := s.getPlayerDataKey(appId, leaderboardType, playerId)

	// 开始事务
	pipe := s.client.TxPipeline()

	// 检查是否需要根据更新策略处理分数
	currentScore, err := s.client.ZScore(s.ctx, leaderboardKey, playerId).Result()
	shouldUpdate := false

	if err == redis.Nil {
		// 玩家不存在，直接添加
		shouldUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取当前分数失败: %v", err)
	} else {
		// 根据更新策略决定是否更新
		switch config.UpdateStrategy {
		case 0: // 最高分
			if config.ScoreType == "higher_better" && score > int64(currentScore) {
				shouldUpdate = true
			} else if config.ScoreType == "lower_better" && score < int64(currentScore) {
				shouldUpdate = true
			}
		case 1: // 最新分
			shouldUpdate = true
		case 2: // 累计分
			score = score + int64(currentScore)
			shouldUpdate = true
		}
	}

	if shouldUpdate {
		// 更新排行榜分数
		pipe.ZAdd(s.ctx, leaderboardKey, &redis.Z{
			Score:  float64(score),
			Member: playerId,
		})

		// 存储玩家详细数据
		now := time.Now()
		playerData := PlayerData{
			PlayerId:  playerId,
			Score:     score,
			ExtraData: extraData,
			CreatedAt: now, // 新数据设置为当前时间
			UpdatedAt: now,
		}

		// 如果是更新，保留原创建时间
		existingData := s.getPlayerData(appId, leaderboardType, playerId)
		if existingData != nil && !existingData.CreatedAt.IsZero() {
			playerData.CreatedAt = existingData.CreatedAt
		}

		dataJson, _ := json.Marshal(playerData)
		pipe.Set(s.ctx, playerDataKey, dataJson, time.Hour*24*7) // 7天过期

		// 限制排行榜大小
		if config.MaxRank > 0 {
			// 保留前MaxRank名，删除其余的
			pipe.ZRemRangeByRank(s.ctx, leaderboardKey, 0, -(int64(config.MaxRank) + 1))
		}

		// 执行事务
		_, err = pipe.Exec(s.ctx)
		if err != nil {
			return fmt.Errorf("更新Redis排行榜失败: %v", err)
		}

		// 异步同步到MySQL
		go s.syncToMySQL(appId, leaderboardType, playerId, score, extraData)
	}

	return nil
}

// GetLeaderboard 获取排行榜数据
func (s *LeaderboardRedisService) GetLeaderboard(appId, leaderboardType string, start, stop int64) ([]*LeaderboardEntry, error) {
	leaderboardKey := s.getLeaderboardKey(appId, leaderboardType)

	// 获取排行榜配置
	config, err := s.getLeaderboardConfig(appId, leaderboardType)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜配置失败: %v", err)
	}

	// 根据排序方式获取数据
	var results []redis.Z
	if config.Sort == 1 { // 降序
		results, err = s.client.ZRevRangeWithScores(s.ctx, leaderboardKey, start, stop).Result()
	} else { // 升序
		results, err = s.client.ZRangeWithScores(s.ctx, leaderboardKey, start, stop).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("获取排行榜数据失败: %v", err)
	}

	entries := make([]*LeaderboardEntry, len(results))
	for i, result := range results {
		playerId := result.Member.(string)
		score := int64(result.Score)

		// 获取玩家详细数据
		playerData := s.getPlayerData(appId, leaderboardType, playerId)

		// 获取用户详细信息
		userInfo := s.getUserInfo(appId, playerId)

		entries[i] = &LeaderboardEntry{
			Rank:      int(start) + i + 1,
			PlayerId:  playerId,
			Score:     score,
			ExtraData: playerData.ExtraData,
			CreatedAt: playerData.CreatedAt,
			UpdatedAt: playerData.UpdatedAt,
			// 玩家详细信息
			Token:    userInfo["token"],
			Nickname: userInfo["nickname"],
			Avatar:   userInfo["avatar"],
			Data:     userInfo["data"],
			Level:    userInfo["level"],
			Exp:      userInfo["exp"],
			Coin:     userInfo["coin"],
			Diamond:  userInfo["diamond"],
			VipLevel: userInfo["vipLevel"],
		}
	}

	return entries, nil
}

// GetPlayerRank 获取玩家排名和分数
func (s *LeaderboardRedisService) GetPlayerRank(appId, leaderboardType, playerId string) (score int64, rank int, err error) {
	leaderboardKey := s.getLeaderboardKey(appId, leaderboardType)

	// 获取分数
	scoreFloat, err := s.client.ZScore(s.ctx, leaderboardKey, playerId).Result()
	if err == redis.Nil {
		return 0, 0, nil // 玩家不在排行榜中
	} else if err != nil {
		return 0, 0, fmt.Errorf("获取玩家分数失败: %v", err)
	}

	score = int64(scoreFloat)

	// 获取排名
	config, err := s.getLeaderboardConfig(appId, leaderboardType)
	if err != nil {
		return score, 0, fmt.Errorf("获取排行榜配置失败: %v", err)
	}

	var rankResult int64
	if config.Sort == 1 { // 降序
		rankResult, err = s.client.ZRevRank(s.ctx, leaderboardKey, playerId).Result()
	} else { // 升序
		rankResult, err = s.client.ZRank(s.ctx, leaderboardKey, playerId).Result()
	}

	if err != nil {
		return score, 0, fmt.Errorf("获取玩家排名失败: %v", err)
	}

	rank = int(rankResult) + 1 // Redis排名从0开始，转换为从1开始
	return score, rank, nil
}

// RemovePlayer 移除玩家
func (s *LeaderboardRedisService) RemovePlayer(appId, leaderboardType, playerId string) error {
	leaderboardKey := s.getLeaderboardKey(appId, leaderboardType)
	playerDataKey := s.getPlayerDataKey(appId, leaderboardType, playerId)

	// 开始事务
	pipe := s.client.TxPipeline()
	pipe.ZRem(s.ctx, leaderboardKey, playerId)
	pipe.Del(s.ctx, playerDataKey)

	_, err := pipe.Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("移除玩家失败: %v", err)
	}

	// 异步从MySQL删除
	go s.removeFromMySQL(appId, leaderboardType, playerId)

	return nil
}

// GetLeaderboardSize 获取排行榜大小
func (s *LeaderboardRedisService) GetLeaderboardSize(appId, leaderboardType string) (int64, error) {
	leaderboardKey := s.getLeaderboardKey(appId, leaderboardType)
	return s.client.ZCard(s.ctx, leaderboardKey).Result()
}

// ClearLeaderboard 清空排行榜
func (s *LeaderboardRedisService) ClearLeaderboard(appId, leaderboardType string) error {
	leaderboardKey := s.getLeaderboardKey(appId, leaderboardType)

	// 获取所有玩家ID用于清除详细数据
	playerIds, err := s.client.ZRange(s.ctx, leaderboardKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("获取玩家列表失败: %v", err)
	}

	// 开始事务
	pipe := s.client.TxPipeline()
	pipe.Del(s.ctx, leaderboardKey)

	// 删除所有玩家详细数据
	for _, playerId := range playerIds {
		playerDataKey := s.getPlayerDataKey(appId, leaderboardType, playerId)
		pipe.Del(s.ctx, playerDataKey)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("清空排行榜失败: %v", err)
	}

	return nil
}

// getLeaderboardConfig 获取排行榜配置
func (s *LeaderboardRedisService) getLeaderboardConfig(appId, leaderboardType string) (*models.LeaderboardConfig, error) {
	configKey := s.getLeaderboardConfigKey(appId, leaderboardType)

	// 先从Redis缓存获取
	configJson, err := s.client.Get(s.ctx, configKey).Result()
	if err == nil {
		var config models.LeaderboardConfig
		if json.Unmarshal([]byte(configJson), &config) == nil {
			return &config, nil
		}
	}

	// 从数据库获取并缓存
	config, err := s.getConfigFromDB(appId, leaderboardType)
	if err != nil {
		return nil, err
	}

	// 缓存到Redis
	if configData, err := json.Marshal(config); err == nil {
		s.client.Set(s.ctx, configKey, configData, time.Minute*10) // 10分钟缓存
	}

	return config, nil
}

// getConfigFromDB 从数据库获取配置
func (s *LeaderboardRedisService) getConfigFromDB(appId, leaderboardType string) (*models.LeaderboardConfig, error) {
	// 这里需要调用models包的函数来获取配置
	// 为了避免循环依赖，我们直接使用ORM查询
	var config models.LeaderboardConfig
	err := models.GetOrm().QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("leaderboard_type", leaderboardType).
		One(&config)

	if err != nil {
		return nil, fmt.Errorf("排行榜配置不存在")
	}

	return &config, nil
}

// getPlayerData 获取玩家详细数据
func (s *LeaderboardRedisService) getPlayerData(appId, leaderboardType, playerId string) *PlayerData {
	playerDataKey := s.getPlayerDataKey(appId, leaderboardType, playerId)

	dataJson, err := s.client.Get(s.ctx, playerDataKey).Result()
	if err != nil {
		// 返回默认数据
		return &PlayerData{
			PlayerId:  playerId,
			ExtraData: make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	var data PlayerData
	if json.Unmarshal([]byte(dataJson), &data) != nil {
		return &PlayerData{
			PlayerId:  playerId,
			ExtraData: make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return &data
}

// getUserInfo 获取用户详细信息（从MySQL用户表）
func (s *LeaderboardRedisService) getUserInfo(appId, playerId string) map[string]interface{} {
	// 默认用户信息
	defaultInfo := map[string]interface{}{
		"token":    "",
		"nickname": "",
		"avatar":   "",
		"data":     map[string]interface{}{},
		"level":    0,
		"exp":      0,
		"coin":     0,
		"diamond":  0,
		"vipLevel": 0,
	}

	// 从用户表中获取玩家数据
	userTableName := fmt.Sprintf("user_%s", appId)
	o := orm.NewOrm()

	var userData []orm.Params
	userSQL := fmt.Sprintf("SELECT data FROM %s WHERE player_id = ?", userTableName)
	_, err := o.Raw(userSQL, playerId).Values(&userData)

	if err != nil || len(userData) == 0 {
		return defaultInfo
	}

	if dataStr, ok := userData[0]["data"].(string); ok && dataStr != "" {
		var parsedData map[string]interface{}
		if json.Unmarshal([]byte(dataStr), &parsedData) == nil {
			// 解析并设置玩家信息字段
			if token, exists := parsedData["token"]; exists {
				defaultInfo["token"] = token
			}
			if nickname, exists := parsedData["nickname"]; exists {
				defaultInfo["nickname"] = nickname
			}
			if avatar, exists := parsedData["avatar"]; exists {
				defaultInfo["avatar"] = avatar
			}
			if level, exists := parsedData["level"]; exists {
				defaultInfo["level"] = level
			}
			if exp, exists := parsedData["exp"]; exists {
				defaultInfo["exp"] = exp
			}
			if coin, exists := parsedData["coin"]; exists {
				defaultInfo["coin"] = coin
			}
			if diamond, exists := parsedData["diamond"]; exists {
				defaultInfo["diamond"] = diamond
			}
			if vipLevel, exists := parsedData["vipLevel"]; exists {
				defaultInfo["vipLevel"] = vipLevel
			}
			// 保存完整的游戏数据
			defaultInfo["data"] = parsedData
		}
	}

	return defaultInfo
}

// syncToMySQL 异步同步到MySQL
func (s *LeaderboardRedisService) syncToMySQL(appId, leaderboardType, playerId string, score int64, extraData map[string]interface{}) {
	// 这里实现异步同步到MySQL的逻辑
	// 可以使用消息队列或者定时同步策略
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 记录错误日志
				fmt.Printf("同步到MySQL失败: %v\n", r)
			}
		}()

		// 调用原有的MySQL更新函数
		extraDataJson, _ := json.Marshal(extraData)
		models.UpdateLeaderboardScoreWithExtra(appId, leaderboardType, playerId, score, string(extraDataJson))
	}()
}

// removeFromMySQL 异步从MySQL删除
func (s *LeaderboardRedisService) removeFromMySQL(appId, leaderboardType, playerId string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("从MySQL删除失败: %v\n", r)
			}
		}()

		models.DeleteLeaderboardScore(appId, leaderboardType, playerId)
	}()
}

// SyncFromMySQL 从MySQL同步数据到Redis（用于冷启动）
func (s *LeaderboardRedisService) SyncFromMySQL(appId, leaderboardType string) error {
	// 清空Redis中的排行榜
	s.ClearLeaderboard(appId, leaderboardType)

	// 从MySQL获取数据
	data, _, err := models.GetLeaderboardData(appId, leaderboardType, 1, 10000) // 获取前10000名
	if err != nil {
		return fmt.Errorf("从MySQL获取数据失败: %v", err)
	}

	// 批量写入Redis
	leaderboardKey := s.getLeaderboardKey(appId, leaderboardType)
	pipe := s.client.TxPipeline()

	for _, item := range data {
		playerId := item["playerId"].(string)
		score, _ := strconv.ParseInt(fmt.Sprintf("%v", item["score"]), 10, 64)

		// 添加到排行榜
		pipe.ZAdd(s.ctx, leaderboardKey, &redis.Z{
			Score:  float64(score),
			Member: playerId,
		})

		// 存储玩家详细数据
		playerDataKey := s.getPlayerDataKey(appId, leaderboardType, playerId)
		playerData := map[string]interface{}{
			"playerId":  playerId,
			"score":     score,
			"extraData": item["extraData"],
			"updatedAt": time.Now(),
		}
		dataJson, _ := json.Marshal(playerData)
		pipe.Set(s.ctx, playerDataKey, dataJson, time.Hour*24*7)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("同步到Redis失败: %v", err)
	}

	return nil
}
