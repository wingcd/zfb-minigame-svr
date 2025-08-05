const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission } = require("./common/auth");

/**
 * 函数：createCounter
 * 说明：创建计数器
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | key | string | 是 | 计数器key |
    | locations | array | 否 | 点位标识数组，如['default', 'beijing', 'shanghai']，默认为['default'] |
    | resetType | string | 否 | 重置类型：daily(每日)、weekly(每周)、monthly(每月)、custom(自定义)、permanent(永久)，默认permanent |
    | resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |
    | description | string | 否 | 计数器描述 |
 */

async function createCounterHandler(event, context) {
    let appId = event.appId;
    let key = event.key;
    let locations = event.locations || ['default'];
    let resetType = event.resetType || 'permanent';
    let resetValue = event.resetValue;
    let description = event.description || '';

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

    if (!Array.isArray(locations) || locations.length === 0 || locations.length > 1000) {
        ret.code = 4001;
        ret.msg = "参数[locations]错误，必须是1-1000个点位的数组";
        return ret;
    }

    // 验证点位标识格式
    for (let location of locations) {
        if (!location || typeof location !== "string" || location.length > 50) {
            ret.code = 4001;
            ret.msg = "点位标识必须是非空字符串且长度不超过50";
            return ret;
        }
    }

    if (!["daily", "weekly", "monthly", "custom", "permanent"].includes(resetType)) {
        ret.code = 4001;
        ret.msg = "参数[resetType]错误，支持的值：daily、weekly、monthly、custom、permanent";
        return ret;
    }

    if (resetType === "custom" && (!resetValue || typeof resetValue !== "number" || resetValue <= 0)) {
        ret.code = 4001;
        ret.msg = "参数[resetValue]错误，自定义重置类型必须提供大于0的重置时间(小时)";
        return ret;
    }

    try {
        const db = cloud.database();
        let counterTableName = `counter_${appId}`;
        
        // 确保计数器表存在，如果不存在则创建
        try {
            await db.getCollection(counterTableName);
        } catch (e) {
            if (e.message == "not found collection") {
                await db.createCollection(counterTableName);
                console.log(`计数器表 ${counterTableName} 创建成功`);
            } else {
                throw e;
            }
        }
        
        let collection = db.collection(counterTableName);

        // 检查key是否已存在
        let existingCounter = await collection
            .where({
                "key": key
            })
            .get();

        if (existingCounter.length > 0) {
            ret.code = 4002;
            ret.msg = `计数器[${key}]已存在`;
            return ret;
        }

        let now = moment();
        // 计算重置时间
        let resetTime = null;
        if (resetType !== 'permanent') {
            switch (resetType) {
                case "daily":
                    resetTime = moment().startOf('day').add(1, 'day');
                    break;
                case "weekly":
                    resetTime = moment().startOf('week').add(1, 'week');
                    break;
                case "monthly":
                    resetTime = moment().startOf('month').add(1, 'month');
                    break;
                case "custom":
                    if (resetValue) {
                        resetTime = moment().add(resetValue, 'hours');
                    }
                    break;
            }
        }

        // 构建点位数据
        const locationsMap = {};
        if (Array.isArray(locations)) {
            // 多个点位
            locations.forEach(loc => {
                locationsMap[loc] = {
                    value: 0
                };
            });
        } else {
            // 单个点位
            locationsMap[locations] = {
                value: 0
            };
        }

        // 构建计数器数据
        let counterData = {
            "key": key,
            "locations": locationsMap,
            "resetType": resetType,
            "resetValue": resetValue,
            "description": description,
            "resetTime": resetTime ? resetTime.format("YYYY-MM-DD HH:mm:ss") : null,
            "gmtCreate": now.format("YYYY-MM-DD HH:mm:ss"),
            "gmtModify": now.format("YYYY-MM-DD HH:mm:ss")
        };

        await collection.add({
            data: counterData
        });

        ret.data = {
            key: key,
            locations: Object.keys(locationsMap),
            locationsMap: locationsMap,
            resetType: resetType,
            resetValue: resetValue || null,
            resetTime: resetTime ? resetTime.format("YYYY-MM-DD HH:mm:ss") : null,
            description: description
        };

    } catch (e) {
        console.error('创建计数器失败:', e);
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 导出处理函数
exports.main = requirePermission(createCounterHandler, ['counter_manage']); 