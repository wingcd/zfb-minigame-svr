const cloud = require("@alipay/faas-server-sdk");
const { requirePermission } = require("./common/auth");

/**
 * 函数：getMailStats
 * 说明：获取邮件统计数据
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | mailId | string | 否 | 邮件ID，为空则获取全部邮件统计 |
    | appId | string | 否 | 游戏ID，为空则获取全部游戏统计 |

    返回数据：
    {
        code: 0,
        msg: "获取邮件统计成功",
        timestamp: Date.now(),
        data: {
            mail: {
                mailId: mail.mailId,
                title: mail.title,
                type: mail.type,
                targetType: mail.targetType,
                status: mail.status,
                publishTime: mail.publishTime,
                expireTime: mail.expireTime,
                potentialRecipients: mail.potentialRecipients || 0      
            },
            stats: {
                totalUsers: 0,
                readUsers: 0,
                receivedUsers: 0,
                deletedUsers: 0,
                readRate: 0,
                receiveRate: 0,
                deleteRate: 0
            }
        }
    }
 */

async function getMailStatsHandler(event, context) {

    // 请求参数
    const { mailId, appId } = event;

    try {
        const db = cloud.database();
        
        if (mailId) {
            // 获取单个邮件的统计数据
            return await getSingleMailStats(db, mailId);
        } else {
            // 获取整体统计数据
            return await getOverallMailStats(db, appId);
        }
    } catch (error) {
        console.error('获取邮件统计失败:', error);
        return {
            code: 500,
            msg: "获取统计数据失败",
            timestamp: Date.now()
        };
    }
}

async function getSingleMailStats(db, mailId) {
    // 获取邮件信息
    const mailResult = await db.collection('mails').where({ mailId }).get();
    if (!mailResult || mailResult.length === 0) {
        return {
            code: 404,
            msg: "邮件不存在",
            timestamp: Date.now()
        };
    }

    const mail = mailResult[0];
    
    // 获取用户状态统计
    const statusResult = await db.collection('user_mail_status')
        .where({ mailId })
        .get();
    
    const stats = {
        totalUsers: 0,
        readUsers: 0,
        receivedUsers: 0,
        deletedUsers: 0,
        readRate: 0,
        receiveRate: 0,
        deleteRate: 0
    };

    if (statusResult && statusResult.length > 0) {
        stats.totalUsers = statusResult.length;
        stats.readUsers = statusResult.filter(item => item.isRead).length;
        stats.receivedUsers = statusResult.filter(item => item.isReceived).length;
        stats.deletedUsers = statusResult.filter(item => item.isDeleted).length;
        
        stats.readRate = stats.totalUsers > 0 ? (stats.readUsers / stats.totalUsers * 100).toFixed(2) : 0;
        stats.receiveRate = stats.totalUsers > 0 ? (stats.receivedUsers / stats.totalUsers * 100).toFixed(2) : 0;
        stats.deleteRate = stats.totalUsers > 0 ? (stats.deletedUsers / stats.totalUsers * 100).toFixed(2) : 0;
    }

    return {
        code: 0,
        msg: "获取邮件统计成功",
        timestamp: Date.now(),
        data: {
            mail: {
                mailId: mail.mailId,
                title: mail.title,
                type: mail.type,
                targetType: mail.targetType,
                status: mail.status,
                publishTime: mail.publishTime,
                expireTime: mail.expireTime,
                potentialRecipients: mail.potentialRecipients || 0
            },
            stats
        }
    };
}

async function getOverallMailStats(db, appId) {
    // 构建查询条件
    let whereCondition = {};
    if (appId) {
        whereCondition.appId = appId;
    }

    // 获取邮件总数统计
    const [
        totalMailsResult,
        activeMailsResult,
        draftMailsResult,
        expiredMailsResult
    ] = await Promise.all([
        db.collection('mails').where(whereCondition).count(),
        db.collection('mails').where({ ...whereCondition, status: 'active' }).count(),
        db.collection('mails').where({ ...whereCondition, status: 'draft' }).count(),
        db.collection('mails').where({ ...whereCondition, status: 'expired' }).count()
    ]);

    // 获取用户邮件状态统计
    let userStatsCondition = {};
    if (appId) {
        userStatsCondition.appId = appId;
    }

    const [
        totalInteractionsResult,
        readCountResult,
        receivedCountResult,
        deletedCountResult
    ] = await Promise.all([
        db.collection('user_mail_status').where(userStatsCondition).count(),
        db.collection('user_mail_status').where({ ...userStatsCondition, isRead: true }).count(),
        db.collection('user_mail_status').where({ ...userStatsCondition, isReceived: true }).count(),
        db.collection('user_mail_status').where({ ...userStatsCondition, isDeleted: true }).count()
    ]);

    const stats = {
        mailStats: {
            total: totalMailsResult.total,
            active: activeMailsResult.total,
            draft: draftMailsResult.total,
            expired: expiredMailsResult.total
        },
        interactionStats: {
            totalInteractions: totalInteractionsResult.total,
            readCount: readCountResult.total,
            receivedCount: receivedCountResult.total,
            deletedCount: deletedCountResult.total,
            readRate: totalInteractionsResult.total > 0 ? 
                (readCountResult.total / totalInteractionsResult.total * 100).toFixed(2) : 0,
            receiveRate: totalInteractionsResult.total > 0 ? 
                (receivedCountResult.total / totalInteractionsResult.total * 100).toFixed(2) : 0,
            deleteRate: totalInteractionsResult.total > 0 ? 
                (deletedCountResult.total / totalInteractionsResult.total * 100).toFixed(2) : 0
        }
    };

    return {
        code: 0,
        msg: "获取邮件统计成功",
        timestamp: Date.now(),
        data: stats
    };
}

// 导出处理函数
exports.main = requirePermission(getMailStatsHandler, 'mail_manage'); 