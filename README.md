# 小游戏管理后台系统

基于 Vue.js 3 + Element Plus 构建的小游戏管理后台系统，用于管理小游戏应用、用户数据、排行榜等功能。

## 功能特性

### 🎮 应用管理
- ✅ 应用创建、编辑、删除
- ✅ 多平台支持（微信小程序、支付宝小程序、抖音小程序）
- ✅ 应用状态管理（启用/停用）
- ✅ 应用详情查看（用户统计、排行榜列表）
- ✅ 应用搜索与筛选

### 👥 用户管理
- ✅ 用户列表查看与搜索
- ✅ 用户游戏数据查看与编辑
- ✅ 用户封禁/解封功能
- ✅ 用户删除功能
- ✅ 用户统计信息（总数、新增、活跃等）

### 🏆 排行榜管理
- ✅ 排行榜配置创建与编辑
- ✅ 多种更新策略（历史最高值、最近记录、历史总和）
- ✅ 排行榜数据查看与管理
- ✅ 玩家分数编辑与删除
- ✅ 排行榜统计信息

### 📊 数据统计
- ✅ 仪表板总览
- ✅ 用户增长趋势图表
- ✅ 平台分布统计
- ✅ 热门应用排行
- ✅ 最近活动记录

## 技术栈

- **前端框架**: Vue.js 3 (Composition API)
- **UI 组件库**: Element Plus
- **路由管理**: Vue Router 4
- **图表库**: ECharts
- **HTTP 客户端**: Axios
- **构建工具**: Vite
- **语言**: JavaScript

## 项目结构

```
game-web-admin/
├── public/                 # 静态资源
├── src/
│   ├── components/         # 公共组件
│   ├── config/            # 配置文件
│   │   └── index.js       # API 配置
│   ├── router/            # 路由配置
│   │   └── index.js       # 路由定义
│   ├── services/          # API 服务
│   │   ├── api.js         # API 接口封装
│   │   └── statsService.js # 统计服务
│   ├── views/             # 页面组件
│   │   ├── Dashboard.vue      # 仪表板
│   │   ├── AppManagement.vue  # 应用管理
│   │   ├── UserManagement.vue # 用户管理
│   │   └── LeaderboardManagement.vue # 排行榜管理
│   ├── App.vue            # 根组件
│   ├── main.js            # 入口文件
│   └── style.css          # 全局样式
├── package.json           # 项目依赖
├── vite.config.js         # Vite 配置
└── README.md              # 项目说明
```

## API 接口对应

### 云函数接口映射

项目已根据 `alipay-server` 中的云函数接口进行了完整的接口封装：

#### 用户相关接口
- `POST /user/login` → `alipay-server/user/login.js`
- `POST /user/login.wx` → `alipay-server/user/login.wx.js`
- `POST /user/getData` → `alipay-server/user/getData.js`
- `POST /user/saveData` → `alipay-server/user/saveData.js`

#### 应用相关接口
- `POST /app/appInit` → `alipay-server/app/appInit.js`
- `POST /app/queryApp` → `alipay-server/app/queryApp.js`

#### 排行榜相关接口
- `POST /leaderboard/commitScore` → `alipay-server/leaderboard/commitScore.js`
- `POST /leaderboard/queryScore` → `alipay-server/leaderboard/queryScore.js`
- `POST /leaderboard/getLeaderboardTopRank` → `alipay-server/leaderboard/getLeaderboardTopRank.js`
- `POST /leaderboard/deleteScore` → `alipay-server/leaderboard/deleteScore.js`

## 安装与运行

### 环境要求
- Node.js >= 16.0.0
- npm >= 7.0.0

### 安装依赖
```bash
cd game-web-admin
npm install
```

### 开发模式
```bash
npm run dev
```

### 生产构建
```bash
npm run build
```

### 预览构建结果
```bash
npm run preview
```

## 配置说明

### API 配置
编辑 `src/config/index.js` 文件，配置云函数服务的 API 地址：

```javascript
export default {
  api: {
    baseURL: process.env.NODE_ENV === 'production' 
      ? 'https://your-alipay-function-domain.com' 
      : 'http://localhost:3000/api',
    timeout: 10000
  }
}
```

### 云函数部署
确保 `alipay-server` 中的云函数已正确部署到支付宝云服务，并配置好相应的路由映射。

## 主要功能使用指南

### 1. 应用管理
1. 点击"创建应用"按钮
2. 填写应用名称、选择平台、输入渠道应用ID和密钥
3. 保存后即可在应用列表中看到新创建的应用
4. 可对应用进行编辑、启用/停用、查看详情等操作

### 2. 用户管理
1. 选择要管理的应用
2. 在用户列表中可以查看所有用户信息
3. 点击"查看数据"可以查看和编辑用户的游戏数据
4. 可对用户进行封禁/解封、删除等操作

### 3. 排行榜管理
1. 选择要管理的应用
2. 在排行榜配置中创建新的排行榜
3. 设置排行榜类型、名称、排序方式和更新策略
4. 可查看排行榜数据、编辑玩家分数、删除记录等

### 4. 数据统计
- 仪表板展示系统整体运营数据
- 包含用户增长趋势、平台分布等图表
- 显示热门应用排行和最近系统活动

## 开发说明

### 添加新功能
1. 在 `src/services/api.js` 中添加新的 API 接口
2. 在相应的 Vue 组件中调用接口
3. 更新路由配置（如需要）

### 自定义主题
可通过修改 Element Plus 的 CSS 变量来自定义主题颜色。

### 响应式设计
所有页面都已适配移动端，支持响应式布局。

## 注意事项

1. **安全性**: 生产环境请配置正确的 CORS 策略和身份验证
2. **性能**: 大量数据时建议实现虚拟滚动或分页加载
3. **错误处理**: 已实现基础错误处理，可根据需求扩展
4. **数据校验**: 重要操作已添加确认提示

## 许可证

MIT License

## 支持

如有问题或建议，请创建 Issue 或 Pull Request。 