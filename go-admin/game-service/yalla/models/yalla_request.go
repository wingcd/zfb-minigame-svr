package models

// YallaBaseRequest Yalla基础请求结构
type YallaBaseRequest struct {
	AppID     string `json:"app_id" validate:"required"`
	Timestamp int64  `json:"timestamp" validate:"required"`
	Sign      string `json:"sign" validate:"required"`
}

// YallaAuthRequest 用户认证请求
type YallaAuthRequest struct {
	YallaBaseRequest
	UserID    string `json:"user_id" validate:"required"`
	AuthToken string `json:"auth_token" validate:"required"`
}

// YallaUserInfoRequest 获取用户信息请求
type YallaUserInfoRequest struct {
	YallaBaseRequest
	YallaUserID string `json:"yalla_user_id" validate:"required"`
}

// YallaRewardRequest 发放奖励请求
type YallaRewardRequest struct {
	YallaBaseRequest
	YallaUserID  string                 `json:"yalla_user_id" validate:"required"`
	RewardType   string                 `json:"reward_type" validate:"required"`
	RewardAmount int64                  `json:"reward_amount" validate:"required"`
	RewardData   map[string]interface{} `json:"reward_data,omitempty"`
	Description  string                 `json:"description,omitempty"`
}

// YallaGameDataRequest 游戏数据同步请求
type YallaGameDataRequest struct {
	YallaBaseRequest
	YallaUserID string                 `json:"yalla_user_id" validate:"required"`
	DataType    string                 `json:"data_type" validate:"required"`
	GameData    map[string]interface{} `json:"game_data" validate:"required"`
}

// YallaEventRequest 事件上报请求
type YallaEventRequest struct {
	YallaBaseRequest
	YallaUserID string                 `json:"yalla_user_id" validate:"required"`
	EventType   string                 `json:"event_type" validate:"required"`
	EventData   map[string]interface{} `json:"event_data,omitempty"`
}
