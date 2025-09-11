package models

import (
	"fmt"
	"time"
)

// LeaderboardConfig 排行榜配置结构（管理表）
type LeaderboardConfig struct {
	BaseModel
	AppId            string    `orm:"size(100)" json:"appId"`
	LeaderboardType  string    `orm:"size(100);column(leaderboard_type)" json:"leaderboardType"`
	Name             string    `orm:"size(200)" json:"name"`
	Description      string    `orm:"type(text)" json:"description"`
	ScoreType        string    `orm:"size(20);default(higher_better)" json:"scoreType"` // higher_better, lower_better
	MaxRank          int       `orm:"default(1000);column(max_rank)" json:"maxRank"`
	Enabled          bool      `orm:"default(true)" json:"enabled"`
	Category         string    `orm:"size(100)" json:"category"`
	ResetType        string    `orm:"size(50);default(permanent)" json:"resetType"` // permanent, daily, weekly, monthly, custom
	ResetValue       int       `orm:"default(0);column(reset_value)" json:"resetValue"`
	ResetTime        time.Time `orm:"null;type(datetime);column(reset_time)" json:"resetTime"`
	UpdateStrategy   int       `orm:"default(0);column(update_strategy)" json:"updateStrategy"` // 0=最高分, 1=最新分, 2=累计分
	Sort             int       `orm:"default(1)" json:"sort"`                                   // 0=升序, 1=降序
	ScoreCount       int       `orm:"default(0);column(score_count)" json:"scoreCount"`
	ParticipantCount int       `orm:"default(0);column(participant_count)" json:"participantCount"`
	LastResetTime    time.Time `orm:"null;type(datetime);column(last_reset_time)" json:"lastResetTime"`
	CreatedBy        string    `orm:"size(100);column(created_by)" json:"createdBy"`
}

// Leaderboard 排行榜数据结构（动态表）
type Leaderboard struct {
	Id        int64                  `orm:"auto" json:"id"`
	Type      string                 `orm:"size(50)" json:"type"`
	UserId    string                 `orm:"size(100);column(player_id)" json:"user_id"`
	Score     int64                  `orm:"default(0)" json:"score"`
	ExtraData string                 `orm:"type(text);column(extra_data)" json:"extra_data"`
	UserInfo  map[string]interface{} `orm:"-" json:"userInfo,omitempty"` // 用户信息，不存储到数据库
	CreatedAt string                 `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt string                 `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// TableName 返回配置表名
func (l *LeaderboardConfig) TableName() string {
	return "leaderboard_config"
}

// TableName 获取动态表名（需要appId参数）
func (l *Leaderboard) TableName() string {
	return "leaderboard_dynamic" // 基础表名，实际使用时需要添加appId后缀
}

// GetTableName 获取动态表名
func (l *Leaderboard) GetTableName(appId string) string {
	// 清理appId，确保表名安全
	cleanAppId := appId
	// 这里可以添加更多的清理逻辑
	return fmt.Sprintf("leaderboard_%s", cleanAppId)
}

// Register 注册模型
func (l *LeaderboardConfig) Register() error {
	return RegisterModel(new(LeaderboardConfig))
}

// RegisterWithSuffix 带后缀注册模型
func (l *LeaderboardConfig) RegisterWithSuffix(suffix string) error {
	return RegisterModelWithSuffix(new(LeaderboardConfig), suffix)
}

// Init 初始化模型
func (l *LeaderboardConfig) Init() error {
	return l.Register()
}

// Register 注册模型
func (l *Leaderboard) Register() error {
	return RegisterModel(new(Leaderboard))
}

// RegisterWithSuffix 带后缀注册模型
func (l *Leaderboard) RegisterWithSuffix(suffix string) error {
	return RegisterModelWithSuffix(new(Leaderboard), suffix)
}

// Init 初始化模型
func (l *Leaderboard) Init() error {
	return l.Register()
}

// LeaderboardQuery 排行榜查询结构
type LeaderboardQuery struct {
	AppId    string `json:"app_id"`
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	MinScore int64  `json:"min_score"`
	MaxScore int64  `json:"max_score"`
	Level    int    `json:"level"`
	Status   int    `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	OrderBy  string `json:"order_by"` // score_desc, score_asc, time_desc, time_asc
}

// LeaderboardResponse 排行榜响应结构
type LeaderboardResponse struct {
	List     []Leaderboard `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// PlayerRank 玩家排名信息
type PlayerRank struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Score    int64  `json:"score"`
	Rank     int    `json:"rank"`
	Level    int    `json:"level"`
}
