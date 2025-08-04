const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");

/**
 * 权限校验中间件
 * @param {Object} event - 请求事件
 * @param {Array|String} requiredPermissions - 需要的权限（数组或字符串）
 * @returns {Object} - 验证结果 { valid: boolean, adminInfo: Object, error: Object }
 */
async function checkPermission(event, requiredPermissions) {
    const result = {
        valid: false,
        adminInfo: null,
        error: null
    };

    try {
        // 从请求头获取token
        let token = null;
        
        // 尝试从不同位置获取token
        if (event.headers && event.headers.authorization) {
            // 从Authorization头获取 "Bearer token"
            const authHeader = event.headers.authorization;
            if (authHeader.startsWith('Bearer ')) {
                token = authHeader.substring(7);
            }
        }
        
        if (!token) {
            // 直接从参数获取
            token = event.token;
        }

        if (!token) {
            result.error = {
                code: 4001,
                msg: "缺少认证token"
            };
            return result;
        }

        const db = cloud.database();

        // 验证token
        const adminList = await db.collection('admin_users')
            .where({ 
                token: token,
                status: 'active'
            })
            .get();

        if (adminList.length === 0) {
            result.error = {
                code: 4001,
                msg: "无效的token"
            };
            return result;
        }

        const admin = adminList[0];

        // 检查token是否过期
        if (admin.tokenExpire && moment(admin.tokenExpire).isBefore(moment())) {
            result.error = {
                code: 4002,
                msg: "token已过期"
            };
            return result;
        }

        // 获取管理员角色和权限
        const roleList = await db.collection('admin_roles')
            .where({ roleId: admin.roleId })
            .get();

        if (roleList.length === 0) {
            result.error = {
                code: 4003,
                msg: "用户角色不存在"
            };
            return result;
        }

        const role = roleList[0];
        const permissions = role.permissions || [];

        // 检查权限
        if (requiredPermissions) {
            const required = Array.isArray(requiredPermissions) ? requiredPermissions : [requiredPermissions];
            const hasPermission = required.some(perm => permissions.includes(perm) || permissions.includes('*'));
            
            if (!hasPermission) {
                result.error = {
                    code: 4003,
                    msg: "权限不足"
                };
                return result;
            }
        }

        // 验证成功
        result.valid = true;
        result.adminInfo = {
            id: admin._id,
            username: admin.username,
            roleId: admin.roleId,
            roleName: role.roleName,
            permissions: permissions
        };

    } catch (e) {
        result.error = {
            code: 5001,
            msg: e.message
        };
    }

    return result;
}

/**
 * 权限装饰器
 * @param {Function} handler - 原始处理函数
 * @param {Array|String} requiredPermissions - 需要的权限
 * @returns {Function} - 包装后的函数
 */
function requirePermission(handler, requiredPermissions) {
    return async (event, context) => {
        // 权限校验
        const authResult = await checkPermission(event, requiredPermissions);
        
        if (!authResult.valid) {
            return {
                code: authResult.error.code,
                msg: authResult.error.msg,
                timestamp: Date.now(),
                data: null
            };
        }

        // 将管理员信息添加到event中
        event.adminInfo = authResult.adminInfo;

        // 调用原始处理函数
        return await handler(event, context);
    };
}

/**
 * 记录操作日志
 * @param {Object} adminInfo - 管理员信息
 * @param {String} action - 操作类型
 * @param {String} resource - 操作资源
 * @param {Object} details - 操作详情
 */
async function logOperation(adminInfo, action, resource, details = {}) {
    try {
        const db = cloud.database();
        
        await db.collection('admin_operation_logs').add({
            data: {
                adminId: adminInfo.id,
                username: adminInfo.username,
                action: action,
                resource: resource,
                details: details,
                ip: details.ip || 'unknown',
                userAgent: details.userAgent || 'unknown',
                createTime: moment().format("YYYY-MM-DD HH:mm:ss")
            }
        });
    } catch (e) {
        console.error('记录操作日志失败:', e);
    }
}

module.exports = {
    checkPermission,
    requirePermission,
    logOperation
}; 