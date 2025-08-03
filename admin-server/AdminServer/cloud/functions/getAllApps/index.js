const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

// 原始处理函数
async function queryAppHandler(event, context) {
    //请求参数
    //app 名字
    let appName;
    let appId;
    let channelAppId;

    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    //参数校验 字段存在  为空   类型
    appName = event.appName === undefined ? null : event.appName;
    appId = event.appId === undefined ? null : event.appId;
    channelAppId = event.channelAppId === undefined ? null : event.channelAppId;
    
    if (!appName && !appId && !channelAppId) {
        ret.code = 4001;
        ret.msg = "参数[appName]和[appId]和[channelAppId]不能同时为空";
        return ret;
    }

    if (appName && typeof appName != "string") {
        ret.code = 4001;
        ret.msg = "参数[appName]错误";
        return ret;
    }

    if (appId && typeof appId != "string") {
        ret.code = 4001;
        ret.msg = "参数[appId]错误";
        return ret;
    }

    if (channelAppId && typeof channelAppId != "string") {
        ret.code = 4001;
        ret.msg = "参数[channelAppId]错误";
        return ret;
    }

    //数据库实例
    const db = cloud.database();

    try {
        let appList = null;
        let queryCondition = {};

        if (appId) {
            queryCondition = { appId: appId };
            appList = await db.collection(`app_config`)
                .where(queryCondition)
                .get();
        } else if (channelAppId) {
            queryCondition = { channelAppId: channelAppId };
            appList = await db.collection(`app_config`)
                .where(queryCondition)
                .get();
        } else {
            queryCondition = { appName: appName };
            appList = await db.collection(`app_config`)
                .where(queryCondition)
                .get();
        }

        if (appList.length === 0) {
            ret.msg = "未查询到您的数据";
            return ret;
        }

        const appData = appList[0];
        ret.data = appData;

        // 查询排行榜数据
        let leaderBoardList = await db.collection(`leaderboard_config`)
            .where({
                appId: appData.appId
            })
            .get();
        ret.data.leaderBoardList = leaderBoardList;

        // 统计用户数量
        try {
            const userTableName = `user_${appData.appId}`;
            const userCount = await db.collection(userTableName).count();
            ret.data.userCount = userCount.total;
        } catch (e) {
            ret.data.userCount = 0;
        }

        // 记录查询操作日志
        await logOperation(event.adminInfo, 'VIEW', 'APP_QUERY', {
            queryCondition: queryCondition,
            foundApp: {
                appId: appData.appId,
                appName: appData.appName,
                platform: appData.platform
            },
            leaderboardCount: leaderBoardList.length,
            userCount: ret.data.userCount
        });

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(queryAppHandler, 'app_manage');
exports.main = mainFunc;