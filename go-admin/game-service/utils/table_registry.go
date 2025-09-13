package utils

import (
	"fmt"
	"sync"

	"github.com/beego/beego/v2/client/orm"
)

var (
	registeredTables = make(map[string]bool)
	tableMutex       sync.RWMutex
)

// RegisterCounterTable 注册计数器表模型
func RegisterCounterTable(tableName, appId string, model interface{}) error {
	tableMutex.Lock()
	defer tableMutex.Unlock()

	// 检查是否已注册
	if registeredTables[tableName] {
		return nil
	}

	// 注册模型
	orm.RegisterModelWithSuffix("_"+appId, model)

	// 标记为已注册
	registeredTables[tableName] = true

	return nil
}

// IsTableRegistered 检查表是否已注册
func IsTableRegistered(tableName string) bool {
	tableMutex.RLock()
	defer tableMutex.RUnlock()
	return registeredTables[tableName]
}

// EnsureTableRegistered 确保计数器表已注册
func EnsureTableRegistered(tableName, appId string, model interface{}) error {
	if !IsTableRegistered(tableName) {
		return RegisterCounterTable(tableName, appId, model)
	}

	return nil
}

// RegisterAllDynamicTables 注册所有动态表类型
func RegisterAllDynamicTables(tableName, appId string, model interface{}) error {
	// 注册计数器表
	if err := RegisterCounterTable(tableName, appId, model); err != nil {
		return fmt.Errorf("注册计数器表失败: %v", err)
	}

	// 可以在这里添加其他动态表的注册
	// RegisterUserTable(appId)
	// RegisterLeaderboardTable(appId)
	// RegisterMailTable(appId)

	return nil
}

// GetRegisteredTables 获取已注册的表列表
func GetRegisteredTables() []string {
	tableMutex.RLock()
	defer tableMutex.RUnlock()

	tables := make([]string, 0, len(registeredTables))
	for tableName := range registeredTables {
		tables = append(tables, tableName)
	}

	return tables
}
