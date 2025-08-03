# API 自动注册使用说明

## 概述

新的 API 工厂系统支持自动注册功能，无需在 `index.js` 中统一注册所有 API。每个 API 函数可以在文件末尾主动注册自己。

## 使用方法

### 方法一：使用 register 函数（推荐）

在 API 文件末尾添加：

```javascript
const { register } = require('../api-factory');

// 导出带权限校验的函数
const mainFunc = requirePermission(yourHandler, 'permission_name');
exports.main = mainFunc;

// 自动注册API
register('api.name', mainFunc);
```

### 方法二：使用装饰器方式

在 API 文件末尾添加：

```javascript
const { autoRegister } = require('../api-factory');

// 使用装饰器方式自动注册
exports.main = autoRegister('api.name')(requirePermission(yourHandler, 'permission_name'));
```

## 示例

### getAdminList.js 示例（方法一）

```javascript
// 导出带权限校验的函数
const mainFunc = requirePermission(getAdminListHandler, 'admin_manage');
exports.main = mainFunc;

// 自动注册API
const { register } = require('../api-factory');
register('admin.getList', mainFunc);
```

### createAdmin.js 示例（方法二）

```javascript
const { autoRegister } = require('../api-factory');

// 使用装饰器方式自动注册
exports.main = autoRegister('admin.create')(requirePermission(createAdminHandler, 'admin_manage'));
```

## API 工厂函数

### registerAPI(apiName, func)
- 手动注册 API
- 参数：apiName (string), func (function)
- 返回：void

### getAPI(apiName)
- 获取已注册的 API
- 参数：apiName (string)
- 返回：function 或 undefined

### getAPIList()
- 获取所有已注册的 API 列表
- 返回：object

### autoRegister(apiName)
- 装饰器函数，用于自动注册
- 参数：apiName (string)
- 返回：function (装饰器)

### register(apiName, func)
- 便捷注册函数
- 参数：apiName (string), func (function)
- 返回：function (原函数)

## 优势

1. **分散管理**：每个 API 文件负责注册自己，无需集中管理
2. **自动发现**：系统启动时自动发现所有已注册的 API
3. **类型安全**：支持装饰器模式，代码更清晰
4. **向后兼容**：保持原有的导出方式不变

## 迁移指南

1. 在现有 API 文件末尾添加注册代码
2. 移除 `index.js` 中的手动注册
3. 使用 `getAPIList()` 获取所有已注册的 API

## 注意事项

- API 名称应该保持一致性，建议使用 `module.action` 格式
- 注册时会在控制台输出日志，便于调试
- 确保在文件末尾进行注册，避免循环依赖 