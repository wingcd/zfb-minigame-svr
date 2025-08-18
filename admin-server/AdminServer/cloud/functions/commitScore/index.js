const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：commitScore
 * 说明：提交分数
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | appId |
    | playerId | string | 是 | 玩家id |
    | type | string | 是 | 排行榜类型 |
    | id | string | 否 | 具体记录id，用于更新 |
    | score | number | 是 | 当前分数 |

  * 测试数据
    {
        "playerId": "player001",
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "type": "easy",
        "score": 100,
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
    }
 */

exports.main = async (event, context) => {

  //玩家id
  let playerId;
  //appid
  let appId;
  //排行榜类型
  let leaderboardType;
  //具体记录id
  let id;
  //当前分数
  let currentScore;

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

  //参数校验 字段存在  为空   类型
  if (!event.hasOwnProperty("appId") || !event.appId || typeof event.appId != "string") {
    ret.code = 4001;
    ret.msg = "参数[appId]错误"
    return ret;
  }

  if (!event.hasOwnProperty("playerId") || !event.playerId || typeof event.playerId != "string") {
    ret.code = 4001;
    ret.msg = "参数[playerId]错误"
    return ret;
  }

  if (!event.hasOwnProperty("type") || !event.type || typeof event.type != "string") {
    ret.code = 4001;
    ret.msg = "参数[type]错误"
    return ret;
  }
  if (event.hasOwnProperty("id") && (!event.id || typeof event.id != "string")) {
    ret.code = 4001;
    ret.msg = "参数[id]错误"
    return ret;
  }
  if (event.hasOwnProperty("score") && !Number.isInteger(event.score)) {
    ret.code = 4001;
    ret.msg = "参数[currentScore]错误"
    return ret;
  }


  //请求参数
  playerId = event.playerId.trim();  //玩家id

  appId = event.appId.trim();//app id

  leaderboardType = event.type.trim();//排行榜具体类型

  id = event.hasOwnProperty("id") ? event.id.trim() : "";//分数记录id

  currentScore = event.score; //分数

  // 获取 cloud 环境中的 mongoDB 数据库对象
  const db = cloud.database();

  // 获取用户数据
  let userTableName = `user_${appId}`;
  let collection = db.collection(userTableName);

  // 查询用户是否存在
  let userList = await collection
      .where({ playerId: playerId })
      .get();

  if (userList.length === 0) {
      ret.code = 4004;
      ret.msg = "用户不存在";
      return ret;
  }

  let user = userList[0];

  //获取更新策略
  let queryList = await db.collection('leaderboard_config')
    .where({
      appId: appId,
      leaderboardType: leaderboardType,
    })
    .get();
  let updateStrategy = queryList[0].updateStrategy
  if (updateStrategy != 0 && updateStrategy != 1 && updateStrategy != 2) {
    ret.code = 6001;
    ret.msg = "更新策略异常";
    return ret;
  }

  // 检查是否需要重置排行榜
  let leaderboardConfig = queryList[0];
  let now = moment();
  
  if (leaderboardConfig.resetTime && leaderboardConfig.resetType !== 'permanent') {
    let resetTime = moment(leaderboardConfig.resetTime);
    
    if (now.isAfter(resetTime)) {
      // 清空排行榜数据
      try {
        await db.collection('leaderboard_score')
          .where({
            appId: appId,
            leaderboardType: leaderboardType
          })
          .remove();
      } catch(e) {
        ret.code = 5001;
        ret.msg = "重置排行榜失败: " + e.message;
        return ret;
      }
      
      // 重新计算下次重置时间
      let newResetTime = null;
      if (leaderboardConfig.resetType) {
        switch (leaderboardConfig.resetType) {
          case "daily":
            newResetTime = moment().startOf('day').add(1, 'day').format("YYYY-MM-DD HH:mm:ss");
            break;
          case "weekly":
            newResetTime = moment().startOf('week').add(1, 'week').format("YYYY-MM-DD HH:mm:ss");
            break;
          case "monthly":
            newResetTime = moment().startOf('month').add(1, 'month').format("YYYY-MM-DD HH:mm:ss");
            break;
          case "custom":
            if (leaderboardConfig.resetValue) {
              newResetTime = moment().add(leaderboardConfig.resetValue, 'hours').format("YYYY-MM-DD HH:mm:ss");
            }
            break;
        }
      }
      
      // 更新配置中的重置时间
      if (newResetTime) {
        try {
          await db.collection('leaderboard_config')
            .doc(leaderboardConfig._id)
            .update({
              data: {
                resetTime: newResetTime,
                gmtModify: now.format("YYYY-MM-DD HH:mm:ss")
              }
            });
        } catch(e) {
          ret.code = 5001;
          ret.msg = "更新重置时间失败: " + e.message;
          return ret;
        }
      }
    }
  }

  if(!id) {
    id = "";
    let recordList = await db.collection('leaderboard_score').where({
      appId: appId,
      playerId: playerId,
      leaderboardType: leaderboardType
    }).get();
    if(recordList.length > 0) {
      id = recordList[0]._id;
    }
  }

  //更新 or 插入 
  if (id.length === 0) {
    //插入
    try {
      let now = moment().format("YYYY-MM-DD HH:mm:ss");
      await db.collection('leaderboard_score').add({
        data: {
          "appId": appId,
          "playerId": playerId,
          "test": user.test,
          "leaderboardType": leaderboardType,
          "score": currentScore,
          "hasUserInfo": user.userInfo ? 1 : 0,
          "gmtCreate": now,
          "gmtModify": now,
        }
      });
    } catch (e) {
      ret.code = 5001;
      ret.msg = e.message;
      return ret;
    }
  } else {
    //更新
    let oldRecord;
    try {
      oldRecord = await db.collection('leaderboard_score').doc(id).get();
    } catch (e) {
      ret.code = 5001;
      ret.msg = e.message;
      return ret;
    }
    let oldScore = oldRecord.score;
    let flag = false;
    let score = currentScore;
    //历史最高值
    if (updateStrategy == 0) {
      flag = oldScore <= currentScore ? true : false;
      //最近记录
    } else if (updateStrategy == 1) {
      flag = true;
    } else {
      //历史总和
      flag = true;
      score += oldScore;
    }
    if (flag) {
      try {
        let now = moment().format("YYYY-MM-DD HH:mm:ss");
        await db.collection('leaderboard_score')
          .doc(id)
          .update({
            data: {
              "test": user.test,
              "score": score,
              "hasUserInfo": user.userInfo ? 1 : 0,
              "gmtModify": now,
            }
          });
      } catch (e) {
        ret.code = 5001;
        ret.msg = e.message;
        return ret;
      }
    }
  }

  return ret;
};