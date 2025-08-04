const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

/**
 * 函数：updateAdmin
 * 说明：更新管理员信息
 * 权限：需要 admin_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | id | string | 是 | 管理员ID |
    | nickname | string | 否 | 昵称 |
    | role | string | 否 | 角色代码 |
    | email | string | 否 | 邮箱地址 |
    | phone | string | 否 | 手机号码 |
    | status | string | 否 | 状态（active/inactive） |
 * 
 * 权限规则：
 * - 超级管理员可以修改任何人的信息
 * - 管理员不能修改超级管理员或提升为超级管理员
 * - 不能修改自己的角色
 * - 邮箱不能与其他用户重复
 * 
 * 测试数据：
    {
        "id": "admin_id_123456",
        "nickname": "新昵称",
        "role": "admin",
        "email": "newemail@example.com",
        "phone": "13900139000",
        "status": "active"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "更新成功",
        "timestamp": 1603991234567,
        "data": {
            "id": "admin_id_123456",
            "updatedFields": ["nickname", "email", "phone", "updateTime", "updatedBy"]
        }
    }
    
 * 错误码：
 * - 4001: 参数错误
 * - 4002: 邮箱已被其他用户使用
 * - 4003: 权限不足
 * - 4004: 管理员或角色不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function updateAdminHandler(event, context) {
    let id = event.id;
    let nickname = event.nickname;
    let role = event.role;
    let email = event.email;
    let phone = event.phone;
    let status = event.status;

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
        // 检查要更新的管理员是否存在
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

        // 权限检查：不能修改比自己权限更高的管理员
        if (currentAdmin.role !== 'super_admin') {
            // 非超级管理员不能修改超级管理员
            if (targetAdmin.role === 'super_admin') {
                ret.code = 4003;
                ret.msg = "无权限修改超级管理员";
                return ret;
            }
            
            // 不能把管理员提升为超级管理员
            if (role === 'super_admin') {
                ret.code = 4003;
                ret.msg = "无权限设置超级管理员角色";
                return ret;
            }
            
            // 不能修改自己的角色
            if (targetAdmin.username === currentAdmin.username && role && role !== targetAdmin.role) {
                ret.code = 4003;
                ret.msg = "不能修改自己的角色";
                return ret;
            }
        }

        // 验证新角色是否存在
        if (role && role !== targetAdmin.role) {
            const roleList = await db.collection('admin_roles')
                .where({ roleCode: role })
                .get();

            if (roleList.length === 0) {
                ret.code = 4004;
                ret.msg = "指定的角色不存在";
                return ret;
            }
        }

        // 邮箱格式验证
        if (email && email !== targetAdmin.email) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(email)) {
                ret.code = 4001;
                ret.msg = "邮箱格式不正确";
                return ret;
            }

            // 检查邮箱是否已被其他用户使用
            const existingEmail = await db.collection('admin_users')
                .where({ 
                    email: email,
                    _id: { $ne: id }
                })
                .get();

            if (existingEmail.length > 0) {
                ret.code = 4002;
                ret.msg = "邮箱已被其他用户使用";
                return ret;
            }
        }

        // 手机号格式验证
        if (phone && phone !== targetAdmin.phone) {
            const phoneRegex = /^1[3-9]\d{9}$/;
            if (!phoneRegex.test(phone)) {
                ret.code = 4001;
                ret.msg = "手机号格式不正确";
                return ret;
            }
        }

        // 状态值验证
        if (status && !['active', 'inactive'].includes(status)) {
            ret.code = 4001;
            ret.msg = "状态值只能是 active 或 inactive";
            return ret;
        }

        // 昵称长度验证
        if (nickname !== undefined && (nickname.length < 1 || nickname.length > 50)) {
            ret.code = 4001;
            ret.msg = "昵称长度必须在 1-50 个字符之间";
            return ret;
        }

        // 构建更新数据
        let updateData = {
            updateTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            updatedBy: currentAdmin.username
        };

        let updatedFields = ['updateTime', 'updatedBy'];

        if (nickname !== undefined && nickname !== targetAdmin.nickname) {
            updateData.nickname = nickname;
            updatedFields.push('nickname');
        }

        if (role !== undefined && role !== targetAdmin.role) {
            updateData.role = role;
            updatedFields.push('role');
        }

        if (email !== undefined && email !== targetAdmin.email) {
            updateData.email = email;
            updatedFields.push('email');
        }

        if (phone !== undefined && phone !== targetAdmin.phone) {
            updateData.phone = phone;
            updatedFields.push('phone');
        }

        if (status !== undefined && status !== targetAdmin.status) {
            updateData.status = status;
            updatedFields.push('status');
            
            // 如果禁用管理员，清除其token
            if (status === 'inactive') {
                updateData.token = null;
                updateData.tokenExpire = null;
                updatedFields.push('token', 'tokenExpire');
            }
        }

        // 如果没有实际更新的字段，直接返回成功
        if (updatedFields.length <= 2) { // 只有 updateTime 和 updatedBy
            ret.msg = "没有需要更新的字段";
            ret.data = {
                id: id,
                updatedFields: []
            };
            return ret;
        }

        // 更新管理员信息
        await db.collection('admin_users')
            .doc(id)
            .update({
                data: updateData
            });

        // 记录操作日志
        await logOperation(currentAdmin, 'UPDATE', 'ADMIN', {
            targetAdminId: id,
            targetAdminUsername: targetAdmin.username,
            originalData: {
                nickname: targetAdmin.nickname,
                role: targetAdmin.role,
                email: targetAdmin.email,
                phone: targetAdmin.phone,
                status: targetAdmin.status
            },
            changes: updateData,
            severity: 'HIGH'  // 修改管理员信息是高风险操作
        });

        ret.msg = "更新成功";
        ret.data = {
            id: id,
            updatedFields: updatedFields.filter(field => !['updateTime', 'updatedBy'].includes(field))
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(updateAdminHandler, 'admin_manage');
exports.main = mainFunc; 