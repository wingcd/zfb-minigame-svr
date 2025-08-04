const cloud = require("@alipay/faas-server-sdk");
const { requirePermission } = require("./common/auth");

/**
 * 函数：getAllMails
 * 说明：获取邮件列表
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |
    | title | string | 否 | 邮件标题搜索 |
    | type | string | 否 | 邮件类型筛选 |
    | status | string | 否 | 状态筛选 |
    | appId | string | 否 | 应用ID筛选 |
 */

async function getAllMailsHandler(event, context) {
    // 请求参数
    const page = event.page || 1;
    const pageSize = Math.min(event.pageSize || 20, 100); // 限制最大100条
    const { title, type, status, appId } = event;

    try {
        const db = cloud.database();
        const _ = db.command;
        
        // 构建查询条件
        let where = {};
        
        if (title) {
            where.title = new RegExp(title, 'i');
        }
        
        if (type) {
            where.type = type;
        }
        
        if (status) {
            where.status = status;
        }
        
        if (appId) {
            where.appId = appId;
        }

        // 查询总数
        const countResult = await db.collection('mails').where(where).count();
        const total = countResult.total;

        // 查询列表数据
        const listResult = await db.collection('mails')
            .where(where)
            .orderBy('createTime', 'desc')
            .skip((page - 1) * pageSize)
            .limit(pageSize)
            .get();

        return {
            code: 0,
            msg: "success",
            timestamp: Date.now(),
            data: {
                list: listResult || [],
                total,
                page,
                pageSize,
                totalPages: Math.ceil(total / pageSize)
            }
        };
    } catch (error) {
        console.error('获取邮件列表失败:', error);
        return {
            code: 500,
            msg: "获取邮件列表失败",
            timestamp: Date.now()
        };
    }
}

// 导出处理函数
exports.main = requirePermission(getAllMailsHandler, 'mail_manage');