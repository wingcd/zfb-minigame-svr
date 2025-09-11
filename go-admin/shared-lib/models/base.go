package models

import (
	"time"
)

// BaseModel 基础模型结构
type BaseModel struct {
	ID        int       `orm:"auto;pk" json:"id"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"updated_at"`
}

// TableName 返回默认表名
func (m *BaseModel) TableName() string {
	return "base_model"
}

// BeforeInsert 插入前的钩子函数
func (m *BaseModel) BeforeInsert() {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
}

// BeforeUpdate 更新前的钩子函数
func (m *BaseModel) BeforeUpdate() {
	m.UpdatedAt = time.Now()
}

// ModelInitializer 模型初始化接口
type ModelInitializer interface {
	Init() error
}

// Registrable 可注册接口
type Registrable interface {
	Register() error
	TableName() string
}

// ServiceModels 服务模型接口
type ServiceModels interface {
	RegisterGameModels() error
	RegisterAdminModels() error
	RegisterCommonModels() error
}
