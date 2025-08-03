const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：banUser
 * 说明：封禁用户
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | playerId | string | 是 | 玩家ID |
    | reason | string | 否 | 封禁原因 |
    | duration | number | 否 | 封禁时长(小时)，0表示永久封禁 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "playerId": "player001",
        "reason": "违规行为",
        "duration": 24
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {}
    }
 */

exports.main = async (event, context) => {
    let appId = event.appId;
    let playerId = event.playerId;
    let reason = event.reason || "违规行为";
    let duration = event.duration || 0; // 0表示永久封禁

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    var parmErr = common.hash.CheckParams(event);
    if(parmErr) {
        ret.code = 4001;
        ret.msg = "参数错误, error code:" + parmErr;
        return ret;
    }

    // 参数校验
    if (!appId || typeof appId !== "string") {
        ret.code = 4001;
        ret.msg = "参数[appId]错误";
        return ret;
    }

    if (!playerId || typeof playerId !== "string") {
        ret.code = 4001;
        ret.msg = "参数[playerId]错误";
        return ret;
    }

    const db = cloud.database();
    const userTableName = `user_${appId}`;

    try {
        let collection = db.collection(userTableName);

        // 查询用户是否存在
        let userList = await collection
            .where({ playerId: playerId })
            .get();

        if (userList.length === 0) {
            ret.code = 4004;
            ret.msg = "用户不存在";
            return ret;
        }

        let user = userList[0];

        // 检查用户是否已被封禁
        if (user.banned) {
            ret.code = 4003;
            ret.msg = "用户已被封禁";
            return ret;
        }

        // 计算封禁到期时间
        let banUntil = null;
        if (duration > 0) {
            banUntil = moment().add(duration, 'hours').format("YYYY-MM-DD HH:mm:ss");
        }

        // 更新用户封禁状态
        await collection.doc(user._id)
            .update({
                data: {
                    banned: true,
                    banReason: reason,
                    banTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                    banUntil: banUntil,
                    gmtModify: moment().format("YYYY-MM-DD HH:mm:ss")
                }
            });

        ret.msg = "封禁成功";
        ret.data = {
            playerId: playerId,
            banReason: reason,
            banTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            banUntil: banUntil,
            permanent: duration === 0
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 