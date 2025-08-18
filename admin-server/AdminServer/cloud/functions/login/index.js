const cloud = require("@alipay/faas-server-sdk");
const { randomUUID } = require('crypto');
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：login
 * 说明：登录
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 小程序id |
    | code | string | 是 | 玩家id |
  * 测试数据
    {
        "appId": "13c6994c-9beb-4d6d-8fcb-8500a1814f66",
        "code": "player001"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "token": "5f9b3b7b7b4b4b0001b4b4b4",
            "playerId": "600015",
            "isNew": true,
            "data": null
        }
    }
 */

exports.main = async (event, context) => {
    let appId;
    let openId;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    var parmErr = common.hash.CheckParams(event, true);
    if(parmErr) {
        ret.code = 4001;
        ret.msg = "参数错误, error code:" + parmErr;
        return ret;
    }

    //参数校验 字段存在  为空   类型
    if (!event.hasOwnProperty("appId") || (!event.appId || typeof event.appId != "string")) {
        ret.code = 4001;
        ret.msg = "参数[appId]错误"
        return ret;
    }

    if (!event.hasOwnProperty("code") || (!event.code || typeof event.code != "string")) {
        ret.code = 4001;
        ret.msg = "参数[code]错误"
        return ret;
    }

    appId = event.appId;
    openId = event.code;

    const db = cloud.database();

    let userTableName = `user_${appId}`;
    let collection = null;
    try {
        collection = db.collection(userTableName);
    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    //获取玩家信息
    let queryList = await collection
        .where({
            "openId": openId,
        })
        .get();
    let token = randomUUID();
    
    let row = null;
    //更新 or 插入 
    if (queryList.length === 0) {
        // 生成唯一的playerId，采用重试机制确保唯一性
        let playerId;
        let maxRetries = 5;
        let retryCount = 0;
        
        while (retryCount < maxRetries) {
            try {
                // 使用时间戳 + 随机数 + 计数器的方式生成ID
                let count = await collection.where({}).count();
                let timestamp = Date.now().toString().slice(-6); // 取时间戳后6位
                let random = Math.floor(Math.random() * 1000).toString().padStart(3, '0'); // 3位随机数
                playerId = `6${timestamp}${random}`;
                
                // 检查ID是否已存在
                let existingUser = await collection.where({ "playerId": playerId }).get();
                if (existingUser.length === 0) {
                    break; // ID唯一，跳出循环
                }
                
                retryCount++;
                if (retryCount >= maxRetries) {
                    // 如果重试多次仍然冲突，使用UUID作为后备方案
                    playerId = `6${randomUUID().replace(/-/g, '').slice(0, 9)}`;
                }
            } catch (e) {
                retryCount++;
                if (retryCount >= maxRetries) {
                    throw e;
                }
            }
        }

        //插入
        try {
            let now = moment().format("YYYY-MM-DD HH:mm:ss");
            let dt = await collection.add({
                data: {
                    "openId": openId,
                    "playerId": playerId,
                    "token": token,
                    "test": 1,
                    "data": null,
                    "gmtCreate": now,
                    "gmtModify": now,
                }
            });

            row = await collection.doc(dt._id).get();
        } catch (e) {
            ret.code = 5001;
            ret.msg = e.message;
            ret.data.playerId = playerId;
            return ret;
        }
        
        ret.data.isNew = true;
    } else {
        //更新
        try {
            let now = moment().format("YYYY-MM-DD HH:mm:ss");
            await collection.doc(queryList[0]._id)
                .update({
                    data: {
                        "token": token,
                        "gmtModify": now,
                        "test": 1,
                    }
                });

            row = queryList[0];
        } catch (e) {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }

    ret.data.data = row.data;
    ret.data.playerId = row.playerId;
    ret.data.token = token;
    ret.data.openId = row.openId;
    return ret;
};