const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：updateRole
 * 说明：更新角色信息
 * 权限：需要 role_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | roleCode | string | 是 | 角色代码 |
    | roleName | string | 否 | 角色名称 |
    | description | string | 否 | 角色描述 |
    | permissions | array | 否 | 权限列表 |
    | status | string | 否 | 状态 (active/inactive) |
 * 
 * 测试数据：
    {
        "roleCode": "custom_admin",
        "roleName": "更新后的管理员",
        "description": "更新后的描述",
        "permissions": ["app_manage", "user_manage", "stats_view"]
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
 * - 4005: 不能修改系统预设角色
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function updateRoleHandler(event, context) {
    let roleCode = event.roleCode;
    let roleName = event.roleName;
    let description = event.description;
    let permissions = event.permissions;
    let status = event.status;

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

    // 验证权限列表
    if (permissions && Array.isArray(permissions)) {
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

        const oldRoleData = roleList[0];

        // 检查是否为系统预设角色
        const systemRoles = ['super_admin', 'admin', 'operator', 'viewer'];
        if (systemRoles.includes(roleCode)) {
            ret.code = 4005;
            ret.msg = "不能修改系统预设角色";
            return ret;
        }

        // 构建更新数据
        let updateData = {
            updateTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            updatedBy: event.adminInfo.username
        };

        if (roleName !== undefined) {
            updateData.roleName = roleName;
        }

        if (description !== undefined) {
            updateData.description = description;
        }

        if (permissions !== undefined) {
            updateData.permissions = permissions;
        }

        if (status !== undefined) {
            updateData.status = status;
        }

        // 更新角色信息
        await db.collection('admin_roles')
            .where({ roleCode: roleCode })
            .update({
                data: updateData
            });

        // 记录操作日志
        await logOperation(event.adminInfo, 'UPDATE', 'ROLE', {
            roleCode: roleCode,
            roleName: oldRoleData.roleName,
            changes: updateData,
            updatedBy: event.adminInfo.username,
            severity: 'HIGH'
        });

        ret.msg = "更新成功";

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(updateRoleHandler, 'role_manage');
exports.main = mainFunc; 