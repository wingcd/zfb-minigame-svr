const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getTopApps
 * 说明：获取热门应用排行数据
 * 权限：需要 stats_view 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | sortBy | string | 否 | 排序方式 (users/scores/activity/retention) |
    | timeRange | string | 否 | 时间范围 (week/month/quarter) |
    | limit | number | 否 | 返回数量 (默认10，最大50) |
    | category | string | 否 | 应用分类筛选 |
 * 
 * 测试数据：
    {
        "sortBy": "users",
        "timeRange": "month",
        "limit": 10,
        "category": "game"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "topApps": [
                {
                    "rank": 1,
                    "appId": "super_game_001",
                    "appName": "超级游戏",
                    "category": "game",
                    "userCount": 15000,
                    "scoreCount": 45000,
                    "dailyActiveUsers": 2800,
                    "retentionRate": 78.5,
                    "growthRate": 15.2,
                    "createTime": "2023-08-15 10:30:00"
                }
            ],
            "summary": {
                "totalApps": 25,
                "totalUsers": 65000,
                "avgGrowthRate": 12.8,
                "topCategory": "game"
            },
            "criteria": {
                "sortBy": "users",
                "timeRange": "month",
                "category": "game"
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getTopAppsHandler(event, context) {
    let sortBy = event.sortBy || 'users';
    let timeRange = event.timeRange || 'month';
    let limit = event.limit || 10;
    let category = event.category;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    const validSortBy = ['users', 'scores', 'activity', 'retention'];
    if (!validSortBy.includes(sortBy)) {
        ret.code = 4001;
        ret.msg = "无效的排序方式";
        return ret;
    }

    const validTimeRanges = ['week', 'month', 'quarter'];
    if (!validTimeRanges.includes(timeRange)) {
        ret.code = 4001;
        ret.msg = "无效的时间范围";
        return ret;
    }

    // 限制返回数量
    if (limit > 50) {
        limit = 50;
    }
    if (limit < 1) {
        limit = 10;
    }

    const db = cloud.database();

    try {
        // 构建应用查询条件
        let appQuery = {};
        if (category) {
            appQuery.category = category;
        }

        // 获取应用列表
        const appList = await db.collection('app_config')
            .where(appQuery)
            .get();

        if (appList.length === 0) {
            ret.data = {
                topApps: [],
                summary: {
                    totalApps: 0,
                    totalUsers: 0,
                    avgGrowthRate: 0,
                    topCategory: null
                },
                criteria: {
                    sortBy: sortBy,
                    timeRange: timeRange,
                    category: category
                }
            };
            return ret;
        }

        // 计算时间范围
        const now = moment();
        let startDate, endDate, prevStartDate, prevEndDate;
        
        switch (timeRange) {
            case 'week':
                startDate = now.clone().subtract(7, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                endDate = now.format('YYYY-MM-DD') + ' 23:59:59';
                prevStartDate = now.clone().subtract(14, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                prevEndDate = now.clone().subtract(7, 'days').format('YYYY-MM-DD') + ' 23:59:59';
                break;
            case 'month':
                startDate = now.clone().subtract(30, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                endDate = now.format('YYYY-MM-DD') + ' 23:59:59';
                prevStartDate = now.clone().subtract(60, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                prevEndDate = now.clone().subtract(30, 'days').format('YYYY-MM-DD') + ' 23:59:59';
                break;
            case 'quarter':
                startDate = now.clone().subtract(90, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                endDate = now.format('YYYY-MM-DD') + ' 23:59:59';
                prevStartDate = now.clone().subtract(180, 'days').format('YYYY-MM-DD') + ' 00:00:00';
                prevEndDate = now.clone().subtract(90, 'days').format('YYYY-MM-DD') + ' 23:59:59';
                break;
        }

        let appStatsArray = [];
        let totalUsers = 0;
        let categoryStats = {};

        // 收集每个应用的统计数据
        for (let app of appList) {
            const userTableName = `user_${app.appId}`;
            
            let appStats = {
                appId: app.appId,
                appName: app.appName,
                category: app.category || 'other',
                userCount: 0,
                scoreCount: 0,
                dailyActiveUsers: 0,
                retentionRate: 0,
                growthRate: 0,
                createTime: app.createTime
            };

            try {
                // 总用户数
                const totalUsersCount = await db.collection(userTableName).count();
                appStats.userCount = totalUsersCount.total;
                totalUsers += totalUsersCount.total;

                // 当前周期活跃用户
                const currentActiveUsers = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: startDate,
                            $lte: endDate
                        }
                    })
                    .count();
                appStats.dailyActiveUsers = currentActiveUsers.total;

                // 上一周期活跃用户（用于计算增长率）
                const prevActiveUsers = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: prevStartDate,
                            $lte: prevEndDate
                        }
                    })
                    .count();

                // 计算增长率
                if (prevActiveUsers.total > 0) {
                    appStats.growthRate = Math.round(((currentActiveUsers.total - prevActiveUsers.total) / prevActiveUsers.total) * 100 * 100) / 100;
                }

                // 计算留存率
                if (appStats.userCount > 0) {
                    appStats.retentionRate = Math.round((appStats.dailyActiveUsers / appStats.userCount) * 100 * 100) / 100;
                }

            } catch (e) {
                // 用户表不存在，使用默认值
                console.log(`User table ${userTableName} not found:`, e.message);
            }

            // 获取分数统计
            try {
                const scoreStats = await db.collection('leaderboard_score')
                    .where({ appId: app.appId })
                    .count();
                appStats.scoreCount = scoreStats.total;
            } catch (e) {
                // 分数表查询失败，使用默认值
            }

            // 分类统计
            const categoryKey = appStats.category;
            if (!categoryStats[categoryKey]) {
                categoryStats[categoryKey] = { count: 0, users: 0 };
            }
            categoryStats[categoryKey].count++;
            categoryStats[categoryKey].users += appStats.userCount;

            appStatsArray.push(appStats);
        }

        // 根据指定条件排序
        switch (sortBy) {
            case 'users':
                appStatsArray.sort((a, b) => b.userCount - a.userCount);
                break;
            case 'scores':
                appStatsArray.sort((a, b) => b.scoreCount - a.scoreCount);
                break;
            case 'activity':
                appStatsArray.sort((a, b) => b.dailyActiveUsers - a.dailyActiveUsers);
                break;
            case 'retention':
                appStatsArray.sort((a, b) => b.retentionRate - a.retentionRate);
                break;
        }

        // 取前N个应用并添加排名
        const topApps = appStatsArray.slice(0, limit).map((app, index) => ({
            rank: index + 1,
            ...app
        }));

        // 计算汇总数据
        const avgGrowthRate = appStatsArray.length > 0 ? 
            Math.round((appStatsArray.reduce((sum, app) => sum + app.growthRate, 0) / appStatsArray.length) * 100) / 100 : 0;

        // 找出用户数最多的分类
        let topCategory = null;
        let maxUsers = 0;
        for (let [category, stats] of Object.entries(categoryStats)) {
            if (stats.users > maxUsers) {
                maxUsers = stats.users;
                topCategory = category;
            }
        }

        ret.data = {
            topApps: topApps,
            summary: {
                totalApps: appList.length,
                totalUsers: totalUsers,
                avgGrowthRate: avgGrowthRate,
                topCategory: topCategory
            },
            criteria: {
                sortBy: sortBy,
                timeRange: timeRange,
                category: category || null
            }
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.05; // 5% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'TOP_APPS', {
                sortBy: sortBy,
                timeRange: timeRange,
                limit: limit,
                category: category || 'ALL',
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

// 导出带权限校验的函数
const mainFunc = requirePermission(getTopAppsHandler, 'stats_view');
exports.main = mainFunc; 