import { Md5 } from "../framework/libs/md5";

/**
 * 生成hash
 * obj: 请求参数对象
 * 1. 排除对象中的空值,0,undefined,null
 * 2. 对象按key排序
 * 3. sign,ver不参与生成hash
 * 4. 生成hash
 * @param {Object} obj 
 */
export function GenHash(obj: any) {
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
    return Md5.hashStr(hash);
}

// test
// let obj = {
//     "appId": "123",
//     "playerId": "456",
//     "data": "789",
//     "sign": "123456",
//     "ver": "1.0"
// };

// let hash = GenHash(obj);
// console.log(hash, hash == "2042eb87f638e871be764e436fad871d");