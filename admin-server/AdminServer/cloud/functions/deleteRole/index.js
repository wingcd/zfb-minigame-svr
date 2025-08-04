const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：deleteRole
 * 说明：删除角色
 * 权限：需要 role_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | roleCode | string | 是 | 角色代码 |
 * 
 * 测试数据：
    {
        "roleCode": "custom_admin"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {}
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 角色不存在
 * - 4005: 不能删除系统预设角色
 * - 4006: 角色正在使用中，不能删除
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function deleteRoleHandler(event, context) {
    let roleCode = event.roleCode;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!roleCode || typeof roleCode !== "string") {
        ret.code = 4001;
        ret.msg = "角色代码不能为空";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查角色是否存在
        const roleList = await db.collection('admin_roles')
            .where({ roleCode: roleCode })
            .get();

        if (roleList.length === 0) {
            ret.code = 4004;
            ret.msg = "角色不存在";
            return ret;
        }

        const roleData = roleList[0];

        // 检查是否为系统预设角色
        const systemRoles = ['super_admin', 'admin', 'operator', 'viewer'];
        if (systemRoles.includes(roleCode)) {
            ret.code = 4005;
            ret.msg = "不能删除系统预设角色";
            return ret;
        }

        // 检查是否有管理员正在使用此角色
        const adminsUsingRole = await db.collection('admin_users')
            .where({ role: roleCode })
            .get();

        if (adminsUsingRole.length > 0) {
            ret.code = 4006;
            ret.msg = `角色正在被${adminsUsingRole.length}个管理员使用，不能删除`;
            return ret;
        }

        // 删除角色
        await db.collection('admin_roles')
            .where({ roleCode: roleCode })
            .remove();

        // 记录操作日志
        await logOperation(event.adminInfo, 'DELETE', 'ROLE', {
            roleCode: roleCode,
            roleName: roleData.roleName,
            deletedBy: event.adminInfo.username,
            severity: 'HIGH'
        });

        ret.msg = "删除成功";

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(deleteRoleHandler, 'role_manage');
exports.main = mainFunc; 