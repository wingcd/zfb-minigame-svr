const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");

/**
 * 函数：incrementCounter
 * 说明：增加计数器值
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | key | string | 是 | 计数器key |
    | increment | number | 否 | 增加的数量，默认1 |
    | location | string | 否 | 点位标识，默认为"default" |
 */

async function incrementCounterHandler(event, context) {
    let appId = event.appId;
    let key = event.key;
    let increment = event.increment || 1;
    let location = event.location || "default";

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

    if (!key || typeof key !== "string") {
        ret.code = 4001;
        ret.msg = "参数[key]错误";
        return ret;
    }

    if (typeof increment !== "number" || increment <= 0) {
        ret.code = 4001;
        ret.msg = "参数[increment]错误，必须是大于0的数字";
        return ret;
    }

    if (typeof location !== "string") {
        ret.code = 4001;
        ret.msg = "参数[location]错误，必须是字符串";
        return ret;
    }

    try {
        const db = cloud.database();
        let counterTableName = `counter_${appId}`;

        // 确保计数器表存在
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

        // 查询计数器记录
        let queryList = await collection
            .where({ key: key })
            .get();

        if (queryList.length === 0) {
            ret.code = 4004;
            ret.msg = `计数器[${key}]不存在，请先在管理后台创建`;
            return ret;
        }

        const record = queryList[0];
        const locations = record.locations || {};
        let now = moment();

        // 检查点位是否存在
        if (!locations[location]) {
            ret.code = 4004;
            ret.msg = `计数器[${key}]的点位[${location}]不存在`;
            return ret;
        }

        let locationData = locations[location];
        let currentValue = locationData.value || 0;
        let shouldReset = false;

        // 检查是否需要重置
        if (locationData.resetTime) {
            let resetTime = moment(locationData.resetTime);
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

                // 更新重置时间
                if (newResetTime) {
                    locationData.resetTime = newResetTime.format("YYYY-MM-DD HH:mm:ss");
                }
            }
        }

        // 计算新值
        let newValue = currentValue + increment;

        // 更新数据库
        let updateData = {
            [`locations.${location}.value`]: newValue,
            "gmtModify": now.format("YYYY-MM-DD HH:mm:ss")
        };

        // 如果重置了时间，也更新重置时间
        if (shouldReset && locationData.resetTime) {
            updateData[`locations.${location}.resetTime`] = locationData.resetTime;
        }

        await collection.doc(record._id).update({
            data: updateData
        });

        ret.data = {
            key: key,
            location: location,
            currentValue: newValue,
            resetTime: locationData.resetTime
        };

    } catch (e) {
        console.error('增加计数器值失败:', e);
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出处理函数
exports.main = incrementCounterHandler; 