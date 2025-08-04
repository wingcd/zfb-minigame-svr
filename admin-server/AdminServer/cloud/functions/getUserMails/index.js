const cloud = require("@alipay/faas-server-sdk");

/**
 * 函数：getUserMails
 * 说明：获取玩家邮件列表
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | playerId | string | 是 | 用户playerId |
    | appId | string | 是 | 游戏ID |
    | status | string | 否 | 邮件状态筛选 |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |

    返回数据：
    {
        code: 0,
        msg: "获取邮件列表成功",
        timestamp: Date.now(),
        data: {
            list: paginatedMails,
            total: total,
            page: page,
            pageSize: pageSize,
            hasMore: start + pageSize < total
        }
    }   
 */

async function getUserMailsHandler(event, context) {
    // 请求参数
    const { playerId, appId, status, page = 1, pageSize = 20 } = event;

    // 参数验证
    if (!playerId || !appId) {
        return {
            code: 400,
            msg: "缺少必要参数",
            timestamp: Date.now()
        };
    }

    try {
        const db = cloud.database();
        const now = new Date();
        
        // 获取用户信息
        const userTableName = `user_${appId}`;
        const userResult = await db
            .collection(userTableName)
            .where({ playerId })
            .get();
            
        if (!userResult || userResult.length === 0) {
            return {
                code: 404,
                msg: "用户不存在",
                timestamp: Date.now()
            };
        }
        
        const user = userResult[0];
        const userLevel = user.level || 0;

        // 构建查询条件
        let whereCondition = {
            appId,
            status: 'active'
        };

        // 查询可用的邮件
        const mailsQuery = db.collection('mails').where(whereCondition);
        const mailsResult = await mailsQuery.get();
        
        // 筛选符合条件的邮件
        const availableMails = mailsResult.filter(mail => {
            // 检查过期时间
            if (mail.expireTime && new Date(mail.expireTime) < now) {
                return false;
            }
            
            // 检查发布时间
            if (mail.publishTime && new Date(mail.publishTime) > now) {
                return false;
            }
            
            // 检查目标类型
            if (mail.targetType === 'all') {
                return true;
            } else if (mail.targetType === 'user') {
                return mail.targetUsers && mail.targetUsers.includes(playerId);
            } else if (mail.targetType === 'level') {
                const minLevel = mail.minLevel || 0;
                const maxLevel = mail.maxLevel || 999;
                return userLevel >= minLevel && userLevel <= maxLevel;
            }
            
            return false;
        });

        // 获取用户邮件状态
        const userMailsResult = await db.collection('user_mail_status')
            .where({ playerId, appId })
            .get();
        
        const userMailStatus = {};
        userMailsResult.forEach(item => {
            userMailStatus[item.mailId] = item;
        });

        // 合并邮件数据和用户状态
        let userMails = availableMails.map(mail => {
            const status = userMailStatus[mail.mailId] || {
                isRead: false,
                isReceived: false,
                isDeleted: false,
                readTime: null,
                receiveTime: null
            };
            
            return {
                mailId: mail.mailId,
                title: mail.title,
                content: mail.content,
                type: mail.type,
                rewards: mail.rewards || [],
                publishTime: mail.publishTime,
                expireTime: mail.expireTime,
                isRead: status.isRead,
                isReceived: status.isReceived,
                isDeleted: status.isDeleted,
                readTime: status.readTime,
                receiveTime: status.receiveTime,
                status: status.isDeleted ? 'deleted' : 
                       status.isReceived ? 'received' : 
                       status.isRead ? 'read' : 'unread'
            };
        });

        // 过滤已删除的邮件（除非特别查询）
        if (status !== 'deleted') {
            userMails = userMails.filter(mail => !mail.isDeleted);
        }

        // 按状态筛选
        if (status) {
            userMails = userMails.filter(mail => mail.status === status);
        }

        // 排序：未读优先，然后按发布时间倒序
        userMails.sort((a, b) => {
            if (a.isRead !== b.isRead) {
                return a.isRead ? 1 : -1;
            }
            return new Date(b.publishTime) - new Date(a.publishTime);
        });

        // 分页
        const total = userMails.length;
        const start = (page - 1) * pageSize;
        const paginatedMails = userMails.slice(start, start + pageSize);

        return {
            code: 0,
            msg: "获取邮件列表成功",
            timestamp: Date.now(),
            data: {
                list: paginatedMails,
                total,
                page,
                pageSize,
                hasMore: start + pageSize < total
            }
        };
    } catch (error) {
        console.error('获取用户邮件失败:', error);
        return {
            code: 500,
            msg: "获取邮件失败",
            timestamp: Date.now()
        };
    }
}

// 导出处理函数
exports.main = getUserMailsHandler; 