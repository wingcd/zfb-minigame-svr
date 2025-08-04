const cloud = require("@alipay/faas-server-sdk");
const { requirePermission, logOperation } = require("./common/auth");

// 请求参数
/**
 * 函数：getRoleList
 * 说明：获取角色列表（支持分页和筛选，包含管理员数量统计）
 * 权限：需要 role_manage 权限
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | page | number | 否 | 页码（默认：1） |
    | pageSize | number | 否 | 每页数量（默认：20，最大：100） |
    | roleName | string | 否 | 角色名称筛选（模糊搜索） |
 * 
 * 测试数据：
    {
        "page": 1,
        "pageSize": 10,
        "roleName": "管理员"
    }
    
 * 返回结果：
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "list": [
                {
                    "_id": "role_id_123456",
                    "roleCode": "super_admin",
                    "roleName": "超级管理员",
                    "description": "系统最高权限，拥有所有操作权限",
                    "permissions": ["admin_manage", "role_manage", "app_manage"],
                    "createTime": "2023-10-01 10:00:00",
                    "adminCount": 2
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
async function getRoleListHandler(event, context) {
    let page = event.page || 1;
    let pageSize = event.pageSize || 20;
    let roleName = event.roleName;

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
        
        if (roleName) {
            whereCondition.roleName = new RegExp(roleName, 'i'); // 模糊搜索，忽略大小写
        }

        // 查询总数
        const countResult = await db.collection('admin_roles').where(whereCondition).count();
        const total = countResult.total;

        // 分页查询
        const skip = (page - 1) * pageSize;
        let roleList = await db.collection('admin_roles')
            .where(whereCondition)
            .orderBy('createTime', 'desc')
            .skip(skip)
            .limit(pageSize)
            .get();

        // 为每个角色添加管理员数量统计
        for (let role of roleList) {
            try {
                const adminCount = await db.collection('admin_users')
                    .where({ role: role.roleCode })
                    .count();
                role.adminCount = adminCount.total;
            } catch (e) {
                role.adminCount = 0;
            }
        }

        ret.data.list = roleList;
        ret.data.total = total;

        // 记录操作日志
        await logOperation(event.adminInfo, 'VIEW', 'ROLE_LIST', {
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
const mainFunc = requirePermission(getRoleListHandler, 'role_manage');
exports.main = mainFunc;