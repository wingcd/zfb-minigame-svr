const cloud = require("@alipay/faas-server-sdk");

// 请求参数
/**
 * 函数：getRoleList
 * 说明：获取角色列表
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | page | number | 否 | 页码，默认1 |
    | pageSize | number | 否 | 每页数量，默认20 |
    | roleName | string | 否 | 角色名称搜索 |
  * 测试数据
    {
        "page": 1,
        "pageSize": 20
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "list": [...],
            "total": 5,
            "page": 1,
            "pageSize": 20
        }
    }
 */

exports.main = async (event, context) => {
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

    // 限制最大每页数量
    if (pageSize > 100) {
        pageSize = 100;
    }

    const db = cloud.database();

    try {
        // 构建查询条件
        let whereCondition = {};
        
        if (roleName) {
            whereCondition.roleName = new RegExp(roleName, 'i'); // 模糊搜索
        }

        // 查询总数
        const countResult = await db.collection('admin_roles').where(whereCondition).count();
        const total = countResult.total;

        // 分页查询
        const skip = (page - 1) * pageSize;
        let roleList = await db.collection('admin_roles')
            .where(whereCondition)
            .orderBy('sort', 'asc')
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

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}; 