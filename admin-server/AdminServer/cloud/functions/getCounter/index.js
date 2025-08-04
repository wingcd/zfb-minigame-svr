const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：getCounter
 * 说明：获取计数器值
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 小程序id |
    | key | string | 否 | 计数器key，不传则获取该游戏所有计数器 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "key": "daily_challenge"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "key": "daily_challenge",
            "value": 5,
            "resetType": "daily",
            "resetTime": "2023-10-29 00:00:00",
            "timeToReset": 36000000
        }
    }
    
    或者获取所有计数器：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": [
            {
                "key": "daily_challenge",
                "value": 5,
                "resetType": "daily",
                "resetTime": "2023-10-29 00:00:00",
                "timeToReset": 36000000
            },
            {
                "key": "weekly_battle",
                "value": 10,
                "resetType": "weekly",
                "resetTime": "2023-10-30 00:00:00",
                "timeToReset": 122400000
            }
        ]
    }
*/

exports.main = async (event, context) => {
    let appId;
    let key = null;

    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": null
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

    //请求参数
    appId = event.appId.trim();
    key = event.key.trim();

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
        } else {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }

    try {
        // 构建查询条件
        let whereCondition = {
            key: key
        };

        // 查询计数器记录
        let queryList = await collection
            .where(whereCondition)
            .get();

        if (queryList.length === 0) {
            // 如果查询特定key但不存在，返回错误
            ret.code = 4004;
            ret.msg = `计数器[${key}]不存在，请先在管理后台创建`;
            return ret;
        }

        let now = moment();
        let results = [];

        // 处理每个计数器记录
        for (let record of queryList) {
            let currentValue = record.value || 0;
            let shouldReset = false;
            let timeToReset = null;

            // 检查是否需要重置
            if (record.resetTime) {
                let resetTime = moment(record.resetTime);
                timeToReset = resetTime.diff(now);
                
                if (now.isAfter(resetTime)) {
                    shouldReset = true;
                    currentValue = 0;
                    
                    // 重新计算下次重置时间
                    let newResetTime = null;
                    if (record.resetType && record.resetType !== 'permanent') {
                        switch (record.resetType) {
                            case "daily":
                                newResetTime = moment().startOf('day').add(1, 'day');
                                break;
                            case "weekly":
                                newResetTime = moment().startOf('week').add(1, 'week');
                                break;
                            case "monthly":
                                newResetTime = moment().startOf('month').add(1, 'month');
                                break;
                            case "custom":
                                if (record.resetValue) {
                                    newResetTime = moment().add(record.resetValue, 'hours');
                                }
                                break;
                        }
                    }

                    // 更新数据库中的记录
                    let updateData = {
                        "value": 0,
                        "gmtModify": now.format("YYYY-MM-DD HH:mm:ss")
                    };

                    if (newResetTime) {
                        updateData.resetTime = newResetTime.format("YYYY-MM-DD HH:mm:ss");
                        timeToReset = newResetTime.diff(now);
                    }

                    await collection.doc(record._id).update({
                        data: updateData
                    });
                }
            }

            let counterData = {
                key: record.key,
                value: currentValue
            };

            results.push(counterData);
        }

        // 返回单个对象
        ret.data = results[0];

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 