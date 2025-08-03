const cloud = require("@alipay/faas-server-sdk");
const common = require("./common");

// 请求参数
/**
 * 函数：getAllUsers
 * 说明：获取用户列表
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |
    | playerId | string | 否 | 玩家ID搜索 |
    | openId | string | 否 | OpenID搜索 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "page": 1,
        "pageSize": 20,
        "playerId": "player001"
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

exports.main = async (event, context) => {
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

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 