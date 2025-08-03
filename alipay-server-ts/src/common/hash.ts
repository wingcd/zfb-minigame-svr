const { createHash } = require('crypto');

class Hash {
    public CheckSign(obj) {
        let hash = this.GenHash(obj);
        let sign = obj.sign;
        return sign === hash;
    }

    /**
     * 检查时间戳
     * @param {*} obj 
     * @param {number} timeout seconds, default 10s
     * @returns 
     */
    public CheckTimestamp(obj, timeout = 10) {
        let timestamp = obj.timestamp;
        if(!timestamp) {
            return false;
        }

        let now = new Date().getTime();
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
    public GenHash(obj) {
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
        let newObj2 = {};
        for (let key of keys) {
            newObj2[key] = newObj[key];
        }

        // 生成hash
        let hash = '';
        for (let key in newObj2) {
            hash += key + newObj2[key];
        }
        return createHash('md5').update(hash).digest('hex');
    }
}

export default new Hash();