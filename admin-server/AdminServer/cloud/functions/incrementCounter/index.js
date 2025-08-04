const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：incrementCounter
 * 说明：增加计数器值
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 小程序id |
    | key | string | 是 | 计数器key |
    | increment | number | 否 | 增加的数量，默认1 |
    | resetType | string | 否 | 重置类型：daily(每日)、weekly(每周)、monthly(每月)、custom(自定义)、permanent(永久) |
    | resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "key": "daily_challenge",
        "increment": 1,
        "resetType": "daily"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "key": "daily_challenge",
            "currentValue": 5,
            "resetTime": "2023-10-29 00:00:00"
        }
    }
*/

exports.main = async (event, context) => {
    let appId;
    let key;
    let increment = 1;
    let resetType = null;
    let resetValue = null;

    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    var parmErr = common.hash.CheckParams(event);
    if(parmErr) {
        ret.code = 4001;
        ret.msg = "参数错误, error code:" + parmErr;
        return ret;
    }

    //参数校验 字段存在 为空 类型
    if (!event.hasOwnProperty("appId") || !event.appId || typeof event.appId != "string") {
        ret.code = 4001;
        ret.msg = "参数[appId]错误";
        return ret;
    }



    if (!event.hasOwnProperty("key") || !event.key || typeof event.key != "string") {
        ret.code = 4001;
        ret.msg = "参数[key]错误";
        return ret;
    }

    if (event.hasOwnProperty("increment") && (typeof event.increment != "number" || event.increment <= 0)) {
        ret.code = 4001;
        ret.msg = "参数[increment]错误，必须是大于0的数字";
        return ret;
    }

    if (event.hasOwnProperty("resetType") && !["daily", "weekly", "monthly", "custom", "permanent"].includes(event.resetType)) {
        ret.code = 4001;
        ret.msg = "参数[resetType]错误，支持的值：daily、weekly、monthly、custom、permanent";
        return ret;
    }

    if (event.hasOwnProperty("resetValue") && (typeof event.resetValue != "number" || event.resetValue <= 0)) {
        ret.code = 4001;
        ret.msg = "参数[resetValue]错误，必须是大于0的数字(小时)";
        return ret;
    }

    //请求参数
    appId = event.appId.trim();
    key = event.key.trim();
    if (event.hasOwnProperty("increment")) {
        increment = event.increment;
    }
    if (event.hasOwnProperty("resetType")) {
        resetType = event.resetType.trim();
    }
    if (event.hasOwnProperty("resetValue")) {
        resetValue = event.resetValue;
    }

    // 获取 cloud 环境中的 mongoDB 数据库对象
    const db = cloud.database();
    let counterTableName = `counter_${appId}`;

    try {
        var collection = db.collection(counterTableName);
    } catch (e) {
        if (e.message == "not found collection") {
            ret.code = 4004;
            ret.msg = `计数器[${key}]不存在，请先在管理后台创建`;
            return ret;
        }
        else {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }

    try {
        // 查询现有计数器记录
        let queryList = await collection
            .where({
                "key": key
            })
            .get();

        let now = moment();
        let shouldReset = false;

        if (queryList.length === 0) {
            // 计数器不存在，返回错误
            ret.code = 4004;
            ret.msg = `计数器[${key}]不存在，请先在管理后台创建`;
            return ret;
        } else {
            // 更新现有记录
            let existingRecord = queryList[0];
            let currentValue = existingRecord.value || 0;

            // 检查是否需要重置
            if (existingRecord.resetTime) {
                let lastResetTime = moment(existingRecord.resetTime);
                if (now.isAfter(lastResetTime)) {
                    shouldReset = true;
                    currentValue = 0;
                }
            }

            let newValue = currentValue + increment;

            let updateData = {
                "value": newValue,
                "gmtModify": now.format("YYYY-MM-DD HH:mm:ss")
            };

            await collection.doc(existingRecord._id).update({
                data: updateData
            });

            ret.data = {
                key: key,
                currentValue: newValue
            };
        }

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 