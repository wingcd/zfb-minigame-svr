const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：setDetail
 * 说明：设置用户详细信息
 * 权限：需要 user_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | openId | string | 是 | 用户openId |
    | playerId | string | 否 | 玩家ID |
    | nickName | string | 否 | 用户昵称 |
    | avatarUrl | string | 否 | 头像URL |
    | gender | number | 否 | 性别(0:未知,1:男,2:女) |
    | province | string | 否 | 省份 |
    | city | string | 否 | 城市 |
    | userData | string | 否 | 用户游戏数据(JSON字符串) |
    | banned | boolean | 否 | 是否封禁 |
    | banReason | string | 否 | 封禁原因 |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "openId": "user_openid_123456",
        "playerId": "player_123",
        "nickName": "新昵称",
        "avatarUrl": "https://example.com/avatar.jpg",
        "gender": 1,
        "province": "广东省",
        "city": "深圳市",
        "userData": "{\"level\": 15, \"coins\": 5000, \"items\": [\"sword\", \"shield\", \"potion\"]}"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "updated": true,
            "updatedFields": ["nickName", "userData"]
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 用户不存在
 * - 4005: 无效的JSON数据
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function setDetailHandler(event, context) {
    let appId = event.appId;
    let openId = event.openId;
    let playerId = event.playerId;
    let nickName = event.nickName;
    let avatarUrl = event.avatarUrl;
    let gender = event.gender;
    let province = event.province;
    let city = event.city;
    let userData = event.userData;
    let banned = event.banned;
    let banReason = event.banReason;

    console.log("setDetail event:", event);

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
    if (userData !== undefined && userData !== null && userData !== "") {
        try {
            JSON.parse(userData);
        } catch (e) {
            ret.code = 4005;
            ret.msg = "无效的JSON数据格式";
            return ret;
        }
    }

    // 验证gender值
    if (gender !== undefined && gender !== null) {
        if (![0, 1, 2].includes(Number(gender))) {
            ret.code = 4001;
            ret.msg = "性别参数无效，必须为0、1或2";
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
        const updatedFields = [];

        // 构建更新数据
        let updateData = {
            gmtModify: moment().format("YYYY-MM-DD HH:mm:ss")
        };

        // 更新各个字段
        if (playerId !== undefined && playerId !== null) {
            updateData.playerId = playerId;
            updatedFields.push('playerId');
        }

        if (nickName !== undefined && nickName !== null) {
            updateData.nickName = nickName;
            updatedFields.push('nickName');
        }

        if (avatarUrl !== undefined && avatarUrl !== null) {
            updateData.avatarUrl = avatarUrl;
            updatedFields.push('avatarUrl');
        }

        if (gender !== undefined && gender !== null) {
            updateData.gender = Number(gender);
            updatedFields.push('gender');
        }

        if (province !== undefined && province !== null) {
            updateData.province = province;
            updatedFields.push('province');
        }

        if (city !== undefined && city !== null) {
            updateData.city = city;
            updatedFields.push('city');
        }

        if (userData !== undefined && userData !== null && userData !== "") {
            let userInfo = userData.userInfo;
            if(userInfo) {
                let keys = Object.keys(userInfo);
                for (let key of keys) {
                    updateData[key] = userInfo[key];
                    updatedFields.push(key);
                }
            }
        }

        // 处理封禁状态
        if (banned !== undefined && banned !== null) {
            updateData.banned = Boolean(banned);
            updatedFields.push('banned');
            
            if (banned && banReason) {
                updateData.banReason = banReason;
                updateData.banTime = moment().format("YYYY-MM-DD HH:mm:ss");
                updateData.bannedBy = event.adminInfo.username;
                updatedFields.push('banReason', 'banTime', 'bannedBy');
            } else if (!banned) {
                updateData.banReason = null;
                updateData.banTime = null;
                updateData.bannedBy = null;
                updateData.unbanTime = moment().format("YYYY-MM-DD HH:mm:ss");
                updateData.unbannedBy = event.adminInfo.username;
                updatedFields.push('unbanTime', 'unbannedBy');
            }
        }

        // 如果没有任何字段需要更新
        if (updatedFields.length === 0) {
            ret.code = 4001;
            ret.msg = "没有提供需要更新的字段";
            return ret;
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
            playerId: playerId,
            userNickName: oldUserData.nickName,
            updatedFields: updatedFields,
            changes: updateData,
            updatedBy: event.adminInfo.username
        };

        // 如果是封禁/解封操作，提高日志级别
        const severity = banned !== undefined ? 'HIGH' : 'MEDIUM';
        logData.severity = severity;

        await logOperation(event.adminInfo, 'UPDATE', 'USER_DETAIL', logData);

        ret.data = {
            updated: true,
            updatedFields: updatedFields
        };
        ret.msg = `成功更新用户信息，更新字段：${updatedFields.join(', ')}`;

    } catch (e) {
        console.error('setDetail error:', e);
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(setDetailHandler, 'user_manage');
exports.main = mainFunc; 