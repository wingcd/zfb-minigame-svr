const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：createRole
 * 说明：创建新角色
 * 权限：需要 role_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | roleCode | string | 是 | 角色代码（唯一标识） |
    | roleName | string | 是 | 角色名称 |
    | description | string | 否 | 角色描述 |
    | permissions | array | 是 | 权限列表 |
 * 
 * 测试数据：
    {
        "roleCode": "custom_admin",
        "roleName": "自定义管理员",
        "description": "自定义权限的管理员角色",
        "permissions": ["app_manage", "user_manage"]
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "id": "role_id_123456",
            "roleCode": "custom_admin",
            "roleName": "自定义管理员",
            "description": "自定义权限的管理员角色",
            "permissions": ["app_manage", "user_manage"],
            "createTime": "2023-10-01 10:00:00"
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4002: 角色代码已存在
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function createRoleHandler(event, context) {
    let roleCode = event.roleCode;
    let roleName = event.roleName;
    let description = event.description || '';
    let permissions = event.permissions || [];

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!roleCode || typeof roleCode !== "string" || roleCode.length < 2) {
        ret.code = 4001;
        ret.msg = "角色代码必须至少2个字符";
        return ret;
    }

    if (!roleName || typeof roleName !== "string" || roleName.length < 2) {
        ret.code = 4001;
        ret.msg = "角色名称必须至少2个字符";
        return ret;
    }

    if (!Array.isArray(permissions)) {
        ret.code = 4001;
        ret.msg = "权限列表必须是数组";
        return ret;
    }

    // 验证权限列表
    const validPermissions = [
        'admin_manage', 'role_manage', 'app_manage', 'user_manage', 
        'leaderboard_manage', 'mail_manage', 'stats_view', 'system_config'
    ];
    
    for (let permission of permissions) {
        if (!validPermissions.includes(permission)) {
            ret.code = 4001;
            ret.msg = `无效的权限: ${permission}`;
            return ret;
        }
    }

    const db = cloud.database();

    try {
        // 检查角色代码是否已存在
        const existingRoles = await db.collection('admin_roles')
            .where({ roleCode: roleCode })
            .get();

        if (existingRoles.length > 0) {
            ret.code = 4002;
            ret.msg = "角色代码已存在";
            return ret;
        }

        // 创建角色
        const newRole = {
            roleCode: roleCode,
            roleName: roleName,
            description: description,
            permissions: permissions,
            createTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            createdBy: event.adminInfo.username,
            status: 'active'
        };

        const addResult = await db.collection('admin_roles').add({
            data: newRole
        });

        // 记录操作日志
        await logOperation(event.adminInfo, 'CREATE', 'ROLE', {
            roleCode: roleCode,
            roleName: roleName,
            permissions: permissions,
            createdBy: event.adminInfo.username,
            severity: 'HIGH'
        });

        ret.msg = "创建成功";
        ret.data = {
            id: addResult._id,
            roleCode: roleCode,
            roleName: roleName,
            description: description,
            permissions: permissions,
            createTime: newRole.createTime
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(createRoleHandler, 'role_manage');
exports.main = mainFunc; 