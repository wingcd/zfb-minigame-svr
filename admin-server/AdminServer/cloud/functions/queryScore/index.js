const cloud = require("@alipay/faas-server-sdk");
const common = require("./common");

// 请求参数
/**
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 排行榜id |
    | playerId | string | 是 | 玩家id |
    | leaderboardType | string | 是 | 排行榜类型 |
 */
// 测试数据
/**
    {
      "playerId": "player001",
      "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
      "leaderboardType": "easy"
    }
 */

exports.main = async (event, context) => {
  //请求参数
  //玩家id
  let playerId;
  //排行榜id
  let appId;
  //排行榜类型
  let leaderboardType;

  //返回结果
  let ret = {
    "code": 0,
    "msg": "success",
    "timestamp": Date.now(),
    "data" : {}
  }

  var parmErr = common.hash.CheckParams(event);
  if(parmErr) {
      ret.code = 4001;
      ret.msg = "参数错误, error code:" + parmErr;
      return ret;
  }

  //参数校验 字段存在  为空   类型
  if(!event.hasOwnProperty("appId") || !event.appId || typeof event.appId != "string") {
    ret.code = 4001;
    ret.msg = "参数[appId]错误"
    return ret;
  }

  if(!event.hasOwnProperty("playerId") || !event.playerId || typeof event.playerId != "string") {
    ret.code = 4001;
    ret.msg = "参数[playerId]错误"
    return ret;
  }

  if(!event.hasOwnProperty("leaderboardType") || !event.leaderboardType || typeof event.leaderboardType != "string") {
    ret.code = 4001;
    ret.msg = "参数[leaderboardType]错误"
    return ret;
  }

  playerId = event.playerId.trim();
  appId = event.appId.trim();
  leaderboardType = event.leaderboardType.trim();

  //数据库实例
  const db = cloud.database();

   try {
    let recordList = await db.collection(`leaderboard_score`)
      .where({
        appId : appId,
        playerId : playerId,
        leaderboardType : leaderboardType
      })
      .get();
      if(recordList.length === 0) {
        ret.msg = "未查询到您的数据";
        return ret;
      }
      let data = recordList[0];
      delete data._id;
      delete data._openid;
      delete data.appId;
      delete data.playerId;
      delete data.leaderboardType;
      delete data.gmtCreate;
      delete data.gmtModify;
      ret.data = data;
   } catch (e) {
      ret.code = 5001;
      ret.msg = e.message;
      return ret;
   }
 
  return ret;
};