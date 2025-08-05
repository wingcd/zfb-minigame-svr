const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：getGameConfigList
 * 说明：获取游戏配置列表（管理后台使用）
 * 权限：需要 app_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | version | string | 否 | 过滤特定版本（为空时显示所有版本） |
    | configKey | string | 否 | 过滤特定配置键名 |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "page": 1,
        "pageSize": 20
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "list": [
                {
                    "id": "config_id_123456",
                    "appId": "test_game_001",
                    "configKey": "max_level",
                    "configValue": 100,
                    "version": "1.0.0",
                    "description": "最大关卡数",
                    "configType": "number",
                    "isActive": true,
                    "createTime": "2023-10-01 10:00:00",
                    "updateTime": "2023-10-01 10:00:00"
                }
            ],
            "total": 10,
            "page": 1,
            "pageSize": 20,
            "versions": ["1.0.0", "1.1.0", "2.0.0"]
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 应用不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getGameConfigListHandler(event, context) {
    let appId = event.appId;
    let version = event.version || null;
    let configKey = event.configKey || null;
    let page = event.page || 1;
    let pageSize = event.pageSize || 20;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {
            "list": [],
            "total": 0,
            "page": page,
            "pageSize": pageSize,
            "versions": []
        }
    };

    // 参数校验
    if (!appId || typeof appId !== "string") {
        ret.code = 4001;
        ret.msg = "应用ID不能为空";
        return ret;
    }

    if (page < 1) {
        ret.code = 4001;
        ret.msg = "页码必须大于0";
        return ret;
    }

    if (pageSize < 1 || pageSize > 100) {
        ret.code = 4001;
        ret.msg = "每页数量必须在1-100之间";
        return ret;
    }

    const db = cloud.database();

    // 创建必要的集合（表）
    const gameConfigCollection = 'game_config'
    try {
        await db.getCollection(gameConfigCollection);
    } catch (e) {
        if (e.message == "not found collection") {
            await db.createCollection(gameConfigCollection);
        }
    }

    try {
        // 检查应用是否存在
        const appsResult = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (appsResult.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        // 构建查询条件
        const whereCondition = { appId: appId };

        if (version !== null) {
            if (version === '') {
                // 查询全局配置（没有version字段）
                whereCondition.version = db.command.exists(false);
            } else {
                // 查询特定版本配置
                whereCondition.version = version;
            }
        }

        if (configKey) {
            whereCondition.configKey = configKey;
        }

        // 获取总数
        const countResult = await db.collection(gameConfigCollection)
            .where(whereCondition)
            .count();

        ret.data.total = countResult.total;

        // 获取配置列表
        const skip = (page - 1) * pageSize;
        const configList = await db.collection(gameConfigCollection)
            .where(whereCondition)
            .orderBy('createTime', 'desc')
            .skip(skip)
            .limit(pageSize)
            .get();

        // 将_id映射为id字段，方便前端使用
        ret.data.list = configList.map(config => {
            let id = config._id;
            delete config._id;
            return {
                ...config,
                id: id
            }
        });

        // 获取所有版本列表
        const allConfigs = await db.collection(gameConfigCollection)
            .where({ appId: appId })
            .field({ version: true })
            .get();

        const versionSet = new Set();
        allConfigs.forEach(config => {
            if (config.version) {
                versionSet.add(config.version);
            }
        });

        ret.data.versions = Array.from(versionSet).sort();

        return ret;

    } catch (error) {
        console.error('获取游戏配置列表失败:', error);
        ret.code = 5001;
        ret.msg = "服务器内部错误";
        return ret;
    }
}

// 导出函数
exports.main = requirePermission(getGameConfigListHandler, ['app_manage']); 