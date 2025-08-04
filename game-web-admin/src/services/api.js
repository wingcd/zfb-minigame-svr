import axios from 'axios'
import config from '@/config'

// 创建axios实例
const api = axios.create({
  baseURL: config.api.baseURL,
  timeout: config.api.timeout || 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加token
api.interceptors.request.use(
  config => {
    // console.log('API Request:', config.method?.toUpperCase(), config.url, config.data)
    
    const token = localStorage.getItem('admin_token')
    // console.log('Token Debug:', {
    //   tokenExists: !!token,
    //   tokenLength: token?.length,
    //   tokenPreview: token ? `${token.substring(0, 8)}...${token.substring(token.length-8)}` : null
    // })
    
    if (token) {
      config.headers.authorization = `Bearer ${token}`
      if(typeof config.data === 'object') {
        config.data.token = token;
      }
      // console.log('Authorization header set:', config.headers.authorization.substring(0, 20) + '...')
    } else {
      console.warn('No token found in localStorage')
    }
    return config
  },
  error => {
    console.error('Request Error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理401错误和数据格式
api.interceptors.response.use(
  response => {
    const data = response.data
    // 确保返回的数据有基本结构
    if (typeof data === 'object' && data !== null) {
      return data
    }
    // 如果数据格式异常，返回标准错误格式
    return {
      code: 5001,
      msg: '服务器返回数据格式异常',
      data: null
    }
  },
  error => {
    console.error('API Error:', error)
    console.error('Error Response:', error.response)
    console.error('Error Config:', error.config)
    
    if (error.response?.status === 401) {
      // token过期或无效，清除本地存储
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      
      // 避免重复跳转，只有当前不在登录页时才跳转
      if (window.location.pathname !== '/login') {
        console.warn('Token已失效，即将跳转到登录页')
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// === 统一API调用 ===
export const unifiedAPI = {  
  // 认证相关
  auth: {
    login: (params) => api.post('/admin/login', params),
    verifyToken: (params) => api.post('/admin/verifyToken', params),
    logout: () => {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      return Promise.resolve()
    }
  },
  
  // 管理员相关
  admin: {
    getList: (params) => api.post('/admin/getList', params || {}),
    create: (params) => api.post('/admin/create', params),
    delete: (params) => api.post('/admin/delete', params),
    resetPassword: (params) => api.post('/admin/resetPwd', params),
    update: (params) => api.post('/admin/update', params),
    init: (params) => api.post('/admin/init', params || {})
  },
  
  // 角色相关
  role: {
    getList: (params) => api.post('/role/getList', params || {}),
    getAll: (params) => api.post('/role/getAll', params || {}),    
    create: (data) => api.post('/role/create', data),
    update: (data) => api.post('/role/update', data),
    delete: (roleCode) => api.post('/role/delete', { roleCode }),
  },
  
  // 应用相关
  app: {
    init: (params) => api.post('/app/init', params),
    query: (params) => api.post('/app/query', params),
    getAll: (params) => api.post('/app/getAll', params || {}),
    update: (params) => api.post('/app/update', params),
    delete: (params) => api.post('/app/delete', params),    
    create: (data) => api.post('/app/create', data),
    getDetail: (appId) => api.post('/app/getDetail', { appId }),
  },
  
  // 用户相关
  user: {
    getAll: (params) => api.post('/user/getAll', params),
    ban: (params) => api.post('/user/ban', params),
    unban: (params) => api.post('/user/unban', params),
    delete: (params) => api.post('/user/delete', params),    
    getDetail: (data) => api.post('/user/getDetail', data),
    setDetail: (data) => api.post('/user/setDetail', data),
    getStats: (params) => api.post('/user/getStats', params),
  },
  
  // 排行榜相关
  leaderboard: {
    getAll: (params) => api.post('/leaderboard/getAll', params),
    update: (params) => api.post('/leaderboard/update', params),
    delete: (params) => api.post('/leaderboard/delete', params),    
    create: (data) => api.post('/leaderboard/create', data),
    getData: (params) => api.post('/leaderboard/getData', params),
    updateScore: (data) => api.post('/leaderboard/updateScore', data),
    deleteScore: (data) => api.post('/leaderboard/deleteScore', data),
  },
  
  // 计数器相关
  counter: {
    getList: (params) => api.post('/counter/getList', params),
    create: (data) => api.post('/counter/create', data),
    update: (data) => api.post('/counter/update', data),
    delete: (params) => api.post('/counter/delete', params),
  },
  
  // 统计相关
  stats: {
    getDashboardStats: (params) => api.post('/stat/dashboard', params || {}),
    getTopApps: (params) => api.post('/stat/getTopApps', params || {}),
    getRecentActivity: (params) => api.post('/stat/getRecentActivity', params || {}),    
    userGrowth: (params) => api.post('/stat/getUserGrowth', params || {}),
    getAppStats: (params) => api.post('/stat/getAppStats', params || {}),
    leaderboardStats: (params) => api.post('/stat/getLeaderboardStats', params || {}),
    getUserStats: (params) => api.post('/stat/getUserStats', params || {})
  }
}

// === 为了向后兼容，保留原有的API导出 ===
// 认证相关API
export const authAPI = unifiedAPI.auth

// 管理员管理API
export const adminAPI = unifiedAPI.admin

// 角色管理API
export const roleAPI = unifiedAPI.role

// 应用管理API
export const appAPI = unifiedAPI.app

// 用户管理API
export const userAPI = unifiedAPI.user

// 排行榜管理API
export const leaderboardAPI = unifiedAPI.leaderboard

// 计数器管理API
export const counterAPI = unifiedAPI.counter

// 统计数据API
export const statsAPI = unifiedAPI.stats

export default api