import axios from 'axios'
import config from '../config/index.js'

const api = axios.create({
  baseURL: config.api.baseURL,
  timeout: config.api.timeout,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 添加token到请求头
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    console.error('API Error:', error)
    // 如果是401错误，清除token并跳转到登录页
    if (error.response?.status === 401) {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// 管理员认证API
export const authAPI = {
  login: (data) => api.post('/admin/adminLogin', data), // 管理员登录
  verifyToken: (token) => api.post('/admin/verifyToken', { token }), // 验证token
  logout: () => {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
    return Promise.resolve()
  }
}

// 管理员管理API
export const adminAPI = {
  getAdminList: (params) => api.post('/admin/getAdminList', params || {}), // 获取管理员列表
  createAdmin: (data) => api.post('/admin/createAdmin', data), // 创建管理员
  updateAdmin: (data) => api.post('/admin/updateAdmin', data), // 更新管理员
  deleteAdmin: (id) => api.post('/admin/deleteAdmin', { id }), // 删除管理员
  resetPassword: (data) => api.post('/admin/resetPassword', data), // 重置密码
  initAdmin: (data) => api.post('/admin/initAdmin', data || {}) // 初始化管理员系统
}

// 角色管理API
export const roleAPI = {
  getRoleList: (params) => api.post('/admin/getRoleList', params || {}), // 获取角色列表
  createRole: (data) => api.post('/admin/createRole', data), // 创建角色
  updateRole: (data) => api.post('/admin/updateRole', data), // 更新角色
  deleteRole: (roleCode) => api.post('/admin/deleteRole', { roleCode }), // 删除角色
  getAllRoles: () => api.post('/admin/getAllRoles', {}) // 获取所有角色（用于下拉框）
}

// 应用管理API - 对应云函数
export const appAPI = {
  initApp: (data) => api.post('/app/appInit', data), // 对应 appInit.js
  queryApp: (params) => api.post('/app/queryApp', params), // 对应 queryApp.js
  getAllApps: (params) => api.post('/app/getAllApps', params || {}), // 对应 getAllApps.js
  deleteApp: (appId) => api.post('/app/deleteApp', { appId }), // 对应 deleteApp.js
  updateApp: (data) => api.post('/app/updateApp', data) // 对应 updateApp.js
}

// 排行榜管理API - 对应云函数
export const leaderboardAPI = {
  initLeaderboard: (data) => api.post('/leaderboard/leaderboardInit', data), // 对应 leaderboardInit.js
  commitScore: (data) => api.post('/leaderboard/commitScore', data), // 对应 commitScore.js
  getTopRank: (params) => api.post('/leaderboard/getLeaderboardTopRank', params), // 对应 getLeaderboardTopRank.js
  queryScore: (params) => api.post('/leaderboard/queryScore', params), // 对应 queryScore.js
  deleteScore: (data) => api.post('/leaderboard/deleteScore', data), // 对应 deleteScore.js
  getAllLeaderboards: (params) => api.post('/leaderboard/getAllLeaderboards', params || {}), // 对应 getAllLeaderboards.js
  createLeaderboard: (data) => api.post('/leaderboard/leaderboardInit', data), // 复用 leaderboardInit.js
  updateLeaderboard: (data) => api.post('/leaderboard/updateLeaderboard', data), // 对应 updateLeaderboard.js
  deleteLeaderboard: (data) => api.post('/leaderboard/deleteLeaderboard', data) // 对应 deleteLeaderboard.js
}

// 用户管理API - 对应云函数
export const userAPI = {
  login: (data) => api.post('/user/login', data), // 对应 login.js
  loginWechat: (data) => api.post('/user/login.wx', data), // 对应 login.wx.js
  getUserData: (params) => api.post('/user/getData', params), // 对应 getData.js
  saveUserData: (data) => api.post('/user/saveData', data), // 对应 saveData.js
  getAllUsers: (params) => api.post('/user/getAllUsers', params || {}), // 对应 getAllUsers.js
  banUser: (data) => api.post('/user/banUser', data), // 对应 banUser.js
  unbanUser: (data) => api.post('/user/unbanUser', data), // 对应 unbanUser.js
  deleteUser: (data) => api.post('/user/deleteUser', data), // 对应 deleteUser.js
  queryUser: (params) => api.post('/user/getData', params) // 复用 getData.js
}

// 统计API - 对应云函数
export const statsAPI = {
  getDashboardStats: (params) => api.post('/stats/getDashboardStats', params || {}), // 对应 getDashboardStats.js
  getAppStats: (appId) => api.post('/stats/getAppStats', { appId }), // 需要创建
  getUserStats: (appId) => api.post('/stats/getUserStats', { appId }), // 需要创建
  getLeaderboardStats: (appId) => api.post('/stats/getLeaderboardStats', { appId }) // 需要创建
}

export default api