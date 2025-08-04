const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("../common/auth");
const { formatTime } = require("./common");

/**
 * 函数：updateMail
 * 说明：更新邮件
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | mailId | string | 是 | 邮件ID |
    | title | string | 否 | 邮件标题 |
    | content | string | 否 | 邮件内容 |
    | type | string | 否 | 邮件类型 |
    | targetType | string | 否 | 目标类型 |
    | targetUsers | array | 否 | 目标用户列表 |
    | minLevel | number | 否 | 最小等级 |
    | maxLevel | number | 否 | 最大等级 |
    | rewards | array | 否 | 奖励列表 |
    | sendTime | string | 否 | 发送时间 |
    | expireTime | string | 否 | 过期时间 |
    | status | string | 否 | 状态 |
 */

async function updateMailHandler(event, context) {
    // 请求参数
    const { mailId, ...updateData } = event;

    // 参数验证
    if (!mailId) {
        return {
            code: 400,
            msg: "缺少邮件ID",
            timestamp: Date.now()
        };
    }

    try {
        const db = cloud.database();
        const now = new Date();
        
        // 查找邮件是否存在
        const mailResult = await db.collection('mails').where({ mailId }).get();
        if (!mailResult.data || mailResult.data.length === 0) {
            return {
                code: 404,
                msg: "邮件不存在",
                timestamp: Date.now()
            };
        }

        const currentMail = mailResult.data[0];
        
        // 如果邮件已发布，只允许更新状态和过期时间
        if (currentMail.status === 'active' && Object.keys(updateData).some(key => !['status', 'expireTime'].includes(key))) {
            return {
                code: 400,
                msg: "已发布的邮件只能更新状态和过期时间",
                timestamp: Date.now()
            };
        }

        // 准备更新数据
        const updateFields = {
            ...updateData,
            updateTime: formatTime(now)
        };

        // 处理时间字段
        if (updateData.sendTime) {
            updateFields.sendTime = formatTime(new Date(updateData.sendTime));
        }
        if (updateData.expireTime) {
            updateFields.expireTime = formatTime(new Date(updateData.expireTime));
        }

        // 更新邮件
        await db.collection('mails').where({ mailId }).update(updateFields);
        
        // 记录操作日志
        await logOperation(authResult.admin.username, 'update_mail', {
            mailId,
            changes: Object.keys(updateData)
        });

        return {
            code: 0,
            msg: "邮件更新成功",
            timestamp: Date.now()
        };
    } catch (error) {
        console.error('更新邮件失败:', error);
        return {
            code: 500,
            msg: "更新邮件失败",
            timestamp: Date.now()
        };
    }
}

// 导出处理函数
exports.main = requirePermission(updateMailHandler, 'mail_manage');