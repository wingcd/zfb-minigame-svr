const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：deleteGameConfig
 * 说明：删除游戏配置
 * 权限：需要 app_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | id | string | 是 | 配置ID |
 * 
 * 测试数据：
    {
        "id": "config_id_123456"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "id": "config_id_123456",
            "deleted": true
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 配置不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function deleteGameConfigHandler(event, context) {
    let id = event.id;

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

    const db = cloud.database();

    try {
        // 检查配置是否存在
        const existingConfig = await db.collection('game_config')
            .doc(id)
            .get();

        if (!existingConfig) {
            ret.code = 4004;
            ret.msg = "配置不存在";
            return ret;
        }

        // 删除配置
        await db.collection('game_config')
            .doc(id)
            .remove();

        ret.data = {
            id: id,
            deleted: true
        };

        return ret;

    } catch (error) {
        console.error('删除游戏配置失败:', error);
        ret.code = 5001;
        ret.msg = "服务器内部错误";
        return ret;
    }
}

// 导出函数
exports.main = requirePermission(deleteGameConfigHandler, ['app_manage']); 