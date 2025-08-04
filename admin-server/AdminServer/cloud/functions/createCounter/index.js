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
    | resetType | string | 是 | 重置类型：daily(每日)、weekly(每周)、monthly(每月)、custom(自定义)、permanent(永久) |
    | resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |
    | description | string | 否 | 计数器描述 |
 */

async function createCounterHandler(event, context) {
    let appId = event.appId;
    let key = event.key;
    let resetType = event.resetType;
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

    if (!resetType || !["daily", "weekly", "monthly", "custom", "permanent"].includes(resetType)) {
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
        let resetTime = null;

        // 计算重置时间
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
                    resetTime = moment().add(resetValue, 'hours');
                    break;
            }
        }

        // 插入新记录
        let insertData = {
            "appId": appId,
            "key": key,
            "value": 0,
            "resetType": resetType,
            "resetValue": resetValue || null,
            "description": description,
            "gmtCreate": now.format("YYYY-MM-DD HH:mm:ss"),
            "gmtModify": now.format("YYYY-MM-DD HH:mm:ss")
        };

        if (resetTime) {
            insertData.resetTime = resetTime.format("YYYY-MM-DD HH:mm:ss");
        }

        await collection.add({
            data: insertData
        });

        ret.data = {
            key: key,
            value: 0,
            resetType: resetType,
            resetValue: resetValue || null,
            description: description,
            resetTime: resetTime ? resetTime.format("YYYY-MM-DD HH:mm:ss") : null
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 包装权限校验
exports.main = requirePermission(createCounterHandler, ['leaderboard_manage']); 