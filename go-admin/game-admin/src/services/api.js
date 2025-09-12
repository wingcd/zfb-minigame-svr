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
      }else{
        // request token
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

// 用于跟踪是否正在刷新token
let isRefreshing = false
// 存储等待刷新token的请求队列
let failedQueue = []

// 处理队列中的请求
const processQueue = (error, token = null) => {
  failedQueue.forEach(({ resolve, reject, config }) => {
    if (error) {
      reject(error)
    } else {
      // 更新请求的token
      if (token) {
        config.headers.authorization = `Bearer ${token}`
        if (typeof config.data === 'object') {
          config.data.token = token
        }
      }
      resolve(api(config))
    }
  })
  failedQueue = []
}

// 响应拦截器 - 处理401错误和数据格式，支持自动token刷新
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
      msg: '数据格式错误',
      data: null
    }
  },
  async error => {
    console.error('Response Error:', error)
    
    // 处理网络错误
    if (!error.response) {
      return Promise.resolve({
        code: 5001,
        msg: '网络连接失败',
        data: null
      })
    }
    
    const originalRequest = error.config
    const { status, data } = error.response
    
    if (status === 401 && !originalRequest._retry) {
      // 避免无限递归
      if (originalRequest.url?.includes('/admin/login') || originalRequest.url?.includes('/admin/verifyToken')) {
        // 登录或验证token接口返回401，直接跳转到登录页
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_info')
        window.location.href = '/login'
        return Promise.resolve({
          code: 4001,
          msg: '登录已过期，请重新登录',
          data: null
        })
      }

      if (isRefreshing) {
        // 如果正在刷新token，将请求加入队列
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject, config: originalRequest })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        // 尝试验证当前token是否仍有效
        const currentToken = localStorage.getItem('admin_token')
        if (currentToken) {
          const verifyResponse = await api.post('/admin/verifyToken', { token: currentToken })
          
          if (verifyResponse.code === 0 && verifyResponse.data?.valid) {
            // token仍然有效，可能是临时网络问题
            isRefreshing = false
            processQueue(null, currentToken)
            
            // 重试原始请求
            originalRequest.headers.authorization = `Bearer ${currentToken}`
            if (typeof originalRequest.data === 'object') {
              originalRequest.data.token = currentToken
            }
            return api(originalRequest)
          }
        }
        
        // token已失效，清除存储并跳转到登录页
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_info')
        
        isRefreshing = false
        processQueue(new Error('Token expired'), null)
        
        // 延迟跳转，给用户一些提示时间
        setTimeout(() => {
          if (window.location.pathname !== '/login') {
            window.location.href = '/login'
          }
        }, 1000)
        
        return Promise.resolve({
          code: 4001,
          msg: '登录已过期，正在跳转到登录页...',
          data: null
        })
        
      } catch (refreshError) {
        // 刷新失败，清除token并跳转
        isRefreshing = false
        processQueue(refreshError, null)
        
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_info')
        window.location.href = '/login'
        
        return Promise.resolve({
          code: 4001,
          msg: '登录已过期，请重新登录',
          data: null
        })
      }
    }
    
    // 返回服务器错误信息或默认错误
    return Promise.resolve(data || {
      code: status,
      msg: `HTTP ${status} 错误`,
      data: null
    })
  }
)

// 统一的API对象
const unifiedAPI = {
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
    resetPassword: (params) => api.post('/admin/resetPassword', params),
    update: (params) => api.post('/admin/update', params),
    init: (params) => api.post('/admin/init', params || {})
  },
  
  // 系统安装相关
  install: {
    // 检查系统状态
    checkStatus: () => api.get('/install/status'),
    // 系统初始化安装
    initSystem: (params) => api.post('/install/init', params),
    // 数据库连接测试
    testDB: (params) => api.post('/install/test-db', params),
    // Redis连接测试  
    testRedis: (params) => api.post('/install/test-redis', params)
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
    getStats: (appId) => api.post('/user/getStats', {appId}),
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
    restore: (params) => api.post('/counter/restore', params),
    getAllStats: (params) => api.post('/counter/getAllStats', params),
    toggleStatus: (params) => api.post('/counter/toggleStatus', params),
  },
  
  // 统计相关
  stats: {
    getDashboardStats: (params) => api.post('/stat/dashboard', params || {}),
    getTopApps: (params) => api.post('/stat/getTopApps', params || {}),
    getRecentActivity: (params) => api.post('/stat/getRecentActivity', params || {}),    
    userGrowth: (params) => api.post('/stat/getUserGrowth', params || {}),
    getAppStats: (params) => api.post('/stat/getAppStats', params || {}),
    leaderboardStats: (params) => api.post('/stat/getLeaderboardStats', params || {}),
    getPlatformDistribution: (params) => api.post('/stat/getPlatformDistribution', params || {})
  },
  
  // 邮件相关
  mail: {
    getAll: (params) => api.post('/mail/getAll', params || {}),
    create: (data) => api.post('/mail/create', data),
    update: (data) => api.post('/mail/update', data),
    delete: (params) => api.post('/mail/delete', params),
    publish: (mailData) => api.post('/mail/publish', mailData),
    publishExisting: (params) => api.post('/mail/publishExisting', params),
    getStats: (params) => api.post('/mail/getStats', params || {}),
    getUserMails: (params) => api.post('/mail/getUserMails', params),
    initSystem: (params) => api.post('/mail/initSystem', params || {})
  },
  
  // 游戏配置相关
  gameConfig: {
    getList: (params) => api.post('/gameConfig/getList', params || {}),
    create: (data) => api.post('/gameConfig/create', data),
    update: (data) => api.post('/gameConfig/update', data),
    delete: (data) => api.post('/gameConfig/delete', data),
    get: (params) => api.post('/gameConfig/get', params || {})
  },
  
  // Yalla配置相关
  yallaConfig: {
    getList: (params) => api.post('/yallaConfig/getList', params || {}),
    create: (data) => api.post('/yallaConfig/create', data),
    update: (data) => api.post('/yallaConfig/update', data),
    delete: (data) => api.post('/yallaConfig/delete', data),
    get: (params) => api.post('/yallaConfig/get', params || {}),
    testConnection: (params) => api.post('/yallaConfig/testConnection', params || {})
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

// 邮件管理API
export const mailAPI = unifiedAPI.mail

// 游戏配置管理API
export const gameConfigAPI = unifiedAPI.gameConfig

// Yalla配置管理API
export const yallaConfigAPI = unifiedAPI.yallaConfig

// === 便捷方法导出（向后兼容） ===
export const login = unifiedAPI.auth.login
export const verifyToken = unifiedAPI.auth.verifyToken
export const logout = unifiedAPI.auth.logout

export const getAdminList = unifiedAPI.admin.getList
export const createAdmin = unifiedAPI.admin.create
export const deleteAdmin = unifiedAPI.admin.delete
export const resetAdminPassword = unifiedAPI.admin.resetPassword
export const updateAdmin = unifiedAPI.admin.update
export const initAdmin = unifiedAPI.admin.init

// 安装相关API
export const checkInstallStatus = unifiedAPI.install.checkStatus
export const initSystemInstall = unifiedAPI.install.initSystem
export const testDBConnection = unifiedAPI.install.testDB
export const testRedisConnection = unifiedAPI.install.testRedis

export const getRoleList = unifiedAPI.role.getList
export const getAllRoles = unifiedAPI.role.getAll
export const createRole = unifiedAPI.role.create
export const updateRole = unifiedAPI.role.update
export const deleteRole = unifiedAPI.role.delete

export const initApp = unifiedAPI.app.init
export const queryApp = unifiedAPI.app.query
export const getAppList = unifiedAPI.app.getAll
export const updateApp = unifiedAPI.app.update
export const deleteApp = unifiedAPI.app.delete
export const createApp = unifiedAPI.app.create
export const getAppDetail = unifiedAPI.app.getDetail

export const getUserList = unifiedAPI.user.getAll
export const banUser = unifiedAPI.user.ban
export const unbanUser = unifiedAPI.user.unban
export const deleteUser = unifiedAPI.user.delete
export const getUserDetail = unifiedAPI.user.getDetail
export const setUserDetail = unifiedAPI.user.setDetail
export const getUserStats = unifiedAPI.user.getStats

export const getLeaderboardList = unifiedAPI.leaderboard.getAll
export const updateLeaderboard = unifiedAPI.leaderboard.update
export const deleteLeaderboard = unifiedAPI.leaderboard.delete
export const createLeaderboard = unifiedAPI.leaderboard.create
export const getLeaderboardData = unifiedAPI.leaderboard.getData
export const updateLeaderboardScore = unifiedAPI.leaderboard.updateScore
export const deleteLeaderboardScore = unifiedAPI.leaderboard.deleteScore

export const getCounterList = unifiedAPI.counter.getList
export const createCounter = unifiedAPI.counter.create
export const updateCounter = unifiedAPI.counter.update
export const deleteCounter = unifiedAPI.counter.delete
export const getCounterStats = unifiedAPI.counter.getAllStats
export const toggleCounterStatus = unifiedAPI.counter.toggleStatus

export const getDashboardStats = unifiedAPI.stats.getDashboardStats
export const getTopApps = unifiedAPI.stats.getTopApps
export const getRecentActivity = unifiedAPI.stats.getRecentActivity
export const getUserGrowth = unifiedAPI.stats.userGrowth
export const getAppStats = unifiedAPI.stats.getAppStats
export const getLeaderboardStats = unifiedAPI.stats.leaderboardStats

export const getMailList = unifiedAPI.mail.getAll
export const createMail = unifiedAPI.mail.create
export const updateMail = unifiedAPI.mail.update
export const deleteMail = unifiedAPI.mail.delete
export const publishMail = unifiedAPI.mail.publish
export const getMailStats = unifiedAPI.mail.getStats
export const getUserMails = unifiedAPI.mail.getUserMails
export const initMailSystem = unifiedAPI.mail.initSystem

// 游戏配置相关便捷方法
export const getGameConfigList = unifiedAPI.gameConfig.getList
export const createGameConfig = unifiedAPI.gameConfig.create
export const updateGameConfig = unifiedAPI.gameConfig.update
export const deleteGameConfig = unifiedAPI.gameConfig.delete
export const getGameConfig = unifiedAPI.gameConfig.get

export default api