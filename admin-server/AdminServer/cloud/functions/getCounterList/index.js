const cloud = require("@alipay/faas-server-sdk");
const { requirePermission } = require("./common/auth");

/**
 * 函数：getCounterList
 * 说明：获取计数器列表（支持分页和筛选）
 * 权限：需要 counter_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | page | number | 否 | 页码（默认：1） |
    | pageSize | number | 否 | 每页数量（默认：20，最大：100） |
    | key | string | 否 | 计数器key筛选（模糊搜索） |
    | resetType | string | 否 | 重置类型筛选 |
    | groupByKey | boolean | 否 | 是否按key分组（默认：false） |
 */

async function getCounterListHandler(event, context) {
    let appId = event.appId;
    let page = event.page || 1;
    let pageSize = Math.min(event.pageSize || 20, 100);
    let key = event.key;
    let resetType = event.resetType;
    let groupByKey = event.groupByKey || false;

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

        if (groupByKey) {
            // 分组模式：返回按key分组的数据
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

            // 格式化数据为分组格式
            const list = queryList.map(item => {
                // 计算总值和点位数量
                const locations = item.locations || {};
                const locationKeys = Object.keys(locations);
                const totalValue = locationKeys.reduce((sum, key) => sum + (locations[key].value || 0), 0);
                
                // 转换locations为数组格式（前端期望的格式）
                const locationsArray = locationKeys.map(locationKey => ({
                    location: locationKey,
                    value: locations[locationKey].value || 0
                }));
                
                return {
                    _id: item._id,
                    key: item.key,
                    locations: locationsArray,
                    locationCount: locationKeys.length,
                    totalValue: totalValue,
                    resetType: item.resetType || 'permanent',
                    resetValue: item.resetValue || null,
                    resetTime: item.resetTime || null,
                    description: item.description || '',
                    gmtCreate: item.gmtCreate,
                    gmtModify: item.gmtModify
                };
            });

            ret.data = {
                list: list,
                total: total,
                page: page,
                pageSize: pageSize
            };
        } else {
            // 列表模式：返回扁平化的数据
            // 获取所有数据
            const allCounters = await collection.where(whereCondition).get();
            
            // 扁平化数据
            const flatList = [];
            allCounters.forEach(counter => {
                const locations = counter.locations || {};
                Object.keys(locations).forEach(locationKey => {
                    flatList.push({
                        _id: counter._id,
                        key: counter.key,
                        location: locationKey,
                        value: locations[locationKey].value || 0,
                        resetType: counter.resetType || 'permanent',
                        resetValue: counter.resetValue || null,
                        resetTime: counter.resetTime || null,
                        description: counter.description || '',
                        gmtCreate: counter.gmtCreate,
                        gmtModify: counter.gmtModify
                    });
                });
            });

            // 排序
            flatList.sort((a, b) => new Date(b.gmtModify) - new Date(a.gmtModify));

            // 分页
            const total = flatList.length;
            const skip = (page - 1) * pageSize;
            const paginatedList = flatList.slice(skip, skip + pageSize);

            ret.data = {
                list: paginatedList,
                total: total,
                page: page,
                pageSize: pageSize
            };
        }

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 导出处理函数
exports.main = requirePermission(getCounterListHandler, ['counter_manage']); 