const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getLeaderboardStats
 * 说明：获取排行榜统计数据
 * 权限：需要 stats_view 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 否 | 应用ID（为空时统计所有应用） |
    | leaderboardId | string | 否 | 排行榜ID（为空时统计所有排行榜） |
    | timeRange | string | 否 | 时间范围 (today/week/month) |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "leaderboardId": "weekly_score",
        "timeRange": "week"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "overview": {
                "totalLeaderboards": 50,
                "activeLeaderboards": 35,
                "totalScores": 125000,
                "totalParticipants": 8500,
                "avgScoresPerLeaderboard": 2500
            },
            "performance": {
                "topLeaderboards": [
                    {
                        "appId": "game_001",
                        "leaderboardId": "weekly_score",
                        "name": "周榜",
                        "scoreCount": 5000,
                        "participantCount": 800,
                        "avgScore": 1250.5,
                        "activityLevel": "high"
                    }
                ],
                "scoreDistribution": {
                    "ranges": [
                        { "range": "0-100", "count": 1500 },
                        { "range": "101-500", "count": 2800 },
                        { "range": "501-1000", "count": 1200 },
                        { "range": "1000+", "count": 500 }
                    ]
                }
            },
            "trends": {
                "newScores": 1200,
                "scoreGrowth": 15.5,
                "participantGrowth": 8.3
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 应用或排行榜不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getLeaderboardStatsHandler(event, context) {
    let appId = event.appId;
    let leaderboardId = event.leaderboardId;
    let timeRange = event.timeRange || 'week';

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
        // 构建查询条件
        let leaderboardQuery = {};
        if (appId) {
            leaderboardQuery.appId = appId;
        }
        if (leaderboardId) {
            leaderboardQuery.leaderboardId = leaderboardId;
        }

        // 获取排行榜配置
        const leaderboardConfigs = await db.collection('leaderboard_config')
            .where(leaderboardQuery)
            .get();

        if (leaderboardConfigs.length === 0) {
            ret.code = 4004;
            ret.msg = "未找到符合条件的排行榜";
            return ret;
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
            totalLeaderboards: leaderboardConfigs.length,
            activeLeaderboards: 0,
            totalScores: 0,
            totalParticipants: 0,
            avgScoresPerLeaderboard: 0
        };

        let leaderboardPerformanceList = [];
        let scoreDistribution = {
            ranges: [
                { range: "0-100", count: 0 },
                { range: "101-500", count: 0 },
                { range: "501-1000", count: 0 },
                { range: "1000+", count: 0 }
            ]
        };

        let participantSet = new Set(); // 用于统计总参与人数

        // 遍历每个排行榜收集统计数据
        for (let leaderboard of leaderboardConfigs) {
            let leaderboardStats = {
                appId: leaderboard.appId,
                leaderboardId: leaderboard.leaderboardId,
                name: leaderboard.name,
                scoreCount: 0,
                participantCount: 0,
                avgScore: 0,
                maxScore: 0,
                minScore: 0,
                activityLevel: 'low'
            };

            try {
                // 获取排行榜的所有分数记录
                const allScores = await db.collection('leaderboard_score')
                    .where({
                        appId: leaderboard.appId,
                        leaderboardId: leaderboard.leaderboardId
                    })
                    .get();

                leaderboardStats.scoreCount = allScores.length;
                overview.totalScores += allScores.length;

                if (allScores.length > 0) {
                    // 计算参与人数（去重）
                    const participants = new Set(allScores.map(score => score.openId));
                    leaderboardStats.participantCount = participants.size;
                    
                    // 添加到总参与人数统计
                    participants.forEach(openId => participantSet.add(`${leaderboard.appId}:${openId}`));

                    // 计算分数统计
                    const scores = allScores.map(s => s.score || 0);
                    const totalScore = scores.reduce((sum, score) => sum + score, 0);
                    leaderboardStats.avgScore = Math.round((totalScore / scores.length) * 100) / 100;
                    leaderboardStats.maxScore = Math.max(...scores);
                    leaderboardStats.minScore = Math.min(...scores);

                    // 计算分数分布
                    scores.forEach(score => {
                        if (score <= 100) {
                            scoreDistribution.ranges[0].count++;
                        } else if (score <= 500) {
                            scoreDistribution.ranges[1].count++;
                        } else if (score <= 1000) {
                            scoreDistribution.ranges[2].count++;
                        } else {
                            scoreDistribution.ranges[3].count++;
                        }
                    });

                    // 判断活跃度
                    const recentScores = allScores.filter(score => {
                        const scoreDate = score.gmtCreate || score.createTime;
                        return scoreDate >= startDate && scoreDate <= endDate;
                    });

                    if (recentScores.length > 0) {
                        overview.activeLeaderboards++;
                        
                        if (recentScores.length >= 100) {
                            leaderboardStats.activityLevel = 'high';
                        } else if (recentScores.length >= 20) {
                            leaderboardStats.activityLevel = 'medium';
                        } else {
                            leaderboardStats.activityLevel = 'low';
                        }
                    }
                }

            } catch (e) {
                console.log(`Error processing leaderboard ${leaderboard.leaderboardId}:`, e.message);
            }

            leaderboardPerformanceList.push(leaderboardStats);
        }

        // 计算总参与人数
        overview.totalParticipants = participantSet.size;

        // 计算平均分数
        overview.avgScoresPerLeaderboard = overview.totalLeaderboards > 0 ? 
            Math.round(overview.totalScores / overview.totalLeaderboards) : 0;

        // 排序排行榜性能列表（按分数数量降序）
        leaderboardPerformanceList.sort((a, b) => b.scoreCount - a.scoreCount);

        // 计算趋势数据
        let trends = await calculateLeaderboardTrends(db, leaderboardConfigs, timeRange, startDate, endDate);

        ret.data = {
            overview: overview,
            performance: {
                topLeaderboards: leaderboardPerformanceList.slice(0, 10), // 只返回前10个排行榜
                scoreDistribution: scoreDistribution
            },
            trends: trends
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.05; // 5% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'LEADERBOARD_STATS', {
                appId: appId || 'ALL',
                leaderboardId: leaderboardId || 'ALL',
                timeRange: timeRange,
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

// 计算排行榜趋势数据
async function calculateLeaderboardTrends(db, leaderboardConfigs, timeRange, startDate, endDate) {
    let trends = {
        newScores: 0,
        scoreGrowth: 0,
        participantGrowth: 0
    };

    try {
        // 计算当前周期和上一周期的时间范围
        const currentStart = startDate;
        const currentEnd = endDate;
        
        const period = moment(endDate).diff(moment(startDate), 'days');
        const previousStart = moment(startDate).subtract(period, 'days').format('YYYY-MM-DD') + ' 00:00:00';
        const previousEnd = startDate;

        let currentPeriodScores = 0;
        let previousPeriodScores = 0;
        let currentParticipants = new Set();
        let previousParticipants = new Set();

        for (let leaderboard of leaderboardConfigs) {
            try {
                // 当前周期的新分数
                const currentScores = await db.collection('leaderboard_score')
                    .where({
                        appId: leaderboard.appId,
                        leaderboardId: leaderboard.leaderboardId,
                        gmtCreate: {
                            $gte: currentStart,
                            $lte: currentEnd
                        }
                    })
                    .get();

                currentPeriodScores += currentScores.length;
                currentScores.forEach(score => {
                    currentParticipants.add(`${leaderboard.appId}:${score.openId}`);
                });

                // 上一周期的分数
                const previousScores = await db.collection('leaderboard_score')
                    .where({
                        appId: leaderboard.appId,
                        leaderboardId: leaderboard.leaderboardId,
                        gmtCreate: {
                            $gte: previousStart,
                            $lte: previousEnd
                        }
                    })
                    .get();

                previousPeriodScores += previousScores.length;
                previousScores.forEach(score => {
                    previousParticipants.add(`${leaderboard.appId}:${score.openId}`);
                });

            } catch (e) {
                console.log(`Error calculating trends for leaderboard ${leaderboard.leaderboardId}:`, e.message);
            }
        }

        trends.newScores = currentPeriodScores;

        // 计算增长率
        if (previousPeriodScores > 0) {
            trends.scoreGrowth = Math.round(((currentPeriodScores - previousPeriodScores) / previousPeriodScores) * 100 * 100) / 100;
        }

        if (previousParticipants.size > 0) {
            trends.participantGrowth = Math.round(((currentParticipants.size - previousParticipants.size) / previousParticipants.size) * 100 * 100) / 100;
        }

    } catch (e) {
        console.log('Trend calculation failed:', e.message);
    }

    return trends;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getLeaderboardStatsHandler, 'stats_view');
exports.main = mainFunc; 