const cloud = require("@alipay/faas-server-sdk");
const crypto = require("crypto");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：resetPassword
 * 说明：重置管理员密码
 * 权限：需要 admin_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | id | string | 是 | 管理员ID |
    | newPassword | string | 否 | 新密码（默认：123456，最少6位） |
 * 
 * 权限规则：
 * - 超级管理员可以重置任何人的密码
 * - 管理员只能重置自己的密码或下级的密码
 * - 不能重置比自己权限更高的管理员密码
 * 
 * 测试数据：
    {
        "id": "admin_id_123456",
        "newPassword": "newpassword123"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "密码重置成功",
        "timestamp": 1603991234567,
        "data": {
            "adminId": "admin_id_123456",
            "adminUsername": "testadmin",
            "newPassword": "newpassword123",
            "message": "管理员需要重新登录"
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4003: 权限不足
 * - 4004: 管理员不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function resetPasswordHandler(event, context) {
    let id = event.id;
    let newPassword = event.newPassword || "123456";

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

    if (!newPassword || typeof newPassword !== "string" || newPassword.length < 6) {
        ret.code = 4001;
        ret.msg = "密码必须至少6个字符";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查要重置密码的管理员是否存在
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

        // 权限检查：不能重置比自己权限更高的管理员密码
        if (currentAdmin.role !== 'super_admin') {
            // 非超级管理员不能重置超级管理员密码
            if (targetAdmin.role === 'super_admin') {
                ret.code = 4003;
                ret.msg = "无权限重置超级管理员密码";
                return ret;
            }
            
            // 只能重置自己的密码或下级管理员的密码
            if (targetAdmin.username !== currentAdmin.username && currentAdmin.role !== 'admin') {
                ret.code = 4003;
                ret.msg = "权限不足";
                return ret;
            }
        }

        // 密码加密
        const passwordHash = crypto.createHash('md5').update(newPassword).digest('hex');

        // 更新密码并清除token（强制重新登录）
        await db.collection('admin_users')
            .doc(id)
            .update({
                data: {
                    password: passwordHash,
                    token: null,
                    tokenExpire: null,
                    updateTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                    passwordResetBy: currentAdmin.username,
                    passwordResetTime: moment().format("YYYY-MM-DD HH:mm:ss")
                }
            });

        // 记录操作日志
        await logOperation(currentAdmin, 'RESET_PASSWORD', 'ADMIN', {
            targetAdminId: id,
            targetAdminUsername: targetAdmin.username,
            resetBy: currentAdmin.username,
            isSelfReset: targetAdmin.username === currentAdmin.username,
            severity: 'HIGH'  // 重置密码是高风险操作
        });

        ret.msg = "密码重置成功";
        ret.data = {
            adminId: id,
            adminUsername: targetAdmin.username,
            newPassword: newPassword,
            message: "管理员需要重新登录"
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(resetPasswordHandler, 'admin_manage');
exports.main = mainFunc;