const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：updateScore
 * 说明：更新用户分数记录
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | leaderboardId | string | 是 | 排行榜ID |
    | openId | string | 是 | 用户openId |
    | score | number | 是 | 新分数 |
    | operation | string | 否 | 操作类型 (set/add/subtract) |
    | reason | string | 否 | 操作原因 |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "leaderboardId": "weekly_score",
        "openId": "user_openid_123456",
        "score": 1000,
        "operation": "add",
        "reason": "管理员补发积分"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "oldScore": 5000,
            "newScore": 6000,
            "scoreChange": 1000,
            "operation": "add"
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 排行榜或用户不存在
 * - 4005: 分数操作无效
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function updateScoreHandler(event, context) {
    let appId = event.appId;
    let leaderboardId = event.leaderboardId;
    let openId = event.openId;
    let score = event.score;
    let operation = event.operation || 'set'; // 默认直接设置分数
    let reason = event.reason || '管理员操作';

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

    if (!leaderboardId || typeof leaderboardId !== "string") {
        ret.code = 4001;
        ret.msg = "排行榜ID不能为空";
        return ret;
    }

    if (!openId || typeof openId !== "string") {
        ret.code = 4001;
        ret.msg = "用户openId不能为空";
        return ret;
    }

    if (typeof score !== 'number' || isNaN(score)) {
        ret.code = 4001;
        ret.msg = "分数必须是有效数字";
        return ret;
    }

    // 验证操作类型
    const validOperations = ['set', 'add', 'subtract'];
    if (!validOperations.includes(operation)) {
        ret.code = 4001;
        ret.msg = "无效的操作类型";
        return ret;
    }

    const db = cloud.database();

    try {
        // 验证排行榜是否存在
        const leaderboardList = await db.collection('leaderboard_config')
            .where({ 
                appId: appId,
                leaderboardId: leaderboardId 
            })
            .get();

        if (leaderboardList.length === 0) {
            ret.code = 4004;
            ret.msg = "排行榜不存在";
            return ret;
        }

        const leaderboardConfig = leaderboardList[0];

        // 验证用户是否存在
        const userTableName = `user_${appId}`;
        const userList = await db.collection(userTableName)
            .where({ openId: openId })
            .get();

        if (userList.length === 0) {
            ret.code = 4004;
            ret.msg = "用户不存在";
            return ret;
        }

        // 查找现有分数记录
        const existingScores = await db.collection('leaderboard_score')
            .where({ 
                appId: appId,
                leaderboardId: leaderboardId,
                openId: openId 
            })
            .orderBy('gmtCreate', 'desc')
            .limit(1)
            .get();

        let oldScore = 0;
        let newScore = score;
        let scoreRecord = null;

        if (existingScores.length > 0) {
            scoreRecord = existingScores[0];
            oldScore = scoreRecord.score || 0;

            // 根据操作类型计算新分数
            switch (operation) {
                case 'set':
                    newScore = score;
                    break;
                case 'add':
                    newScore = oldScore + score;
                    break;
                case 'subtract':
                    newScore = oldScore - score;
                    break;
            }
        } else {
            // 没有现有记录，只有 set 和 add 操作有效
            if (operation === 'subtract') {
                ret.code = 4005;
                ret.msg = "用户没有分数记录，不能执行减分操作";
                return ret;
            }
            if (operation === 'add') {
                newScore = score; // 没有基础分数时，add 等于 set
            }
        }

        // 验证新分数的合理性
        if (newScore < 0) {
            ret.code = 4005;
            ret.msg = "分数不能为负数";
            return ret;
        }

        const scoreChange = newScore - oldScore;

        if (scoreRecord) {
            // 更新现有记录
            await db.collection('leaderboard_score')
                .where({ 
                    _id: scoreRecord._id
                })
                .update({
                    data: {
                        score: newScore,
                        gmtModify: moment().format("YYYY-MM-DD HH:mm:ss"),
                        modifiedBy: event.adminInfo.username,
                        modifyReason: reason,
                        operation: operation,
                        scoreChange: scoreChange
                    }
                });
        } else {
            // 创建新记录
            await db.collection('leaderboard_score').add({
                data: {
                    appId: appId,
                    leaderboardId: leaderboardId,
                    openId: openId,
                    score: newScore,
                    gmtCreate: moment().format("YYYY-MM-DD HH:mm:ss"),
                    gmtModify: moment().format("YYYY-MM-DD HH:mm:ss"),
                    createdBy: event.adminInfo.username,
                    createReason: reason,
                    operation: operation,
                    scoreChange: scoreChange
                }
            });
        }

        // 记录操作日志
        await logOperation(event.adminInfo, 'UPDATE', 'SCORE', {
            appId: appId,
            leaderboardId: leaderboardId,
            openId: openId,
            oldScore: oldScore,
            newScore: newScore,
            scoreChange: scoreChange,
            operation: operation,
            reason: reason,
            updatedBy: event.adminInfo.username,
            severity: 'MEDIUM'
        });

        ret.data = {
            oldScore: oldScore,
            newScore: newScore,
            scoreChange: scoreChange,
            operation: operation
        };

        ret.msg = "分数更新成功";

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(updateScoreHandler, 'leaderboard_manage');
exports.main = mainFunc; 