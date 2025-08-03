const cloud = require("@alipay/faas-server-sdk");
const crypto = require('crypto');
const moment = require("moment");

// 请求参数
/**
 * 函数：initAdmin
 * 说明：初始化管理员系统（创建默认角色和超级管理员）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | force | boolean | 否 | 是否强制重新初始化 |
    | adminPassword | string | 否 | 超级管理员密码，默认123456 |
  * 测试数据
    {
        "force": true,
        "adminPassword": "admin123"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "message": "初始化完成",
            "adminUsername": "admin",
            "adminPassword": "123456"
        }
    }
 */

exports.main = async (event, context) => {
    let force = event.force || false;
    let adminPassword = event.adminPassword || "123456";

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    const db = cloud.database();

    try {
        // 检查是否已经初始化
        if (!force) {
            const existingAdmin = await db.collection('admin_users').get();
            if (existingAdmin.length > 0) {
                ret.code = 4003;
                ret.msg = "管理员系统已初始化，如需重新初始化请设置force=true";
                return ret;
            }
        }

        // 定义默认角色
        const defaultRoles = [
            {
                roleCode: 'super_admin',
                roleName: '超级管理员',
                description: '拥有所有权限的超级管理员',
                permissions: [
                    'admin_manage',    // 管理员管理
                    'role_manage',     // 角色管理
                    'app_manage',      // 应用管理
                    'user_manage',     // 用户管理
                    'leaderboard_manage', // 排行榜管理
                    'stats_view',      // 统计查看
                    'system_config'    // 系统配置
                ],
                sort: 1,
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            },
            {
                roleCode: 'admin',
                roleName: '管理员',
                description: '普通管理员，拥有大部分管理权限',
                permissions: [
                    'app_manage',
                    'user_manage',
                    'leaderboard_manage',
                    'stats_view'
                ],
                sort: 2,
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            },
            {
                roleCode: 'operator',
                roleName: '运营人员',
                description: '运营人员，拥有查看和基础操作权限',
                permissions: [
                    'user_manage',
                    'leaderboard_manage',
                    'stats_view'
                ],
                sort: 3,
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            },
            {
                roleCode: 'viewer',
                roleName: '查看者',
                description: '只读权限，仅可查看数据',
                permissions: [
                    'stats_view'
                ],
                sort: 4,
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            }
        ];

        // 清空现有角色（如果强制初始化）
        if (force) {
            await db.collection('admin_roles').where({}).remove();
            await db.collection('admin_users').where({}).remove();
        }

        // 创建默认角色
        for (let role of defaultRoles) {
            await db.collection('admin_roles').add({
                data: role
            });
        }

        // 创建超级管理员账户
        const passwordHash = crypto.createHash('md5').update(adminPassword).digest('hex');
        
        await db.collection('admin_users').add({
            data: {
                username: 'admin',
                password: passwordHash,
                nickname: '超级管理员',
                role: 'super_admin',
                email: 'admin@example.com',
                phone: '',
                status: 'active',
                createTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                updateTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                lastLoginTime: null,
                token: null,
                tokenExpire: null
            }
        });

        ret.msg = "初始化完成";
        ret.data = {
            message: "管理员系统初始化完成",
            adminUsername: "admin",
            adminPassword: adminPassword,
            rolesCreated: defaultRoles.length,
            roles: defaultRoles.map(role => ({
                roleCode: role.roleCode,
                roleName: role.roleName,
                permissions: role.permissions
            }))
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 