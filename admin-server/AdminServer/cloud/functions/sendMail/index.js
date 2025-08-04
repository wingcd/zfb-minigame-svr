const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 格式化时间
 * @param {Date|string} date - 日期对象或字符串
 * @returns {string} - 格式化后的时间字符串
 */
function formatTime(date = new Date()) {
    const d = new Date(date);
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    const seconds = String(d.getSeconds()).padStart(2, '0');
    
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}


/**
 * 函数：sendMail
 * 说明：发布邮件（设置为可用状态，玩家可主动获取）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | mailId | string | 是 | 邮件ID |
 */

async function sendMailHandler(event, context) {
    // 请求参数
    const { mailId } = event;

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
        
        // 查找邮件
        const mailResult = await db.collection('mails').where({ mailId }).get();
        if (!mailResult || mailResult.length === 0) {
            return {
                code: 404,
                msg: "邮件不存在",
                timestamp: Date.now()
            };
        }

        const mail = mailResult[0];
        
        // 检查邮件状态
        if (mail.status === 'active') {
            return {
                code: 400,
                msg: "邮件已发布",
                timestamp: Date.now()
            };
        }

        if (mail.status === 'expired') {
            return {
                code: 400,
                msg: "邮件已过期",
                timestamp: Date.now()
            };
        }

        // 验证必要字段
        if (!mail.appId) {
            return {
                code: 400,
                msg: "邮件缺少游戏ID",
                timestamp: Date.now()
            };
        }

        // 统计潜在收件人数量（用于显示）
        let potentialRecipients = 0;
        const { targetType, targetUsers, minLevel, maxLevel, appId } = mail;

        try {
            if (targetType === 'all') {
                // 该游戏的所有用户
                const usersResult = await db.collection('users').where({ appId }).count();
                potentialRecipients = usersResult.total;
            } else if (targetType === 'user') {
                // 指定用户
                potentialRecipients = (targetUsers || []).length;
            } else if (targetType === 'level') {
                // 指定等级范围的用户
                const usersResult = await db.collection('users')
                    .where({
                        appId,
                        level: db.command.gte(minLevel || 0).and(db.command.lte(maxLevel || 999))
                    })
                    .count();
                potentialRecipients = usersResult.total;
            }
        } catch (error) {
            console.warn('统计潜在收件人失败:', error);
            potentialRecipients = 0;
        }

        // 更新邮件状态为已发布
        await db.collection('mails').where({ mailId }).update({
            data: {
                status: 'active',
                publishTime: formatTime(now),
                potentialRecipients,
                updateTime: formatTime(now)
            }
        });
        
        // 记录操作日志
        const adminInfo = context.adminInfo || { username: 'system' };
        await logOperation(adminInfo.username, 'publish_mail', {
            mailId,
            title: mail.title,
            potentialRecipients
        });

        return {
            code: 0,
            msg: "邮件发布成功",
            timestamp: Date.now(),
            data: {
                mailId,
                potentialRecipients
            }
        };
    } catch (error) {
        console.error('发布邮件失败:', error);
        return {
            code: 500,
            msg: "发布邮件失败",
            timestamp: Date.now()
        };
    }
}

// 导出处理函数
exports.main = requirePermission(sendMailHandler, 'mail_manage');