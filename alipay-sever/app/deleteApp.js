const cloud = require("@alipay/faas-server-sdk");

// 请求参数
/**
 * 函数：deleteApp
 * 说明：删除应用（危险操作，会删除所有相关数据）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | force | boolean | 否 | 是否强制删除，默认false |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
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
    let force = event.force || false;

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
        ret.msg = "参数[appId]错误";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查应用是否存在
        const appList = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (appList.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        // 如果不是强制删除，检查是否有用户数据
        if (!force) {
            const userTableName = `user_${appId}`;
            try {
                const userCount = await db.collection(userTableName).count();
                if (userCount.total > 0) {
                    ret.code = 4003;
                    ret.msg = `应用下还有 ${userCount.total} 个用户，请设置 force=true 强制删除`;
                    return ret;
                }
            } catch (e) {
                // 用户表不存在，继续删除
            }
        }

        // 开始删除相关数据
        let deletedCount = {
            app: 0,
            users: 0,
            leaderboardConfigs: 0,
            leaderboardScores: 0
        };

        // 1. 删除应用配置
        await db.collection('app_config')
            .where({ appId: appId })
            .remove();
        deletedCount.app = 1;

        // 2. 删除用户数据表
        const userTableName = `user_${appId}`;
        try {
            const userCount = await db.collection(userTableName).count();
            deletedCount.users = userCount.total;
            
            // 批量删除用户数据
            await db.collection(userTableName).where({}).remove();
        } catch (e) {
            // 用户表不存在，忽略错误
        }

        // 3. 删除排行榜配置
        try {
            const leaderboardConfigCount = await db.collection('leaderboard_config')
                .where({ appId: appId })
                .count();
            deletedCount.leaderboardConfigs = leaderboardConfigCount.total;
            
            await db.collection('leaderboard_config')
                .where({ appId: appId })
                .remove();
        } catch (e) {
            // 排行榜配置表不存在，忽略错误
        }

        // 4. 删除排行榜分数数据
        try {
            const leaderboardScoreCount = await db.collection('leaderboard_score')
                .where({ appId: appId })
                .count();
            deletedCount.leaderboardScores = leaderboardScoreCount.total;
            
            await db.collection('leaderboard_score')
                .where({ appId: appId })
                .remove();
        } catch (e) {
            // 排行榜分数表不存在，忽略错误
        }

        ret.msg = "删除成功";
        ret.data = {
            deletedCount: deletedCount,
            message: `已删除应用及其所有相关数据：用户${deletedCount.users}个，排行榜配置${deletedCount.leaderboardConfigs}个，排行榜记录${deletedCount.leaderboardScores}个`
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 