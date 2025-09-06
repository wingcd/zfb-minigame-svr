# 数据格式迁移总结

## 概述
在从云函数后端迁移到Go后端的过程中，发现了多个数据格式不一致的问题。本文档记录了这些问题及其修复方案。

## 主要数据格式差异

### 1. 状态字段格式不一致

#### 应用状态 (applications)
- **云函数 (旧)**: 使用字符串 `'active'` / `'inactive'`
- **Go后端 (新)**: 使用数字 `1` / `0`
- **影响文件**: 
  - `AppManagement.vue`
  - `admin-service/models/application.go`

#### 管理员状态 (admin_users)  
- **云函数 (旧)**: 使用字符串 `'active'` / `'inactive'`
- **Go后端 (新)**: 使用数字 `1` / `0`
- **影响文件**:
  - `AdminManagement.vue`
  - `admin-service/models/admin_user.go`

### 2. 已修复的具体变更

#### AppManagement.vue
1. 表格状态显示: `scope.row.status === 'active'` → `scope.row.status === 1`
2. 状态切换按钮: `scope.row.status === 'active'` → `scope.row.status === 1`
3. 表单默认值: `status: 'active'` → `status: 1`
4. 开关组件: `active-value="active"` → `:active-value="1"`
5. 详情页状态: `detailDialog.app.status === 'active'` → `detailDialog.app.status === 1`
6. 状态切换逻辑: `app.status === 'active' ? 'inactive' : 'active'` → `app.status === 1 ? 0 : 1`

#### AdminManagement.vue
1. 搜索选项: `value="active"` → `:value="1"`
2. 表格状态显示: `row.status === 'active'` → `row.status === 1`
3. 表单默认值: `status: 'active'` → `status: 1`
4. 单选按钮: `label="active"` → `:label="1"`

#### 后端模型文件
1. `admin-service/models/application.go`: 状态字段从 string 改为 int
2. `admin-service/models/admin_user.go`: 状态字段从 string 改为 int

### 3. 保持不变的状态格式

#### 邮件状态 (保持字符串格式)
邮件系统使用特殊的状态值，保持原有字符串格式：
- `'pending'` - 待发布
- `'scheduled'` - 定时发布  
- `'active'` - 已发布
- `'expired'` - 已过期
- `'draft'` - 草稿

#### 游戏配置状态 (保持布尔格式)
游戏配置使用布尔值 `true/false`，无需修改。

## 注意事项

1. **API兼容性**: 前端和后端必须使用相同的数据格式
2. **数据库迁移**: 如果有现有数据，需要运行数据库迁移脚本
3. **测试**: 需要全面测试状态切换功能
4. **文档更新**: API文档需要同步更新状态字段类型

## 验证清单

- [x] AppManagement.vue 状态格式修复
- [x] AdminManagement.vue 状态格式修复  
- [x] 后端模型状态字段类型调整
- [ ] 数据库迁移脚本 (如需要)
- [ ] API文档更新
- [ ] 功能测试验证

## 云函数需要更新的文件

如果继续使用云函数，以下文件需要更新以匹配新的数据格式：

### 应用相关
- `getAllApps/index.js`: 返回数字状态而非字符串
- `createApp/index.js`: 接受数字状态参数
- `updateApp/index.js`: 处理数字状态更新

### 管理员相关  
- `getAdminList/index.js`: 返回数字状态
- `adminLogin/index.js`: 状态查询使用数字格式
- `createAdmin/index.js`: 接受数字状态参数
- `updateAdmin/index.js`: 处理数字状态更新

每个文件需要将:
```javascript
// 旧格式
status: 'active'
status === 'active'
status: 'inactive'

// 改为新格式  
status: 1
status === 1
status: 0
```
