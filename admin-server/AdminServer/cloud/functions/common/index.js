const crypto = require('crypto');

/**
 * 生成随机字符串
 * @param {number} length - 字符串长度
 * @returns {string} - 随机字符串
 */
function generateRandomString(length = 32) {
    return crypto.randomBytes(length).toString('hex');
}

/**
 * 生成哈希值
 * @param {string} data - 待哈希的数据
 * @param {string} algorithm - 哈希算法，默认为sha256
 * @returns {string} - 哈希值
 */
function generateHash(data, algorithm = 'sha256') {
    return crypto.createHash(algorithm).update(data).digest('hex');
}

/**
 * 验证参数是否为空
 * @param {any} value - 待验证的值
 * @returns {boolean} - 是否为空
 */
function isEmpty(value) {
    return value === null || value === undefined || value === '';
}

/**
 * 格式化时间
 * @param {Date|string} date - 日期对象或字符串
 * @returns {string} - 格式化后的时间字符串
 */
function formatTime(date = new Date()) {
    const d = new Date(date);
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    const seconds = String(d.getSeconds()).padStart(2, '0');
    
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

module.exports = {
    generateRandomString,
    generateHash,
    isEmpty,
    formatTime
}; 