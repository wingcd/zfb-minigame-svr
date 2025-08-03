const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：unbanUser
 * 说明：解封用户
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | playerId | string | 是 | 玩家ID |
    | reason | string | 否 | 解封原因 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "playerId": "player001",
        "reason": "申诉成功"
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
    let reason = event.reason || "管理员解封";

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

        // 检查用户是否被封禁
        if (!user.banned) {
            ret.code = 4003;
            ret.msg = "用户未被封禁";
            return ret;
        }

        // 更新用户解封状态
        await collection.doc(user._id)
            .update({
                data: {
                    banned: false,
                    unbanReason: reason,
                    unbanTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                    // 清除封禁相关字段
                    banReason: null,
                    banTime: null,
                    banUntil: null,
                    gmtModify: moment().format("YYYY-MM-DD HH:mm:ss")
                }
            });

        ret.msg = "解封成功";
        ret.data = {
            playerId: playerId,
            unbanReason: reason,
            unbanTime: moment().format("YYYY-MM-DD HH:mm:ss")
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 