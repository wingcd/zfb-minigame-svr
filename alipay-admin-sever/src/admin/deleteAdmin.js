const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：deleteAdmin
 * 说明：删除管理员账户
 * 权限：需要 admin_manage 权限（仅超级管理员可删除）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | id | string | 是 | 要删除的管理员ID |
 * 
 * 权限规则：
 * - 只有超级管理员可以删除管理员
 * - 不能删除自己的账户
 * - 不能删除最后一个超级管理员
 * 
 * 测试数据：
    {
        "id": "admin_id_123456"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "删除成功",
        "timestamp": 1603991234567,
        "data": {
            "deletedAdmin": {
                "id": "admin_id_123456",
                "username": "testadmin",
                "nickname": "测试管理员"
            }
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足或业务限制
 * - 4004: 管理员不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function deleteAdminHandler(event, context) {
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
        ret.msg = "管理员ID不能为空";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查要删除的管理员是否存在
        const adminList = await db.collection('admin_users')
            .where({ _id: id })
            .get();

        if (adminList.length === 0) {
            ret.code = 4004;
            ret.msg = "管理员不存在";
            return ret;
        }

        const targetAdmin = adminList[0];
        const currentAdmin = event.adminInfo;

        // 权限检查：只有超级管理员可以删除管理员
        if (currentAdmin.role !== 'super_admin') {
            ret.code = 4003;
            ret.msg = "删除管理员需要超级管理员权限";
            return ret;
        }

        // 不能删除自己
        if (targetAdmin.username === currentAdmin.username) {
            ret.code = 4003;
            ret.msg = "不能删除自己的账户";
            return ret;
        }

        // 检查是否是最后一个超级管理员
        if (targetAdmin.role === 'super_admin') {
            const superAdminCount = await db.collection('admin_users')
                .where({ 
                    role: 'super_admin',
                    status: 'active'
                })
                .count();

            if (superAdminCount.total <= 1) {
                ret.code = 4003;
                ret.msg = "不能删除最后一个超级管理员";
                return ret;
            }
        }

        // 保存要删除的管理员信息用于日志
        const deletedAdminInfo = {
            id: targetAdmin._id,
            username: targetAdmin.username,
            nickname: targetAdmin.nickname,
            role: targetAdmin.role,
            email: targetAdmin.email,
            createTime: targetAdmin.createTime
        };

        // 删除管理员
        await db.collection('admin_users').doc(id).remove();

        // 记录操作日志
        await logOperation(currentAdmin, 'DELETE', 'ADMIN', {
            deletedAdmin: deletedAdminInfo,
            deletedBy: currentAdmin.username,
            severity: 'CRITICAL'  // 删除管理员是关键操作
        });

        ret.msg = "删除成功";
        ret.data = {
            deletedAdmin: {
                id: deletedAdminInfo.id,
                username: deletedAdminInfo.username,
                nickname: deletedAdminInfo.nickname
            }
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
exports.main = requirePermission(deleteAdminHandler, 'admin_manage');