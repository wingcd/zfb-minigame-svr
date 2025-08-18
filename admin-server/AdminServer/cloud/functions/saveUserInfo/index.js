const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：saveUserInfo
 * 说明：保存玩家基本信息
 * 参数：
   | 参数名 | 类型 | 必选 | 说明 |
   | userInfo | string | 否 | 用户信息 |
   {
       "nickName": "用户昵称",
       "avatarUrl": "头像URL",
       "gender": 1,
       "province": "广东省",
       "city": "深圳市"
   }
 * 测试数据
   {
       "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
       "playerId": "600015",
       "userInfo": {
           "nickName": "用户昵称",
           "avatarUrl": "头像URL",
           "gender": 1,
           "province": "广东省",
           "city": "深圳市"
       }
   }
   
   * 返回结果
   {
       "code": 0,
       "msg": "success",
       "timestamp": 1603991234567,
   }
*/

exports.main = async (event, context) => {

    let appId;
    let playerId;

    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now(),
    }

    var parmErr = common.hash.CheckParams(event);
    if(parmErr) {
        ret.code = 4001;
        ret.msg = "参数错误, error code:" + parmErr;
        return ret;
    }

    try {
        appId = event.appId.trim();  //小程序id
        playerId = event.playerId.trim();  //玩家id
        let commitUserInfo = event.userInfo ? JSON.parse(event.userInfo) : {};

        // 获取 cloud 环境中的 mongoDB 数据库对象
        const db = cloud.database();
        let userTableName = `user_${appId}`;
        const collection = db.collection(userTableName);

        //获取玩家信息
        let queryList = await collection
        .where({
            "playerId": playerId,
        })
        .get();
        
        
        if (queryList.length > 0) {
            // 玩家已存在，更新用户信息
            let user = queryList[0];

            let userInfo = user.userInfo || {};
            let keys = Object.keys(commitUserInfo);
            for (let key of keys) {
                if (commitUserInfo[key] !== undefined && commitUserInfo[key] !== null) {
                    userInfo[key] = commitUserInfo[key];
                }
            }
            
            //更新
            try {
                let now = moment().format("YYYY-MM-DD HH:mm:ss");
                await collection.doc(user._id)
                        .update({
                            data: {
                                "gmtModify": now,
                                "data": user.data,
                                "userInfo": userInfo
                            }
                        });
            } catch (e) {
                ret.code = 5001;
                ret.msg = e.message;
                return ret;
            }
        }else {
            ret.code = 4004;
            ret.msg = "用户不存在";
            return ret;
        }
    } catch (e) {
        console.error(e);
        ret.code = 5001;
        ret.msg = "数据库操作失败: " + e.message;
        return ret;
    }

    return ret;
}; 