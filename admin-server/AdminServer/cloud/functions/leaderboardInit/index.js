const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 | 参数名 | 类型 | 必选 | 说明 |
 | --- | --- | --- | --- |
 | leaderboardName | string | 是 | 排行榜名字 |
 | appId | string | 是 | 小程序id |
 | leaderboardTypeList | array | 是 | 排行榜类型数组 |
 | updateStrategy | number | 否 | 更新策略 |
 | sort | number | 否 | 排序方式 |
 | resetType | string | 否 | 重置类型：daily(每日)、weekly(每周)、monthly(每月)、custom(自定义)、permanent(永久) |
 | resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |
 */

// 测试数据
/**
{
     "leaderboardName": "难度排行榜",
    "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
    "leaderboardType": "easy",
    "updateStrategy": 0,
    "sort": 1
 }
*/

// 原始处理函数
async function leaderboardInitHandler(event, context) {
    //排行榜名字
    let leaderboardName;
    //排行榜类型
    let leaderboardType;
    //更新策略
    let updateStrategy;
    //排序方式
    let sort;
    //appId   
    let appId;
    //重置类型
    let resetType;
    //重置间隔值
    let resetValue;

    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    }

    //参数校验 字段存在  为空   类型
    if (event.hasOwnProperty("leaderboardName") && (!event.leaderboardName || typeof event.leaderboardName != "string")) {
        ret.code = 4001;
        ret.msg = "参数[leaderboardName]错误"
        return ret;
    }

    if (event.hasOwnProperty("appId") && (!event.appId || typeof event.appId != "string")) {
        ret.code = 4001;
        ret.msg = "参数[appId]错误"
        return ret;
    }

    if (event.hasOwnProperty("leaderboardType") && (!event.leaderboardType || typeof event.leaderboardName != "string")) {
        ret.code = 4001;
        ret.msg = "参数[leaderboardType]错误"
        return ret;
    }


    leaderboardName = event.hasOwnProperty("leaderboardName") ? event.leaderboardName.trim() : "排行榜";

    appId = event.hasOwnProperty("appId") ? event.appId.trim() : "myGame";

    leaderboardType = event.hasOwnProperty("leaderboardType") ? event.leaderboardType : "default";

    updateStrategy = event.hasOwnProperty("updateStrategy") ? event.updateStrategy : 0;

    sort = event.hasOwnProperty("sort") ? event.sort : 1;

    resetType = event.hasOwnProperty("resetType") ? event.resetType : "permanent";
    
    resetValue = event.hasOwnProperty("resetValue") ? event.resetValue : 24;

    const db = cloud.database();
    //创建集合
    try {
        await db.getCollection("leaderboard_config")
    } catch (e) {
        if (e.message == "not found collection") {
            await db.createCollection("leaderboard_config");
        } else {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }
    try {
        await db.getCollection("leaderboard_score")
    } catch (e) {
        if (e.message == "not found collection") {
            await db.createCollection("leaderboard_score");
        } else {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }
    try {
        await db.getCollection("leaderboard_segment")
    } catch (e) {
        if (e.message == "not found collection") {
            await db.createCollection("leaderboard_segment");
        } else {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }

    try {
        let now = moment().format("YYYY-MM-DD HH:mm:ss");
        
        // 计算重置时间
        let resetTime = null;
        if (resetType !== 'permanent') {
            switch (resetType) {
                case "daily":
                    resetTime = moment().startOf('day').add(1, 'day').format("YYYY-MM-DD HH:mm:ss");
                    break;
                case "weekly":
                    resetTime = moment().startOf('week').add(1, 'week').format("YYYY-MM-DD HH:mm:ss");
                    break;
                case "monthly":
                    resetTime = moment().startOf('month').add(1, 'month').format("YYYY-MM-DD HH:mm:ss");
                    break;
                case "custom":
                    resetTime = moment().add(resetValue, 'hours').format("YYYY-MM-DD HH:mm:ss");
                    break;
                default:
                    resetTime = null;
                    break;
            }
        }
        
        await db.collection("leaderboard_config")
            .add({
                data: {
                    "appId": appId,
                    "leaderboardName": leaderboardName,
                    "leaderboardType": leaderboardType,
                    "sort": sort,
                    "updateStrategy": updateStrategy,
                    "resetType": resetType,
                    "resetValue": resetValue,
                    "resetTime": resetTime,
                    "gmtCreate": now,
                    "gmtModify": now,
                }
            });
    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    // 记录操作日志
    await logOperation(event.adminInfo, 'CREATE', 'LEADERBOARD', {
        appId: appId,
        leaderboardName: leaderboardName,
        leaderboardType: leaderboardType,
        updateStrategy: updateStrategy,
        sort: sort
    });

    ret.data = {
        "appId": appId,
        "leaderboardName": leaderboardName,
        "leaderboardType": leaderboardType,
    }
    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(leaderboardInitHandler, 'leaderboard_manage');
exports.main = mainFunc;