const cloud = require("@alipay/faas-server-sdk");
const { requirePermission } = require("./common/auth");

/**
 * 函数：deleteCounter
 * 说明：删除计数器
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | key | string | 是 | 计数器key |
 */

async function deleteCounterHandler(event, context) {
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

        // 删除计数器
        await collection.doc(existingCounter[0]._id).remove();

        ret.data = {
            key: key,
            deleted: true
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 包装权限校验
exports.main = requirePermission(deleteCounterHandler, ['leaderboard_manage']); 