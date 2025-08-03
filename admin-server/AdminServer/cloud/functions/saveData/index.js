const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：saveData
 * 说明：保存数据
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 小程序id |
    | playerId | string | 是 | 玩家id |
    | data | string | 是 | 数据 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "playerId": "600015",
        "data": "{\"score\": 100}"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
    }
 */

exports.main = async (event, context) => {

    let appId;
    //玩家id
    let playerId;
    let data;

    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
    }

    var parmErr = common.hash.CheckParams(event);
    if(parmErr) {
        ret.code = 4001;
        ret.msg = "参数错误, error code:" + parmErr;
        return ret;
    }

    //参数校验 字段存在  为空   类型
    if (!event.hasOwnProperty("appId") || !event.appId || typeof event.appId != "string") {
        ret.code = 4001;
        ret.msg = "参数[appId]错误"
        return ret;
    }

    if (!event.hasOwnProperty("playerId") || !event.playerId || typeof event.playerId != "string") {
        ret.code = 4001;
        ret.msg = "参数[playerId]错误"
        return ret;
    }

    if (!event.hasOwnProperty("data") || !event.data || typeof event.data != "string") {
        ret.code = 4001;
        ret.msg = "参数[data]错误"
        return ret;
    }

    //请求参数
    appId = event.appId.trim();  //app id
    playerId = event.playerId.trim();  //玩家id
    data = event.data.trim();  //数据

    // 获取 cloud 环境中的 mongoDB 数据库对象
    const db = cloud.database();
    let userTableName = `user_${appId}`;
    let collection = db.collection(userTableName);
    //获取玩家信息
    let queryList = await collection
        .where({
            "playerId": playerId,
        })
        .get();

    if(queryList.length === 0){
        ret.code = 4004;
        ret.msg = "用户不存在";
        return ret;
    }

    //更新
    try {
        let now = moment().format("YYYY-MM-DD HH:mm:ss");
        await collection.doc(queryList[0]._id)
            .update({
                data: {
                    "gmtModify": now,
                    "data": data
                }
            });
    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
};