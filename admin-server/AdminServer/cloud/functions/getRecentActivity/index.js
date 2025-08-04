const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getRecentActivity
 * 说明：获取最近活动数据
 * 权限：需要 stats_view 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | activityType | string | 否 | 活动类型 (all/user/score/admin) |
    | timeRange | string | 否 | 时间范围 (hour/day/week) |
    | limit | number | 否 | 返回数量 (默认20，最大100) |
    | appId | string | 否 | 应用ID筛选 |
 * 
 * 测试数据：
    {
        "activityType": "all",
        "timeRange": "day",
        "limit": 20,
        "appId": "test_game_001"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "activities": [
                {
                    "id": "activity_123456",
                    "type": "USER_REGISTER",
                    "description": "新用户注册",
                    "appId": "game_001",
                    "appName": "超级游戏",
                    "userId": "user_openid_123",
                    "userName": "玩家A",
                    "timestamp": "2023-10-02 15:30:00",
                    "details": {
                        "userCount": 1001,
                        "isNewRecord": false
                    }
                },
                {
                    "id": "activity_123457",
                    "type": "HIGH_SCORE",
                    "description": "新的高分记录",
                    "appId": "game_001",
                    "appName": "超级游戏",
                    "userId": "user_openid_456",
                    "userName": "玩家B",
                    "timestamp": "2023-10-02 15:25:00",
                    "details": {
                        "score": 9999,
                        "leaderboardId": "weekly_score",
                        "previousBest": 8888
                    }
                }
            ],
            "summary": {
                "totalActivities": 150,
                "userActivities": 85,
                "scoreActivities": 45,
                "adminActivities": 20,
                "peakHour": "15:00-16:00"
            },
            "criteria": {
                "activityType": "all",
                "timeRange": "day",
                "appId": "test_game_001"
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getRecentActivityHandler(event, context) {
    let activityType = event.activityType || 'all';
    let timeRange = event.timeRange || 'day';
    let limit = event.limit || 20;
    let appId = event.appId;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    const validActivityTypes = ['all', 'user', 'score', 'admin'];
    if (!validActivityTypes.includes(activityType)) {
        ret.code = 4001;
        ret.msg = "无效的活动类型";
        return ret;
    }

    const validTimeRanges = ['hour', 'day', 'week'];
    if (!validTimeRanges.includes(timeRange)) {
        ret.code = 4001;
        ret.msg = "无效的时间范围";
        return ret;
    }

    // 限制返回数量
    if (limit > 100) {
        limit = 100;
    }
    if (limit < 1) {
        limit = 20;
    }

    const db = cloud.database();

    try {
        // 计算时间范围
        const now = moment();
        let startDate, endDate;
        
        switch (timeRange) {
            case 'hour':
                startDate = now.subtract(1, 'hour').format('YYYY-MM-DD HH:mm:ss');
                endDate = now.format('YYYY-MM-DD HH:mm:ss');
                break;
            case 'day':
                startDate = now.subtract(1, 'day').format('YYYY-MM-DD HH:mm:ss');
                endDate = now.format('YYYY-MM-DD HH:mm:ss');
                break;
            case 'week':
                startDate = now.subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss');
                endDate = now.format('YYYY-MM-DD HH:mm:ss');
                break;
        }

        let activities = [];
        let summary = {
            totalActivities: 0,
            userActivities: 0,
            scoreActivities: 0,
            adminActivities: 0,
            peakHour: null
        };

        // 收集用户活动
        if (activityType === 'all' || activityType === 'user') {
            const userActivities = await collectUserActivities(db, appId, startDate, endDate, limit);
            activities.push(...userActivities);
            summary.userActivities = userActivities.length;
        }

        // 收集分数活动
        if (activityType === 'all' || activityType === 'score') {
            const scoreActivities = await collectScoreActivities(db, appId, startDate, endDate, limit);
            activities.push(...scoreActivities);
            summary.scoreActivities = scoreActivities.length;
        }

        // 收集管理员活动
        if (activityType === 'all' || activityType === 'admin') {
            const adminActivities = await collectAdminActivities(db, startDate, endDate, limit);
            activities.push(...adminActivities);
            summary.adminActivities = adminActivities.length;
        }

        // 按时间排序并限制数量
        activities.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
        activities = activities.slice(0, limit);

        // 添加活动ID
        activities = activities.map((activity, index) => ({
            id: `activity_${Date.now()}_${index}`,
            ...activity
        }));

        summary.totalActivities = activities.length;

        // 计算活跃时段
        summary.peakHour = calculatePeakHour(activities);

        ret.data = {
            activities: activities,
            summary: summary,
            criteria: {
                activityType: activityType,
                timeRange: timeRange,
                appId: appId || null
            }
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.1; // 10% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'RECENT_ACTIVITY', {
                activityType: activityType,
                timeRange: timeRange,
                limit: limit,
                appId: appId || 'ALL',
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

// 收集用户活动
async function collectUserActivities(db, appId, startDate, endDate, limit) {
    let activities = [];
    
    try {
        // 获取目标应用
        let targetApps = [];
        if (appId) {
            const appList = await db.collection('app_config')
                .where({ appId: appId })
                .get();
            targetApps = appList;
        } else {
            targetApps = await db.collection('app_config').limit(10).get(); // 限制应用数量以提高性能
        }

        for (let app of targetApps) {
            const userTableName = `user_${app.appId}`;
            
            try {
                // 获取新注册用户
                const newUsers = await db.collection(userTableName)
                    .where({
                        gmtCreate: {
                            $gte: startDate,
                            $lte: endDate
                        }
                    })
                    .orderBy('gmtCreate', 'desc')
                    .limit(Math.min(limit, 20))
                    .get();

                newUsers.forEach(user => {
                    activities.push({
                        type: 'USER_REGISTER',
                        description: '新用户注册',
                        appId: app.appId,
                        appName: app.appName,
                        userId: user.openId,
                        userName: user.nickName || '匿名用户',
                        timestamp: user.gmtCreate,
                        details: {
                            userDevice: user.deviceInfo || 'Unknown',
                            userLocation: `${user.province || ''} ${user.city || ''}`.trim() || 'Unknown'
                        }
                    });
                });

                // 获取活跃用户（最近修改时间）
                const activeUsers = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: startDate,
                            $lte: endDate
                        }
                    })
                    .orderBy('gmtModify', 'desc')
                    .limit(Math.min(limit, 15))
                    .get();

                activeUsers.forEach(user => {
                    // 避免重复（新注册用户）
                    const isNewUser = new Date(user.gmtCreate) >= new Date(startDate);
                    if (!isNewUser) {
                        activities.push({
                            type: 'USER_ACTIVE',
                            description: '用户活跃',
                            appId: app.appId,
                            appName: app.appName,
                            userId: user.openId,
                            userName: user.nickName || '匿名用户',
                            timestamp: user.gmtModify,
                            details: {
                                lastActionTime: user.gmtModify,
                                userData: user.userData ? '已更新游戏数据' : '基础活跃'
                            }
                        });
                    }
                });

            } catch (e) {
                console.log(`Error collecting user activities for ${userTableName}:`, e.message);
            }
        }
    } catch (e) {
        console.log('Error in collectUserActivities:', e.message);
    }

    return activities;
}

// 收集分数活动
async function collectScoreActivities(db, appId, startDate, endDate, limit) {
    let activities = [];
    
    try {
        let scoreQuery = {
            gmtCreate: {
                $gte: startDate,
                $lte: endDate
            }
        };
        
        if (appId) {
            scoreQuery.appId = appId;
        }

        const recentScores = await db.collection('leaderboard_score')
            .where(scoreQuery)
            .orderBy('gmtCreate', 'desc')
            .limit(limit)
            .get();

        for (let score of recentScores) {
            // 获取应用信息
            let appName = score.appId;
            try {
                const appInfo = await db.collection('app_config')
                    .where({ appId: score.appId })
                    .get();
                if (appInfo.length > 0) {
                    appName = appInfo[0].appName;
                }
            } catch (e) {
                // 应用信息获取失败，使用默认值
            }

            // 获取用户信息
            let userName = '匿名用户';
            try {
                const userTableName = `user_${score.appId}`;
                const userInfo = await db.collection(userTableName)
                    .where({ openId: score.openId })
                    .get();
                if (userInfo.length > 0) {
                    userName = userInfo[0].nickName || '匿名用户';
                }
            } catch (e) {
                // 用户信息获取失败，使用默认值
            }

            // 判断是否是高分记录
            const isHighScore = score.score >= 1000; // 简化判断
            
            activities.push({
                type: isHighScore ? 'HIGH_SCORE' : 'NEW_SCORE',
                description: isHighScore ? '新的高分记录' : '提交新分数',
                appId: score.appId,
                appName: appName,
                userId: score.openId,
                userName: userName,
                timestamp: score.gmtCreate,
                details: {
                    score: score.score,
                    leaderboardId: score.leaderboardId,
                    isRecord: isHighScore
                }
            });
        }
    } catch (e) {
        console.log('Error in collectScoreActivities:', e.message);
    }

    return activities;
}

// 收集管理员活动
async function collectAdminActivities(db, startDate, endDate, limit) {
    let activities = [];
    
    try {
        // 这里应该查询操作日志表，但由于示例中没有，我们模拟一些管理员活动
        // 实际实现时应该从 admin_operation_logs 表查询
        
        // 获取最近创建的应用（管理员操作）
        const recentApps = await db.collection('app_config')
            .where({
                createTime: {
                    $gte: startDate,
                    $lte: endDate
                }
            })
            .orderBy('createTime', 'desc')
            .limit(Math.min(limit, 10))
            .get();

        recentApps.forEach(app => {
            activities.push({
                type: 'ADMIN_CREATE_APP',
                description: '管理员创建应用',
                appId: app.appId,
                appName: app.appName,
                userId: app.createdBy || 'system',
                userName: app.createdBy || '系统管理员',
                timestamp: app.createTime,
                details: {
                    operation: 'CREATE_APP',
                    target: app.appName,
                    category: app.category || 'default'
                }
            });
        });

        // 获取最近创建的排行榜
        const recentLeaderboards = await db.collection('leaderboard_config')
            .where({
                createTime: {
                    $gte: startDate,
                    $lte: endDate
                }
            })
            .orderBy('createTime', 'desc')
            .limit(Math.min(limit, 10))
            .get();

        recentLeaderboards.forEach(leaderboard => {
            activities.push({
                type: 'ADMIN_CREATE_LEADERBOARD',
                description: '管理员创建排行榜',
                appId: leaderboard.appId,
                appName: leaderboard.appId, // 这里应该查询app名称
                userId: leaderboard.createdBy || 'system',
                userName: leaderboard.createdBy || '系统管理员',
                timestamp: leaderboard.createTime,
                details: {
                    operation: 'CREATE_LEADERBOARD',
                    target: leaderboard.name,
                    leaderboardId: leaderboard.leaderboardId
                }
            });
        });

    } catch (e) {
        console.log('Error in collectAdminActivities:', e.message);
    }

    return activities;
}

// 计算活跃时段
function calculatePeakHour(activities) {
    if (activities.length === 0) {
        return null;
    }

    const hourCounts = {};
    
    activities.forEach(activity => {
        const hour = moment(activity.timestamp).format('HH');
        hourCounts[hour] = (hourCounts[hour] || 0) + 1;
    });

    let peakHour = null;
    let maxCount = 0;
    
    for (let [hour, count] of Object.entries(hourCounts)) {
        if (count > maxCount) {
            maxCount = count;
            peakHour = hour;
        }
    }

    return peakHour ? `${peakHour}:00-${(parseInt(peakHour) + 1).toString().padStart(2, '0')}:00` : null;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getRecentActivityHandler, 'stats_view');
exports.main = mainFunc; 