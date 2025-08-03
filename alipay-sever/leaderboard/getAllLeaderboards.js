const cloud = require("@alipay/faas-server-sdk");
const common = require("./common");

// 请求参数
/**
 * 函数：getAllLeaderboards
 * 说明：获取排行榜配置列表
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |
    | leaderboardType | string | 否 | 排行榜类型搜索 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "page": 1,
        "pageSize": 20
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "list": [...],
            "total": 10,
            "page": 1,
            "pageSize": 20
        }
    }
 */

exports.main = async (event, context) => {
    let appId = event.appId;
    let page = event.page || 1;
    let pageSize = event.pageSize || 20;
    let leaderboardType = event.leaderboardType;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {
            list: [],
            total: 0,
            page: page,
            pageSize: pageSize
        }
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

    // 限制最大每页数量
    if (pageSize > 100) {
        pageSize = 100;
    }

    const db = cloud.database();

    try {
        // 构建查询条件
        let whereCondition = { appId: appId };
        
        if (leaderboardType) {
            whereCondition.leaderboardType = new RegExp(leaderboardType, 'i'); // 模糊搜索
        }

        // 查询总数
        const countResult = await db.collection('leaderboard_config').where(whereCondition).count();
        const total = countResult.total;

        // 分页查询
        const skip = (page - 1) * pageSize;
        let leaderboardList = await db.collection('leaderboard_config')
            .where(whereCondition)
            .orderBy('createTime', 'desc')
            .skip(skip)
            .limit(pageSize)
            .get();

        // 为每个排行榜添加统计信息
        for (let leaderboard of leaderboardList) {
            try {
                // 统计参与人数
                const participantCount = await db.collection('leaderboard_score')
                    .where({ 
                        appId: appId,
                        leaderboardType: leaderboard.leaderboardType 
                    })
                    .count();
                leaderboard.participantCount = participantCount.total;

                // 获取最高分数
                const topScore = await db.collection('leaderboard_score')
                    .where({ 
                        appId: appId,
                        leaderboardType: leaderboard.leaderboardType 
                    })
                    .orderBy('score', leaderboard.sort === 1 ? 'desc' : 'asc')
                    .limit(1)
                    .get();
                
                leaderboard.topScore = topScore.length > 0 ? topScore[0].score : 0;
                
                // 设置默认状态
                leaderboard.enabled = leaderboard.enabled !== false; // 默认启用
            } catch (e) {
                // 如果统计出错，设置默认值
                leaderboard.participantCount = 0;
                leaderboard.topScore = 0;
                leaderboard.enabled = true;
            }
        }

        ret.data.list = leaderboardList;
        ret.data.total = total;

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 