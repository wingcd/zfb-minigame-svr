const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：createApp
 * 说明：创建新应用
 * 权限：需要 app_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID（唯一标识） |
    | appName | string | 是 | 应用名称 |
    | description | string | 否 | 应用描述 |
    | channelAppKey | string | 否 | 渠道应用密钥 |
    | appSecret | string | 否 | 应用密钥 |
    | category | string | 否 | 应用分类 |
    | status | string | 否 | 状态 (active/inactive) |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "appName": "测试游戏",
        "description": "一个测试用的小游戏",
        "channelAppKey": "channel_key_123",
        "category": "game"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "id": "app_config_id_123456",
            "appId": "test_game_001",
            "appName": "测试游戏",
            "description": "一个测试用的小游戏",
            "status": "active",
            "createTime": "2023-10-01 10:00:00"
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4002: 应用ID已存在
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function createAppHandler(event, context) {
    let appId = event.appId;
    let appName = event.appName;
    let description = event.description || '';
    let channelAppKey = event.channelAppKey || '';
    let appSecret = event.appSecret || '';
    let category = event.category || 'game';
    let status = event.status || 'active';

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!appId || typeof appId !== "string" || appId.length < 3) {
        ret.code = 4001;
        ret.msg = "应用ID必须至少3个字符";
        return ret;
    }

    if (!appName || typeof appName !== "string" || appName.length < 2) {
        ret.code = 4001;
        ret.msg = "应用名称必须至少2个字符";
        return ret;
    }

    // 验证状态
    const validStatuses = ['active', 'inactive', 'pending'];
    if (!validStatuses.includes(status)) {
        ret.code = 4001;
        ret.msg = "无效的状态值";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查应用ID是否已存在
        const existingApps = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (existingApps.length > 0) {
            ret.code = 4002;
            ret.msg = "应用ID已存在";
            return ret;
        }

        // 生成应用密钥（如果没有提供）
        if (!appSecret) {
            appSecret = require('crypto').randomBytes(32).toString('hex');
        }

        // 创建应用
        const newApp = {
            appId: appId,
            appName: appName,
            description: description,
            channelAppKey: channelAppKey,
            appSecret: appSecret,
            category: category,
            status: status,
            createTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            createdBy: event.adminInfo.username,
            userCount: 0,
            scoreCount: 0
        };

        const addResult = await db.collection('app_config').add({
            data: newApp
        });

        // 创建对应的用户表
        const userTableName = `user_${appId}`;
        try {
            // 创建用户表的索引
            await db.collection(userTableName).createIndex({
                keys: { openId: 1 },
                unique: true
            });
        } catch (e) {
            // 表可能已存在，忽略错误
        }

        // 记录操作日志
        await logOperation(event.adminInfo, 'CREATE', 'APP', {
            appId: appId,
            appName: appName,
            category: category,
            createdBy: event.adminInfo.username,
            severity: 'MEDIUM'
        });

        ret.msg = "创建成功";
        ret.data = {
            id: addResult._id,
            appId: appId,
            appName: appName,
            description: description,
            status: status,
            createTime: newApp.createTime
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(createAppHandler, 'app_manage');
exports.main = mainFunc; 