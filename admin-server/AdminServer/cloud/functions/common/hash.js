const { createHash } = require('crypto');

/**
 * 检查请求参数
 * 1. 检查token
 * 2. 检查时间戳
 * 3. 检查签名
 * @param {*} obj 
 * @returns 0: success, 1: obj is null, 2: token is null, 3: timestamp is invalid, 4: sign is invalid
 */
function CheckParams(obj, isLogin = false) {
    if(!obj) {
        return 1;
    }

    if(!obj.token && !isLogin) {
        return 2;
    }

    // if(!CheckTimestamp(obj)) {
    //     return 3;
    // }

    if(!CheckSign(obj)) {
        return 4;
    }

    return 0;
}

function CheckSign(obj) {
    let hash = GenHash(obj);
    let sign = obj.sign;
    return sign === hash;
}

/**
 * 检查时间戳
 * @param {*} obj 
 * @param {number} timeout seconds, default 10s
 * @returns 
 */
function CheckTimestamp(obj, timeout = 10) {
    let timestamp = obj.timestamp;
    if(!timestamp) {
        return false;
    }

    let now = Date.now();
    return now - timestamp < timeout * 1000;
}

/**
 * 生成hash
 * obj: 请求参数对象
 * 1. 排除对象中的空值,0,undefined,null
 * 2. 对象按key排序
 * 3. sign,ver不参与生成hash
 * 4. 生成hash
 * @param {Object} obj 
 */
function GenHash(obj) {
    // 排除对象中的空值,0,undefined,null
    let newObj = {};
    for (let key in obj) {
        if(key == 'sign' || key == "ver") {
            continue;
        }

        if (obj[key] !== null && obj[key] !== undefined && obj[key] !== '' && obj[key] !== 0) {
            newObj[key] = obj[key];
        }
    }

    // 对象按key排序
    let keys = Object.keys(newObj).sort();
    // 生成hash
    let hash = '';
    for (let key of keys) {
        hash += key + newObj[key];
    }
    return createHash('md5').update(hash).digest('hex');
}

// test
// let obj = {
//     "appId": "123",
//     "playerId": "456",
//     "data": "789",
//     "sign": "123456",
//     "ver": "1.0",
// };
// console.log(GenHash(obj));

// let obj = {
//     "appId":"13c6994c-9beb-4d6d-8fcb-8500a1814f66",
//     "timestamp":1736079046148,
//     "ver":"1.0.0",
//     "code":"wing",
//     "sign":"86bd81b4499df1fee145e7f4359e2b8f"
// }
// console.log(GenHash(obj));

module.exports = {
    GenHash,
    CheckTimestamp,
    CheckSign,
    CheckParams,
}