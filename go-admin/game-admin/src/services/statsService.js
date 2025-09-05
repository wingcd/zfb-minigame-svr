import api from './api'

export const statsService = {
  // 获取仪表盘统计数据
  async getDashboardStats() {
    try {
      const response = await api.get('/stats/dashboard')
      return response
    } catch (error) {
      throw new Error('获取统计数据失败: ' + error.message)
    }
  },

  // 获取实时统计数据
  async getRealtimeStats() {
    try {
      const response = await api.get('/stats/realtime')
      return response
    } catch (error) {
      throw new Error('获取实时数据失败: ' + error.message)
    }
  }
}