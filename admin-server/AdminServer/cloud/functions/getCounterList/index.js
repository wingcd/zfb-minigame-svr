const cloud = require("@alipay/faas-server-sdk");
const { requirePermission } = require("./common/auth");

/**
 * 函数：getCounterList
 * 说明：获取计数器列表（支持分页和筛选）
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | page | number | 否 | 页码（默认：1） |
    | pageSize | number | 否 | 每页数量（默认：20，最大：100） |
    | key | string | 否 | 计数器key筛选（模糊搜索） |
    | resetType | string | 否 | 重置类型筛选 |
 */

async function getCounterListHandler(event, context) {
    let appId = event.appId;
    let page = event.page || 1;
    let pageSize = Math.min(event.pageSize || 20, 100);
    let key = event.key;
    let resetType = event.resetType;

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

    try {
        const db = cloud.database();
        let counterTableName = `counter_${appId}`;        

        // 创建集合
        try {
            var collection = db.collection(counterTableName);
        } catch (e) {
            if (e.message == "not found collection") {
                ret.data = {
                    list: [],
                    total: 0,
                    page: 0,
                    pageSize: 10
                };
                return ret;
            } else {
                ret.code = 5001;
                ret.msg = e.message;
                return ret;
            }
        }

        // 构建查询条件
        let whereCondition = {};
        
        if (key) {
            whereCondition.key = { $regex: key, $options: 'i' };
        }
        
        if (resetType) {
            whereCondition.resetType = resetType;
        }

        // 获取总数
        const totalResult = await collection.where(whereCondition).count();
        const total = totalResult.total;

        // 分页查询
        const skip = (page - 1) * pageSize;
        const queryList = await collection
            .where(whereCondition)
            .orderBy('gmtModify', 'desc')
            .skip(skip)
            .limit(pageSize)
            .get();

        // 格式化数据
        const list = queryList.map(item => ({
            _id: item._id,
            key: item.key,
            value: item.value || 0,
            resetType: item.resetType || 'permanent',
            resetValue: item.resetValue || null,
            resetTime: item.resetTime || null,
            description: item.description || '',
            gmtCreate: item.gmtCreate,
            gmtModify: item.gmtModify
        }));

        ret.data = {
            list: list,
            total: total,
            page: page,
            pageSize: pageSize
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 包装权限校验
exports.main = requirePermission(getCounterListHandler, ['leaderboard_manage']); 