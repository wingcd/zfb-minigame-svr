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
    // 可以在这里添加token等认证信息
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
    return Promise.reject(error)
  }
)

// 应用管理API
export const appAPI = {
  initApp: (data) => api.post('/app/init', data),
  queryApp: (params) => api.get('/app/query', { params }),
  getAllApps: () => api.get('/apps'),
  createApp: (data) => api.post('/apps', data),
  updateApp: (id, data) => api.put(`/apps/${id}`, data),
  deleteApp: (id) => api.delete(`/apps/${id}`)
}

// 排行榜管理API
export const leaderboardAPI = {
  initLeaderboard: (data) => api.post('/leaderboard/init', data),
  commitScore: (data) => api.post('/leaderboard/commit', data),
  getTopRank: (params) => api.get('/leaderboard/top', { params }),
  queryScore: (params) => api.get('/leaderboard/score', { params }),
  deleteScore: (data) => api.post('/leaderboard/delete', data),
  getAllLeaderboards: () => api.get('/leaderboards'),
  createLeaderboard: (data) => api.post('/leaderboards', data),
  deleteLeaderboard: (id) => api.delete(`/leaderboards/${id}`)
}

// 用户管理API
export const userAPI = {
  login: (data) => api.post('/user/login', data),
  getUserData: (params) => api.get('/user/data', { params }),
  saveUserData: (data) => api.post('/user/save', data),
  getAllUsers: (params) => api.get('/users', { params }),
  banUser: (id) => api.post(`/users/${id}/ban`),
  unbanUser: (id) => api.post(`/users/${id}/unban`)
}

export default api