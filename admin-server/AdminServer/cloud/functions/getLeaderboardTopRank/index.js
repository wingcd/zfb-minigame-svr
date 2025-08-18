const cloud = require("@alipay/faas-server-sdk");
const moment = require("moment");
const common = require("./common");

// 请求参数
/**
 * 函数：getLeaderboardTopRank
 * 说明：获取排行榜top排名
 * 参数：
    | 参数名 | 类型 | 必选 | 说明 |
    | --- | --- | --- | --- |
    | appId | string | 是 | 排行榜id |
    | type | string | 是 | 排行榜类型 |
    | startRank | number | 否 | 起始排名 |
    | count | number | 否 | top数量 |
    | sort | number | 否 | 排序方式 |
  * 测试数据
    {
        "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
        "type": "easy",
        "startRank": 0,
        "count": 10,
        "sort": 1
    }
    
    * 返回结果
    {
        "code": 0,
        "msg": "success",
        "timestamp": 1603991234567,
        "data": {
            "type": "easy",
            "count": 10,
            "list": [
                {
                    "playerId": "player001",
                    "score": 100,
                    "userInfo": {
                        "nickName": "小明",
                        "avatarUrl": "https://xxx.com/xxx.jpg"
                    }
                }
            ]
        }
    }
*/

const getLeaderboardTopRankHandler = async (event, context) => {

  //排行榜id
  let appId;
  //排行榜类型
  let leaderboardType;
  //起始排名 可空
  let startRank = 0;
  //排序方式
  let sort = 1;
  //top数量 可空
  let count = 10;


  //返回结果
  let ret = {
    "code": 0,
    "msg": "success",
    "timestamp": Date.now(),
  }

  var parmErr = common.hash.CheckParams(event);
  if (parmErr) {
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

  if (!event.hasOwnProperty("type") || !event.type || typeof event.type != "string") {
    ret.code = 4001;
    ret.msg = "参数[type]错误"
    return ret;
  }
  if (event.hasOwnProperty("startRank") && !Number.isInteger(event.startRank)) {
    ret.code = 4001;
    ret.msg = "参数[startRank]错误"
    return ret;
  }
  if (event.hasOwnProperty("count") && !Number.isInteger(event.count)) {
    ret.code = 4001;
    ret.msg = "参数[count]错误"
    return ret;
  }
  if (event.hasOwnProperty("sort") && !Number.isInteger(event.sort)) {
    ret.code = 4001;
    ret.msg = "参数[sort]错误"
    return ret;
  }

  //请求参数
  appId = event.appId.trim();
  leaderboardType = event.type.trim();
  startRank = event.hasOwnProperty("startRank") ? event.startRank : startRank;
  count = event.hasOwnProperty("count") ? event.count : count;
  let test = event.hasOwnProperty("test") ? event.test : false;

  //数据库实例
  const db = cloud.database();

  let config;
  if (!event.hasOwnProperty("sort") || typeof event.sort == undefined || event.sort == null) {
    try {
      config = await db.collection('leaderboard_config')
        .where({
          appId: appId,
          leaderboardType: leaderboardType
        })
        .get();
      sort = config[0].sort;
    } catch (e) {
      ret.code = 5001;
      ret.msg = e.message;
      return ret;
    }
  } else {
    sort = event.sort;
    // 仍需获取配置以检查重置
    try {
      config = await db.collection('leaderboard_config')
        .where({
          appId: appId,
          leaderboardType: leaderboardType
        })
        .get();
    } catch (e) {
      ret.code = 5001;
      ret.msg = e.message;
      return ret;
    }
  }

  // 检查是否需要重置排行榜
  if (config && config.length > 0) {
    let leaderboardConfig = config[0];
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
        } catch (e) {
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
          } catch (e) {
            ret.code = 5001;
            ret.msg = "更新重置时间失败: " + e.message;
            return ret;
          }
        }
      }
    }
  }


  try {
    let whereInfo = {
      appId: appId,
      leaderboardType: leaderboardType,
      hasUserInfo: 1
    }
    if (!test) {
      whereInfo.test = 0;
    }
    let topList = await db.collection('leaderboard_score')
      .where(whereInfo)
      .orderBy("score", sort == 1 ? cloud.Sort.DESC : cloud.Sort.ASC)
      .skip(startRank)
      .limit(count)
      .get()

    // 获取用户信息
    let userInfoMap = {};
    if (topList.length > 0) {
      const ids = [...new Set(topList.map(score => score.playerId))];
      const userTableName = `user_${appId}`;

      if (ids.length > 0) {
        try {
          const userList = await db.collection(userTableName)
            .where({
              playerId: db.command.in(ids)
            })
            .get();

          userList.forEach(user => {
            if(userInfoMap[user.playerId] && !user.userInfo) {
              console.log("userInfoMap[user.playerId]:", user.playerId, userInfoMap[user.playerId]);
              return;
            }

            let userInfo = user.userInfo || {};
            let nickName = userInfo.nickName || '';
            let avatarUrl = userInfo.avatarUrl || '';
            userInfoMap[user.playerId] = {
              nickName: nickName,
              avatarUrl: avatarUrl
            };
          });
        } catch (e) {
          // 用户表查询失败，忽略用户信息
          console.log('User info query failed:', e.message);
        }
      }
    }

    // 删除不需要的字段并添加用户信息
    topList = topList.map(item => {
      const result = {
        playerId: item.playerId,
        score: item.score,
      };

      // 添加用户信息
      result.userInfo = userInfoMap[item.playerId] || {};
      return result;
    });

    ret.data = {
      "type": leaderboardType,
      "count": topList.length,
      "list": topList
    }
  } catch (e) {
    ret.code = 5001;
    ret.msg = e.message;
  }

  return ret;
};

exports.main = getLeaderboardTopRankHandler;