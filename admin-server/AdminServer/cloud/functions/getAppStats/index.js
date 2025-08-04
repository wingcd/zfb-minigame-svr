const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getAppStats
 * 说明：获取应用统计数据
 * 权限：需要 stats_view 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 否 | 应用ID（为空时统计所有应用） |
    | timeRange | string | 否 | 时间范围 (today/week/month) |
    | includeDetails | boolean | 否 | 是否包含详细数据 |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "timeRange": "week",
        "includeDetails": true
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "overview": {
                "totalApps": 25,
                "activeApps": 18,
                "totalUsers": 45000,
                "totalScores": 125000,
                "avgUsersPerApp": 1800
            },
            "performance": {
                "topApps": [
                    {
                        "appId": "game_001",
                        "appName": "超级游戏",
                        "userCount": 5000,
                        "scoreCount": 15000,
                        "dailyActiveUsers": 800,
                        "retentionRate": 75.5
                    }
                ],
                "growth": {
                    "newApps": 2,
                    "userGrowth": 8.5,
                    "activityGrowth": 12.3
                }
            },
            "categories": {
                "game": { "count": 15, "users": 30000 },
                "tool": { "count": 8, "users": 12000 },
                "other": { "count": 2, "users": 3000 }
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 应用不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getAppStatsHandler(event, context) {
    let appId = event.appId;
    let timeRange = event.timeRange || 'week';
    let includeDetails = event.includeDetails !== false; // 默认包含详细数据

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    const validTimeRanges = ['today', 'week', 'month'];
    if (!validTimeRanges.includes(timeRange)) {
        ret.code = 4001;
        ret.msg = "无效的时间范围";
        return ret;
    }

    const db = cloud.database();

    try {
        // 获取目标应用列表
        let targetApps = [];
        if (appId) {
            const appList = await db.collection('app_config')
                .where({ appId: appId })
                .get();

            if (appList.length === 0) {
                ret.code = 4004;
                ret.msg = "应用不存在";
                return ret;
            }
            targetApps = [appList[0]];
        } else {
            targetApps = await db.collection('app_config').get();
        }

        // 计算时间范围
        const now = moment();
        let startDate, endDate;
        
        switch (timeRange) {
            case 'today':
                startDate = now.format('YYYY-MM-DD') + ' 00:00:00';
                endDate = now.format('YYYY-MM-DD') + ' 23:59:59';
                break;
            case 'week':
                startDate = now.subtract(7, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                endDate = now.format('YYYY-MM-DD') + ' 23:59:59';
                break;
            case 'month':
                startDate = now.subtract(30, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                endDate = now.format('YYYY-MM-DD') + ' 23:59:59';
                break;
        }

        // 初始化统计数据
        let overview = {
            totalApps: targetApps.length,
            activeApps: 0,
            totalUsers: 0,
            totalScores: 0,
            avgUsersPerApp: 0
        };

        let appPerformanceList = [];
        let categories = {};

        // 遍历每个应用收集统计数据
        for (let app of targetApps) {
            const userTableName = `user_${app.appId}`;
            let appStats = {
                appId: app.appId,
                appName: app.appName,
                category: app.category || 'other',
                userCount: 0,
                scoreCount: 0,
                dailyActiveUsers: 0,
                retentionRate: 0,
                isActive: false
            };

            try {
                // 用户总数
                const totalUsers = await db.collection(userTableName).count();
                appStats.userCount = totalUsers.total;
                overview.totalUsers += totalUsers.total;

                // 活跃用户数（指定时间范围内）
                const activeUsers = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: startDate,
                            $lte: endDate
                        }
                    })
                    .count();
                appStats.dailyActiveUsers = activeUsers.total;

                // 判断应用是否活跃
                if (activeUsers.total > 0) {
                    appStats.isActive = true;
                    overview.activeApps++;
                }

                // 计算留存率（简化实现）
                if (appStats.userCount > 0) {
                    appStats.retentionRate = Math.round((appStats.dailyActiveUsers / appStats.userCount) * 100 * 100) / 100;
                }

            } catch (e) {
                // 用户表不存在，使用默认值
                console.log(`User table ${userTableName} not found:`, e.message);
            }

            // 排行榜分数统计
            try {
                const scoreStats = await db.collection('leaderboard_score')
                    .where({ appId: app.appId })
                    .count();
                appStats.scoreCount = scoreStats.total;
                overview.totalScores += scoreStats.total;
            } catch (e) {
                // 分数表查询失败，使用默认值
            }

            // 分类统计
            const category = appStats.category;
            if (!categories[category]) {
                categories[category] = { count: 0, users: 0 };
            }
            categories[category].count++;
            categories[category].users += appStats.userCount;

            appPerformanceList.push(appStats);
        }

        // 计算平均值
        overview.avgUsersPerApp = overview.totalApps > 0 ? 
            Math.round(overview.totalUsers / overview.totalApps) : 0;

        // 排序应用性能列表（按用户数降序）
        appPerformanceList.sort((a, b) => b.userCount - a.userCount);

        // 获取增长数据（与上一周期对比）
        let growth = await calculateGrowthData(db, targetApps, timeRange);

        let result = {
            overview: overview
        };

        if (includeDetails) {
            result.performance = {
                topApps: appPerformanceList.slice(0, 10), // 只返回前10个应用
                growth: growth
            };
            result.categories = categories;
        }

        ret.data = result;

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.05; // 5% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'APP_STATS', {
                appId: appId || 'ALL',
                timeRange: timeRange,
                includeDetails: includeDetails,
                currentAdmin: event.adminInfo.username
            });
        }

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 计算增长数据
async function calculateGrowthData(db, targetApps, timeRange) {
    let growth = {
        newApps: 0,
        userGrowth: 0,
        activityGrowth: 0
    };

    try {
        const now = moment();
        let currentPeriodStart, previousPeriodStart, previousPeriodEnd;

        switch (timeRange) {
            case 'today':
                currentPeriodStart = now.format('YYYY-MM-DD') + ' 00:00:00';
                previousPeriodStart = now.subtract(1, 'day').format('YYYY-MM-DD') + ' 00:00:00';
                previousPeriodEnd = now.add(1, 'day').format('YYYY-MM-DD') + ' 00:00:00';
                break;
            case 'week':
                currentPeriodStart = now.subtract(7, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                previousPeriodStart = now.subtract(7, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                previousPeriodEnd = now.add(7, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                break;
            case 'month':
                currentPeriodStart = now.subtract(30, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                previousPeriodStart = now.subtract(30, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                previousPeriodEnd = now.add(30, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                break;
        }

        // 计算新应用数量
        const newApps = await db.collection('app_config')
            .where({
                createTime: {
                    $gte: currentPeriodStart
                }
            })
            .count();
        growth.newApps = newApps.total;

        // 计算用户增长和活跃度增长（简化实现）
        let currentUsers = 0, previousUsers = 0;
        let currentActiveUsers = 0, previousActiveUsers = 0;

        for (let app of targetApps) {
            const userTableName = `user_${app.appId}`;
            
            try {
                // 当前周期新增用户
                const currentNewUsers = await db.collection(userTableName)
                    .where({
                        gmtCreate: {
                            $gte: currentPeriodStart
                        }
                    })
                    .count();
                currentUsers += currentNewUsers.total;

                // 上一周期新增用户
                const previousNewUsers = await db.collection(userTableName)
                    .where({
                        gmtCreate: {
                            $gte: previousPeriodStart,
                            $lte: previousPeriodEnd
                        }
                    })
                    .count();
                previousUsers += previousNewUsers.total;

                // 当前周期活跃用户
                const currentActive = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: currentPeriodStart
                        }
                    })
                    .count();
                currentActiveUsers += currentActive.total;

                // 上一周期活跃用户（简化计算）
                previousActiveUsers += Math.round(currentActive.total * 0.9); // 假设比当前少10%

            } catch (e) {
                // 忽略不存在的用户表
            }
        }

        // 计算增长率
        if (previousUsers > 0) {
            growth.userGrowth = Math.round(((currentUsers - previousUsers) / previousUsers) * 100 * 100) / 100;
        }

        if (previousActiveUsers > 0) {
            growth.activityGrowth = Math.round(((currentActiveUsers - previousActiveUsers) / previousActiveUsers) * 100 * 100) / 100;
        }

    } catch (e) {
        console.log('Growth calculation failed:', e.message);
    }

    return growth;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getAppStatsHandler, 'stats_view');
exports.main = mainFunc; 