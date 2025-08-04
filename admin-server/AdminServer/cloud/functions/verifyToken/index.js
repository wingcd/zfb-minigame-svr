const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");

// 请求参数
/**
 * 函数：verifyToken
 * 说明：验证管理员Token有效性
 * 权限：无需权限验证
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | token | string | 是 | 身份验证Token |
 * 
 * 说明：
 * - 验证Token是否有效且未过期
 * - 返回管理员信息和权限列表
 * - 更新最后活跃时间
 * - 自动记录验证日志（低频率）
 * 
 * 测试数据：
    {
        "token": "abc123def456ghi789"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "Token验证成功",
        "timestamp": 1603991234567,
        "data": {
            "valid": true,
            "adminInfo": {
                "id": "admin_id_123456",
                "username": "admin",
                "nickname": "系统管理员",
                "role": "super_admin",
                "roleName": "超级管理员",
                "permissions": ["admin_manage", "role_manage", "app_manage"],
                "email": "admin@example.com",
                "phone": "13800138000",
                "lastLoginTime": "2023-10-01 10:00:00",
                "createTime": "2023-09-01 10:00:00",
                "tokenExpire": "2023-10-08 10:00:00"
            }
        }
    }
    
 * 错误码：
 * - 4001: Token为空、无效或已过期
 * - 5001: 服务器内部错误
 */

// Token验证（不需要权限校验，但需要安全处理）
const verifyTokenHandler = async (event, context) => {
    let token = event.token;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!token || typeof token !== "string") {
        ret.code = 4001;
        ret.msg = "Token不能为空";
        return ret;
    }

    // 获取客户端信息
    const clientIp = context.headers ? context.headers['x-forwarded-for'] : 'unknown';
    const userAgent = context.headers ? context.headers['user-agent'] : 'unknown';

    const db = cloud.database();

    try {
        // 查询token对应的管理员
        const adminList = await db.collection('admin_users')
            .where({ 
                token: token,
                status: 'active'
            })
            .get();

        if (adminList.length === 0) {
            // 记录无效token尝试
            try {
                await db.collection('admin_operation_logs').add({
                    data: {
                        adminId: 'UNKNOWN',
                        username: 'UNKNOWN',
                        action: 'TOKEN_INVALID',
                        resource: 'AUTH',
                        details: {
                            reason: 'token_not_found',
                            ip: clientIp,
                            userAgent: userAgent,
                            severity: 'MEDIUM'
                        },
                        createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                    }
                });
            } catch (e) {
                // 忽略日志记录错误
            }

            ret.code = 4001;
            ret.msg = "无效的Token";
            return ret;
        }

        const admin = adminList[0];

        // 检查token是否过期
        if (admin.tokenExpire) {
            const now = moment();
            const tokenExpire = moment(admin.tokenExpire);
            
            if (now.isAfter(tokenExpire)) {
                // 记录token过期
                try {
                    await db.collection('admin_operation_logs').add({
                        data: {
                            adminId: admin._id,
                            username: admin.username,
                            action: 'TOKEN_EXPIRED',
                            resource: 'AUTH',
                            details: {
                                expiredAt: admin.tokenExpire,
                                ip: clientIp,
                                userAgent: userAgent,
                                severity: 'LOW'
                            },
                            createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                        }
                    });
                } catch (e) {
                    // 忽略日志记录错误
                }

                ret.code = 4001;
                ret.msg = "Token已过期";
                return ret;
            }
        }

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

        // 更新最后活跃时间（可选，用于活跃度统计）
        try {
            await db.collection('admin_users').doc(admin._id)
                .update({
                    data: {
                        lastActiveTime: moment().format("YYYY-MM-DD HH:mm:ss"),
                        lastActiveIp: clientIp
                    }
                });
        } catch (e) {
            // 忽略更新错误
        }

        // 记录Token验证成功（仅在需要时记录，避免日志过多）
        const shouldLog = Math.random() < 0.1; // 10% 概率记录，减少日志量
        if (shouldLog) {
            try {
                await db.collection('admin_operation_logs').add({
                    data: {
                        adminId: admin._id,
                        username: admin.username,
                        action: 'TOKEN_VERIFY',
                        resource: 'AUTH',
                        details: {
                            ip: clientIp,
                            userAgent: userAgent,
                            severity: 'LOW'
                        },
                        createTime: moment().format("YYYY-MM-DD HH:mm:ss")
                    }
                });
            } catch (e) {
                // 忽略日志记录错误
            }
        }

        ret.msg = "Token验证成功";
        ret.data = {
            valid: true,
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
                createTime: admin.createTime,
                tokenExpire: admin.tokenExpire
            }
        };

    } catch (e) {
        // 记录系统错误
        try {
            await db.collection('admin_operation_logs').add({
                data: {
                    adminId: 'SYSTEM',
                    username: 'SYSTEM',
                    action: 'TOKEN_VERIFY_ERROR',
                    resource: 'AUTH',
                    details: {
                        error: e.message,
                        ip: clientIp,
                        userAgent: userAgent,
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

exports.main = verifyTokenHandler;