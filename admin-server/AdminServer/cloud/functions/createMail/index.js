const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

function formatTime(date) {
    return date.toISOString().replace('T', ' ').substring(0, 19);
}
function generateRandomString(length) {
    return Math.random().toString(36).substring(2, 2 + length);
}

/**
 * 函数：createMail
 * 说明：创建邮件
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | title | string | 是 | 邮件标题 |
    | content | string | 是 | 邮件内容 |
    | type | string | 否 | 邮件类型 (system/notice/reward) |
    | targetType | string | 是 | 目标类型 (all/user/level) |
    | targetUsers | array | 否 | 目标用户列表，targetType为user时必填 |
    | minLevel | number | 否 | 最小等级，targetType为level时必填 |
    | maxLevel | number | 否 | 最大等级，targetType为level时必填 |
    | rewards | array | 否 | 奖励列表 |
    | publishTime | string | 否 | 发送时间，为空则立即发送 |
    | expireTime | string | 否 | 过期时间，为空则7天后过期 |
    | appId | string | 是 | 应用ID |
 */

async function createMailHandler(event, context) {
    // 获取管理员信息
    const adminInfo = context.adminInfo || { username: 'system' };
    
    // 请求参数
    const {
        title,
        content,
        type = 'system',
        targetType,
        targetUsers = [],
        minLevel,
        maxLevel,
        rewards = [],
        publishTime,
        expireTime,
        appId
    } = event;

    // 参数验证
    if (!title || !content || !targetType || !appId) {
        return {
            code: 400,
            msg: "缺少必要参数",
            timestamp: Date.now()
        };
    }

    const validTypes = ['system', 'notice', 'reward'];
    if (!validTypes.includes(type)) {
        return {
            code: 400,
            msg: "邮件类型无效",
            timestamp: Date.now()
        };
    }

    const validTargetTypes = ['all', 'user', 'level'];
    if (!validTargetTypes.includes(targetType)) {
        return {
            code: 400,
            msg: "目标类型无效",
            timestamp: Date.now()
        };
    }

    if (targetType === 'user' && (!targetUsers || targetUsers.length === 0)) {
        return {
            code: 400,
            msg: "指定用户时必须提供用户列表",
            timestamp: Date.now()
        };
    }

    if (targetType === 'level' && (minLevel === undefined || maxLevel === undefined)) {
        return {
            code: 400,
            msg: "按等级发送时必须指定等级范围",
            timestamp: Date.now()
        };
    }

    try {
        const db = cloud.database();
        const now = new Date();
        
        // 处理发送时间逻辑
        let actualpublishTime;
        let status;
        
        if (publishTime) {
            // 指定了发送时间
            actualpublishTime = new Date(publishTime);
            if (actualpublishTime <= now) {
                // 指定时间已过期，立即发送
                actualpublishTime = now;
                status = 'pending'; // 待发送状态，可立即发布
            } else {
                // 定时发送
                status = 'scheduled'; // 定时状态
            }
        } else {
            // 未指定发送时间，立即发送
            actualpublishTime = now;
            status = 'pending'; // 待发送状态，可立即发布
        }
        
        // 设置过期时间（默认7天后）
        const defaultExpireTime = expireTime ? new Date(expireTime) : new Date(actualpublishTime.getTime() + 7 * 24 * 60 * 60 * 1000);
        
        // 生成邮件ID
        const mailId = `mail_${generateRandomString(16)}`;
        
        // 准备邮件数据
        const mailData = {
            mailId,
            title,
            content,
            type,
            targetType,
            targetUsers,
            minLevel,
            maxLevel,
            rewards,
            publishTime: formatTime(actualpublishTime),
            expireTime: formatTime(defaultExpireTime),
            appId,
            status, // pending(待发送)/scheduled(定时)/active(已发布)/expired(已过期)
            createTime: formatTime(now),
            updateTime: formatTime(now),
            createdBy: adminInfo.username
        };

        // 插入邮件记录
        await db.collection('mails').add(mailData);
        
        // 记录操作日志
        await logOperation(adminInfo.username, 'create_mail', {
            mailId,
            title,
            appId,
            publishTime: formatTime(actualpublishTime),
            status
        });

        return {
            code: 0,
            msg: "邮件创建成功",
            timestamp: Date.now(),
            data: {
                mailId,
                ...mailData
            }
        };
    } catch (error) {
        console.error('创建邮件失败:', error);
        return {
            code: 500,
            msg: "创建邮件失败",
            timestamp: Date.now()
        };
    }
}

// 导出处理函数
exports.main = requirePermission(createMailHandler, 'mail_manage');