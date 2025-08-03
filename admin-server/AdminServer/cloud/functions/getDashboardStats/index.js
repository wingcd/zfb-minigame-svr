const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");

// 请求参数
/**
 * 函数：getDashboardStats
 * 说明：获取仪表板统计数据
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | timeRange | string | 否 | 时间范围: today, week, month |
  * 测试数据
    {
        "timeRange": "week"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "apps": {
                "total": 10,
                "active": 8,
                "newThisMonth": 2,
                "totalUsers": 1000,
                "change": 5.2
            },
            "users": {
                "total": 1000,
                "newToday": 50,
                "activeToday": 200,
                "banned": 5,
                "change": 8.5
            },
            "activity": {
                "daily": 200,
                "weekly": 800,
                "monthly": 3000,
                "change": 12.3
            },
            "leaderboards": {
                "total": 25,
                "active": 20,
                "totalScores": 5000,
                "change": -2.1
            }
        }
    }
 */

exports.main = async (event, context) => {
    let timeRange = event.timeRange || 'week';

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    const db = cloud.database();

    try {
        // 时间范围计算
        const now = moment();
        const today = now.format('YYYY-MM-DD');
        const yesterday = now.subtract(1, 'day').format('YYYY-MM-DD');
        const thisMonth = now.format('YYYY-MM');
        const lastMonth = now.subtract(1, 'month').format('YYYY-MM');

        // 应用统计
        const appStats = await getAppStats(db, today, thisMonth);
        
        // 用户统计
        const userStats = await getUserStats(db, today, yesterday);
        
        // 活跃度统计
        const activityStats = await getActivityStats(db, today, timeRange);
        
        // 排行榜统计
        const leaderboardStats = await getLeaderboardStats(db);

        ret.data = {
            apps: appStats,
            users: userStats,
            activity: activityStats,
            leaderboards: leaderboardStats
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
};

// 获取应用统计
async function getAppStats(db, today, thisMonth) {
    try {
        // 总应用数
        const totalApps = await db.collection('app_config').count();
        
        // 活跃应用数（本月有用户活动的应用）
        const activeApps = await db.collection('app_config')
            .where({ status: 'active' })
            .count();
        
        // 本月新增应用
        const newThisMonth = await db.collection('app_config')
            .where({
                createTime: {
                    $gte: thisMonth + '-01 00:00:00',
                    $lte: thisMonth + '-31 23:59:59'
                }
            })
            .count();

        // 计算总用户数（所有应用的用户数之和）
        let totalUsers = 0;
        const apps = await db.collection('app_config').get();
        
        for (let app of apps) {
            try {
                const userTableName = `user_${app.appId}`;
                const userCount = await db.collection(userTableName).count();
                totalUsers += userCount.total;
            } catch (e) {
                // 用户表不存在，忽略
            }
        }

        return {
            total: totalApps.total,
            active: activeApps.total,
            newThisMonth: newThisMonth.total,
            totalUsers: totalUsers,
            change: Math.random() * 10 - 5 // 模拟变化率
        };
    } catch (e) {
        return {
            total: 0,
            active: 0,
            newThisMonth: 0,
            totalUsers: 0,
            change: 0
        };
    }
}

// 获取用户统计
async function getUserStats(db, today, yesterday) {
    try {
        let totalUsers = 0;
        let newToday = 0;
        let activeToday = 0;
        let banned = 0;

        const apps = await db.collection('app_config').get();
        
        for (let app of apps) {
            try {
                const userTableName = `user_${app.appId}`;
                
                // 总用户数
                const userCount = await db.collection(userTableName).count();
                totalUsers += userCount.total;
                
                // 今日新增
                const newTodayCount = await db.collection(userTableName)
                    .where({
                        gmtCreate: {
                            $gte: today + ' 00:00:00',
                            $lte: today + ' 23:59:59'
                        }
                    })
                    .count();
                newToday += newTodayCount.total;
                
                // 今日活跃
                const activeTodayCount = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: today + ' 00:00:00',
                            $lte: today + ' 23:59:59'
                        }
                    })
                    .count();
                activeToday += activeTodayCount.total;
                
                // 封禁用户
                const bannedCount = await db.collection(userTableName)
                    .where({ banned: true })
                    .count();
                banned += bannedCount.total;
                
            } catch (e) {
                // 用户表不存在，忽略
            }
        }

        return {
            total: totalUsers,
            newToday: newToday,
            activeToday: activeToday,
            banned: banned,
            change: Math.random() * 15 - 7.5 // 模拟变化率
        };
    } catch (e) {
        return {
            total: 0,
            newToday: 0,
            activeToday: 0,
            banned: 0,
            change: 0
        };
    }
}

// 获取活跃度统计
async function getActivityStats(db, today, timeRange) {
    try {
        let daily = 0;
        let weekly = 0;
        let monthly = 0;

        const apps = await db.collection('app_config').get();
        
        for (let app of apps) {
            try {
                const userTableName = `user_${app.appId}`;
                
                // 今日活跃
                const dailyCount = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: today + ' 00:00:00',
                            $lte: today + ' 23:59:59'
                        }
                    })
                    .count();
                daily += dailyCount.total;
                
                // 本周活跃（简化实现）
                const weekStart = moment().startOf('week').format('YYYY-MM-DD');
                const weeklyCount = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: weekStart + ' 00:00:00',
                            $lte: today + ' 23:59:59'
                        }
                    })
                    .count();
                weekly += weeklyCount.total;
                
                // 本月活跃
                const monthStart = moment().startOf('month').format('YYYY-MM-DD');
                const monthlyCount = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: monthStart + ' 00:00:00',
                            $lte: today + ' 23:59:59'
                        }
                    })
                    .count();
                monthly += monthlyCount.total;
                
            } catch (e) {
                // 用户表不存在，忽略
            }
        }

        return {
            daily: daily,
            weekly: weekly,
            monthly: monthly,
            change: Math.random() * 20 - 10 // 模拟变化率
        };
    } catch (e) {
        return {
            daily: 0,
            weekly: 0,
            monthly: 0,
            change: 0
        };
    }
}

// 获取排行榜统计
async function getLeaderboardStats(db) {
    try {
        // 总排行榜数
        const totalLeaderboards = await db.collection('leaderboard_config').count();
        
        // 活跃排行榜数
        const activeLeaderboards = await db.collection('leaderboard_config')
            .where({ enabled: true })
            .count();
        
        // 总分数记录数
        const totalScores = await db.collection('leaderboard_score').count();

        return {
            total: totalLeaderboards.total,
            active: activeLeaderboards.total,
            totalScores: totalScores.total,
            change: Math.random() * 8 - 4 // 模拟变化率
        };
    } catch (e) {
        return {
            total: 0,
            active: 0,
            totalScores: 0,
            change: 0
        };
    }
} 