const cloud = require("@alipay/faas-server-sdk");
const common = require("./common");
const { requirePermission, logOperation } = require("./common/auth");

// 原始处理函数
async function getAllLeaderboardsHandler(event, context) {
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

        // 记录操作日志
        await logOperation(event.adminInfo, 'VIEW', 'LEADERBOARDS', {
            appId: appId,
            searchCondition: whereCondition,
            resultCount: total
        });

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getAllLeaderboardsHandler, 'leaderboard_manage');
exports.main = mainFunc;