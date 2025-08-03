const cloud = require("@alipay/faas-server-sdk");
const { randomUUID } = require('crypto');
const moment = require("moment");
const { requirePermission, logOperation } = require("./common/auth");

// 原始处理函数
async function appInitHandler(event, context) {
    // app名字
    let appName;
    let platform;
    let force;
    let channelAppId;
    let channelAppKey;

    // 返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
        "data": {}
    };

    // 参数校验 字段存在  为空   类型
    if (event.hasOwnProperty("appName") && (!event.appName || typeof event.appName != "string") || !event.hasOwnProperty("platform")) {
        ret.code = 4001;
        ret.msg = "参数[appName]错误";
        return ret;
    }

    if (event.hasOwnProperty("platform") && (!event.platform || typeof event.platform != "string") || !event.hasOwnProperty("platform")) {
        ret.code = 4001;
        ret.msg = "参数[platform]错误";
        return ret;
    }

    if(event.hasOwnProperty("channelAppId") && (!event.channelAppId || typeof event.channelAppId != "string") || !event.hasOwnProperty("channelAppKey")) {
        ret.code = 4001;
        ret.msg = "参数[channelAppId]错误";
        return ret;
    }

    if(event.hasOwnProperty("channelAppKey") && (!event.channelAppKey || typeof event.channelAppKey != "string") || !event.hasOwnProperty("channelAppKey")) {
        ret.code = 4001;
        ret.msg = "参数[channelAppKey]错误";
        return ret;
    }

    // 必填参数检查
    if (!event.appName || !event.platform || !event.channelAppId || !event.channelAppKey) {
        ret.code = 4001;
        ret.msg = "appName、platform、channelAppId、channelAppKey为必填参数";
        return ret;
    }

    appName = event.appName.trim();
    platform = event.platform.trim();
    channelAppId = event.channelAppId.trim();
    channelAppKey = event.channelAppKey.trim();
    force = event.force || false;

    // 额外安全检查：创建应用需要管理员以上权限
    if (!['super_admin', 'admin'].includes(event.adminInfo.role)) {
        ret.code = 4003;
        ret.msg = "创建应用需要管理员或超级管理员权限";
        return ret;
    }

    const db = cloud.database();
    
    // 创建集合
    try {
        await db.getCollection("app_config");
    } catch (e) {
        if (e.message == "not found collection") {
            await db.createCollection("app_config");
        } else {
            ret.code = 5001;
            ret.msg = e.message;
            return ret;
        }
    }

    let innerAppId = randomUUID();
    
    try {
        // 查询是否存在相同的应用
        let appList = await db.collection(`app_config`).where({
            "platform": platform,
            "channelAppId": channelAppId,
        }).get();
        
        if (appList.length > 0) {
            if (!force) {
                ret.code = 4003;
                ret.msg = "应用已存在，如需更新请设置 force=true";
                ret.data = {
                    existingApp: {
                        appName: appList[0].appName,
                        innerAppId: appList[0].appId,
                        platform: appList[0].platform,
                        channelAppId: appList[0].channelAppId
                    }
                };
                return ret;
            }

            // 强制更新模式
            await db.collection(`app_config`).where({
                "platform": platform,
                "channelAppId": channelAppId,
            }).update({
                data: {
                    "appName": appName,
                    "channelAppKey": channelAppKey,
                    "updateTime": moment().format("YYYY-MM-DD HH:mm:ss"),
                    "updatedBy": event.adminInfo.username
                }
            });

            // 记录更新操作日志
            await logOperation(event.adminInfo, 'UPDATE', 'APP_INIT', {
                action: 'force_update_existing_app',
                appName: appName,
                platform: platform,
                channelAppId: channelAppId,
                innerAppId: appList[0].appId,
                severity: 'HIGH'
            });

            ret.data = {
                "appName": appName,
                "innerAppId": appList[0].appId,
                "action": "updated",
                "message": "应用信息已更新"
            };
            return ret;
        }

        // 插入新应用
        await db.collection(`app_config`).add({
            data: {
                "appId": innerAppId,
                "channelAppId": channelAppId,
                "appName": appName,
                "platform": platform,
                "channelAppKey": channelAppKey,
                "status": "active",
                "description": "",
                "createTime": moment().format("YYYY-MM-DD HH:mm:ss"),
                "createdBy": event.adminInfo.username
            }
        });    

        // 创建用户表
        try {
            let userTableName = `user_${innerAppId}`;
            await db.getCollection(userTableName);
        } catch (e) {
            if (e.message == "not found collection") {
                await db.createCollection(`user_${innerAppId}`);
            }
        }

        // 记录创建操作日志
        await logOperation(event.adminInfo, 'CREATE', 'APP_INIT', {
            action: 'create_new_app',
            appName: appName,
            platform: platform,
            channelAppId: channelAppId,
            innerAppId: innerAppId,
            userTableCreated: true,
            severity: 'HIGH'
        });

        ret.data = {
            "appName": appName,
            "innerAppId": innerAppId,
            "action": "created",
            "message": "应用创建成功"
        };

    } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
    }

    return ret;
}

// 导出带权限校验的函数
const mainFunc = requirePermission(appInitHandler, 'app_manage');
exports.main = mainFunc;