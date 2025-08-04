const cloud = require("@alipay/faas-server-sdk");
const { formatTime, generateId } = require("./common");

/**
 * 函数：updateMailStatus
 * 说明：更新玩家邮件状态
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | openId | string | 是 | 用户openId |
    | appId | string | 是 | 游戏ID |
    | mailId | string | 是 | 邮件ID |
    | action | string | 是 | 操作类型：read/receive/delete |
 */

async function updateMailStatusHandler(event, context) {
    // 请求参数
    const { openId, appId, mailId, action } = event;

    const validActions = ['read', 'receive', 'delete'];
    if (!validActions.includes(action)) {
        return {
            code: 400,
            msg: "无效的操作类型",
            timestamp: Date.now()
        };
    }

    try {
        const db = cloud.database();
        const now = new Date();
        
        // 验证用户存在
        const userResult = await db.collection('users').where({ openId, appId }).get();
        if (!userResult.data || userResult.data.length === 0) {
            return {
                code: 404,
                msg: "用户不存在",
                timestamp: Date.now()
            };
        }

        // 验证邮件存在且可用
        const mailResult = await db.collection('mails').where({ 
            mailId, 
            appId, 
            status: 'active' 
        }).get();
        
        if (!mailResult.data || mailResult.data.length === 0) {
            return {
                code: 404,
                msg: "邮件不存在或不可用",
                timestamp: Date.now()
            };
        }

        const mail = mailResult.data[0];
        const user = userResult.data[0];
        const userLevel = user.level || 0;

        // 检查邮件是否过期
        if (mail.expireTime && new Date(mail.expireTime) < now) {
            return {
                code: 400,
                msg: "邮件已过期",
                timestamp: Date.now()
            };
        }

        // 检查用户是否有权限接收此邮件
        let hasPermission = false;
        if (mail.targetType === 'all') {
            hasPermission = true;
        } else if (mail.targetType === 'user') {
            hasPermission = mail.targetUsers && mail.targetUsers.includes(openId);
        } else if (mail.targetType === 'level') {
            const minLevel = mail.minLevel || 0;
            const maxLevel = mail.maxLevel || 999;
            hasPermission = userLevel >= minLevel && userLevel <= maxLevel;
        }

        if (!hasPermission) {
            return {
                code: 403,
                msg: "没有权限操作此邮件",
                timestamp: Date.now()
            };
        }

        // 查找用户邮件状态记录
        const statusResult = await db.collection('user_mail_status')
            .where({ openId, appId, mailId })
            .get();

        let statusRecord = statusResult.data[0];
        const updateData = {
            updateTime: formatTime(now)
        };

        // 根据操作类型更新状态
        if (action === 'read') {
            updateData.isRead = true;
            updateData.readTime = formatTime(now);
        } else if (action === 'receive') {
            // 领取奖励前必须先阅读
            if (!statusRecord || !statusRecord.isRead) {
                return {
                    code: 400,
                    msg: "请先阅读邮件",
                    timestamp: Date.now()
                };
            }
            
            // 检查是否已领取
            if (statusRecord.isReceived) {
                return {
                    code: 400,
                    msg: "奖励已领取",
                    timestamp: Date.now()
                };
            }
            
            updateData.isReceived = true;
            updateData.receiveTime = formatTime(now);
            
            // 这里可以添加奖励发放逻辑
            // TODO: 根据 mail.rewards 发放奖励给用户
            
        } else if (action === 'delete') {
            updateData.isDeleted = true;
            updateData.deleteTime = formatTime(now);
        }

        // 更新或创建状态记录
        if (statusRecord) {
            // 更新现有记录
            await db.collection('user_mail_status')
                .where({ openId, appId, mailId })
                .update(updateData);
        } else {
            // 创建新记录
            const newStatus = {
                statusId: generateId(),
                openId,
                appId,
                mailId,
                isRead: action === 'read',
                isReceived: false,
                isDeleted: action === 'delete',
                readTime: action === 'read' ? formatTime(now) : null,
                receiveTime: null,
                deleteTime: action === 'delete' ? formatTime(now) : null,
                createTime: formatTime(now),
                updateTime: formatTime(now)
            };
            
            await db.collection('user_mail_status').add(newStatus);
        }

        // 构建返回数据
        const responseData = {
            mailId,
            action,
            timestamp: Date.now()
        };

        // 如果是领取奖励，返回奖励信息
        if (action === 'receive' && mail.rewards && mail.rewards.length > 0) {
            responseData.rewards = mail.rewards;
        }

        return {
            code: 0,
            msg: `邮件${action === 'read' ? '已读' : action === 'receive' ? '已领取' : '已删除'}`,
            timestamp: Date.now(),
            data: responseData
        };
    } catch (error) {
        console.error('更新邮件状态失败:', error);
        return {
            code: 500,
            msg: "操作失败",
            timestamp: Date.now()
        };
    }
}

// 导出处理函数
exports.main = requirePermission(updateMailStatusHandler, 'mail_manage');