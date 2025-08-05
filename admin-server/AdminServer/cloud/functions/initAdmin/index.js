const cloud = require("@alipay/faas-server-sdk");
const crypto = require("crypto");
const moment = require("moment");

// 请求参数
/**
 * 函数：initAdmin
 * 说明：初始化管理员系统（创建默认角色和管理员）
 * 权限：无需权限验证（仅限首次初始化）
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | force | boolean | 否 | 强制重新初始化（默认：false） |
 * 
 * 说明：
 * - 自动创建必要的数据库集合：admin_users, admin_roles, admin_operation_logs
 * - 创建默认角色：super_admin, admin, operator, viewer
 * - 创建默认超级管理员账户：用户名admin，密码123456
 * - 如果系统已初始化，需要设置force=true强制重置
 * - 强制模式会清除所有现有管理员和角色数据
 * 
 * 默认角色权限：
 * - super_admin: 所有权限
 * - admin: app_manage, user_manage, leaderboard_manage, stats_view
 * - operator: user_manage, leaderboard_manage, stats_view
 * - viewer: stats_view
 * 
 * 测试数据：
    {
        "force": false
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "初始化完成",
        "timestamp": 1603991234567,
        "data": {
            "createdCollections": 3,
            "createdRoles": 4,
            "createdAdmins": 1,
            "defaultCredentials": {
                "username": "admin",
                "password": "123456",
                "warning": "请立即登录并修改默认密码！"
            }
        }
    }
    
 * 错误码：
 * - 4003: 系统已初始化（非强制模式）
 * - 5001: 服务器内部错误
 */

// 初始化管理员系统（不需要权限校验，但需要安全控制）
const initAdminHandler = async (event, context) => {
    let force = event.force || false;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    const db = cloud.database();

    try {
        // 创建必要的集合（表）
        const requiredCollections = [
            'admin_users',      // 管理员用户表
            'admin_roles',      // 角色表
            'admin_operation_logs'  // 操作日志表
        ];

        let createdCollections = 0;
        
        for (let collectionName of requiredCollections) {
            try {
                await db.getCollection(collectionName);
            } catch (e) {
                if (e.message == "not found collection") {
                    await db.createCollection(collectionName);
                    createdCollections++;
                    console.log(`集合 ${collectionName} 创建成功`);
                } else {
                    console.log(`集合 ${collectionName} 创建失败:`, e.message);
                }
            }
        }

        // 安全检查：如果系统已经初始化，且非强制模式，则拒绝
        if (!force) {
            const existingAdmins = await db.collection('admin_users').count();
            if (existingAdmins.total > 0) {
                ret.code = 4003;
                ret.msg = "系统已初始化，如需重新初始化请设置 force=true";
                return ret;
            }
        }

        // 默认角色配置
        const defaultRoles = [
            {
                roleCode: 'super_admin',
                roleName: '超级管理员',
                permissions: ['admin_manage', 'role_manage', 'app_manage', 'user_manage', 'leaderboard_manage', 'mail_manage', 'stats_view', 'system_config', 'counter_manage'],
                description: '系统最高权限，拥有所有操作权限',
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            },
            {
                roleCode: 'admin',
                roleName: '管理员',
                permissions: ['app_manage', 'user_manage', 'leaderboard_manage', 'mail_manage', 'stats_view', 'counter_manage'],
                description: '管理员权限，可管理应用、用户、排行榜',
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            },
            {
                roleCode: 'operator',
                roleName: '运营人员',
                permissions: ['user_manage', 'leaderboard_manage', 'mail_manage', 'stats_view', 'counter_manage'],
                description: '运营人员权限，可管理用户和排行榜',
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            },
            {
                roleCode: 'viewer',
                roleName: '查看者',
                permissions: ['stats_view'],
                description: '只读权限，仅可查看统计数据',
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            }
        ];

        let createdRoles = 0;
        let createdAdmins = 0;

        // 如果是强制模式，先清理现有数据
        if (force) {
            try {
                await db.collection('admin_roles').where({}).remove();
                await db.collection('admin_users').where({}).remove();
                // 记录强制重置操作
                await db.collection('admin_operation_logs').add({
                    data: {
                        adminId: 'SYSTEM',
                        username: 'SYSTEM',
                        action: 'FORCE_RESET',
                        resource: 'ADMIN_SYSTEM',
                        details: {
                            severity: 'CRITICAL',
                            ip: context.headers ? context.headers['x-forwarded-for'] : 'unknown'
                        },
                        createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                    }
                });
            } catch (e) {
                // 忽略删除错误（表可能不存在）
            }
        }

        // 创建默认角色
        for (let role of defaultRoles) {
            try {
                // 检查角色是否已存在
                const existingRoles = await db.collection('admin_roles')
                    .where({ roleCode: role.roleCode })
                    .get();
                
                if (existingRoles.length === 0) {
                    await db.collection('admin_roles').add({
                        data: role
                    });
                    createdRoles++;
                }
            } catch (e) {
                console.log(`创建角色 ${role.roleCode} 失败:`, e.message);
            }
        }

        // 创建默认超级管理员账户
        const defaultPassword = '123456';
        const passwordHash = crypto.createHash('md5').update(defaultPassword).digest('hex');

        const defaultAdmin = {
            username: 'admin',
            password: passwordHash,
            nickname: '系统管理员',
            role: 'super_admin',
            email: null,
            phone: null,
            status: 'active',
            createTime: moment().format("YYYY-MM-DD HH:mm:ss"),
            createdBy: 'SYSTEM',
            lastLoginTime: null,
            token: null,
            tokenExpire: null
        };

        try {
            // 检查默认管理员是否已存在
            const existingAdmins = await db.collection('admin_users')
                .where({ username: 'admin' })
                .get();
            
            if (existingAdmins.length === 0) {
                await db.collection('admin_users').add({
                    data: defaultAdmin
                });
                createdAdmins++;
            }
        } catch (e) {
            console.log('创建默认管理员失败:', e.message);
        }

        // 记录初始化操作日志
        try {
            await db.collection('admin_operation_logs').add({
                data: {
                    adminId: 'SYSTEM',
                    username: 'SYSTEM',
                    action: 'INIT',
                    resource: 'ADMIN_SYSTEM',
                    details: {
                        force: force,
                        createdRoles: createdRoles,
                        createdAdmins: createdAdmins,
                        severity: 'CRITICAL',
                        ip: context.headers ? context.headers['x-forwarded-for'] : 'unknown'
                    },
                    createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                }
            });
        } catch (e) {
            // 忽略日志记录错误
        }

        ret.msg = "初始化完成";
        ret.data = {
            createdCollections: createdCollections,
            createdRoles: createdRoles,
            createdAdmins: createdAdmins,
            defaultCredentials: {
                username: 'admin',
                password: defaultPassword,
                warning: '请立即登录并修改默认密码！'
            }
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 

exports.main = initAdminHandler;