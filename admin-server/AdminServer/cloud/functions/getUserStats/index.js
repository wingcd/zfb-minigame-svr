const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

// 原始处理函数
async function getUserStatsHandler(event, context) {
    let appId = event.appId;

    console.log("getUserStats event:", event);

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

    const db = cloud.database();

    try {
        // 时间范围计算
        const now = moment();
        const today = now.format('YYYY-MM-DD');
        const yesterday = now.clone().subtract(1, 'day').format('YYYY-MM-DD');
        
        const userTableName = `user_${appId}`;

        // 检查用户表是否存在
        let collection;
        try {
            collection = db.collection(userTableName);
        } catch (e) {
            ret.code = 4004;
            ret.msg = "应用不存在或用户表不存在";
            return ret;
        }

        // 获取用户统计数据
        const userStats = await calculateUserStats(collection, today, yesterday);

        ret.data = userStats;

        // 记录操作日志
        await logOperation(event.adminInfo, 'VIEW', 'USER_STATS', {
            appId: appId,
            statsGenerated: Object.keys(ret.data)
        });

    } catch (e) {
        console.error("getUserStats error:", e);
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 计算用户统计信息
async function calculateUserStats(collection, today, yesterday) {
    try {
        // 总用户数
        const totalResult = await collection.count();
        const total = totalResult.total;
        
        // 今日新增用户
        const newTodayResult = await collection
            .where({
                gmtCreate: {
                    $gte: today + ' 00:00:00',
                    $lte: today + ' 23:59:59'
                }
            })
            .count();
        const newToday = newTodayResult.total;
        
        // 今日活跃用户（今天有登录或活动的用户）
        const activeTodayResult = await collection
            .where({
                gmtModify: {
                    $gte: today + ' 00:00:00',
                    $lte: today + ' 23:59:59'
                }
            })
            .count();
        const activeToday = activeTodayResult.total;
        
        // 昨日新增用户（用于计算增长率）
        const newYesterdayResult = await collection
            .where({
                gmtCreate: {
                    $gte: yesterday + ' 00:00:00',
                    $lte: yesterday + ' 23:59:59'
                }
            })
            .count();
        const newYesterday = newYesterdayResult.total;
        
        // 昨日活跃用户（用于计算增长率）
        const activeYesterdayResult = await collection
            .where({
                gmtModify: {
                    $gte: yesterday + ' 00:00:00',
                    $lte: yesterday + ' 23:59:59'
                }
            })
            .count();
        const activeYesterday = activeYesterdayResult.total;
        
        // 封禁用户数
        const bannedResult = await collection
            .where({ banned: true })
            .count();
        const banned = bannedResult.total;
        
        // 本周活跃用户
        const weekStart = moment().startOf('week').format('YYYY-MM-DD');
        const weeklyActiveResult = await collection
            .where({
                gmtModify: {
                    $gte: weekStart + ' 00:00:00',
                    $lte: today + ' 23:59:59'
                }
            })
            .count();
        const weeklyActive = weeklyActiveResult.total;
        
        // 本月活跃用户
        const monthStart = moment().startOf('month').format('YYYY-MM-DD');
        const monthlyActiveResult = await collection
            .where({
                gmtModify: {
                    $gte: monthStart + ' 00:00:00',
                    $lte: today + ' 23:59:59'
                }
            })
            .count();
        const monthlyActive = monthlyActiveResult.total;
        
        // 计算增长率
        const newTodayGrowth = newYesterday > 0 ? ((newToday - newYesterday) / newYesterday * 100).toFixed(2) : 0;
        const activeTodayGrowth = activeYesterday > 0 ? ((activeToday - activeYesterday) / activeYesterday * 100).toFixed(2) : 0;

        return {
            total: total,
            newToday: newToday,
            activeToday: activeToday,
            banned: banned,
            weeklyActive: weeklyActive,
            monthlyActive: monthlyActive,
            growth: {
                newTodayGrowth: parseFloat(newTodayGrowth),
                activeTodayGrowth: parseFloat(activeTodayGrowth)
            },
            comparison: {
                newYesterday: newYesterday,
                activeYesterday: activeYesterday
            }
        };
    } catch (e) {
        console.error("calculateUserStats error:", e);
        return {
            total: 0,
            newToday: 0,
            activeToday: 0,
            banned: 0,
            weeklyActive: 0,
            monthlyActive: 0,
            growth: {
                newTodayGrowth: 0,
                activeTodayGrowth: 0
            },
            comparison: {
                newYesterday: 0,
                activeYesterday: 0
            }
        };
    }
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getUserStatsHandler, 'stats_view');
exports.main = mainFunc; 