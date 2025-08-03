const cloud = require("@alipay/faas-server-sdk");
const crypto = require('crypto');
const moment = require("moment");

// 请求参数
/**
 * 函数：adminLogin
 * 说明：管理员登录
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | username | string | 是 | 用户名 |
    | password | string | 是 | 密码 |
  * 测试数据
    {
        "username": "admin",
        "password": "123456"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "token": "jwt_token_here",
            "adminInfo": {
                "id": "admin_id",
                "username": "admin",
                "role": "super_admin",
                "permissions": ["user_manage", "app_manage", "leaderboard_manage"]
            }
        }
    }
 */

exports.main = async (event, context) => {
    let username = event.username;
    let password = event.password;

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

    if (!password || typeof password !== "string") {
        ret.code = 4001;
        ret.msg = "参数[password]错误";
        return ret;
    }

    const db = cloud.database();

    try {
        // 密码加密（使用MD5，实际项目建议使用bcrypt）
        const passwordHash = crypto.createHash('md5').update(password).digest('hex');

        // 查询管理员
        const adminList = await db.collection('admin_users')
            .where({ 
                username: username,
                password: passwordHash,
                status: 'active'
            })
            .get();

        if (adminList.length === 0) {
            ret.code = 4001;
            ret.msg = "用户名或密码错误";
            return ret;
        }

        const admin = adminList[0];

        // 生成token（简化版，实际项目建议使用JWT）
        const token = crypto.randomBytes(32).toString('hex');
        const tokenExpire = moment().add(7, 'days').format("YYYY-MM-DD HH:mm:ss");

        // 更新管理员登录信息
        await db.collection('admin_users')
            .doc(admin._id)
            .update({
                data: {
                    lastLoginTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                    token: token,
                    tokenExpire: tokenExpire
                }
            });

        // 获取角色权限
        const roleList = await db.collection('admin_roles')
            .where({ roleCode: admin.role })
            .get();
        
        let permissions = [];
        if (roleList.length > 0) {
            permissions = roleList[0].permissions || [];
        }

        ret.data = {
            token: token,
            adminInfo: {
                id: admin._id,
                username: admin.username,
                nickname: admin.nickname || admin.username,
                role: admin.role,
                permissions: permissions,
                lastLoginTime: admin.lastLoginTime
            }
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 