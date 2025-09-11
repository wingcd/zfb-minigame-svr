package models

import (
	"time"
)

// DashboardStats 仪表板统计数据
type DashboardStats struct {
	TotalUsers        int64 `json:"totalUsers"`
	TotalApplications int64 `json:"totalApplications"`
	TotalAdmins       int64 `json:"totalAdmins"`
	TodayOperations   int64 `json:"todayOperations"`
}

// TopApp 热门应用
type TopApp struct {
	AppId       string `json:"appId"`
	AppName     string `json:"appName"`
	UserCount   int64  `json:"userCount"`
	AccessCount int64  `json:"accessCount"`
}

// UserActivity 用户活动数据
type UserActivity struct {
	Date       string `json:"date"`
	UserCount  int64  `json:"userCount"`
	LoginCount int64  `json:"loginCount"`
}

// DataTrend 数据趋势
type DataTrend struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
	Type  string `json:"type"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Version      string `json:"version"`
	StartTime    string `json:"startTime"`
	DatabaseSize string `json:"databaseSize"`
	CacheStatus  string `json:"cacheStatus"`
}

// RecentActivity 最近活动
type RecentActivity struct {
	Id          int64     `json:"id"`
	UserId      int64     `json:"userId"`
	Username    string    `json:"username"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// UserGrowth 用户增长数据
type UserGrowth struct {
	Date       string `json:"date"`
	NewUsers   int64  `json:"newUsers"`
	TotalUsers int64  `json:"totalUsers"`
}

// PlatformDistribution 平台分布数据
type PlatformDistribution struct {
	Platform string `json:"platform"`
	Name     string `json:"name"`
	Count    int64  `json:"count"`
	Value    int64  `json:"value"` // 用于图表显示
}

// Statistics 统计模型（基础数据模型）
type Statistics struct {
	BaseModel
	AppId       string    `orm:"size(50)" json:"appId"`
	StatType    string    `orm:"size(50)" json:"statType"`        // daily, weekly, monthly, total
	StatKey     string    `orm:"size(100)" json:"statKey"`        // 统计键名
	StatValue   int64     `orm:"default(0)" json:"statValue"`     // 统计值
	StatData    string    `orm:"type(text);null" json:"statData"` // 详细数据JSON
	StatDate    time.Time `orm:"type(date)" json:"statDate"`      // 统计日期
	Description string    `orm:"size(255);null" json:"description"`
}

// TableName 返回表名
func (s *Statistics) TableName() string {
	return "statistics"
}

// Register 注册模型
func (s *Statistics) Register() error {
	return RegisterModel(new(Statistics))
}

// RegisterWithSuffix 带后缀注册模型
func (s *Statistics) RegisterWithSuffix(suffix string) error {
	return RegisterModelWithSuffix(new(Statistics), suffix)
}

// Init 初始化模型
func (s *Statistics) Init() error {
	return s.Register()
}

// GameSession 游戏会话模型
type GameSession struct {
	BaseModel
	AppId      string    `orm:"size(50)" json:"app_id"`
	UserId     string    `orm:"size(100)" json:"user_id"`
	SessionId  string    `orm:"size(255);unique" json:"session_id"`
	StartTime  time.Time `orm:"type(datetime)" json:"start_time"`
	EndTime    time.Time `orm:"type(datetime);null" json:"end_time"`
	Duration   int       `orm:"default(0)" json:"duration"` // 秒
	Score      int64     `orm:"default(0)" json:"score"`
	Level      int       `orm:"default(1)" json:"level"`
	GameData   string    `orm:"type(text);null" json:"game_data"`
	ClientInfo string    `orm:"type(text);null" json:"client_info"`
	Status     int       `orm:"default(1)" json:"status"` // 1: 进行中, 2: 已结束, 0: 异常
}

// TableName 返回表名
func (g *GameSession) TableName() string {
	return "game_sessions"
}

// Register 注册模型
func (g *GameSession) Register() error {
	return RegisterModel(new(GameSession))
}

// RegisterWithSuffix 带后缀注册模型
func (g *GameSession) RegisterWithSuffix(suffix string) error {
	return RegisterModelWithSuffix(new(GameSession), suffix)
}

// Init 初始化模型
func (g *GameSession) Init() error {
	return g.Register()
}

// IsActive 检查会话是否活跃
func (g *GameSession) IsActive() bool {
	return g.Status == 1
}

// End 结束会话
func (g *GameSession) End() {
	g.EndTime = time.Now()
	g.Duration = int(g.EndTime.Sub(g.StartTime).Seconds())
	g.Status = 2
}

// StatsQuery 统计查询结构
type StatsQuery struct {
	AppId     string    `json:"app_id"`
	StatType  string    `json:"stat_type"`
	StatKey   string    `json:"stat_key"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
}

// StatsResponse 统计响应结构
type StatsResponse struct {
	List     []Statistics `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Summary  StatsSummary `json:"summary"`
}

// StatsSummary 统计摘要
type StatsSummary struct {
	TotalValue   int64   `json:"total_value"`
	AverageValue float64 `json:"average_value"`
	MaxValue     int64   `json:"max_value"`
	MinValue     int64   `json:"min_value"`
	Count        int64   `json:"count"`
}
