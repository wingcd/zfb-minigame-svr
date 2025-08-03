const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：getAllApps
 * 说明：获取所有应用列表
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |
    | appName | string | 否 | 应用名称搜索 |
    | appId | string | 否 | 应用ID搜索 |
    | platform | string | 否 | 平台筛选 |
  * 测试数据
    {
        "page": 1,
        "pageSize": 20,
        "appName": "小游戏"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "list": [...],
            "total": 100,
            "page": 1,
            "pageSize": 20
        }
    }
 */

// 原始处理函数
async function getAllAppsHandler(event, context) {
    // 请求参数
    let page = event.page || 1;
    let pageSize = event.pageSize || 20;
    let appName = event.appName;
    let appId = event.appId;
    let platform = event.platform;

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
    if (pageSize > 100) {
        pageSize = 100; // 限制最大每页数量
    }

    const db = cloud.database();

    try {
        // 构建查询条件
        let whereCondition = {};
        
        if (appName) {
            whereCondition.appName = new RegExp(appName, 'i'); // 模糊搜索，忽略大小写
        }
        
        if (appId) {
            whereCondition.appId = appId;
        }
        
        if (platform) {
            whereCondition.platform = platform;
        }

        // 查询总数
        const countResult = await db.collection('app_config').where(whereCondition).count();
        const total = countResult.total;

        // 分页查询
        const skip = (page - 1) * pageSize;
        let appList = await db.collection('app_config')
            .where(whereCondition)
            .orderBy('createTime', 'desc')
            .skip(skip)
            .limit(pageSize)
            .get();

        // 为每个应用添加统计信息
        for (let app of appList) {
            try {
                // 统计用户数量
                const userTableName = `user_${app.appId}`;
                const userCount = await db.collection(userTableName).count();
                app.userCount = userCount.total;

                // 统计今日活跃用户（简化实现，可根据实际需求优化）
                const today = new Date().toISOString().split('T')[0];
                const dailyActiveCount = await db.collection(userTableName)
                    .where({
                        gmtModify: {
                            $gte: today + ' 00:00:00',
                            $lte: today + ' 23:59:59'
                        }
                    })
                    .count();
                app.dailyActive = dailyActiveCount.total;

                // 统计排行榜数量
                const leaderboardCount = await db.collection('leaderboard_config')
                    .where({ appId: app.appId })
                    .count();
                app.leaderboardCount = leaderboardCount.total;

                // 设置状态
                app.status = app.status || 'active';
            } catch (e) {
                // 如果统计出错，设置默认值
                app.userCount = 0;
                app.dailyActive = 0;
                app.leaderboardCount = 0;
                app.status = 'active';
            }
        }

        ret.data.list = appList;
        ret.data.total = total;

        // 记录操作日志
        await logOperation(event.adminInfo, 'VIEW', 'APPS', {
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
const mainFunc = requirePermission(getAllAppsHandler, 'app_manage');
exports.main = mainFunc;

// 自动注册API
const { autoRegister } = require('../api-factory');
autoRegister('app.getAll')(mainFunc);