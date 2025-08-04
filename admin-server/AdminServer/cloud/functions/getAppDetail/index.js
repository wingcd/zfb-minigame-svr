const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getAppDetail
 * 说明：获取应用详细信息
 * 权限：需要 app_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
 * 
 * 测试数据：
    {
        "appId": "test_game_001"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "appInfo": {
                "id": "app_config_id_123456",
                "appId": "test_game_001",
                "appName": "测试游戏",
                "description": "一个测试用的小游戏",
                "channelAppKey": "channel_key_123",
                "category": "game",
                "status": "active",
                "createTime": "2023-10-01 10:00:00",
                "userCount": 1250,
                "scoreCount": 3500
            },
            "statistics": {
                "dailyActiveUsers": 125,
                "weeklyActiveUsers": 450,
                "monthlyActiveUsers": 890,
                "totalScores": 3500,
                "avgScore": 280.5
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
async function getAppDetailHandler(event, context) {
    let appId = event.appId;

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
        ret.msg = "应用ID不能为空";
        return ret;
    }

    const db = cloud.database();

    try {
        // 获取应用信息
        const appList = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (appList.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        const appInfo = appList[0];

        // 获取用户统计
        const userTableName = `user_${appId}`;
        let userCount = 0;
        let dailyActiveUsers = 0;
        let weeklyActiveUsers = 0;
        let monthlyActiveUsers = 0;

        try {
            // 总用户数
            const totalUsers = await db.collection(userTableName).count();
            userCount = totalUsers.total;

            // 今日活跃用户
            const today = new Date().toISOString().split('T')[0];
            const dailyActive = await db.collection(userTableName)
                .where({
                    gmtModify: {
                        $gte: today + ' 00:00:00',
                        $lte: today + ' 23:59:59'
                    }
                })
                .count();
            dailyActiveUsers = dailyActive.total;

            // 本周活跃用户
            const weekStart = new Date();
            weekStart.setDate(weekStart.getDate() - weekStart.getDay());
            const weekStartStr = weekStart.toISOString().split('T')[0];
            
            const weeklyActive = await db.collection(userTableName)
                .where({
                    gmtModify: {
                        $gte: weekStartStr + ' 00:00:00'
                    }
                })
                .count();
            weeklyActiveUsers = weeklyActive.total;

            // 本月活跃用户
            const monthStart = new Date();
            monthStart.setDate(1);
            const monthStartStr = monthStart.toISOString().split('T')[0];
            
            const monthlyActive = await db.collection(userTableName)
                .where({
                    gmtModify: {
                        $gte: monthStartStr + ' 00:00:00'
                    }
                })
                .count();
            monthlyActiveUsers = monthlyActive.total;

        } catch (e) {
            // 用户表不存在或查询失败，使用默认值
        }

        // 获取排行榜统计
        let totalScores = 0;
        let avgScore = 0;

        try {
            const scoreStats = await db.collection('leaderboard_score')
                .where({ appId: appId })
                .count();
            totalScores = scoreStats.total;

            if (totalScores > 0) {
                // 计算平均分（简化实现）
                const scoreList = await db.collection('leaderboard_score')
                    .where({ appId: appId })
                    .limit(1000)
                    .get();
                
                const totalScoreValue = scoreList.reduce((sum, score) => sum + (score.score || 0), 0);
                avgScore = scoreList.length > 0 ? (totalScoreValue / scoreList.length) : 0;
            }
        } catch (e) {
            // 分数表查询失败，使用默认值
        }

        // 更新应用统计缓存
        await db.collection('app_config')
            .where({ appId: appId })
            .update({
                data: {
                    userCount: userCount,
                    scoreCount: totalScores,
                    updateTime: new Date().toISOString()
                }
            });

        ret.data = {
            appInfo: {
                id: appInfo._id,
                appId: appInfo.appId,
                appName: appInfo.appName,
                description: appInfo.description,
                channelAppKey: appInfo.channelAppKey,
                category: appInfo.category,
                status: appInfo.status,
                createTime: appInfo.createTime,
                userCount: userCount,
                scoreCount: totalScores
            },
            statistics: {
                dailyActiveUsers: dailyActiveUsers,
                weeklyActiveUsers: weeklyActiveUsers,
                monthlyActiveUsers: monthlyActiveUsers,
                totalScores: totalScores,
                avgScore: Math.round(avgScore * 100) / 100
            }
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.1; // 10% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'APP_DETAIL', {
                appId: appId,
                appName: appInfo.appName,
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
const mainFunc = requirePermission(getAppDetailHandler, 'app_manage');
exports.main = mainFunc; 