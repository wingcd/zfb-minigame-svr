const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getUserGrowth
 * 说明：获取用户增长统计数据
 * 权限：需要 stats_view 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 否 | 应用ID（为空时统计所有应用） |
    | timeRange | string | 否 | 时间范围 (week/month/quarter/year) |
    | startDate | string | 否 | 开始日期 (YYYY-MM-DD) |
    | endDate | string | 否 | 结束日期 (YYYY-MM-DD) |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "timeRange": "month",
        "startDate": "2023-09-01",
        "endDate": "2023-10-01"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "summary": {
                "totalUsers": 15000,
                "newUsers": 1200,
                "activeUsers": 8500,
                "retainedUsers": 7200,
                "growthRate": 8.7
            },
            "dailyData": [
                {
                    "date": "2023-09-01",
                    "newUsers": 45,
                    "activeUsers": 320,
                    "retentionRate": 75.5
                }
            ],
            "trends": {
                "newUserTrend": "up",
                "activeUserTrend": "up", 
                "retentionTrend": "stable"
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
async function getUserGrowthHandler(event, context) {
    let appId = event.appId;
    let timeRange = event.timeRange || 'month';
    let startDate = event.startDate;
    let endDate = event.endDate;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验和默认值设置
    const validTimeRanges = ['week', 'month', 'quarter', 'year'];
    if (!validTimeRanges.includes(timeRange)) {
        ret.code = 4001;
        ret.msg = "无效的时间范围";
        return ret;
    }

    // 设置默认日期范围
    if (!startDate || !endDate) {
        const now = moment();
        switch (timeRange) {
            case 'week':
                startDate = now.subtract(7, 'days').format('YYYY-MM-DD');
                endDate = now.format('YYYY-MM-DD');
                break;
            case 'month':
                startDate = now.subtract(30, 'days').format('YYYY-MM-DD');
                endDate = now.format('YYYY-MM-DD');
                break;
            case 'quarter':
                startDate = now.subtract(90, 'days').format('YYYY-MM-DD');
                endDate = now.format('YYYY-MM-DD');
                break;
            case 'year':
                startDate = now.subtract(365, 'days').format('YYYY-MM-DD');
                endDate = now.format('YYYY-MM-DD');
                break;
        }
    }

    const db = cloud.database();

    try {
        // 如果指定了appId，验证应用是否存在
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
            // 获取所有应用
            targetApps = await db.collection('app_config').get();
        }

        let summary = {
            totalUsers: 0,
            newUsers: 0,
            activeUsers: 0,
            retainedUsers: 0,
            growthRate: 0
        };

        let dailyDataMap = new Map();
        let previousPeriodData = {
            totalUsers: 0,
            newUsers: 0
        };

        // 计算上一周期的数据用于对比
        const prevStartDate = moment(startDate).subtract(moment(endDate).diff(moment(startDate), 'days'), 'days').format('YYYY-MM-DD');
        const prevEndDate = startDate;

        for (let app of targetApps) {
            const userTableName = `user_${app.appId}`;
            
            try {
                // 总用户数
                const totalUsers = await db.collection(userTableName).count();
                summary.totalUsers += totalUsers.total;

                // 当前周期新增用户
                const newUsers = await db.collection(userTableName)
                    .where({
                        gmtCreate: {
                            $gte: startDate + ' 00:00:00',
                            $lte: endDate + ' 23:59:59'
                        }
                    })
                    .get();
                summary.newUsers += newUsers.length;

                // 当前周期活跃用户
                const activeUsers = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: startDate + ' 00:00:00',
                            $lte: endDate + ' 23:59:59'
                        }
                    })
                    .count();
                summary.activeUsers += activeUsers.total;

                // 上一周期数据（用于计算增长率）
                const prevNewUsers = await db.collection(userTableName)
                    .where({
                        gmtCreate: {
                            $gte: prevStartDate + ' 00:00:00',
                            $lte: prevEndDate + ' 23:59:59'
                        }
                    })
                    .count();
                previousPeriodData.newUsers += prevNewUsers.total;

                // 计算每日数据
                let currentDate = moment(startDate);
                const endDateMoment = moment(endDate);

                while (currentDate.isSameOrBefore(endDateMoment)) {
                    const dateStr = currentDate.format('YYYY-MM-DD');
                    
                    if (!dailyDataMap.has(dateStr)) {
                        dailyDataMap.set(dateStr, {
                            date: dateStr,
                            newUsers: 0,
                            activeUsers: 0,
                            retentionRate: 0
                        });
                    }

                    // 每日新增用户
                    const dailyNew = await db.collection(userTableName)
                        .where({
                            gmtCreate: {
                                $gte: dateStr + ' 00:00:00',
                                $lte: dateStr + ' 23:59:59'
                            }
                        })
                        .count();
                    
                    const dailyData = dailyDataMap.get(dateStr);
                    dailyData.newUsers += dailyNew.total;

                    // 每日活跃用户
                    const dailyActive = await db.collection(userTableName)
                        .where({
                            gmtModify: {
                                $gte: dateStr + ' 00:00:00',
                                $lte: dateStr + ' 23:59:59'
                            }
                        })
                        .count();
                    
                    dailyData.activeUsers += dailyActive.total;

                    currentDate.add(1, 'day');
                }

            } catch (e) {
                // 用户表不存在，跳过
                console.log(`User table ${userTableName} not found:`, e.message);
            }
        }

        // 计算留存率（简化实现）
        summary.retainedUsers = Math.round(summary.activeUsers * 0.75); // 假设75%的活跃用户为留存用户

        // 计算增长率
        if (previousPeriodData.newUsers > 0) {
            summary.growthRate = Math.round(((summary.newUsers - previousPeriodData.newUsers) / previousPeriodData.newUsers) * 100 * 100) / 100;
        }

        // 转换每日数据
        const dailyData = Array.from(dailyDataMap.values()).sort((a, b) => a.date.localeCompare(b.date));
        
        // 计算每日留存率
        dailyData.forEach(daily => {
            if (daily.newUsers > 0) {
                daily.retentionRate = Math.round((daily.activeUsers / (daily.newUsers + daily.activeUsers * 0.5)) * 100 * 100) / 100;
            }
        });

        // 计算趋势
        const trends = calculateTrends(dailyData);

        ret.data = {
            summary: summary,
            dailyData: dailyData,
            trends: trends,
            dateRange: {
                startDate: startDate,
                endDate: endDate,
                timeRange: timeRange
            }
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.1; // 10% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'USER_GROWTH_STATS', {
                appId: appId || 'ALL',
                timeRange: timeRange,
                dateRange: `${startDate} to ${endDate}`,
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

// 计算趋势
function calculateTrends(dailyData) {
    if (dailyData.length < 2) {
        return {
            newUserTrend: 'stable',
            activeUserTrend: 'stable',
            retentionTrend: 'stable'
        };
    }

    const firstHalf = dailyData.slice(0, Math.floor(dailyData.length / 2));
    const secondHalf = dailyData.slice(Math.floor(dailyData.length / 2));

    const firstHalfNewUsers = firstHalf.reduce((sum, day) => sum + day.newUsers, 0);
    const secondHalfNewUsers = secondHalf.reduce((sum, day) => sum + day.newUsers, 0);

    const firstHalfActiveUsers = firstHalf.reduce((sum, day) => sum + day.activeUsers, 0);
    const secondHalfActiveUsers = secondHalf.reduce((sum, day) => sum + day.activeUsers, 0);

    const firstHalfRetention = firstHalf.reduce((sum, day) => sum + day.retentionRate, 0) / firstHalf.length;
    const secondHalfRetention = secondHalf.reduce((sum, day) => sum + day.retentionRate, 0) / secondHalf.length;

    return {
        newUserTrend: secondHalfNewUsers > firstHalfNewUsers * 1.05 ? 'up' : 
                     secondHalfNewUsers < firstHalfNewUsers * 0.95 ? 'down' : 'stable',
        activeUserTrend: secondHalfActiveUsers > firstHalfActiveUsers * 1.05 ? 'up' : 
                        secondHalfActiveUsers < firstHalfActiveUsers * 0.95 ? 'down' : 'stable',
        retentionTrend: secondHalfRetention > firstHalfRetention * 1.05 ? 'up' : 
                       secondHalfRetention < firstHalfRetention * 0.95 ? 'down' : 'stable'
    };
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getUserGrowthHandler, 'stats_view');
exports.main = mainFunc; 