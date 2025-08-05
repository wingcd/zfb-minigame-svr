const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");

/**
 * 函数：getCounter
 * 说明：获取计数器当前值（返回所有点位）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | key | string | 是 | 计数器key |
 */

async function getCounterHandler(event, context) {
    let appId = event.appId;
    let key = event.key;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": null
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

    try {
        const db = cloud.database();
        let counterTableName = `counter_${appId}`;
        
        // 确保计数器表存在
        try {
            var collection = db.collection(counterTableName);
        } catch (e) {
            if (e.message == "not found collection") {
                ret.code = 4004;
                ret.msg = `计数器表不存在，请先在管理后台创建计数器`;
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
        let shouldUpdateRecord = false;
        let updateData = {
            "gmtModify": now.format("YYYY-MM-DD HH:mm:ss")
        };

        // 检查是否需要重置（所有点位共享同一个重置时间）
        let resetTime = record.resetTime;
        let timeToReset = null;
        let currentResetTime = resetTime;

        if (resetTime) {
            let resetMoment = moment(resetTime);
            timeToReset = resetMoment.diff(now);
            
            if (now.isAfter(resetMoment)) {
                shouldUpdateRecord = true;
                
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

                if (newResetTime) {
                    updateData.resetTime = newResetTime.format("YYYY-MM-DD HH:mm:ss");
                    timeToReset = newResetTime.diff(now);
                    currentResetTime = newResetTime.format("YYYY-MM-DD HH:mm:ss");
                }

                // 重置所有点位的值
                for (let locationKey of Object.keys(locations)) {
                    updateData[`locations.${locationKey}.value`] = 0;
                }
            }
        }

        // 如果需要更新记录，执行更新
        if (shouldUpdateRecord) {
            await collection.doc(record._id).update({
                data: updateData
            });
        }

        // 构建返回数据
        let resultLocations = {};
        for (let [locationKey, locationData] of Object.entries(locations)) {
            resultLocations[locationKey] = {
                value: shouldUpdateRecord ? 0 : (locationData.value || 0)
            };
        }

        ret.data = {
            key: record.key,
            locations: resultLocations,
            resetType: record.resetType || 'permanent',
            resetValue: record.resetValue || null,
            resetTime: currentResetTime,
            timeToReset: timeToReset,
            description: record.description || ''
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
};

// 导出处理函数
exports.main = getCounterHandler; 