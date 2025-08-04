const cloud = require("@alipay/faas-server-sdk");
const common = require("./common");
const { requirePermission, logOperation } = require("./common/auth");

// 原始处理函数
async function getAllUsersHandler(event, context) {
    let appId = event.appId;
    let page = event.page || 1;
    let pageSize = event.pageSize || 20;
    let playerId = event.playerId;
    let openId = event.openId;

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
    const userTableName = `user_${appId}`;

    try {
        // 检查用户表是否存在
        let collection;
        try {
            collection = db.collection(userTableName);
        } catch (e) {
            ret.code = 4004;
            ret.msg = "应用不存在或用户表不存在";
            return ret;
        }

        // 构建查询条件
        let whereCondition = {};
        
        if (playerId) {
            whereCondition.playerId = new RegExp(playerId, 'i'); // 模糊搜索
        }
        
        if (openId) {
            whereCondition.openId = new RegExp(openId, 'i'); // 模糊搜索
        }

        // 查询总数
        const countResult = await collection.where(whereCondition).count();
        const total = countResult.total;

        // 分页查询
        const skip = (page - 1) * pageSize;
        let userList = await collection
            .where(whereCondition)
            .orderBy('gmtCreate', 'desc')
            .skip(skip)
            .limit(pageSize)
            .get();

        // 处理用户数据
        userList = userList.map(user => {
            // 解析游戏数据
            if (user.data && typeof user.data === 'string') {
                try {
                    user.gameData = JSON.parse(user.data);
                } catch (e) {
                    user.gameData = null;
                }
            }

            // 设置默认状态
            user.banned = user.banned || false;
            
            // 删除敏感信息
            delete user.token;
            
            return user;
        });

        ret.data.list = userList;
        ret.data.total = total;

        // 记录操作日志
        await logOperation(event.adminInfo, 'VIEW', 'USERS', {
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
const mainFunc = requirePermission(getAllUsersHandler, 'user_manage');
exports.main = mainFunc;