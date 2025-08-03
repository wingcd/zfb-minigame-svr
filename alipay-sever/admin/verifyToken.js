const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");

// 请求参数
/**
 * 函数：verifyToken
 * 说明：验证管理员token
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | token | string | 是 | 登录token |
  * 测试数据
    {
        "token": "admin_token_here"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "valid": true,
            "adminInfo": {
                "id": "admin_id",
                "username": "admin",
                "role": "super_admin",
                "permissions": ["user_manage", "app_manage"]
            }
        }
    }
 */

exports.main = async (event, context) => {
    let token = event.token;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {
            valid: false
        }
    };

    // 参数校验
    if (!token || typeof token !== "string") {
        ret.code = 4001;
        ret.msg = "参数[token]错误";
        return ret;
    }

    const db = cloud.database();

    try {
        // 查询token
        const adminList = await db.collection('admin_users')
            .where({ 
                token: token,
                status: 'active'
            })
            .get();

        if (adminList.length === 0) {
            ret.code = 4001;
            ret.msg = "无效的token";
            return ret;
        }

        const admin = adminList[0];

        // 检查token是否过期
        const now = moment();
        const tokenExpire = moment(admin.tokenExpire);
        
        if (now.isAfter(tokenExpire)) {
            ret.code = 4001;
            ret.msg = "token已过期";
            return ret;
        }

        // 获取角色权限
        const roleList = await db.collection('admin_roles')
            .where({ roleCode: admin.role })
            .get();
        
        let permissions = [];
        if (roleList.length > 0) {
            permissions = roleList[0].permissions || [];
        }

        ret.data = {
            valid: true,
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