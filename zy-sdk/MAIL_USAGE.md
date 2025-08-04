# 邮件系统 SDK 使用指南

## 初始化

```typescript
import { ZYSDK } from 'zy-sdk';

// 初始化SDK
ZYSDK.init({
    appId: 'your-app-id',
    baseUrl: 'https://your-api-domain.com'
});
```

## 邮件模块功能

### 1. 获取用户邮件列表

```typescript
// 获取所有邮件
const response = await ZYSDK.mail.getMails({
    openId: 'user-open-id'
});

// 分页获取邮件
const response = await ZYSDK.mail.getMails({
    openId: 'user-open-id',
    page: 1,
    pageSize: 20
});

// 按类型筛选
const response = await ZYSDK.mail.getMails({
    openId: 'user-open-id',
    type: 'reward' // 'system' | 'notice' | 'reward'
});

// 按状态筛选
const response = await ZYSDK.mail.getMails({
    openId: 'user-open-id',
    status: 'unread' // 'unread' | 'read' | 'received' | 'deleted'
});
```

响应数据结构：
```typescript
{
    code: 0,
    msg: "success",
    timestamp: 1634567890000,
    data: {
        list: [
            {
                mailId: "mail-123",
                title: "系统奖励",
                content: "恭喜您获得登录奖励！",
                type: "reward",
                rewards: [
                    {
                        type: "coin",
                        name: "金币",
                        amount: 100,
                        description: "游戏金币"
                    }
                ],
                publishTime: "2023-10-01 10:00:00",
                expireTime: "2023-10-31 23:59:59",
                isRead: false,
                isReceived: false,
                isDeleted: false,
                status: "unread"
            }
        ],
        total: 50,
        page: 1,
        pageSize: 20,
        totalPages: 3,
        hasMore: true
    }
}
```

### 2. 阅读邮件

```typescript
const response = await ZYSDK.mail.readMail('user-open-id', 'mail-123');
```

### 3. 领取邮件奖励

```typescript
const response = await ZYSDK.mail.receiveMail('user-open-id', 'mail-123');

// 响应包含奖励信息
if (response.code === 0) {
    const rewards = response.data?.rewards || [];
    rewards.forEach(reward => {
        console.log(`获得 ${reward.name} x${reward.amount}`);
    });
}
```

### 4. 删除邮件

```typescript
const response = await ZYSDK.mail.deleteMail('user-open-id', 'mail-123');
```

### 5. 获取未读邮件数量

```typescript
const response = await ZYSDK.mail.getUnreadCount({
    openId: 'user-open-id'
});

if (response.code === 0) {
    const { unreadCount, unreceiveCount } = response.data;
    console.log(`未读邮件: ${unreadCount}, 未领取奖励: ${unreceiveCount}`);
}
```

### 6. 批量操作

```typescript
// 一键领取所有可领取的奖励
const response = await ZYSDK.mail.receiveAllRewards('user-open-id');

// 一键删除所有已读邮件
const response = await ZYSDK.mail.deleteAllRead('user-open-id');
```

### 7. 通用状态更新

```typescript
// 使用通用方法更新邮件状态
const response = await ZYSDK.mail.updateStatus({
    openId: 'user-open-id',
    mailId: 'mail-123',
    action: 'read' // 'read' | 'receive' | 'delete'
});
```

## 错误处理

```typescript
try {
    const response = await ZYSDK.mail.getMails({
        openId: 'user-open-id'
    });
    
    if (response.code === 0) {
        // 成功处理
        const mails = response.data.list;
    } else {
        // 业务错误
        console.error('获取邮件失败:', response.msg);
    }
} catch (error) {
    // 网络或其他错误
    console.error('请求失败:', error);
}
```

## 邮件类型说明

- **system**: 系统邮件 - 系统自动发送的通知类邮件
- **notice**: 公告邮件 - 游戏公告、活动通知等
- **reward**: 奖励邮件 - 包含游戏道具、货币等奖励的邮件

## 邮件状态说明

- **unread**: 未读 - 邮件尚未被用户阅读
- **read**: 已读 - 邮件已被用户阅读，但奖励未领取
- **received**: 已领取 - 邮件奖励已被用户领取
- **deleted**: 已删除 - 邮件已被用户删除

## 注意事项

1. 所有接口都需要传入用户的 `openId`
2. 奖励邮件需要先阅读才能领取奖励
3. 已过期的邮件无法领取奖励
4. 删除操作不可逆，请谨慎使用
5. 批量操作可能涉及大量数据，建议在适当的时机调用 