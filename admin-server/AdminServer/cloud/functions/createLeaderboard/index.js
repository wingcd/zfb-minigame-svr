const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：createLeaderboard
 * 说明：创建新排行榜
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | leaderboardType | string | 是 | 排行榜ID（唯一标识） |
    | name | string | 是 | 排行榜名称 |
    | description | string | 否 | 排行榜描述 |
    | scoreType | string | 否 | 分数类型 (higher_better/lower_better) |
    | maxRank | number | 否 | 最大排名数量 |
    | enabled | boolean | 否 | 是否启用 |
    | category | string | 否 | 排行榜分类 |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "leaderboardType": "weekly_score",
        "name": "周榜",
        "description": "每周重置的积分排行榜",
        "scoreType": "higher_better",
        "maxRank": 100,
        "enabled": true,
        "category": "weekly"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "id": "leaderboard_id_123456",
            "appId": "test_game_001",
            "leaderboardType": "weekly_score",
            "name": "周榜",
            "description": "每周重置的积分排行榜",
            "scoreType": "higher_better",
            "maxRank": 100,
            "enabled": true,
            "createTime": "2023-10-01 10:00:00"
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4002: 排行榜ID已存在
 * - 4003: 权限不足
 * - 4004: 应用不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function createLeaderboardHandler(event, context) {
    let appId = event.appId;
    let leaderboardType = event.leaderboardType;
    let name = event.name;
    let description = event.description || '';
    let scoreType = event.scoreType || 'higher_better';
    let maxRank = event.maxRank || 100;
    let enabled = event.enabled !== false; // 默认启用
    let category = event.category || 'default';

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

    if (!leaderboardType || typeof leaderboardType !== "string" || leaderboardType.length < 2) {
        ret.code = 4001;
        ret.msg = "排行榜类型必须至少2个字符";
        return ret;
    }

    if (!name || typeof name !== "string" || name.length < 1) {
        ret.code = 4001;
        ret.msg = "排行榜名称不能为空";
        return ret;
    }

    // 验证分数类型
    const validScoreTypes = ['higher_better', 'lower_better'];
    if (!validScoreTypes.includes(scoreType)) {
        ret.code = 4001;
        ret.msg = "无效的分数类型";
        return ret;
    }

    // 验证最大排名数量
    if (typeof maxRank !== 'number' || maxRank < 10 || maxRank > 10000) {
        ret.code = 4001;
        ret.msg = "最大排名数量必须在10-10000之间";
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

        // 检查排行榜ID是否已存在（同一应用下）
        const existingLeaderboards = await db.collection('leaderboard_config')
            .where({ 
                appId: appId,
                leaderboardType: leaderboardType 
            })
            .get();

        if (existingLeaderboards.length > 0) {
            ret.code = 4002;
            ret.msg = "排行榜ID已存在";
            return ret;
        }

        // 创建排行榜
        const newLeaderboard = {
            appId: appId,
            leaderboardType: leaderboardType,
            name: name,
            description: description,
            scoreType: scoreType,
            maxRank: maxRank,
            enabled: enabled,
            category: category,
            createTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            createdBy: event.adminInfo.username,
            scoreCount: 0,
            participantCount: 0,
            lastResetTime: null
        };

        const addResult = await db.collection('leaderboard_config').add({
            data: newLeaderboard
        });

        // 记录操作日志
        await logOperation(event.adminInfo, 'CREATE', 'LEADERBOARD', {
            appId: appId,
            leaderboardType: leaderboardType,
            name: name,
            scoreType: scoreType,
            maxRank: maxRank,
            createdBy: event.adminInfo.username,
            severity: 'MEDIUM'
        });

        ret.msg = "创建成功";
        ret.data = {
            id: addResult._id,
            appId: appId,
            leaderboardType: leaderboardType,
            name: name,
            description: description,
            scoreType: scoreType,
            maxRank: maxRank,
            enabled: enabled,
            createTime: newLeaderboard.createTime
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(createLeaderboardHandler, 'leaderboard_manage');
exports.main = mainFunc; 