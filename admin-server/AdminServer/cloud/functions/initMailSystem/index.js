const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission } = require("./common/auth");

/**
 * 函数：initMailSystem
 * 说明：初始化邮件系统（创建必要的数据库集合）
 * 权限：需要管理员权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | force | boolean | 否 | 强制重新初始化（默认：false） |
 * 
 * 说明：
 * - 自动创建必要的数据库集合：mails, user_mail_status
 * - 创建必要的索引以提高查询性能
 * - 如果集合已存在，默认不会重复创建
 * - 强制模式会清除所有现有邮件数据
 * 
 * 创建的集合：
 * - mails: 邮件信息表
 * - user_mail_status: 用户邮件状态表
 * 
 * 测试数据：
    {
        "force": false
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "邮件系统初始化完成",
        "timestamp": 1603991234567,
        "data": {
            "createdCollections": 2,
            "createdIndexes": 4,
            "warning": "邮件系统已成功初始化！"
        }
    }
    
 * 错误码：
 * - 4003: 系统已初始化（非强制模式）
 * - 5001: 服务器内部错误
 */

async function initMailSystemHandler(event, context) {
    let force = event.force || false;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    const db = cloud.database();

    try {
        // 创建必要的集合（表）
        const requiredCollections = [
            'mails',                // 邮件信息表
            'user_mail_status'      // 用户邮件状态表
        ];

        let createdCollections = 0;
        
        for (let collectionName of requiredCollections) {
            try {
                await db.getCollection(collectionName);
                console.log(`集合 ${collectionName} 已存在`);
            } catch (e) {
                if (e.message == "not found collection") {
                    await db.createCollection(collectionName);
                    createdCollections++;
                    console.log(`集合 ${collectionName} 创建成功`);
                } else {
                    console.log(`集合 ${collectionName} 检查失败:`, e.message);
                }
            }
        }

        // 安全检查：如果系统已经初始化，且非强制模式，则检查
        if (!force) {
            try {
                const existingMails = await db.collection('mails').count();
                if (existingMails.total > 0) {
                    ret.code = 4003;
                    ret.msg = "邮件系统已初始化，如需重新初始化请设置 force=true";
                    return ret;
                }
            } catch (e) {
                // 如果集合不存在，忽略错误
            }
        }

        // 如果是强制模式，先清理现有数据
        if (force) {
            try {
                await db.collection('mails').where({}).remove();
                await db.collection('user_mail_status').where({}).remove();
                console.log('强制模式：已清理现有邮件数据');
            } catch (e) {
                // 忽略删除错误（表可能不存在）
                console.log('清理数据时出错:', e.message);
            }
        }

        // 创建索引以提高查询性能
        let createdIndexes = 0;
        
        try {
            // mails 集合索引
            const mailsIndexes = [
                { appId: 1, status: 1, createTime: -1 },
                { appId: 1, targetType: 1, publishTime: 1 },
                { expireTime: 1 },
                { publishTime: 1 }
            ];

            for (let index of mailsIndexes) {
                try {
                    await db.collection('mails').createIndex(index);
                    createdIndexes++;
                } catch (e) {
                    // 索引可能已存在，忽略错误
                }
            }

            // user_mail_status 集合索引
            const userMailStatusIndexes = [
                { openId: 1, appId: 1 },
                { mailId: 1, openId: 1 }
            ];

            for (let index of userMailStatusIndexes) {
                try {
                    await db.collection('user_mail_status').createIndex(index);
                    createdIndexes++;
                } catch (e) {
                    // 索引可能已存在，忽略错误
                }
            }

        } catch (e) {
            console.log('创建索引时出错:', e.message);
        }

        ret.msg = "邮件系统初始化完成";
        ret.data = {
            createdCollections: createdCollections,
            createdIndexes: createdIndexes,
            warning: createdCollections > 0 ? "邮件系统已成功初始化！" : "邮件系统集合已存在"
        };

    } catch (e) {
        console.error('邮件系统初始化失败:', e);
        ret.code = 5001;
        ret.msg = "邮件系统初始化失败: " + e.message;
        return ret;
    }

    return ret;
}

// 导出处理函数
exports.main = requirePermission(initMailSystemHandler, 'mail_manage'); 