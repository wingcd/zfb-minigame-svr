const cloud = require("@alipay/faas-server-sdk");
const { requirePermission } = require("./common/auth");

/**
 * 函数：deleteCounter
 * 说明：删除计数器
 * 权限：需要 counter_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | key | string | 是 | 计数器key |
    | location | string | 否 | 点位标识，如果指定则只删除该点位，否则删除整个计数器 |
 */

async function deleteCounterHandler(event, context) {
    let appId = event.appId;
    let key = event.key;
    let location = event.location;

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

    if (location && typeof location !== "string") {
        ret.code = 4001;
        ret.msg = "参数[location]错误，必须是字符串";
        return ret;
    }

    try {
        const db = cloud.database();
        let counterTableName = `counter_${appId}`;
        let collection = db.collection(counterTableName);

        // 检查计数器是否存在
        let existingCounter = await collection
            .where({ "key": key })
            .get();

        if (existingCounter.length === 0) {
            ret.code = 4004;
            ret.msg = `计数器[${key}]不存在`;
            return ret;
        }

        let counter = existingCounter[0];

        if (location) {
            // 删除指定点位
            const currentLocations = counter.locations || {};
            if (!currentLocations[location]) {
                ret.code = 4004;
                ret.msg = `计数器[${key}]的点位[${location}]不存在`;
                return ret;
            }

            // 如果只有一个点位，则删除整个计数器
            if (Object.keys(currentLocations).length === 1) {
                await collection.doc(counter._id).remove();
                ret.msg = `计数器[${key}]的最后一个点位[${location}]已删除，整个计数器已删除`;
            } else {
                // 删除指定点位
                await collection.doc(counter._id).update({
                    data: {
                        [`locations.${location}`]: db.command.remove(),
                        "gmtModify": new Date().toISOString().slice(0, 19).replace('T', ' ')
                    }
                });
                ret.msg = `计数器[${key}]的点位[${location}]已删除`;
            }

            ret.data = {
                key: key,
                location: location,
                action: Object.keys(currentLocations).length === 1 ? "deleted_counter" : "deleted_location"
            };
        } else {
            // 删除整个计数器
            await collection.doc(counter._id).remove();
            ret.msg = `计数器[${key}]已删除`;
            ret.data = {
                key: key,
                action: "deleted_counter"
            };
        }

    } catch (e) {
        console.error('删除计数器失败:', e);
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 导出处理函数
exports.main = requirePermission(deleteCounterHandler, ['counter_manage']); 