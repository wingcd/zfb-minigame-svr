const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：createGameConfig
 * 说明：创建游戏配置
 * 权限：需要 app_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | configKey | string | 是 | 配置键名 |
    | configValue | any | 是 | 配置值 |
    | version | string | 否 | 游戏版本（为空时为全局配置） |
    | description | string | 否 | 配置描述 |
    | configType | string | 否 | 配置类型 (string/number/boolean/object/array) |
    | isActive | boolean | 否 | 是否激活，默认true |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "configKey": "max_level",
        "configValue": 100,
        "version": "1.0.0",
        "description": "最大关卡数",
        "configType": "number"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
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
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4002: 配置已存在
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function createGameConfigHandler(event, context) {
    let appId = event.appId;
    let configKey = event.configKey;
    let configValue = event.configValue;
    let version = event.version || null; // null表示全局配置
    let description = event.description || '';
    let configType = event.configType || 'string';
    let isActive = event.isActive !== undefined ? event.isActive : true;

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

    if (!configKey || typeof configKey !== "string") {
        ret.code = 4001;
        ret.msg = "配置键名不能为空";
        return ret;
    }

    if (configValue === undefined || configValue === null) {
        ret.code = 4001;
        ret.msg = "配置值不能为空";
        return ret;
    }

    // 验证配置类型
    const validTypes = ['string', 'number', 'boolean', 'object', 'array'];
    if (!validTypes.includes(configType)) {
        ret.code = 4001;
        ret.msg = "无效的配置类型";
        return ret;
    }

    const db = cloud.database();

    try {
        // 确保 game_config 集合存在
        try {
            await db.getCollection('game_config');
        } catch (e) {
            if (e.message == "not found collection") {
                await db.createCollection('game_config');
                console.log('game_config 集合创建成功');
            } else {
                throw e;
            }
        }

        // 检查应用是否存在
        const apps = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (apps.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        // 检查配置是否已存在（同一个appId、configKey、version的组合）
        const whereCondition = { 
            appId: appId, 
            configKey: configKey 
        };
        
        if (version) {
            whereCondition.version = version;
        } else {
            whereCondition.version = cloud.database().command.exists(false);
        }

        const existingConfigs = await db.collection('game_config')
            .where(whereCondition)
            .get();

        if (existingConfigs.length > 0) {
            ret.code = 4002;
            ret.msg = version ? `版本 ${version} 的配置 ${configKey} 已存在` : `全局配置 ${configKey} 已存在`;
            return ret;
        }

        // 创建配置记录
        const now = moment().format('YYYY-MM-DD HH:mm:ss');
        const configData = {
            appId: appId,
            configKey: configKey,
            configValue: configValue,
            description: description,
            configType: configType,
            isActive: isActive,
            createTime: now,
            updateTime: now
        };

        // 只有指定了版本才添加version字段
        if (version) {
            configData.version = version;
        }

        const result = await db.collection('game_config').add({
            data: configData
        });

        ret.data = {
            id: result.id,
            ...configData
        };

        return ret;

    } catch (error) {
        console.error('创建游戏配置失败:', error);
        ret.code = 5001;
        ret.msg = "服务器内部错误";
        return ret;
    }
}

// 导出函数
exports.main = requirePermission(createGameConfigHandler, ['app_manage']); 