const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

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
exports.main = requirePermission(getAllRolesHandler, 'role_manage'); 