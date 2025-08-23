const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：fixLeaderboardUserInfo
 * 说明：修复排行榜中的hasUserInfo字段
 * 权限：需要 leaderboard_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | leaderboardType | string | 是 | 排行榜类型 |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "leaderboardType": "easy"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "totalProcessed": 150,
            "updatedCount": 45,
            "errorCount": 0
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

async function fixLeaderboardUserInfoHandler(event, context) {
    let appId = event.appId;
    let leaderboardType = event.leaderboardType;

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
        ret.msg = "排行榜类型不能为空";
        return ret;
    }

    const db = cloud.database();

    try {
        console.log(`开始修复排行榜用户信息标记 - appId: ${appId}, leaderboardType: ${leaderboardType}`);
        
        // 分页获取所有排行榜记录
        let scoreList = [];
        let skip = 0;
        const limit = 100; // 每次查询100条
        let hasMore = true;
        
        while (hasMore) {
            const batch = await db.collection('leaderboard_score')
                .where({
                    appId: appId,
                    leaderboardType: leaderboardType
                })
                .skip(skip)
                .limit(limit)
                .get();
                
            if (batch.length > 0) {
                scoreList = scoreList.concat(batch);
                skip += limit;
                console.log(`已获取 ${scoreList.length} 条记录...`);
                
                // 如果返回的记录数少于limit，说明已经是最后一批
                if (batch.length < limit) {
                    hasMore = false;
                }
            } else {
                hasMore = false;
            }
        }
            
        console.log(`总共找到 ${scoreList.length} 条排行榜记录`);

        if (scoreList.length === 0) {
            ret.data = {
                totalProcessed: 0,
                updatedCount: 0,
                errorCount: 0
            };
            return ret;
        }

        // 获取所有唯一的玩家ID
        const playerIds = [...new Set(scoreList.map(score => score.playerId))];
        const userTableName = `user_${appId}`;
        
                // 分批查询用户信息（因为in查询有数量限制，通常是20个）
        let userInfoMap = {};
        if (playerIds.length > 0) {
            try {
                const batchSize = 20; // in查询的批次大小
                let allUsers = [];
                
                for (let i = 0; i < playerIds.length; i += batchSize) {
                    const batchPlayerIds = playerIds.slice(i, i + batchSize);
                    console.log(`查询用户信息批次 ${Math.floor(i/batchSize) + 1}/${Math.ceil(playerIds.length/batchSize)}: ${batchPlayerIds.length} 个玩家`);
                    
                    const userBatch = await db.collection(userTableName)
                        .where({
                            playerId: db.command.in(batchPlayerIds)
                        })
                        .get();
                    
                    allUsers = allUsers.concat(userBatch);
                }

                allUsers.forEach(user => {
                     const hasValidUserInfo = user.userInfo && user.userInfo.nickName;
                     userInfoMap[user.playerId] = hasValidUserInfo ? 1 : 0;
                 });
                 
                 // 对于没有在用户表中找到的玩家，设置hasUserInfo为0
                 playerIds.forEach(playerId => {
                     if (!(playerId in userInfoMap)) {
                         userInfoMap[playerId] = 0;
                     }
                 });
                 
                 console.log(`用户信息查询完成，找到 ${allUsers.length} 个用户，总共需要处理 ${playerIds.length} 个玩家`);
            } catch (e) {
                console.log('User info query failed:', e.message);
                // 如果查询用户表失败，将所有记录的hasUserInfo设为0
                playerIds.forEach(playerId => {
                   userInfoMap[playerId] = 0;
                });
            }
        }

        // 批量更新hasUserInfo字段
        let totalProcessed = 0;
        let updatedCount = 0;
        let errorCount = 0;

        for (const scoreRecord of scoreList) {
            totalProcessed++;
            
            const newHasUserInfo = userInfoMap[scoreRecord.playerId] || 0;
            
            // 只有当hasUserInfo字段需要更新时才执行更新
            if (scoreRecord.hasUserInfo !== newHasUserInfo) {
                try {
                    await db.collection('leaderboard_score')
                        .doc(scoreRecord._id)
                        .update({
                            data: {
                                hasUserInfo: newHasUserInfo
                            }
                        });
                    updatedCount++;
                } catch (e) {
                    console.error(`Failed to update record ${scoreRecord._id}:`, e.message);
                    errorCount++;
                }
            }
        }

                 ret.data = {
             totalProcessed: totalProcessed,
             updatedCount: updatedCount,
             errorCount: errorCount
         };

         console.log(`修复完成 - 处理: ${totalProcessed}, 更新: ${updatedCount}, 错误: ${errorCount}`);

         // 记录操作日志
         await logOperation(event.adminInfo, 'UPDATE', 'LEADERBOARD_FIX', {
             appId: appId,
             leaderboardType: leaderboardType,
             totalProcessed: totalProcessed,
             updatedCount: updatedCount,
             errorCount: errorCount,
             currentAdmin: event.adminInfo.username
         });

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(fixLeaderboardUserInfoHandler, 'leaderboard_manage');
exports.main = mainFunc; 