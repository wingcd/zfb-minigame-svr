const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission } = require("./common/auth");

/**
 * 函数：updateCounter
 * 说明：更新计数器配置
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | key | string | 是 | 计数器key |
    | resetType | string | 否 | 重置类型：daily(每日)、weekly(每周)、monthly(每月)、custom(自定义)、permanent(永久) |
    | resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |
    | description | string | 否 | 计数器描述 |
    | value | number | 否 | 重置计数器值 |
 */

async function updateCounterHandler(event, context) {
    let appId = event.appId;
    let key = event.key;
    let resetType = event.resetType;
    let resetValue = event.resetValue;
    let description = event.description;
    let value = event.value;

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

    if (resetType && !["daily", "weekly", "monthly", "custom", "permanent"].includes(resetType)) {
        ret.code = 4001;
        ret.msg = "参数[resetType]错误，支持的值：daily、weekly、monthly、custom、permanent";
        return ret;
    }

    if (resetType === "custom" && (!resetValue || typeof resetValue !== "number" || resetValue <= 0)) {
        ret.code = 4001;
        ret.msg = "参数[resetValue]错误，自定义重置类型必须提供大于0的重置时间(小时)";
        return ret;
    }

    if (value !== undefined && (typeof value !== "number" || value < 0)) {
        ret.code = 4001;
        ret.msg = "参数[value]错误，必须是大于等于0的数字";
        return ret;
    }

    try {
        const db = cloud.database();
        let counterTableName = `counter_${appId}`;
        let collection = db.collection(counterTableName);

        // 检查计数器是否存在
        let existingCounter = await collection
            .where({
                "key": key
            })
            .get();

        if (existingCounter.length === 0) {
            ret.code = 4004;
            ret.msg = `计数器[${key}]不存在`;
            return ret;
        }

        let counter = existingCounter[0];
        let now = moment();
        let updateData = {
            "gmtModify": now.format("YYYY-MM-DD HH:mm:ss")
        };

        // 更新重置类型
        if (resetType !== undefined) {
            updateData.resetType = resetType;
            
            if (resetType !== 'permanent') {
                let resetTime = null;
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
                if (resetTime) {
                    updateData.resetTime = resetTime.format("YYYY-MM-DD HH:mm:ss");
                }
            } else {
                updateData.resetTime = null;
            }
        }

        // 更新重置值
        if (resetValue !== undefined) {
            updateData.resetValue = resetValue;
        }

        // 更新描述
        if (description !== undefined) {
            updateData.description = description;
        }

        // 更新计数器值
        if (value !== undefined) {
            updateData.value = value;
        }

        await collection.doc(counter._id).update({
            data: updateData
        });

        // 返回更新后的数据
        let updatedCounter = await collection.doc(counter._id).get();
        ret.data = {
            _id: updatedCounter._id,
            key: updatedCounter.key,
            value: updatedCounter.value,
            resetType: updatedCounter.resetType,
            resetValue: updatedCounter.resetValue,
            description: updatedCounter.description,
            resetTime: updatedCounter.resetTime,
            gmtCreate: updatedCounter.gmtCreate,
            gmtModify: updatedCounter.gmtModify
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 包装权限校验
exports.main = requirePermission(updateCounterHandler, ['leaderboard_manage']); 