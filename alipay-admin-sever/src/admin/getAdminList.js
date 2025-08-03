const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：getAdminList
 * 说明：获取管理员列表（支持分页和筛选）
 * 权限：需要 admin_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | page | number | 否 | 页码（默认：1） |
    | pageSize | number | 否 | 每页数量（默认：20，最大：100） |
    | username | string | 否 | 用户名筛选（模糊搜索） |
    | role | string | 否 | 角色筛选 |
    | status | string | 否 | 状态筛选（active/inactive） |
 * 
 * 测试数据：
    {
        "page": 1,
        "pageSize": 10,
        "username": "admin",
        "role": "admin",
        "status": "active"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "list": [
                {
                    "_id": "admin_id_123456",
                    "username": "admin",
                    "nickname": "系统管理员",
                    "role": "super_admin",
                    "roleName": "超级管理员",
                    "rolePermissions": ["admin_manage", "role_manage"],
                    "email": "admin@example.com",
                    "phone": "13800138000",
                    "status": "active",
                    "createTime": "2023-10-01 10:00:00",
                    "lastLoginTime": "2023-10-01 15:30:00"
                }
            ],
            "total": 1,
            "page": 1,
            "pageSize": 10
        }
    }
    
 * 错误码：
 * - 4003: 权限不足
 * - 5001: 服务器内部错误
 */

// 原始处理函数
async function getAdminListHandler(event, context) {
    let page = event.page || 1;
    let pageSize = event.pageSize || 20;
    let username = event.username;
    let role = event.role;
    let status = event.status;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {
            list: [],
            total: 0,
            page: page,
            pageSize: pageSize
        }
    };

    // 参数校验
    if (pageSize > 100) {
        pageSize = 100; // 限制最大每页数量
    }

    const db = cloud.database();

    try {
        // 构建查询条件
        let whereCondition = {};
        
        if (username) {
            whereCondition.username = new RegExp(username, 'i'); // 模糊搜索，忽略大小写
        }
        
        if (role) {
            whereCondition.role = role;
        }
        
        if (status) {
            whereCondition.status = status;
        }

        // 查询总数
        const countResult = await db.collection('admin_users').where(whereCondition).count();
        const total = countResult.total;

        // 分页查询
        const skip = (page - 1) * pageSize;
        let adminList = await db.collection('admin_users')
            .where(whereCondition)
            .orderBy('createTime', 'desc')
            .skip(skip)
            .limit(pageSize)
            .get();

        // 获取角色信息并处理敏感数据
        for (let admin of adminList) {
            // 删除敏感信息
            delete admin.password;
            delete admin.token;
            
            // 获取角色信息
            try {
                const roleList = await db.collection('admin_roles')
                    .where({ roleCode: admin.role })
                    .get();
                
                if (roleList.length > 0) {
                    admin.roleName = roleList[0].roleName;
                    admin.rolePermissions = roleList[0].permissions;
                } else {
                    admin.roleName = admin.role;
                    admin.rolePermissions = [];
                }
            } catch (e) {
                admin.roleName = admin.role;
                admin.rolePermissions = [];
            }
        }

        ret.data.list = adminList;
        ret.data.total = total;

        // 记录操作日志
        await logOperation(event.adminInfo, 'VIEW', 'ADMIN_LIST', {
            searchCondition: whereCondition,
            resultCount: total,
            currentAdmin: event.adminInfo.username
        });

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(getAdminListHandler, 'admin_manage');
exports.main = mainFunc;