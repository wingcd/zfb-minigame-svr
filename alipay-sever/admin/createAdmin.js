const cloud = require("@alipay/faas-server-sdk");
const crypto = require('crypto');
const moment = require("moment");

// 请求参数
/**
 * 函数：createAdmin
 * 说明：创建管理员账户
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | username | string | 是 | 用户名 |
    | password | string | 是 | 密码 |
    | nickname | string | 否 | 昵称 |
    | role | string | 是 | 角色代码 |
    | email | string | 否 | 邮箱 |
    | phone | string | 否 | 手机号 |
  * 测试数据
    {
        "username": "newadmin",
        "password": "123456",
        "nickname": "新管理员",
        "role": "admin",
        "email": "admin@example.com"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "id": "admin_id"
        }
    }
 */

exports.main = async (event, context) => {
    let username = event.username;
    let password = event.password;
    let nickname = event.nickname;
    let role = event.role;
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
    if (!username || typeof username !== "string") {
        ret.code = 4001;
        ret.msg = "参数[username]错误";
        return ret;
    }

    if (!password || typeof password !== "string" || password.length < 6) {
        ret.code = 4001;
        ret.msg = "参数[password]错误，密码至少6位";
        return ret;
    }

    if (!role || typeof role !== "string") {
        ret.code = 4001;
        ret.msg = "参数[role]错误";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查用户名是否已存在
        const existingAdmin = await db.collection('admin_users')
            .where({ username: username })
            .get();

        if (existingAdmin.length > 0) {
            ret.code = 4003;
            ret.msg = "用户名已存在";
            return ret;
        }

        // 检查角色是否存在
        const roleList = await db.collection('admin_roles')
            .where({ roleCode: role })
            .get();

        if (roleList.length === 0) {
            ret.code = 4004;
            ret.msg = "角色不存在";
            return ret;
        }

        // 密码加密
        const passwordHash = crypto.createHash('md5').update(password).digest('hex');

        // 创建管理员
        const result = await db.collection('admin_users').add({
            data: {
                username: username,
                password: passwordHash,
                nickname: nickname || username,
                role: role,
                email: email || '',
                phone: phone || '',
                status: 'active',
                createTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                updateTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                lastLoginTime: null,
                token: null,
                tokenExpire: null
            }
        });

        ret.msg = "创建成功";
        ret.data = {
            id: result._id
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 