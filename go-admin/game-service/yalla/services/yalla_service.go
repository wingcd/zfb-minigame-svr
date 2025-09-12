package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	
	"game-service/yalla/models"
	"game-service/yalla/utils"
)

// YallaService Yalla服务
type YallaService struct {
	config     *models.YallaConfig
	httpClient *utils.YallaHTTPClient
	crypto     *utils.YallaCrypto
	logger     *YallaServiceLogger
}

// YallaServiceLogger Yalla服务日志器
type YallaServiceLogger struct{}

func (l *YallaServiceLogger) Info(msg string, fields ...interface{}) {
	logs.Info("YallaService: "+msg, fields...)
}

func (l *YallaServiceLogger) Error(msg string, fields ...interface{}) {
	logs.Error("YallaService: "+msg, fields...)
}

func (l *YallaServiceLogger) Debug(msg string, fields ...interface{}) {
	logs.Debug("YallaService: "+msg, fields...)
}

// NewYallaService 创建Yalla服务实例
func NewYallaService(appID string) (*YallaService, error) {
	// 获取配置
	config, err := GetYallaConfig(appID)
	if err != nil {
		return nil, fmt.Errorf("get yalla config failed: %v", err)
	}
	
	if config.Status != 1 {
		return nil, fmt.Errorf("yalla service is disabled for app: %s", appID)
	}
	
	logger := &YallaServiceLogger{}
	
	// 创建HTTP客户端
	httpClient := utils.NewYallaHTTPClient(
		config.BaseURL,
		config.Timeout,
		config.RetryCount,
		logger,
	)
	
	// 创建加密工具
	crypto := utils.NewYallaCrypto(config.SecretKey)
	
	return &YallaService{
		config:     config,
		httpClient: httpClient,
		crypto:     crypto,
		logger:     logger,
	}, nil
}

// AuthenticateUser 用户认证
func (s *YallaService) AuthenticateUser(userID, authToken string) (*models.YallaUserInfo, error) {
	startTime := time.Now()
	
	// 构建请求
	request := &models.YallaAuthRequest{
		YallaBaseRequest: models.YallaBaseRequest{
			AppID:     s.config.AppID,
			Timestamp: time.Now().Unix(),
		},
		UserID:    userID,
		AuthToken: authToken,
	}
	
	// 生成签名
	params := map[string]interface{}{
		"app_id":     request.AppID,
		"user_id":    request.UserID,
		"auth_token": request.AuthToken,
	}
	request.Sign = s.crypto.GenerateSign(params, request.Timestamp)
	
	// 发送请求
	var response models.YallaAuthResponse
	err := s.httpClient.PostJSON("/api/auth/verify", request, &response)
	
	// 记录日志
	s.logAPICall("POST", "/api/auth/verify", userID, request, &response, time.Since(startTime), err)
	
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("auth failed: %s", response.Message)
	}
	
	// 更新用户绑定信息
	if response.Data != nil {
		err = s.updateUserBinding(userID, response.Data.YallaUserID, authToken)
		if err != nil {
			s.logger.Error("update user binding failed", "error", err)
		}
	}
	
	return response.Data, nil
}

// GetUserInfo 获取用户信息
func (s *YallaService) GetUserInfo(yallaUserID string) (*models.YallaUserInfo, error) {
	startTime := time.Now()
	
	// 构建请求
	request := &models.YallaUserInfoRequest{
		YallaBaseRequest: models.YallaBaseRequest{
			AppID:     s.config.AppID,
			Timestamp: time.Now().Unix(),
		},
		YallaUserID: yallaUserID,
	}
	
	// 生成签名
	params := map[string]interface{}{
		"app_id":       request.AppID,
		"yalla_user_id": request.YallaUserID,
	}
	request.Sign = s.crypto.GenerateSign(params, request.Timestamp)
	
	// 发送请求
	var response models.YallaUserInfoResponse
	err := s.httpClient.PostJSON("/api/user/info", request, &response)
	
	// 记录日志
	s.logAPICall("POST", "/api/user/info", yallaUserID, request, &response, time.Since(startTime), err)
	
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("get user info failed: %s", response.Message)
	}
	
	return response.Data, nil
}

// SendReward 发放奖励
func (s *YallaService) SendReward(yallaUserID, rewardType string, rewardAmount int64, rewardData map[string]interface{}, description string) (*models.YallaRewardResult, error) {
	startTime := time.Now()
	
	// 构建请求
	request := &models.YallaRewardRequest{
		YallaBaseRequest: models.YallaBaseRequest{
			AppID:     s.config.AppID,
			Timestamp: time.Now().Unix(),
		},
		YallaUserID:  yallaUserID,
		RewardType:   rewardType,
		RewardAmount: rewardAmount,
		RewardData:   rewardData,
		Description:  description,
	}
	
	// 生成签名
	params := map[string]interface{}{
		"app_id":        request.AppID,
		"yalla_user_id": request.YallaUserID,
		"reward_type":   request.RewardType,
		"reward_amount": request.RewardAmount,
	}
	if description != "" {
		params["description"] = description
	}
	request.Sign = s.crypto.GenerateSign(params, request.Timestamp)
	
	// 发送请求
	var response models.YallaRewardResponse
	err := s.httpClient.PostJSON("/api/reward/send", request, &response)
	
	// 记录日志
	s.logAPICall("POST", "/api/reward/send", yallaUserID, request, &response, time.Since(startTime), err)
	
	if err != nil {
		return nil, fmt.Errorf("send reward failed: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("send reward failed: %s", response.Message)
	}
	
	return response.Data, nil
}

// SyncGameData 同步游戏数据
func (s *YallaService) SyncGameData(yallaUserID, dataType string, gameData map[string]interface{}) (*models.YallaGameDataResult, error) {
	startTime := time.Now()
	
	// 构建请求
	request := &models.YallaGameDataRequest{
		YallaBaseRequest: models.YallaBaseRequest{
			AppID:     s.config.AppID,
			Timestamp: time.Now().Unix(),
		},
		YallaUserID: yallaUserID,
		DataType:    dataType,
		GameData:    gameData,
	}
	
	// 生成签名
	params := map[string]interface{}{
		"app_id":       request.AppID,
		"yalla_user_id": request.YallaUserID,
		"data_type":    request.DataType,
	}
	request.Sign = s.crypto.GenerateSign(params, request.Timestamp)
	
	// 发送请求
	var response models.YallaGameDataResponse
	err := s.httpClient.PostJSON("/api/data/sync", request, &response)
	
	// 记录日志
	s.logAPICall("POST", "/api/data/sync", yallaUserID, request, &response, time.Since(startTime), err)
	
	if err != nil {
		return nil, fmt.Errorf("sync game data failed: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("sync game data failed: %s", response.Message)
	}
	
	return response.Data, nil
}

// ReportEvent 上报事件
func (s *YallaService) ReportEvent(yallaUserID, eventType string, eventData map[string]interface{}) (*models.YallaEventResult, error) {
	startTime := time.Now()
	
	// 构建请求
	request := &models.YallaEventRequest{
		YallaBaseRequest: models.YallaBaseRequest{
			AppID:     s.config.AppID,
			Timestamp: time.Now().Unix(),
		},
		YallaUserID: yallaUserID,
		EventType:   eventType,
		EventData:   eventData,
	}
	
	// 生成签名
	params := map[string]interface{}{
		"app_id":       request.AppID,
		"yalla_user_id": request.YallaUserID,
		"event_type":   request.EventType,
	}
	request.Sign = s.crypto.GenerateSign(params, request.Timestamp)
	
	// 发送请求
	var response models.YallaEventResponse
	err := s.httpClient.PostJSON("/api/event/report", request, &response)
	
	// 记录日志
	s.logAPICall("POST", "/api/event/report", yallaUserID, request, &response, time.Since(startTime), err)
	
	if err != nil {
		return nil, fmt.Errorf("report event failed: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("report event failed: %s", response.Message)
	}
	
	return response.Data, nil
}

// logAPICall 记录API调用日志
func (s *YallaService) logAPICall(method, endpoint, userID string, request, response interface{}, duration time.Duration, err error) {
	if !s.config.EnableLog {
		return
	}
	
	requestData, _ := json.Marshal(request)
	responseData, _ := json.Marshal(response)
	
	log := &models.YallaCallLog{
		AppID:        s.config.AppID,
		UserID:       userID,
		Method:       method,
		Endpoint:     endpoint,
		RequestData:  string(requestData),
		ResponseData: string(responseData),
		Duration:     duration.Milliseconds(),
		Success:      err == nil,
	}
	
	if err != nil {
		log.ErrorMsg = err.Error()
	}
	
	// 异步保存日志
	go func() {
		o := orm.NewOrm()
		_, insertErr := o.Insert(log)
		if insertErr != nil {
			s.logger.Error("save api call log failed", "error", insertErr)
		}
	}()
}

// updateUserBinding 更新用户绑定信息
func (s *YallaService) updateUserBinding(gameUserID, yallaUserID, token string) error {
	o := orm.NewOrm()
	
	binding := &models.YallaUserBinding{}
	err := o.QueryTable("yalla_user_bindings").
		Filter("app_id", s.config.AppID).
		Filter("game_user_id", gameUserID).
		One(binding)
	
	if err == orm.ErrNoRows {
		// 创建新绑定
		binding = &models.YallaUserBinding{
			AppID:       s.config.AppID,
			GameUserID:  gameUserID,
			YallaUserID: yallaUserID,
			YallaToken:  token,
			ExpiresAt:   time.Now().Add(24 * time.Hour), // 默认24小时过期
			Status:      1,
		}
		_, err = o.Insert(binding)
	} else if err == nil {
		// 更新现有绑定
		binding.YallaUserID = yallaUserID
		binding.YallaToken = token
		binding.ExpiresAt = time.Now().Add(24 * time.Hour)
		binding.Status = 1
		_, err = o.Update(binding)
	}
	
	return err
}

// GetYallaConfig 获取Yalla配置
func GetYallaConfig(appID string) (*models.YallaConfig, error) {
	o := orm.NewOrm()
	config := &models.YallaConfig{}
	
	err := o.QueryTable("yalla_config").
		Filter("app_id", appID).
		Filter("status", 1).
		One(config)
	
	if err != nil {
		return nil, err
	}
	
	return config, nil
}

// GetUserBinding 获取用户绑定信息
func GetUserBinding(appID, gameUserID string) (*models.YallaUserBinding, error) {
	o := orm.NewOrm()
	binding := &models.YallaUserBinding{}
	
	err := o.QueryTable("yalla_user_bindings").
		Filter("app_id", appID).
		Filter("game_user_id", gameUserID).
		Filter("status", 1).
		One(binding)
	
	if err != nil {
		return nil, err
	}
	
	// 检查token是否过期
	if time.Now().After(binding.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}
	
	return binding, nil
}
