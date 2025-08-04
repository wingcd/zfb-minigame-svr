# ZY-SDK

小游戏开发工具包，提供用户管理、排行榜、计数器、邮件系统等功能。

## 安装

```bash
npm install zy-sdk
```

## 快速开始

```typescript
import { ZYSDK } from 'zy-sdk';

// 初始化SDK
ZYSDK.init({
    appId: 'your-app-id',
    baseUrl: 'https://your-api-domain.com'
});
```

## 功能模块

### 用户模块 (User)
- 用户登录/注册
- 用户信息管理
- 用户数据统计

### 排行榜模块 (Leaderboard)
- 创建和管理排行榜
- 提交分数
- 获取排行榜数据

### 计数器模块 (Counter)
- 全局计数器
- 用户计数器
- 计数器统计

### 邮件模块 (Mail) 🆕
- 获取用户邮件列表
- 阅读邮件
- 领取邮件奖励
- 删除邮件
- 批量操作
- 未读消息统计

## 使用示例

### 邮件系统

```typescript
// 获取用户邮件
const mails = await ZYSDK.mail.getMails({
    openId: 'user-open-id',
    page: 1,
    pageSize: 20
});

// 阅读邮件
await ZYSDK.mail.readMail('user-open-id', 'mail-id');

// 领取奖励
const result = await ZYSDK.mail.receiveMail('user-open-id', 'mail-id');
if (result.code === 0) {
    console.log('奖励领取成功:', result.data.rewards);
}

// 获取未读数量
const count = await ZYSDK.mail.getUnreadCount({
    openId: 'user-open-id'
});
```

### 用户系统

```typescript
// 用户登录
const user = await ZYSDK.user.login({
    openId: 'user-open-id',
    nickname: '用户昵称'
});
```

### 排行榜

```typescript
// 提交分数
await ZYSDK.leaderboard.submitScore({
    openId: 'user-open-id',
    leaderboardId: 'board-id',
    score: 1000
});

// 获取排行榜
const ranking = await ZYSDK.leaderboard.getRanking({
    leaderboardId: 'board-id',
    page: 1,
    pageSize: 10
});
```

### 计数器

```typescript
// 增加计数
await ZYSDK.counter.increment({
    openId: 'user-open-id',
    counterId: 'login-count',
    value: 1
});

// 获取计数
const count = await ZYSDK.counter.get({
    openId: 'user-open-id',
    counterId: 'login-count'
});
```

## 详细文档

- [邮件系统使用指南](./MAIL_USAGE.md)

## API 响应格式

所有API都遵循统一的响应格式：

```typescript
{
    code: number,        // 0表示成功，非0表示错误
    msg: string,         // 响应消息
    timestamp: number,   // 时间戳
    data?: any          // 响应数据（可选）
}
```

## 错误码

- `0`: 成功
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

## 许可证

MIT 