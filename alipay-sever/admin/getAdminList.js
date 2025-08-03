const cloud = require("@alipay/faas-server-sdk");

// 请求参数
/**
 * 函数：getAdminList
 * 说明：获取管理员列表
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |
    | username | string | 否 | 用户名搜索 |
    | role | string | 否 | 角色筛选 |
    | status | string | 否 | 状态筛选 |
  * 测试数据
    {
        "page": 1,
        "pageSize": 20,
        "username": "admin"
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "list": [...],
            "total": 10,
            "page": 1,
            "pageSize": 20
        }
    }
 */

exports.main = async (event, context) => {
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

    // 限制最大每页数量
    if (pageSize > 100) {
        pageSize = 100;
    }

    const db = cloud.database();

    try {
        // 构建查询条件
        let whereCondition = {};
        
        if (username) {
            whereCondition.username = new RegExp(username, 'i'); // 模糊搜索
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

        // 处理返回数据，移除敏感信息
        adminList = adminList.map(admin => {
            delete admin.password;
            delete admin.token;
            return admin;
        });

        // 获取角色信息
        const roleList = await db.collection('admin_roles').get();
        const roleMap = {};
        roleList.forEach(role => {
            roleMap[role.roleCode] = role;
        });

        // 为管理员添加角色信息
        adminList = adminList.map(admin => {
            admin.roleInfo = roleMap[admin.role] || null;
            return admin;
        });

        ret.data.list = adminList;
        ret.data.total = total;

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 