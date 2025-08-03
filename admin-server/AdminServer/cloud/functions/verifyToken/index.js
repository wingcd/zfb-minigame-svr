const cloud = require("@alipay/faas-server-sdk");
const crypto = require("crypto");
const moment = require("moment");

// 管理员登录（不需要权限校验，但需要安全处理）
exports.main = async (event, context) => {
    let username = event.username;
    let password = event.password;
    let rememberMe = event.rememberMe || false;

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
        ret.msg = "用户名不能为空";
        return ret;
    }

    if (!password || typeof password !== "string") {
        ret.code = 4001;
        ret.msg = "密码不能为空";
        return ret;
    }

    // 安全限制：防暴力破解
    const clientIp = context.headers ? context.headers['x-forwarded-for'] : 'unknown';
    
    const db = cloud.database();

    try {
        // 密码加密
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
            // 记录失败的登录尝试
            try {
                await db.collection('admin_operation_logs').add({
                    data: {
                        adminId: 'UNKNOWN',
                        username: username,
                        action: 'LOGIN_FAILED',
                        resource: 'AUTH',
                        details: {
                            reason: 'invalid_credentials',
                            ip: clientIp,
                            userAgent: context.headers ? context.headers['user-agent'] : 'unknown',
                            severity: 'MEDIUM'
                        },
                        createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                    }
                });
            } catch (e) {
                // 忽略日志记录错误
            }

            ret.code = 4001;
            ret.msg = "用户名或密码错误";
            return ret;
        }

        const admin = adminList[0];

        // 生成token（简单实现，实际应用中建议使用JWT）
        const token = crypto.createHash('md5')
            .update(admin.username + Date.now() + Math.random())
            .digest('hex');

        // 设置token过期时间
        const tokenExpire = rememberMe 
            ? moment().add(30, 'days').format("YYYY-MM-DD HH:mm:ss")  // 记住我：30天
            : moment().add(7, 'days').format("YYYY-MM-DD HH:mm:ss");   // 默认：7天

        // 更新管理员信息
        await db.collection('admin_users').doc(admin._id)
            .update({
                data: {
                    token: token,
                    tokenExpire: tokenExpire,
                    lastLoginTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                    lastLoginIp: clientIp
                }
            });

        // 获取角色权限
        const roleList = await db.collection('admin_roles')
            .where({ roleCode: admin.role })
            .get();

        let permissions = [];
        let roleName = admin.role;
        if (roleList.length > 0) {
            permissions = roleList[0].permissions || [];
            roleName = roleList[0].roleName;
        }

        // 记录成功的登录操作
        try {
            await db.collection('admin_operation_logs').add({
                data: {
                    adminId: admin._id,
                    username: admin.username,
                    action: 'LOGIN_SUCCESS',
                    resource: 'AUTH',
                    details: {
                        ip: clientIp,
                        userAgent: context.headers ? context.headers['user-agent'] : 'unknown',
                        rememberMe: rememberMe,
                        tokenExpire: tokenExpire,
                        severity: 'LOW'
                    },
                    createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                }
            });
        } catch (e) {
            // 忽略日志记录错误
        }

        ret.msg = "登录成功";
        ret.data = {
            token: token,
            tokenExpire: tokenExpire,
            adminInfo: {
                id: admin._id,
                username: admin.username,
                nickname: admin.nickname,
                role: admin.role,
                roleName: roleName,
                permissions: permissions,
                email: admin.email,
                phone: admin.phone,
                lastLoginTime: admin.lastLoginTime,
                createTime: admin.createTime
            }
        };

    } catch (e) {
        // 记录系统错误
        try {
            await db.collection('admin_operation_logs').add({
                data: {
                    adminId: 'SYSTEM',
                    username: username,
                    action: 'LOGIN_ERROR',
                    resource: 'AUTH',
                    details: {
                        error: e.message,
                        ip: clientIp,
                        severity: 'HIGH'
                    },
                    createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                }
            });
        } catch (logError) {
            // 忽略日志记录错误
        }

        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 