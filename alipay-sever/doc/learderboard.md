### 排行榜
* commitScore

    * 说明：提交分数
    * 参数：
        | 参数名 | 类型 | 必选 | 说明 |
        | --- | --- | --- | --- |
        | playerId | string | 是 | 玩家id |
        | appId | string | 是 | 排行榜id |
        | leaderboardType | string | 是 | 排行榜类型 |
        | id | string | 否 | 具体记录id |
        | currentScore | number | 否 | 当前分数 |
        | playerInfo | object | 否 | 玩家信息 |

    * 测试数据
    ``` json
        {
            "playerId": "player001",
            "appId": "39788cbe-564c-4b2a-8b74-cc5e0915bcda",
            "leaderboardType": "easy",
            "id": "5f4d4d7b9e3f7c0001e1f3c6",
            "currentScore": 100,
            "playerInfo": {
                "name": "小明",
                "avatar": "https://xxx.com/xxx.jpg"
            }
        }
    ```
        
        * 返回结果
    ``` json
        {
            "res": {
                "code": 200,
                "msg": "提交成功"
            }
        }
    ```
    */