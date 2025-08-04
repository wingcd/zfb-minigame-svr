const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：getAllRoles
 * 说明：获取所有角色列表（用于下拉框选择）
 * 权限：需要 role_manage 权限
 * 参数：无
 * 
 * 测试数据：
    {}
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "roles": [
                {
                    "roleCode": "super_admin",
                    "roleName": "超级管理员",
                    "description": "系统最高权限，拥有所有操作权限",
                    "permissions": ["admin_manage", "role_manage", "app_manage", "user_manage", "leaderboard_manage", "stats_view", "system_config"]
                },
                {
                    "roleCode": "admin",
                    "roleName": "管理员",
                    "description": "管理员权限，可管理应用、用户、排行榜",
                    "permissions": ["app_manage", "user_manage", "leaderboard_manage", "stats_view"]
                }
            ],
            "total": 4
        }
    }
    
 * 错误码：
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getAllRolesHandler(event, context) {
    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    const db = cloud.database();

    try {
        // 获取所有角色（用于下拉框选择）
        let roleList = await db.collection('admin_roles')
            .orderBy('createTime', 'asc')
            .get();

        // 简化角色信息，只返回必要字段
        const roles = roleList.map(role => ({
            roleCode: role.roleCode,
            roleName: role.roleName,
            description: role.description,
            permissions: role.permissions
        }));

        ret.data = {
            roles: roles,
            total: roles.length
        };

        // 记录操作日志（低频率）
        const shouldLog = Math.random() < 0.05; // 5% 概率记录
        if (shouldLog) {
            await logOperation(event.adminInfo, 'VIEW', 'ALL_ROLES', {
                roleCount: roles.length,
                currentAdmin: event.adminInfo.username
            });
        }

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getAllRolesHandler, 'role_manage');
exports.main = mainFunc;