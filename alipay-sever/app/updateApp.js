const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");

// 请求参数
/**
 * 函数：updateApp
 * 说明：更新应用信息
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | appName | string | 否 | 应用名称 |
    | description | string | 否 | 应用描述 |
    | status | string | 否 | 应用状态 |
    | channelAppKey | string | 否 | 渠道应用密钥 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "appName": "更新的小游戏",
        "description": "这是一个更新的小游戏",
        "status": "active"
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
    let appName = event.appName;
    let description = event.description;
    let status = event.status;
    let channelAppKey = event.channelAppKey;

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
        ret.msg = "参数[appId]错误";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查应用是否存在
        const appList = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (appList.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        // 构建更新数据
        let updateData = {
            updateTime: moment().format("YYYY-MM-DD HH:mm:ss")
        };

        if (appName !== undefined) {
            updateData.appName = appName;
        }

        if (description !== undefined) {
            updateData.description = description;
        }

        if (status !== undefined) {
            updateData.status = status;
        }

        if (channelAppKey !== undefined) {
            updateData.channelAppKey = channelAppKey;
        }

        // 更新应用信息
        await db.collection('app_config')
            .where({ appId: appId })
            .update({
                data: updateData
            });

        ret.msg = "更新成功";

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 