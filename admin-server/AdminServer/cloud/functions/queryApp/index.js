const cloud = require("@alipay/faas-server-sdk");

// 请求参数
/**
 * 函数：queryApp
 * 说明：查询app
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appName | string | 否 | app名字 |
    | appId | string | 否 | 小程序id |
  * 测试数据
    {
        "appName": "小程序"
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
    //请求参数
    //app 名字
    let appName;
    let appId;
    let channelAppId;

    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    }

    //参数校验 字段存在  为空   类型
    appName = event.appName === undefined ? null : event.appName;
    appId = event.appId === undefined ? null : event.appId;
    channelAppId = event.channelAppId === undefined ? null : event.channelAppId;
    if (!appName && !appId && !channelAppId) {
        ret.code = 4001;
        ret.msg = "参数[appName]和[appId]和[channelAppId]不能同时为空";
        return ret;
    }

    if (appName && typeof appName != "string") {
        ret.code = 4001;
        ret.msg = "参数[appName]错误"
        return ret;
    }

    if (appId && typeof appId != "string") {
        ret.code = 4001;
        ret.msg = "参数[appId]错误"
        return ret;
    }

    if (channelAppId && typeof channelAppId != "string") {
        ret.code = 4001;
        ret.msg = "参数[channelAppId]错误"
        return ret;
    }

    //数据库实例
    const db = cloud.database();

    try {
        let appList = null;

        if (appId) {
            appList = await db.collection(`app_config`)
                .where({
                    appId: appId
                })
                .get();
        } else if (channelAppId) {
            appList = await db.collection(`app_config`)
                .where({
                    channelAppId: channelAppId
                })
                .get();
        } else {
            // (模糊查询)
            appList = await db.collection(`app_config`)
                .where({
                    appName: appName
                })
                .get();
        }
        if (appList.length === 0) {
            ret.msg = "未查询到您的数据";
            return ret;
        }
        ret.data = appList[0];

        // 查询排行榜数据
        let leaderBoardList = await db.collection(`leaderboard_config`)
            .where({
                appId: ret.data.appId
            })
            .get();
        ret.data.leaderBoardList = leaderBoardList;
    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
};