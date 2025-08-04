const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getUserDetail
 * 说明：获取用户详细信息
 * 权限：需要 user_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | openId | string | 是 | 用户openId |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "openId": "user_openid_123456"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "userInfo": {
                "id": "user_id_123456",
                "openId": "user_openid_123456",
                "nickName": "测试用户",
                "avatarUrl": "https://example.com/avatar.jpg",
                "gender": 1,
                "province": "广东省",
                "city": "深圳市",
                "userData": "{\"level\": 5, \"coins\": 1000}",
                "banned": false,
                "gmtCreate": "2023-10-01 10:00:00",
                "gmtModify": "2023-10-02 15:30:00"
            },
            "gameStats": {
                "totalScores": 15,
                "bestScore": 9999,
                "avgScore": 750.5,
                "lastPlayTime": "2023-10-02 15:30:00",
                "playDays": 3
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 用户不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getUserDetailHandler(event, context) {
    let appId = event.appId;
    let openId = event.openId;

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

    if (!openId || typeof openId !== "string") {
        ret.code = 4001;
        ret.msg = "用户openId不能为空";
        return ret;
    }

    const db = cloud.database();

    try {
        // 验证应用是否存在
        const appList = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (appList.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        // 获取用户信息
        const userTableName = `user_${appId}`;
        const userList = await db.collection(userTableName)
            .where({ openId: openId })
            .get();

        if (userList.length === 0) {
            ret.code = 4004;
            ret.msg = "用户不存在";
            return ret;
        }

        const userInfo = userList[0];

        // 获取用户游戏统计
        let gameStats = {
            totalScores: 0,
            bestScore: 0,
            avgScore: 0,
            lastPlayTime: null,
            playDays: 0
        };

        try {
            // 查询用户的所有分数记录
            const scoreList = await db.collection('leaderboard_score')
                .where({ 
                    appId: appId,
                    openId: openId 
                })
                .orderBy('score', 'desc')
                .get();

            if (scoreList.length > 0) {
                gameStats.totalScores = scoreList.length;
                gameStats.bestScore = scoreList[0].score;
                
                // 计算平均分
                const totalScore = scoreList.reduce((sum, score) => sum + (score.score || 0), 0);
                gameStats.avgScore = Math.round((totalScore / scoreList.length) * 100) / 100;
                
                // 最后游戏时间
                const latestScore = scoreList.reduce((latest, current) => {
                    const currentTime = new Date(current.gmtCreate || current.createTime);
                    const latestTime = new Date(latest.gmtCreate || latest.createTime);
                    return currentTime > latestTime ? current : latest;
                }, scoreList[0]);
                
                gameStats.lastPlayTime = latestScore.gmtCreate || latestScore.createTime;
                
                // 计算游戏天数（不同日期的记录数）
                const playDates = new Set();
                scoreList.forEach(score => {
                    const date = (score.gmtCreate || score.createTime).split(' ')[0];
                    playDates.add(date);
                });
                gameStats.playDays = playDates.size;
            }
        } catch (e) {
            // 分数查询失败，使用默认值
            console.log('Score query failed:', e.message);
        }

        ret.data = {
            userInfo: {
                id: userInfo._id,
                openId: userInfo.openId,
                nickName: userInfo.nickName || '',
                avatarUrl: userInfo.avatarUrl || '',
                gender: userInfo.gender || 0,
                province: userInfo.province || '',
                city: userInfo.city || '',
                userData: userInfo.data || '{}',
                banned: userInfo.banned || false,
                gmtCreate: userInfo.gmtCreate,
                gmtModify: userInfo.gmtModify
            },
            gameStats: gameStats
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.1; // 10% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'USER_DETAIL', {
                appId: appId,
                openId: openId,
                userNickName: userInfo.nickName,
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
const mainFunc = requirePermission(getUserDetailHandler, 'user_manage');
exports.main = mainFunc; 