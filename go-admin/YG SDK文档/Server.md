# YallaGame SDK服务端对接文档

## 概要
**YallaGame SDK（YG SDK）提供了登录、支付、推送等服务。为了快速接入SDK用于生产，请按照如下所示内容申请和提供相关必要的配置和接口。**

## HOST Domain
本文中使用的Host Domain如下

| Env  | domain                           |
|------|----------------------------------|
| test | https://sdkapitest.yallagame.com | 
| pro  | https://sdkapi.yallagame.com     |

推送域名

| Env  | domain                           |
|------|----------------------------------|
| test | https://sdklogapitest.yallagame.com | 
| pro  | https://sdklogapi.yallagame.com     |

## 配置
请从YG SDK相关对接人员获取如下配置参数

| Parameter Name | Description | Example             |
|----------------|-------------|---------------------|
| gameId         | 游戏id       | 101                |
| gameAppId      | 游戏应用id   | 202505162345        |
| gameAppSecret  | 游戏应用密钥  | g5xvf79ko8rfg2bn96u |

## 互认参数sign
在YG SDK接口交互中必须在请求体里添加sign参数来保证请求的可靠和真实性。
sign的生成方式：多个参数使用'|'隔开，使用MD5小写加密两次获取。

| sign |                                       |
|------|---------------------------------------|
| 要求参数 | userId&#124;roleId&#124;timeSpan    |
| 参数排列 | 10001&#124;10001&#124;1747643773000 |
| 生成结果 | c30da41ba017f4b48926073df13bf3a2    |

## 1.YG SDK服务接口

### 1.1订单信息查询接口
Desc : 根据YG SDK订单号获取订单信息(使用sdk订单号查询，二次确认是否真实支付以及金额)<br>
URI : /v1/server/orderInfo<br>
Method : POST<br>
Request Body : Form

| Element Name | Data Type | Required  | Description                                             |
|----------------|-----------|-----------|---------------------------------------------------------|
| payNo          | string    | true      | SDK订单号                                                  |
| gameId         | int       | true      | 游戏应用id                                                  |
| gameAppId      | string    | true      | 游戏应用密钥                                                  |
| sign           | string    | true      | 验签(payNo&#124;gameId&#124;gameAppId&#124;gameAppSecret) |

Response Body : JSON格式

| Element Name | Data Type | Description |
|--------------|-----------|-------------|
| cpNo         | string    | 游戏订单号    |
| payNo        | string    | SDK订单号    |
| payState     | int       | 支付状态      |
| pointId      | int       | 消费点id     |
| pointName    | string    | 消费点名称    |
| amount       | double    | 金额         |

Example :
```json
{
  "ResultCode": 1, 
  "ResultMsg": "成功",
    "Data": {
       "cpNo": "1610593850713", 
       "payNo": "2101140510519491025316244708208",
	   "payState": 0,
	   "pointId": 1,
	   "pointName": "充值(19.99)",
	   "amount": 19.99
  }
}
```

### 1.2用户登录状态监测
Desc : 查询用户当前的登录状态<br>
URI : /v2/server/userInfo<br>
Method : POST<br>
Request Body : Form

| Element Name | Data Type | Required  | Description                                       |
|--------------|-----------|-----------|---------------------------------------------------|
| accesstoken  | string    | true      | 访问密钥                                              |
| gameAppId    | string    | true      | 游戏应用id                                            |
| sign         | string    | true      | 验签(gameAppId&#124;gameAppSecret&#124;accessToken) |

Response Body : JSON格式

| Element Name | Data Type | Description |
|--------------|-----------|-------------|
| openUserId         | string    | sdk用户id     |

Example :
```json
{
  "resultCode": 1,
  "resultMsg": "成功",
  "data": {
    "openUserId": "10001"
  }
}
```
使用方法：客户端SDK登录成功后会返回SDK ID和Token，然后按照流程客户端应拿SDK ID和Token请求游戏服务端生成或者获取角色信息，在此之前服务端需要验证客户端传过来的token是否有效。
注意！当SDK返回失败（过期或错误token）的情况下，游戏服务器需要告知游戏客户端token已经失效；游戏客户端需要重新调用sdk暴露的登陆接口（YG Sdk内部有刷新token机制，如果刷新token也失效了，那么会弹出登陆界面）

### 1.3角色信息查询
Desc : 查询当前用户的角色相关信息<br>
URI : /v1/server/roleInfo<br>
Method : POST<br>
Request Body : Form

| Element Name | Data Type | Required  | Description                                              |
|--------------|-----------|-----------|----------------------------------------------------------|
| gameId       | int       | true      | 游戏id                                                     |
| gameAppId    | string    | true      | 游戏应用id                                                   |
| roleId       | long      | true      | 角色id                                                     |
| sign         | string    | true      | 验签(roleId&#124;gameId&#124;gameAppId&#124;gameAppSecret) |

Response Body : JSON格式

| Element Name  | Data Type | Description       |
|---------------|-----------|-------------------|
| userId        | long      | sdk用户Id           |
| gameServerId  | int       | 区服id              |
| roleId        | long      | 角色id              |
| regIp         | string    | 账号注册ip(可视为角色注册ip) |
| regCountryCode | string    | 账号注册国家(可视为角色注册国家) |
| socMd5Id      | string    | 设备码               |
| socBrand      | string    | 金额                |
| soc           | string    | 设备型号              |
| cpu           | string    | CPU型号             |
| gpu           | string    | GPU型号             |
| ram           | int       | 运行内存大小(单位G)       |
| deviceRating  | int       | 设备评分              |

Example :
```json
{
  "ResultCode": 1,
  "ResultMsg": "成功",
  "Data": {
    "userId": 11031,
    "gameServerId": 2,
    "roleId": 32323244,
    "regIp": "172.20.40.46",
    "regCountryCode": "CN",
    "socMd5Id": "9cbfcf64eca08b6bb79647244c71d14a",
    "socBrand": "HUAWEI",
    "soc": "BAH4-W29",
    "cpu": "vendor Kirin710",
    "gpu": "ARM Mali-G51",
    "ram": 6,
    "deviceRating": 99
  }
}
```

### 1.4游戏推送(ios和android)
- Desc : 推送消息至客户端<br>
- URI : /v2/server/push<br>
- Method : POST<br>
- Request Parameter : Query Param格式
    - String gameAppId : 游戏应用id
    - String sign : 验签(gameId|gameAppId|gameAppSecret)
- Request Body : JSON 格式

| Element Name       | Data Type | Required | Description                                    |
|--------------------|-----------|------|------------------------------------------------|
| gameServiceId      | int       | true | 区服id                                           |
| userGameRoles      | list      |      | 推送角色(按照角色推送时必填 大小范围说明：该数组必须包含至少1个、最多1000个角色ID) |
| userId             | long      | true | 用户id                                           |
| roleId             | string    | true | 角色id                                           |
| data               | string    | true | 此参数用于指定消息载荷的自定义键值对。JSON格式 如：{\"score\":\"123\"} |
| dryrun             | bool      | true | 此参数设置为 true 时，开发者可在不实际发送消息的情况下对请求进行测试。默认值为 true |
| notices            | list      | true | 通知多语言内容                                        |
| region             | int       | true | 语言                                             |
| title              | string    | true | 通知标题                                           |
| body               | string    | true | 通知内容                                           |
| isDefault          | int       | true | 该语言是不是默认语言                                     |
| analyticsLabel     | string    | false | 推送跳转配置                                         |
| clickAction          | object    | true | 跳转跳转设置                                         |
| goType             | int       | true | 跳转类型：0默认 1跳转游戏页面 2跳转H5 3跳转房间                   |
| goContent          | string    | true | 跳转内容:游戏模块/h5地址/房间id                            |

Example :
```json
{
  "gameServiceId": 100,
  "userGameRoles": [{
    "UserId": 10001,
    "RoleId": "10001"
  }],
  "data": "",
  "notices": [{
    "region": 0,
    "title": "title",
    "body": "body",
    "isDefault": 1
  }],
  "dryRun": true,
  "analyticsLabel": "",
  "clickAction": {
    "goType": 0,
    "goContent": ""
  }
}
```
- Response Body : JSON格式

| Element Name | Data Type | Description |
|--------------|-----------|-------------|
| resultCode         | int       | 状态码         |
| resultMsg        | string    | 返回信息        |

Example :
```json
{
  "resultCode": 1,
  "resultMsg": "成功"
}
```

## 标准响应格式
| Element Name | Data Type | Description     |
|--------------|-----------|-----------------|
| resultCode   | int       | 结果状态码 1:成功 0:失败 |
| resultMsg    | int       | 返回结果说明            |
| data         | object    | 结果数据对象            |

## 2.游戏服务接口(游戏方提供)

### 2.1订单发货接口
Desc : 用于用户充值成功后进行发货操作<br>
URI : 游戏服务提供Url<br>
Method : POST<br>
Request Body : JSON 格式

| Element Name | Data Type | Required  | Description                                                                                                                                                        |
|--------------|-----------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| roleId       | string    | true      | 角色id                                                                                                                                                               |
| cpNo         | string    | true      | 游戏订单号                                                                                                                                                              |
| payNo        | string    | true      | sdk订单号                                                                                                                                                             |
| payType      | int       | true      | 支付类型                                                                                                                                                               |
| sandbox      | int       | true      | 是否沙盒                                                                                                                                                               |
| pointId      | int       | true      | 消费点id                                                                                                                                                              |
| sku          | string    | true      | 商品                                                                                                                                                                 |
| skuNum       | int       | true      | 商品数量                                                                                                                                                               |
| amount       | double    | true      | 总价格                                                                                                                                                                |
| payTime      | long      | true      | 支付时间戳                                                                                                                                                              |
| sign         | string    | true      | 验签(payTime&#124;roleId&#124;cpNo&#124;payNo&#124;payType&#124;sandbox&#124;pointId&#124;sku&#124;skuNum&#124;amount&#124;gameId&#124;gameAppId&#124;gameAppSecret) |

Response Body : Text格式
成功:SUCCESS
失败:FAIL|(其他错误信息)

注意：返回不是SUCCESS，SDK服务端会重复请求30次。重复请求发货成功后也需要返回SUCCESS停止请求。

### 2.2封/解封角色接口
Desc : 封禁或者解封角色<br>
URI : 游戏服务提供Url<br>
Method : POST<br>
Request Body : JSON 格式

| Element Name | Data Type | Required  | Description                                                                                             |
|--------------|-----------|-----------|---------------------------------------------------------------------------------------------------------|
| gameServerId       | int       | true      | 游戏区服id                                                                                                  |
| roleId    | string    | true      | 角色id                                                                                                    |
| time       | long      | true      | 操作时间戳                                                                                                   |
| reason       | string    | true      | 被封/解封原因                                                                                                 |
| type       | long      | true      | 封禁类型(1封号 2解封)                                                                                           |
| endTime       | long      | true      | 截至时间戳(秒)0为永久                                                                                            |
| sign         | string    | true      | 验签(time&#124;gameServerId&#124;roleId&#124;reason&#124;type&#124;gameid&#124;gameAppId&#124;gameAppSecret) |

Response Body : JSON格式

| Element Name  | Data Type | Description        |
|---------------|-----------|--------------------|
| result        | int       | 处理结果：1000成功 2000失败 |
| msg  | string    | 结果说明               |


### 2.3禁言/解禁角色接口
Desc : 禁言或解除禁言角色<br>
URI : 游戏服务提供Url<br>
Method : POST<br>
Request Body : JSON 格式

| Element Name | Data Type | Required  | Description                                                                                                |
|--------------|-----------|-----------|------------------------------------------------------------------------------------------------------------|
| gameServerId       | int       | true      | 游戏区服id                                                                                                     |
| roleId    | string    | true      | 角色id                                                                                                       |
| time       | long      | true      | 操作时间戳                                                                                                      |
| reason       | string    | true      | 被封/解封原因                                                                                                    |
| type       | long      | true      | 封禁类型(1禁言 2解除禁言)                                                                                            |
| endTime       | long      | true      | 截至时间戳(秒)0为永久                                                                                               |
| sign         | string    | true      | 验签(time&#124;gameServerId&#124;roleId&#124;reason&#124;type&#124;gameid&#124;gameAppId&#124;gameAppSecret) |

Response Body : JSON格式

| Element Name  | Data Type | Description        |
|---------------|-----------|--------------------|
| result        | int       | 处理结果：1000成功 2000失败 |
| msg  | string    | 结果说明               |

### 2.4查询角色信息接口
Desc : 检测角色对应的区服Id及sdk的用户Id(登录时调用确认游戏角色id当前绑定的sdk账号id和区服id)<br>
URI : 游戏服务提供Url<br>
Method : POST<br>
Request Body : JSON 格式

| Element Name | Data Type | Required  | Description                                                        |
|--------------|-----------|-----------|--------------------------------------------------------------------|
| roleId    | string    | true      | 角色id                                                               |
| time       | long      | true      | 操作时间戳                                                              |
| sign         | string    | true      | 验签(time&#124;roleId&#124;gameId&#124;gameAppId&#124;gameAppSecret) |

Response Body : JSON格式

| Element Name | Data Type | Description |
|--------------|-----------|-------------|
| result       | int       | 处理结果：1000成功 2000失败 |
| msg          | string    | 结果说明              |
| data         | object    |               |
| userId       | long      | sdk账号id     |
| serverId     | int       | 区服id        |

Example :
```json
{
  "result": 1000,
  "msg": "操作成功",
  "data": {
    "userId": 100001,
    "serverId": 2
  }
}
```

### 2.5YallaPay订单发货接口
Desc : 用于用户YallaPay充值成功后进行发货操作<br>
URI : 游戏服务提供Url<br>
Method : POST<br>
Request Body : JSON 格式

| Element Name | Data Type | Required  | Description                                                                                                              |
|--------------|-----------|-----------|--------------------------------------------------------------------------------------------------------------------------|
| roleId       | string    | true      | 角色id                                                                                                                     |
| payNo        | string    | true      | sdk订单号                                                                                                                   |
| payType      | int       | true      | 支付类型(1000-2000)                                                                                                          |
| sandbox      | int       | true      | 是否沙盒                                                                                                                     |
| amount       | double    | true      | 总价格                                                                                                                      |
| vPrice       | double       | true      | 虚拟商品价格/数量(币、金钻、钻石等)                                                                                                                     |
| timeStamp      | long      | true      | 支付时间戳                                                                                                                    |
| sign         | string    | true      | 验签(timeStamp&#124;roleId&#124;payNo&#124;payType&#124;sandbox&#124;amount&#124;vPrice&#124;gameAppId&#124;gameAppSecret) |

Response Body : Text格式
成功:SUCCESS
失败:FAIL|(其他错误信息)

注意：返回不是SUCCESS，SDK服务端会重复请求30次。重复请求发货成功后也需要返回SUCCESS停止请求。

## 参数枚举
| Key      | Value                                                                      | 
|----------|----------------------------------------------------------------------------|
| payState | 0:待支付 1:支付完成 2:发货完成 3:支付失败 4:发货失败                                          |
| payType  | 1:Ios 2:Google 3:Huawei 4:OneStore                                         |
| region   | 1:英语 2:阿语 3:土语 4:简中 5:繁中 6:泰语 7:日语 8:韩语 9:德语 10:法语 11:意大利语 12:葡萄牙语 13:西班牙语 |


## 充值流程
![](Image/pay.jpg)

## 需要同步的一些表（Excel格式）
1、游戏区服配置表（区服id、区服名）<br>
2、商品配置表（商品sku，商品名，商品价格）<br>
3、道具配置表（道具id，道具名称，区服id，道具类别，道具描述）<br>
4、角色等级配置表（等级编号，等级名称，区服id）<br>
5、角色VIP等级配置表（等级编号，等级名称，区服id）<br>
6、角色城堡等级配置表（等级编号，等级名称，区服id）<br>
7、消费点配置表（消费编号，消费点名称，消费点价格）