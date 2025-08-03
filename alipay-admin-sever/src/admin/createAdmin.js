const cloud = require("@alipay/faas-server-sdk");
const crypto = require("crypto");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：createAdmin
 * 说明：创建新管理员账户
 * 权限：需要 admin_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | username | string | 是 | 用户名（至少3个字符） |
    | password | string | 是 | 密码（至少6个字符） |
    | nickname | string | 否 | 昵称（默认同用户名） |
    | role | string | 否 | 角色代码（默认：viewer） |
    | email | string | 否 | 邮箱地址 |
    | phone | string | 否 | 手机号码 |
 * 
 * 角色说明：
 * - super_admin: 超级管理员
 * - admin: 管理员
 * - operator: 运营人员
 * - viewer: 查看者
 * 
 * 权限规则：
 * - 超级管理员可以创建任何角色的账户
 * - 管理员不能创建超级管理员或同级管理员
 * - 用户名和邮箱不能重复
 * 
 * 测试数据：
    {
        "username": "testadmin",
        "password": "password123",
        "nickname": "测试管理员",
        "role": "admin",
        "email": "test@example.com",
        "phone": "13800138000"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "创建成功",
        "timestamp": 1603991234567,
        "data": {
            "id": "admin_id_123456",
            "username": "testadmin",
            "nickname": "测试管理员",
            "role": "admin",
            "status": "active",
            "createTime": "2023-10-01 10:00:00"
        }
    }
    
 * 错误码：
 * - 4001: 参数错误或无效角色
 * - 4002: 用户名或邮箱已存在
 * - 4003: 权限不足
 * - 4004: 指定角色不存在
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function createAdminHandler(event, context) {
    let username = event.username;
    let password = event.password;
    let nickname = event.nickname || username;
    let role = event.role || 'viewer';
    let email = event.email;
    let phone = event.phone;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!username || typeof username !== "string" || username.length < 3) {
        ret.code = 4001;
        ret.msg = "用户名必须至少3个字符";
        return ret;
    }

    if (!password || typeof password !== "string" || password.length < 6) {
        ret.code = 4001;
        ret.msg = "密码必须至少6个字符";
        return ret;
    }

    // 验证角色是否有效
    const validRoles = ['super_admin', 'admin', 'operator', 'viewer'];
    if (!validRoles.includes(role)) {
        ret.code = 4001;
        ret.msg = "无效的角色";
        return ret;
    }

    // 权限检查：不能创建比自己权限更高的管理员
    const currentRole = event.adminInfo.role;
    if (currentRole !== 'super_admin') {
        // 非超级管理员不能创建super_admin
        if (role === 'super_admin') {
            ret.code = 4003;
            ret.msg = "只有超级管理员可以创建超级管理员账户";
            return ret;
        }
        
        // admin 不能创建 admin 角色
        if (currentRole === 'admin' && role === 'admin') {
            ret.code = 4003;
            ret.msg = "管理员不能创建同级管理员账户";
            return ret;
        }
    }

    const db = cloud.database();

    try {
        // 检查用户名是否已存在
        const existingAdmins = await db.collection('admin_users')
            .where({ username: username })
            .get();

        if (existingAdmins.length > 0) {
            ret.code = 4002;
            ret.msg = "用户名已存在";
            return ret;
        }

        // 检查邮箱是否已存在
        if (email) {
            const existingEmail = await db.collection('admin_users')
                .where({ email: email })
                .get();

            if (existingEmail.length > 0) {
                ret.code = 4002;
                ret.msg = "邮箱已被使用";
                return ret;
            }
        }

        // 验证指定的角色是否存在
        const roleList = await db.collection('admin_roles')
            .where({ roleCode: role })
            .get();

        if (roleList.length === 0) {
            ret.code = 4004;
            ret.msg = "指定的角色不存在";
            return ret;
        }

        // 密码加密
        const passwordHash = crypto.createHash('md5').update(password).digest('hex');

        // 创建管理员
        const newAdmin = {
            username: username,
            password: passwordHash,
            nickname: nickname,
            role: role,
            email: email || null,
            phone: phone || null,
            status: 'active',
            createTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            createdBy: event.adminInfo.username,
            lastLoginTime: null,
            token: null,
            tokenExpire: null
        };

        const addResult = await db.collection('admin_users').add({
            data: newAdmin
        });

        // 记录操作日志
        await logOperation(event.adminInfo, 'CREATE', 'ADMIN', {
            newAdminUsername: username,
            newAdminRole: role,
            newAdminNickname: nickname,
            createdBy: event.adminInfo.username,
            severity: 'HIGH'  // 创建管理员是高风险操作
        });

        ret.msg = "创建成功";
        ret.data = {
            id: addResult._id,
            username: username,
            nickname: nickname,
            role: role,
            status: 'active',
            createTime: newAdmin.createTime
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 使用装饰器方式自动注册
exports.main = requirePermission(createAdminHandler, 'admin_manage');