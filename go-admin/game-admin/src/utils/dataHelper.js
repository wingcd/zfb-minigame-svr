/**
 * 数据处理工具函数
 */

/**
 * 确保数据是数组格式
 * @param {any} data - 需要验证的数据
 * @param {Array} defaultValue - 默认值，默认为空数组
 * @returns {Array} 验证后的数组
 */
export function ensureArray(data, defaultValue = []) {
  return Array.isArray(data) ? data : defaultValue
}

/**
 * 安全地设置响应式数组数据
 * @param {Ref} refValue - Vue3的ref对象
 * @param {any} data - 要设置的数据
 * @param {Array} defaultValue - 默认值
 */
export function safeSetArrayRef(refValue, data, defaultValue = []) {
  const validData = ensureArray(data, defaultValue)
  refValue.value = validData
}

/**
 * 处理API响应数据
 * @param {Object} response - API响应对象
 * @param {string} dataPath - 数据路径，如 'data.list'
 * @param {any} defaultValue - 默认值
 * @returns {any} 处理后的数据
 */
export function extractResponseData(response, dataPath = 'data', defaultValue = null) {
  if (!response || typeof response !== 'object') {
    return defaultValue
  }
  
  const paths = dataPath.split('.')
  let current = response
  
  for (const path of paths) {
    if (current && typeof current === 'object' && path in current) {
      current = current[path]
    } else {
      return defaultValue
    }
  }
  
  return current !== undefined ? current : defaultValue
}

/**
 * 安全地处理表格数据响应
 * @param {Object} response - API响应对象
 * @param {Ref} listRef - 列表的ref对象
 * @param {Ref} totalRef - 总数的ref对象（可选）
 * @param {string} listPath - 列表数据路径，默认 'data.list'
 * @param {string} totalPath - 总数数据路径，默认 'data.total'
 */
export function handleTableResponse(response, listRef, totalRef = null, listPath = 'data.list', totalPath = 'data.total') {
  if (response && response.code === 0) {
    // 安全设置列表数据
    const listData = extractResponseData(response, listPath, [])
    safeSetArrayRef(listRef, listData)
    
    // 安全设置总数
    if (totalRef) {
      const totalData = extractResponseData(response, totalPath, 0)
      totalRef.value = typeof totalData === 'number' ? totalData : 0
    }
    
    return true
  } else {
    // 出错时清空数据
    listRef.value = []
    if (totalRef) {
      totalRef.value = 0
    }
    return false
  }
}

/**
 * 验证对象是否具有必需的属性
 * @param {Object} obj - 要验证的对象
 * @param {Array<string>} requiredFields - 必需的字段列表
 * @returns {boolean} 是否验证通过
 */
export function validateRequiredFields(obj, requiredFields) {
  if (!obj || typeof obj !== 'object') {
    return false
  }
  
  return requiredFields.every(field => {
    return field in obj && obj[field] !== null && obj[field] !== undefined
  })
}

/**
 * 安全地获取嵌套对象属性
 * @param {Object} obj - 源对象
 * @param {string} path - 属性路径，如 'user.profile.name'
 * @param {any} defaultValue - 默认值
 * @returns {any} 属性值或默认值
 */
export function safeGet(obj, path, defaultValue = null) {
  if (!obj || typeof obj !== 'object') {
    return defaultValue
  }
  
  const paths = path.split('.')
  let current = obj
  
  for (const p of paths) {
    if (current && typeof current === 'object' && p in current) {
      current = current[p]
    } else {
      return defaultValue
    }
  }
  
  return current !== undefined ? current : defaultValue
} 