const cloud = require("@alipay/faas-server-sdk");

/**
 * 函数：getGameConfig
 * 说明：获取游戏配置（客户端调用，优先返回版本配置）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 应用ID |
    | version | string | 否 | 游戏版本 |
    | configKey | string | 否 | 特定配置键名（为空时返回所有配置） |
 * 
 * 测试数据：
    {
        "appId": "test_game_001",
        "version": "1.0.0",
        "configKey": "max_level"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "configs": {
                "max_level": {
                    "value": 100,
                    "type": "number",
                    "source": "version", // version 或 global
                    "version": "1.0.0"
                },
                "game_name": {
                    "value": "超级游戏",
                    "type": "string",
                    "source": "global",
                    "version": null
                }
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4004: 应用不存在
 * - 5001: 服务器内部错误
 */

exports.main = async (event, context) => {
    let appId = event.appId;
    let version = event.version || null;
    let configKey = event.configKey || null;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {
            "configs": {}
        }
    };

    // 参数校验
    if (!appId || typeof appId !== "string") {
        ret.code = 4001;
        ret.msg = "应用ID不能为空";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查应用是否存在
        const apps = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (apps.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        // 构建查询条件
        const whereCondition = { 
            appId: appId,
            isActive: true
        };

        if (configKey) {
            whereCondition.configKey = configKey;
        }

        // 获取所有相关配置（包括全局和版本配置）
        const allConfigs = await db.collection('game_config')
            .where(whereCondition)
            .get();

        // 按优先级处理配置：版本配置优先于全局配置
        const configMap = {};

        // 先处理全局配置
        allConfigs.forEach(config => {
            if (!config.version) {
                configMap[config.configKey] = {
                    value: config.configValue,
                    type: config.configType,
                    source: 'global',
                    version: null,
                    description: config.description || ''
                };
            }
        });

        // 再处理版本配置（会覆盖同名的全局配置）
        if (version) {
            allConfigs.forEach(config => {
                if (config.version === version) {
                    configMap[config.configKey] = {
                        value: config.configValue,
                        type: config.configType,
                        source: 'version',
                        version: config.version,
                        description: config.description || ''
                    };
                }
            });
        }

        ret.data.configs = configMap;
        return ret;

    } catch (error) {
        console.error('获取游戏配置失败:', error);
        ret.code = 5001;
        ret.msg = "服务器内部错误";
        return ret;
    }
}; 