const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getLeaderboardData
 * 说明：获取排行榜数据
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | leaderboardType | string | 是 | 排行榜Key |
    | limit | number | 否 | 返回记录数量(默认100) |
    | offset | number | 否 | 分页偏移量(默认0) |
    | includeUserInfo | boolean | 否 | 是否包含用户信息(默认true) |
    | hasUserInfo | number | 否 | 用户信息过滤(0=无用户信息,1=有用户信息,null=不过滤) |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "leaderboardType": "weekly_score",
        "limit": 50,
        "offset": 0,
        "includeUserInfo": true,
        "hasUserInfo": 1
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "leaderboardInfo": {
                "id": "leaderboard_id_123456",
                "appId": "test_game_001",
                "leaderboardType": "weekly_score",
                "name": "周榜",
                "description": "每周重置的积分排行榜",
                "scoreType": "higher_better",
                "maxRank": 100,
                "enabled": true,
                "scoreCount": 1500,
                "participantCount": 300
            },
            "scores": [
                {
                    "rank": 1,
                    "openId": "user_openid_123",
                    "score": 9999,
                    "gmtCreate": "2023-10-02 15:30:00",
                    "userInfo": {
                        "nickName": "玩家A",
                        "avatarUrl": "https://example.com/avatar1.jpg"
                    }
                }
            ],
            "pagination": {
                "total": 1500,
                "limit": 50,
                "offset": 0,
                "hasMore": true
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 排行榜不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getLeaderboardDataHandler(event, context) {
    let appId = event.appId;
    let leaderboardType = event.leaderboardType;
    let limit = event.limit || 100;
    let offset = event.offset || 0;
    let includeUserInfo = event.includeUserInfo !== false; // 默认包含用户信息
    let hasUserInfo = event.hasUserInfo; // 用户信息过滤参数

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

    if (!leaderboardType || typeof leaderboardType !== "string") {
        ret.code = 4001;
        ret.msg = "排行榜ID不能为空";
        return ret;
    }

    // 限制查询数量
    if (limit > 1000) {
        limit = 1000;
    }
    if (limit < 1) {
        limit = 100;
    }

    const db = cloud.database();

    try {
        // 获取排行榜配置信息
        const leaderboardList = await db.collection('leaderboard_config')
            .where({ 
                appId: appId,
                leaderboardType: leaderboardType 
            })
            .get();

        if (leaderboardList.length === 0) {
            ret.code = 4004;
            ret.msg = "排行榜不存在";
            return ret;
        }

        const leaderboardInfo = leaderboardList[0];

        // 构建查询条件
        let queryCondition = { 
            appId: appId,
            leaderboardType: leaderboardType 
        };
        
        // 添加用户信息过滤条件
        if (hasUserInfo !== null && hasUserInfo !== undefined) {
            queryCondition.hasUserInfo = hasUserInfo;
        }

        // 获取总记录数
        const totalCount = await db.collection('leaderboard_score')
            .where(queryCondition)
            .count();

        // 根据分数类型决定排序方式
        const orderBy = leaderboardInfo.scoreType === 'higher_better' ? 'desc' : 'asc';

        // 获取分数记录
        let scoreQuery = db.collection('leaderboard_score')
            .where(queryCondition)
            .orderBy('score', orderBy)
            .orderBy('gmtCreate', 'asc') // 同分数时按时间排序
            .skip(offset)
            .limit(limit);

        const scoreList = await scoreQuery.get();

        // 如果需要包含用户信息，查询用户数据
        let userInfoMap = {};
        if (includeUserInfo && scoreList.length > 0) {
            const playerIds = [...new Set(scoreList.map(score => score.playerId))];
            const userTableName = `user_${appId}`;
            
            try {
                // 分批查询用户信息（避免查询过多）
                const userList = await db.collection(userTableName)
                    .where({
                        playerId: db.command.in(playerIds)
                    })
                    .get();

                userList.forEach(user => {
                    let userInfo = user.userInfo || {};
                    userInfoMap[user.playerId] = {
                        nickName: userInfo.nickName || '',
                        avatarUrl: userInfo.avatarUrl || ''
                    };
                });
            } catch (e) {
                // 用户表查询失败，忽略用户信息
                console.log('User info query failed:', e.message);
            }
        }

        // 组装返回数据
        const scores = scoreList.map((score, index) => {
            const item = {
                rank: offset + index + 1,
                playerId: score.playerId,
                openId: score.openId,
                score: score.score,
                gmtCreate: score.gmtCreate,
                gmtModify: score.gmtModify
            };

            if (userInfoMap[score.playerId]) {
                item.userInfo = userInfoMap[score.playerId];
            }

            return item;
        });

        ret.data = {
            leaderboardInfo: {
                id: leaderboardInfo._id,
                appId: leaderboardInfo.appId,
                leaderboardType: leaderboardInfo.leaderboardType,
                name: leaderboardInfo.name,
                description: leaderboardInfo.description,
                scoreType: leaderboardInfo.scoreType,
                maxRank: leaderboardInfo.maxRank,
                enabled: leaderboardInfo.enabled,
                scoreCount: totalCount.total,
                participantCount: leaderboardInfo.participantCount || 0
            },
            scores: scores,
            pagination: {
                total: totalCount.total,
                limit: limit,
                offset: offset,
                hasMore: offset + limit < totalCount.total
            }
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.05; // 5% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'LEADERBOARD_DATA', {
                appId: appId,
                leaderboardType: leaderboardType,
                limit: limit,
                offset: offset,
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
const mainFunc = requirePermission(getLeaderboardDataHandler, 'leaderboard_manage');
exports.main = mainFunc; 