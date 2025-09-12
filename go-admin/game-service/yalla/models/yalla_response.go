package models

// YallaBaseResponse Yalla基础响应结构
type YallaBaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// YallaAuthResponse 用户认证响应
type YallaAuthResponse struct {
	YallaBaseResponse
	Data *YallaUserInfo `json:"data,omitempty"`
}

// YallaUserInfo Yalla用户信息
type YallaUserInfo struct {
	YallaUserID string `json:"yalla_user_id"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	Level       int    `json:"level"`
	VipLevel    int    `json:"vip_level"`
	Points      int64  `json:"points"`
	Coins       int64  `json:"coins"`
	Status      int    `json:"status"`
}

// YallaUserInfoResponse 获取用户信息响应
type YallaUserInfoResponse struct {
	YallaBaseResponse
	Data *YallaUserInfo `json:"data,omitempty"`
}

// YallaRewardResponse 发放奖励响应
type YallaRewardResponse struct {
	YallaBaseResponse
	Data *YallaRewardResult `json:"data,omitempty"`
}

// YallaRewardResult 奖励发放结果
type YallaRewardResult struct {
	RewardID     string `json:"reward_id"`
	YallaUserID  string `json:"yalla_user_id"`
	RewardType   string `json:"reward_type"`
	RewardAmount int64  `json:"reward_amount"`
	Status       string `json:"status"`
	ProcessedAt  string `json:"processed_at"`
}

// YallaGameDataResponse 游戏数据同步响应
type YallaGameDataResponse struct {
	YallaBaseResponse
	Data *YallaGameDataResult `json:"data,omitempty"`
}

// YallaGameDataResult 游戏数据同步结果
type YallaGameDataResult struct {
	DataID      string `json:"data_id"`
	YallaUserID string `json:"yalla_user_id"`
	DataType    string `json:"data_type"`
	Status      string `json:"status"`
	SyncedAt    string `json:"synced_at"`
}

// YallaEventResponse 事件上报响应
type YallaEventResponse struct {
	YallaBaseResponse
	Data *YallaEventResult `json:"data,omitempty"`
}

// YallaEventResult 事件上报结果
type YallaEventResult struct {
	EventID     string `json:"event_id"`
	YallaUserID string `json:"yalla_user_id"`
	EventType   string `json:"event_type"`
	Status      string `json:"status"`
	ReportedAt  string `json:"reported_at"`
}
