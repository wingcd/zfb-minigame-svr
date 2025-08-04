const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：updateUserData
 * 说明：更新用户数据
 * 权限：需要 user_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | openId | string | 是 | 用户openId |
    | userData | string | 否 | 用户游戏数据(JSON字符串) |
    | nickName | string | 否 | 用户昵称 |
    | banned | boolean | 否 | 是否封禁 |
    | banReason | string | 否 | 封禁原因 |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "openId": "user_openid_123456",
        "userData": "{\"level\": 10, \"coins\": 2000, \"items\": [\"sword\", \"shield\"]}",
        "nickName": "更新后的昵称"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {}
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 用户不存在
 * - 4005: 无效的JSON数据
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function updateUserDataHandler(event, context) {
    let appId = event.appId;
    let openId = event.openId;
    let userData = event.userData;
    let nickName = event.nickName;
    let banned = event.banned;
    let banReason = event.banReason;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!appId || typeof appId !== "string") {
        ret.code = 4001;
        ret.msg = "应用ID不能为空";
        return ret;
    }

    if (!openId || typeof openId !== "string") {
        ret.code = 4001;
        ret.msg = "用户openId不能为空";
        return ret;
    }

    // 验证userData是否为有效JSON
    if (userData !== undefined) {
        try {
            JSON.parse(userData);
        } catch (e) {
            ret.code = 4005;
            ret.msg = "无效的JSON数据格式";
            return ret;
        }
    }

    const db = cloud.database();

    try {
        // 验证应用是否存在
        const appList = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (appList.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        // 检查用户是否存在
        const userTableName = `user_${appId}`;
        const userList = await db.collection(userTableName)
            .where({ openId: openId })
            .get();

        if (userList.length === 0) {
            ret.code = 4004;
            ret.msg = "用户不存在";
            return ret;
        }

        const oldUserData = userList[0];

        // 构建更新数据
        let updateData = {
            gmtModify: moment().format("YYYY-MM-DD HH:mm:ss")
        };

        if (userData !== undefined) {
            updateData.userData = userData;
        }

        if (nickName !== undefined) {
            updateData.nickName = nickName;
        }

        if (banned !== undefined) {
            updateData.banned = banned;
            if (banned && banReason) {
                updateData.banReason = banReason;
                updateData.banTime = moment().format("YYYY-MM-DD HH:mm:ss");
                updateData.bannedBy = event.adminInfo.username;
            } else if (!banned) {
                updateData.banReason = null;
                updateData.banTime = null;
                updateData.bannedBy = null;
                updateData.unbanTime = moment().format("YYYY-MM-DD HH:mm:ss");
                updateData.unbannedBy = event.adminInfo.username;
            }
        }

        // 更新用户数据
        await db.collection(userTableName)
            .where({ openId: openId })
            .update({
                data: updateData
            });

        // 记录操作日志
        const logData = {
            appId: appId,
            openId: openId,
            userNickName: oldUserData.nickName,
            changes: updateData,
            updatedBy: event.adminInfo.username
        };

        // 如果是封禁/解封操作，提高日志级别
        const severity = banned !== undefined ? 'HIGH' : 'MEDIUM';
        logData.severity = severity;

        await logOperation(event.adminInfo, 'UPDATE', 'USER_DATA', logData);

        ret.msg = "更新成功";

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(updateUserDataHandler, 'user_manage');
exports.main = mainFunc; 