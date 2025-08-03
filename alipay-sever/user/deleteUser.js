const cloud = require("@alipay/faas-server-sdk");
const common = require("./common");

// 请求参数
/**
 * 函数：deleteUser
 * 说明：删除用户（危险操作，会删除用户的所有数据）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | playerId | string | 是 | 玩家ID |
    | force | boolean | 否 | 是否强制删除，默认false |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "playerId": "player001",
        "force": true
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {}
    }
 */

exports.main = async (event, context) => {
    let appId = event.appId;
    let playerId = event.playerId;
    let force = event.force || false;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    var parmErr = common.hash.CheckParams(event);
    if(parmErr) {
        ret.code = 4001;
        ret.msg = "参数错误, error code:" + parmErr;
        return ret;
    }

    // 参数校验
    if (!appId || typeof appId !== "string") {
        ret.code = 4001;
        ret.msg = "参数[appId]错误";
        return ret;
    }

    if (!playerId || typeof playerId !== "string") {
        ret.code = 4001;
        ret.msg = "参数[playerId]错误";
        return ret;
    }

    const db = cloud.database();
    const userTableName = `user_${appId}`;

    try {
        let collection = db.collection(userTableName);

        // 查询用户是否存在
        let userList = await collection
            .where({ playerId: playerId })
            .get();

        if (userList.length === 0) {
            ret.code = 4004;
            ret.msg = "用户不存在";
            return ret;
        }

        let user = userList[0];

        // 如果不是强制删除，检查用户是否有排行榜记录
        if (!force) {
            try {
                const leaderboardCount = await db.collection('leaderboard_score')
                    .where({ appId: appId, playerId: playerId })
                    .count();
                
                if (leaderboardCount.total > 0) {
                    ret.code = 4003;
                    ret.msg = `用户在排行榜中有 ${leaderboardCount.total} 条记录，请设置 force=true 强制删除`;
                    return ret;
                }
            } catch (e) {
                // 排行榜表不存在，继续删除
            }
        }

        // 开始删除用户相关数据
        let deletedCount = {
            user: 0,
            leaderboardScores: 0
        };

        // 1. 删除用户的排行榜分数记录
        try {
            const leaderboardScores = await db.collection('leaderboard_score')
                .where({ appId: appId, playerId: playerId })
                .get();
            
            deletedCount.leaderboardScores = leaderboardScores.length;
            
            if (leaderboardScores.length > 0) {
                await db.collection('leaderboard_score')
                    .where({ appId: appId, playerId: playerId })
                    .remove();
            }
        } catch (e) {
            // 排行榜分数表不存在，忽略错误
        }

        // 2. 删除用户记录
        await collection.doc(user._id).remove();
        deletedCount.user = 1;

        ret.msg = "删除成功";
        ret.data = {
            playerId: playerId,
            deletedCount: deletedCount,
            message: `已删除用户及其所有相关数据：排行榜记录${deletedCount.leaderboardScores}个`
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 