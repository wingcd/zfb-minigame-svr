const cloud = require("@alipay/faas-server-sdk");
const common = require("./common");

// 请求参数
/**
 * 函数：deleteLeaderboard
 * 说明：删除排行榜配置及相关数据
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | leaderboardType | string | 是 | 排行榜类型 |
    | force | boolean | 否 | 是否强制删除，默认false |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "leaderboardType": "easy",
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
    let leaderboardType = event.leaderboardType;
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

    if (!leaderboardType || typeof leaderboardType !== "string") {
        ret.code = 4001;
        ret.msg = "参数[leaderboardType]错误";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查排行榜配置是否存在
        const leaderboardList = await db.collection('leaderboard_config')
            .where({ 
                appId: appId,
                leaderboardType: leaderboardType 
            })
            .get();

        if (leaderboardList.length === 0) {
            ret.code = 4004;
            ret.msg = "排行榜配置不存在";
            return ret;
        }

        // 如果不是强制删除，检查是否有分数记录
        if (!force) {
            try {
                const scoreCount = await db.collection('leaderboard_score')
                    .where({ 
                        appId: appId,
                        leaderboardType: leaderboardType 
                    })
                    .count();
                
                if (scoreCount.total > 0) {
                    ret.code = 4003;
                    ret.msg = `排行榜中有 ${scoreCount.total} 条分数记录，请设置 force=true 强制删除`;
                    return ret;
                }
            } catch (e) {
                // 分数表不存在，继续删除
            }
        }

        // 开始删除排行榜相关数据
        let deletedCount = {
            config: 0,
            scores: 0
        };

        // 1. 删除排行榜分数记录
        try {
            const scores = await db.collection('leaderboard_score')
                .where({ 
                    appId: appId,
                    leaderboardType: leaderboardType 
                })
                .get();
            
            deletedCount.scores = scores.length;
            
            if (scores.length > 0) {
                await db.collection('leaderboard_score')
                    .where({ 
                        appId: appId,
                        leaderboardType: leaderboardType 
                    })
                    .remove();
            }
        } catch (e) {
            // 分数表不存在，忽略错误
        }

        // 2. 删除排行榜配置
        await db.collection('leaderboard_config')
            .where({ 
                appId: appId,
                leaderboardType: leaderboardType 
            })
            .remove();
        deletedCount.config = 1;

        ret.msg = "删除成功";
        ret.data = {
            appId: appId,
            leaderboardType: leaderboardType,
            deletedCount: deletedCount,
            message: `已删除排行榜配置及其所有分数记录：${deletedCount.scores}条`
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 