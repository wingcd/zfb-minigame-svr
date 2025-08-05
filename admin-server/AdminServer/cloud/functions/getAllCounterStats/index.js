const cloud = require("@alipay/faas-server-sdk");
const { requirePermission } = require("./common/auth");

/**
 * 函数：getAllCounterStats
 * 说明：获取应用的所有计数器统计信息
 * 权限：需要 counter_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
 */

async function getAllCounterStatsHandler(event, context) {
    let appId = event.appId;

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
                    totalCounters: 0,
                    totalLocations: 0,
                    totalValue: 0,
                    resetTypeStats: {},
                    topCounters: [],
                    recentActivity: []
                };
                return ret;
            } else {
                ret.code = 5001;
                ret.msg = e.message;
                return ret;
            }
        }

        // 获取所有计数器数据
        const allCounters = await collection
            .orderBy('gmtModify', 'desc')
            .get();

        if (allCounters.length === 0) {
            ret.data = {
                totalCounters: 0,
                totalLocations: 0,
                totalValue: 0,
                resetTypeStats: {},
                topCounters: [],
                recentActivity: []
            };
            return ret;
        }

        // 按key分组统计
        const counterGroups = {};
        let totalValue = 0;
        const resetTypeStats = {};
        
        allCounters.forEach(counter => {
            const key = counter.key;
            const value = counter.value || 0;
            const resetType = counter.resetType || 'permanent';
            
            totalValue += value;
            
            // 重置类型统计
            if (!resetTypeStats[resetType]) {
                resetTypeStats[resetType] = 0;
            }
            resetTypeStats[resetType]++;
            
            // 按key分组
            if (!counterGroups[key]) {
                counterGroups[key] = {
                    key: key,
                    locations: [],
                    totalValue: 0,
                    locationCount: 0,
                    resetType: resetType,
                    description: counter.description || '',
                    lastModified: counter.gmtModify
                };
            }
            
            counterGroups[key].locations.push({
                location: counter.location || 'default',
                value: value,
                resetTime: counter.resetTime,
                gmtModify: counter.gmtModify
            });
            
            counterGroups[key].totalValue += value;
            counterGroups[key].locationCount++;
            
            // 更新最后修改时间
            if (counter.gmtModify > counterGroups[key].lastModified) {
                counterGroups[key].lastModified = counter.gmtModify;
            }
        });

        // 转换为数组并排序
        const counterList = Object.values(counterGroups);
        
        // 获取TOP10计数器（按总值排序）
        const topCounters = counterList
            .sort((a, b) => b.totalValue - a.totalValue)
            .slice(0, 10)
            .map(counter => ({
                key: counter.key,
                totalValue: counter.totalValue,
                locationCount: counter.locationCount,
                resetType: counter.resetType,
                description: counter.description
            }));

        // 获取最近活动（最近修改的10个点位）
        const recentActivity = allCounters
            .sort((a, b) => new Date(b.gmtModify) - new Date(a.gmtModify))
            .slice(0, 10)
            .map(counter => ({
                key: counter.key,
                location: counter.location || 'default',
                value: counter.value || 0,
                resetType: counter.resetType,
                gmtModify: counter.gmtModify
            }));

        ret.data = {
            totalCounters: counterList.length,
            totalLocations: allCounters.length,
            totalValue: totalValue,
            resetTypeStats: resetTypeStats,
            topCounters: topCounters,
            recentActivity: recentActivity,
            summary: {
                averageValuePerCounter: counterList.length > 0 ? Math.round(totalValue / counterList.length) : 0,
                averageLocationsPerCounter: counterList.length > 0 ? Math.round(allCounters.length / counterList.length) : 0
            }
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
    }

    return ret;
}

// 包装权限校验
exports.main = requirePermission(getAllCounterStatsHandler, ['counter_manage']); 