import axios from 'axios'
import config from '@/config'

// 创建axios实例
const api = axios.create({
  baseURL: config.baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理401错误
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    if (error.response?.status === 401) {
      // token过期或无效，清除本地存储并跳转到登录页
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// === 统一API调用 ===
export const unifiedAPI = {
  // 通用调用方法
  call: (action, params = {}) => api.post('/admin/callAPI', { action, params }),
  
  // 获取API列表
  getAPIList: () => api.post('/admin/callAPI', { action: 'api.list' }),
  
  // 管理员相关
  admin: {
    getList: (params) => api.post('/admin/callAPI', { action: 'admin.getList', params }),
    create: (params) => api.post('/admin/callAPI', { action: 'admin.create', params }),
    update: (params) => api.post('/admin/callAPI', { action: 'admin.update', params }),
    delete: (params) => api.post('/admin/callAPI', { action: 'admin.delete', params }),
    resetPassword: (params) => api.post('/admin/callAPI', { action: 'admin.resetPassword', params })
  },
  
  // 角色相关
  role: {
    getList: (params) => api.post('/admin/callAPI', { action: 'role.getList', params }),
    getAll: (params) => api.post('/admin/callAPI', { action: 'role.getAll', params })
  },
  
  // 认证相关
  auth: {
    login: (params) => api.post('/admin/callAPI', { action: 'auth.login', params }),
    verify: (params) => api.post('/admin/callAPI', { action: 'auth.verify', params }),
    init: (params) => api.post('/admin/callAPI', { action: 'auth.init', params })
  },
  
  // 应用相关
  app: {
    init: (params) => api.post('/admin/callAPI', { action: 'app.init', params }),
    query: (params) => api.post('/admin/callAPI', { action: 'app.query', params }),
    getAll: (params) => api.post('/admin/callAPI', { action: 'app.getAll', params }),
    update: (params) => api.post('/admin/callAPI', { action: 'app.update', params }),
    delete: (params) => api.post('/admin/callAPI', { action: 'app.delete', params })
  },
  
  // 用户相关
  user: {
    getAll: (params) => api.post('/admin/callAPI', { action: 'user.getAll', params }),
    ban: (params) => api.post('/admin/callAPI', { action: 'user.ban', params }),
    unban: (params) => api.post('/admin/callAPI', { action: 'user.unban', params }),
    delete: (params) => api.post('/admin/callAPI', { action: 'user.delete', params })
  },
  
  // 排行榜相关
  leaderboard: {
    getAll: (params) => api.post('/admin/callAPI', { action: 'leaderboard.getAll', params }),
    update: (params) => api.post('/admin/callAPI', { action: 'leaderboard.update', params }),
    delete: (params) => api.post('/admin/callAPI', { action: 'leaderboard.delete', params })
  },
  
  // 统计相关
  stats: {
    dashboard: (params) => api.post('/admin/callAPI', { action: 'stats.dashboard', params })
  }
}

// === 认证相关API ===
export const authAPI = {
  login: (data) => api.post('/admin/adminLogin', data), // 管理员登录
  verifyToken: (token) => api.post('/admin/verifyToken', { token }), // 验证token
  logout: () => {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
    return Promise.resolve()
  }
}

// === 管理员管理API ===
export const adminAPI = {
  getAdminList: (params) => api.post('/admin/getAdminList', params || {}), // 获取管理员列表
  createAdmin: (data) => api.post('/admin/createAdmin', data), // 创建管理员
  updateAdmin: (data) => api.post('/admin/updateAdmin', data), // 更新管理员
  deleteAdmin: (id) => api.post('/admin/deleteAdmin', { id }), // 删除管理员
  resetPassword: (data) => api.post('/admin/resetPassword', data), // 重置密码
  initAdmin: (data) => api.post('/admin/initAdmin', data || {}) // 初始化管理员系统
}

// === 角色管理API ===
export const roleAPI = {
  getRoleList: (params) => api.post('/admin/getRoleList', params || {}), // 获取角色列表
  createRole: (data) => api.post('/admin/createRole', data), // 创建角色
  updateRole: (data) => api.post('/admin/updateRole', data), // 更新角色
  deleteRole: (roleCode) => api.post('/admin/deleteRole', { roleCode }), // 删除角色
  getAllRoles: () => api.post('/admin/getAllRoles', {}) // 获取所有角色（用于下拉框）
}

// === 应用管理API ===
export const appAPI = {
  initApp: (data) => api.post('/app/appInit', data), // 初始化应用
  queryApp: (params) => api.post('/app/queryApp', params), // 查询应用详情
  getAllApps: (params) => api.post('/app/getAllApps', params || {}), // 获取应用列表
  createApp: (data) => api.post('/app/createApp', data), // 创建应用
  updateApp: (data) => api.post('/app/updateApp', data), // 更新应用
  deleteApp: (appId) => api.post('/app/deleteApp', { appId }), // 删除应用
  getAppDetail: (appId) => api.post('/app/getAppDetail', { appId }) // 获取应用详情
}

// === 用户管理API ===
export const userAPI = {
  getAllUsers: (params) => api.post('/user/getAllUsers', params), // 获取用户列表
  getUserDetail: (data) => api.post('/user/getUserDetail', data), // 获取用户详情
  updateUserData: (data) => api.post('/user/updateUserData', data), // 更新用户数据
  banUser: (data) => api.post('/user/banUser', data), // 封禁用户
  unbanUser: (data) => api.post('/user/unbanUser', data), // 解封用户
  deleteUser: (data) => api.post('/user/deleteUser', data) // 删除用户
}

// === 排行榜管理API ===
export const leaderboardAPI = {
  getAllLeaderboards: (params) => api.post('/leaderboard/getAllLeaderboards', params), // 获取排行榜列表
  createLeaderboard: (data) => api.post('/leaderboard/createLeaderboard', data), // 创建排行榜
  updateLeaderboard: (data) => api.post('/leaderboard/updateLeaderboard', data), // 更新排行榜
  deleteLeaderboard: (data) => api.post('/leaderboard/deleteLeaderboard', data), // 删除排行榜
  getLeaderboardData: (params) => api.post('/leaderboard/getLeaderboardData', params), // 获取排行榜数据
  updateScore: (data) => api.post('/leaderboard/updateScore', data), // 更新分数
  deleteScore: (data) => api.post('/leaderboard/deleteScore', data) // 删除分数
}

// === 统计数据API ===
export const statsAPI = {
  getDashboardStats: (params) => api.post('/stats/getDashboardStats', params || {}), // 获取仪表板统计
  getUserGrowth: (params) => api.post('/stats/getUserGrowth', params || {}), // 获取用户增长数据
  getAppStats: (params) => api.post('/stats/getAppStats', params || {}), // 获取应用统计
  getLeaderboardStats: (params) => api.post('/stats/getLeaderboardStats', params || {}) // 获取排行榜统计
}

export default api