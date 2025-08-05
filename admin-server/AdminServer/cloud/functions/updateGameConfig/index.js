const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：updateGameConfig
 * 说明：更新游戏配置
 * 权限：需要 app_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | id | string | 是 | 配置ID |
    | configValue | any | 否 | 配置值 |
    | description | string | 否 | 配置描述 |
    | configType | string | 否 | 配置类型 |
    | isActive | boolean | 否 | 是否激活 |
 * 
 * 测试数据：
    {
        "id": "config_id_123456",
        "configValue": 120,
        "description": "更新后的最大关卡数",
        "isActive": true
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "id": "config_id_123456",
            "updated": true
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 配置不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function updateGameConfigHandler(event, context) {
    let id = event.id;
    let configValue = event.configValue;
    let description = event.description;
    let configType = event.configType;
    let isActive = event.isActive;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!id || typeof id !== "string") {
        ret.code = 4001;
        ret.msg = "配置ID不能为空";
        return ret;
    }

    // 验证配置类型
    if (configType) {
        const validTypes = ['string', 'number', 'boolean', 'object', 'array'];
        if (!validTypes.includes(configType)) {
            ret.code = 4001;
            ret.msg = "无效的配置类型";
            return ret;
        }
    }

    const db = cloud.database();

    try {
        // 检查配置是否存在
        const existingConfig = await db.collection('game_config')
            .doc(id)
            .get();

        if (existingConfig.length === 0) {
            ret.code = 4004;
            ret.msg = "配置不存在";
            return ret;
        }

        // 构建更新数据
        const updateData = {
            updateTime: moment().format('YYYY-MM-DD HH:mm:ss')
        };

        if (configValue !== undefined) {
            updateData.configValue = configValue;
        }

        if (description !== undefined) {
            updateData.description = description;
        }

        if (configType !== undefined) {
            updateData.configType = configType;
        }

        if (isActive !== undefined) {
            updateData.isActive = isActive;
        }

        // 更新配置
        await db.collection('game_config')
            .doc(id)
            .update({
                data: updateData
            });

        ret.data = {
            id: id,
            updated: true
        };

        return ret;

    } catch (error) {
        console.error('更新游戏配置失败:', error);
        ret.code = 5001;
        ret.msg = "服务器内部错误";
        return ret;
    }
}

// 导出函数
exports.main = requirePermission(updateGameConfigHandler, ['app_manage']); 