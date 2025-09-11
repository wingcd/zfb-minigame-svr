package models

import (
	"fmt"
	"log"
)

// ServiceModelRegistry 服务模型注册器
type ServiceModelRegistry struct {
	serviceName string
	models      []Registrable
}

// NewServiceModelRegistry 创建服务模型注册器
func NewServiceModelRegistry(serviceName string) *ServiceModelRegistry {
	return &ServiceModelRegistry{
		serviceName: serviceName,
		models:      make([]Registrable, 0),
	}
}

// AddModel 添加模型到注册器
func (s *ServiceModelRegistry) AddModel(model Registrable) *ServiceModelRegistry {
	s.models = append(s.models, model)
	return s
}

// AddModels 批量添加模型
func (s *ServiceModelRegistry) AddModels(models ...Registrable) *ServiceModelRegistry {
	s.models = append(s.models, models...)
	return s
}

// RegisterAll 注册所有模型
func (s *ServiceModelRegistry) RegisterAll() error {
	for _, model := range s.models {
		// 使用服务名作为后缀注册模型
		if err := RegisterModelWithSuffix(model, s.serviceName); err != nil {
			log.Printf("Failed to register model %T with suffix %s: %v", model, s.serviceName, err)
			return err
		}
	}
	return nil
}

// RegisterCommonModels 注册通用模型（不带后缀）
func (s *ServiceModelRegistry) RegisterCommonModels() error {
	commonModels := []interface{}{
		new(Application),
	}

	for _, model := range commonModels {
		if err := RegisterModel(model); err != nil {
			log.Printf("Failed to register common model %T: %v", model, err)
			return err
		}
	}
	return nil
}

// InitGameServiceModels 初始化游戏服务模型
func InitGameServiceModels() error {
	registry := NewServiceModelRegistry("game")

	// 添加游戏服务特有的模型
	registry.AddModels(
		new(LeaderboardConfig),
		new(Leaderboard),
		new(GameSession),
		new(Statistics),
	)

	// 注册通用模型
	if err := registry.RegisterCommonModels(); err != nil {
		return fmt.Errorf("failed to register common models: %v", err)
	}

	// 注册游戏服务模型
	if err := registry.RegisterAll(); err != nil {
		return fmt.Errorf("failed to register game service models: %v", err)
	}

	log.Printf("Game service models registered successfully")
	return nil
}

// InitAdminServiceModels 初始化管理服务模型
func InitAdminServiceModels() error {
	registry := NewServiceModelRegistry("admin")

	// 添加管理服务特有的模型
	registry.AddModels(
		new(LeaderboardConfig),
		new(Leaderboard),
		new(Statistics),
		new(GameSession),
	)

	// 注册通用模型
	if err := registry.RegisterCommonModels(); err != nil {
		return fmt.Errorf("failed to register common models: %v", err)
	}

	// 注册管理服务模型
	if err := registry.RegisterAll(); err != nil {
		return fmt.Errorf("failed to register admin service models: %v", err)
	}

	log.Printf("Admin service models registered successfully")
	return nil
}

// GetRegistryStatus 获取注册状态
func GetRegistryStatus() map[string]interface{} {
	return map[string]interface{}{
		"registered_models": GetRegisteredModels(),
		"total_count":       len(GetRegisteredModels()),
	}
}

// ValidateModels 验证模型注册状态
func ValidateModels(requiredModels []string) error {
	registered := GetRegisteredModels()
	registeredMap := make(map[string]bool)

	for _, model := range registered {
		registeredMap[model] = true
	}

	var missing []string
	for _, required := range requiredModels {
		if !registeredMap[required] {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required models: %v", missing)
	}

	return nil
}
