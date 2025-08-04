const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

// 原始处理函数
async function updateAppHandler(event, context) {
    let appId = event.appId;
    let appName = event.appName;
    let description = event.description;
    let status = event.status;
    let channelAppKey = event.channelAppKey;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验
    if (!appId || typeof appId !== "string") {
        ret.code = 4001;
        ret.msg = "参数[appId]错误";
        return ret;
    }

    const db = cloud.database();

    try {
        // 检查应用是否存在
        const appList = await db.collection('app_config')
            .where({ appId: appId })
            .get();

        if (appList.length === 0) {
            ret.code = 4004;
            ret.msg = "应用不存在";
            return ret;
        }

        const oldAppData = appList[0];

        // 构建更新数据
        let updateData = {
            updateTime: moment().format("YYYY-MM-DD HH:mm:ss")
        };

        if (appName !== undefined) {
            updateData.appName = appName;
        }

        if (description !== undefined) {
            updateData.description = description;
        }

        if (status !== undefined) {
            updateData.status = status;
        }

        if (channelAppKey !== undefined) {
            updateData.channelAppKey = channelAppKey;
        }

        // 更新应用信息
        await db.collection('app_config')
            .where({ appId: appId })
            .update({
                data: updateData
            });

        // 记录操作日志
        await logOperation(event.adminInfo, 'UPDATE', 'APP', {
            appId: appId,
            appName: oldAppData.appName,
            changes: updateData
        });

        ret.msg = "更新成功";

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(updateAppHandler, 'app_manage');
exports.main = mainFunc;