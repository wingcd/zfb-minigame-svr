const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：deleteMail
 * 说明：删除邮件
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | mailId | string | 是 | 邮件ID |
 */

async function deleteMailHandler(event, context) {
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
        
        // 查找邮件是否存在
        const mailResult = await db.collection('mails').where({ mailId }).get();
        if (!mailResult.data || mailResult.data.length === 0) {
            return {
                code: 404,
                msg: "邮件不存在",
                timestamp: Date.now()
            };
        }

        const mail = mailResult.data[0];
        
        // 检查邮件状态，已发布的邮件不允许删除
        if (mail.status === 'active') {
            return {
                code: 400,
                msg: "已发布的邮件不能删除",
                timestamp: Date.now()
            };
        }

        // 删除邮件
        await db.collection('mails').where({ mailId }).remove();
        
        // 记录操作日志
        await logOperation(authResult.admin.username, 'delete_mail', {
            mailId,
            title: mail.title
        });

        return {
            code: 0,
            msg: "邮件删除成功",
            timestamp: Date.now()
        };
    } catch (error) {
        console.error('删除邮件失败:', error);
        return {
            code: 500,
            msg: "删除邮件失败",
            timestamp: Date.now()
        };
    }
}

// 导出处理函数
exports.main = requirePermission(deleteMailHandler, 'mail_manage');