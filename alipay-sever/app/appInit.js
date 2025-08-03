const cloud = require("@alipay/faas-server-sdk");
const { randomUUID } = require('crypto');
const moment = require("moment")

// 请求参数
/**
 * 函数：appinit
 * 说明：初始化app
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appName | string | 是 | app名字 |
    | platform | string | 是 | 平台 |
    | force | boolean | 否 | 是否强制初始化 |
  * 测试数据
    {
        "appName": "小程序",
        "platform": "wechat",
        "appId": "5f9b3b7b7b4b4b0001b4b4b4",
        "appKey": "5f9b3b7b7b4b4b0001b4b4b4",
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "appName": "小程序",
            "appId": "5f9b3b7b7b4b4b0001b4b4b4"
        }
    }
 */

exports.main = async (event, context) => {
    // app名字
    let appName;
    let platform;
    let force;
    let channelAppId;
    let channelAppKey;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    //参数校验 字段存在  为空   类型
    if (event.hasOwnProperty("appName") && (!event.appName || typeof event.appName != "string")) {
        ret.code = 4001;
        ret.msg = "参数[appName]错误"
        return ret;
    }

    if (event.hasOwnProperty("platform") && (!event.platform || typeof event.platform != "string")) {
        ret.code = 4001;
        ret.msg = "参数[platform]错误"
        return ret;
    }

    if(event.hasOwnProperty("channelAppId") && (!event.appId || typeof event.appId != "string")) {
        ret.code = 4001;
        ret.msg = "参数[appId]错误"
        return ret;
    }

    if(event.hasOwnProperty("appKey") && (!event.appKey || typeof event.appKey != "string")) {
        ret.code = 4001;
        ret.msg = "参数[appKey]错误"
        return ret;
    }

    appName = event.appName.trim();
    platform = event.platform.trim();
    channelAppId = event.appId.trim();
    channelAppKey = event.appKey.trim();
    force = event.force || false;

    const db = cloud.database();
    //创建集合
    try {
        await db.getCollection("app_config")
    } catch (e) {
        if (e.message == "not found collection") {
            await db.createCollection("app_config");
        } else {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }

    let innerAppId = randomUUID();
    try {
        // 查询是否存在
        let appList = await db.collection(`app_config`).where({
            "platform": platform,
            "channelAppId": channelAppId,
        }).get();
        if (appList.length > 0) {
            db.collection(`app_config`).where({
                "platform": platform,
                "channelAppId": channelAppId,
            }).update({
                data: {
                    "appName": appName,
                    "channelAppKey": channelAppKey,
                }
            });
            ret.data = {
                "appName": appName,
                "innerAppId": appList[0].appId,
            }
            return ret;
        }

        // 插入app
        await db.collection(`app_config`).add({
            data: {
                "appId": innerAppId,
                "channelAppId": channelAppId,
                "appName": appName,
                "platform": platform,
                "channelAppKey": channelAppKey,
                "createTime": moment().format("YYYY-MM-DD HH:mm:ss"),
            }
        });    

        // 创建用户表
        try {
            let userTableName = `user_${innerAppId}`;
            await db.getCollection(userTableName);
        } catch (e) {
            if (e.message == "not found collection") {
                await db.createCollection(`user_${innerAppId}`);
            }
        }
    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    ret.data = {
        "appName": appName,
        "innerAppId": innerAppId,
    }
    return ret;
};