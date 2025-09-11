package models

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/beego/beego/v2/client/orm"
)

// ModelRegistry 模型注册器
type ModelRegistry struct {
	mu               sync.RWMutex
	registeredModels map[string]bool
}

var (
	// 全局模型注册器实例
	registry = &ModelRegistry{
		registeredModels: make(map[string]bool),
	}
)

// GetRegistry 获取全局注册器实例
func GetRegistry() *ModelRegistry {
	return registry
}

// IsRegistered 检查模型是否已注册
func (r *ModelRegistry) IsRegistered(modelName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registeredModels[modelName]
}

// RegisterModel 注册单个模型
func (r *ModelRegistry) RegisterModel(model interface{}) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	modelName := modelType.Name()

	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否已注册
	if r.registeredModels[modelName] {
		return fmt.Errorf("model %s already registered", modelName)
	}

	// 注册模型到ORM
	orm.RegisterModel(model)

	// 记录已注册
	r.registeredModels[modelName] = true

	return nil
}

// RegisterModels 批量注册模型
func (r *ModelRegistry) RegisterModels(models ...interface{}) error {
	for _, model := range models {
		if err := r.RegisterModel(model); err != nil {
			return err
		}
	}
	return nil
}

// RegisterModelWithSuffix 带后缀注册模型（用于区分不同服务的同名模型）
func (r *ModelRegistry) RegisterModelWithSuffix(model interface{}, suffix string) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	modelName := fmt.Sprintf("%s_%s", modelType.Name(), suffix)

	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否已注册
	if r.registeredModels[modelName] {
		return fmt.Errorf("model %s already registered", modelName)
	}

	// 注册模型到ORM
	orm.RegisterModelWithSuffix(suffix, model)

	// 记录已注册
	r.registeredModels[modelName] = true

	return nil
}

// RegisterModelsWithSuffix 带后缀批量注册模型
func (r *ModelRegistry) RegisterModelsWithSuffix(suffix string, models ...interface{}) error {
	for _, model := range models {
		if err := r.RegisterModelWithSuffix(model, suffix); err != nil {
			return err
		}
	}
	return nil
}

// GetRegisteredModels 获取所有已注册的模型名称
func (r *ModelRegistry) GetRegisteredModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []string
	for name := range r.registeredModels {
		models = append(models, name)
	}
	return models
}

// Clear 清空注册记录（主要用于测试）
func (r *ModelRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registeredModels = make(map[string]bool)
}

// 便捷函数
func RegisterModel(model interface{}) error {
	return registry.RegisterModel(model)
}

func RegisterModels(models ...interface{}) error {
	return registry.RegisterModels(models...)
}

func RegisterModelWithSuffix(model interface{}, suffix string) error {
	return registry.RegisterModelWithSuffix(model, suffix)
}

func RegisterModelsWithSuffix(suffix string, models ...interface{}) error {
	return registry.RegisterModelsWithSuffix(suffix, models...)
}

func IsRegistered(modelName string) bool {
	return registry.IsRegistered(modelName)
}

func GetRegisteredModels() []string {
	return registry.GetRegisteredModels()
}
