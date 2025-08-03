const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：updateLeaderboard
 * 说明：更新排行榜配置
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | leaderboardType | string | 是 | 排行榜类型 |
    | name | string | 否 | 排行榜名称 |
    | updateStrategy | number | 否 | 更新策略 |
    | sort | number | 否 | 排序方式 |
    | enabled | boolean | 否 | 是否启用 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "leaderboardType": "easy",
        "name": "简单模式排行榜",
        "updateStrategy": 0,
        "sort": 1,
        "enabled": true
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {}
    }
 */

// 原始处理函数
async function updateLeaderboardHandler(event, context) {
    let appId = event.appId;
    let leaderboardType = event.leaderboardType;
    let name = event.name;
    let updateStrategy = event.updateStrategy;
    let sort = event.sort;
    let enabled = event.enabled;

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

        const oldConfig = leaderboardList[0];

        // 构建更新数据
        let updateData = {
            updateTime: moment().format("YYYY-MM-DD HH:mm:ss")
        };

        if (name !== undefined) {
            updateData.name = name;
        }

        if (updateStrategy !== undefined) {
            if (![0, 1, 2].includes(updateStrategy)) {
                ret.code = 4001;
                ret.msg = "参数[updateStrategy]错误，必须为0、1或2";
                return ret;
            }
            updateData.updateStrategy = updateStrategy;
        }

        if (sort !== undefined) {
            if (![0, 1].includes(sort)) {
                ret.code = 4001;
                ret.msg = "参数[sort]错误，必须为0或1";
                return ret;
            }
            updateData.sort = sort;
        }

        if (enabled !== undefined) {
            updateData.enabled = enabled;
        }

        // 更新排行榜配置
        await db.collection('leaderboard_config')
            .where({ 
                appId: appId,
                leaderboardType: leaderboardType 
            })
            .update({
                data: updateData
            });

        // 记录操作日志
        await logOperation(event.adminInfo, 'UPDATE', 'LEADERBOARD', {
            appId: appId,
            leaderboardType: leaderboardType,
            oldConfig: {
                name: oldConfig.name,
                updateStrategy: oldConfig.updateStrategy,
                sort: oldConfig.sort,
                enabled: oldConfig.enabled
            },
            changes: updateData
        });

        ret.msg = "更新成功";
        ret.data = {
            appId: appId,
            leaderboardType: leaderboardType,
            updatedFields: Object.keys(updateData)
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(updateLeaderboardHandler, 'leaderboard_manage');
exports.main = mainFunc;

// 自动注册API
const { autoRegister } = require('../api-factory');
autoRegister('leaderboard.update')(mainFunc);