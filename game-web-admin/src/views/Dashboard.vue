<template>
  <div class="dashboard">
    <div class="page-header">
      <h1>管理台概览</h1>
      <div class="header-actions">
        <el-select v-model="timeRange" @change="handleTimeRangeChange" style="width: 150px;">
          <el-option label="今天" value="today"></el-option>
          <el-option label="最近7天" value="week"></el-option>
          <el-option label="最近30天" value="month"></el-option>
        </el-select>
        <el-button @click="refreshData" :loading="loading">刷新数据</el-button>
      </div>
    </div>

    <!-- 总体统计卡片 -->
    <div class="overview-stats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-header">
                <span class="stat-title">应用总数</span>
                <el-icon class="stat-icon app-icon"><Platform /></el-icon>
              </div>
              <div class="stat-value">{{ dashboardStats?.apps?.total || 0 }}</div>
              <div class="stat-change" :class="getChangeClass(dashboardStats?.apps?.change)">
                <span>{{ getChangeText(dashboardStats?.apps?.change) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-header">
                <span class="stat-title">用户总数</span>
                <el-icon class="stat-icon user-icon"><User /></el-icon>
              </div>
              <div class="stat-value">{{ dashboardStats?.users?.total || 0 }}</div>
              <div class="stat-change" :class="getChangeClass(dashboardStats?.users?.change)">
                <span>{{ getChangeText(dashboardStats?.users?.change) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-header">
                <span class="stat-title">今日活跃</span>
                <el-icon class="stat-icon active-icon"><TrendCharts /></el-icon>
              </div>
              <div class="stat-value">{{ dashboardStats?.activity?.daily || 0 }}</div>
              <div class="stat-change" :class="getChangeClass(dashboardStats?.activity?.change)">
                <span>{{ getChangeText(dashboardStats?.activity?.change) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-header">
                <span class="stat-title">排行榜数</span>
                <el-icon class="stat-icon leaderboard-icon"><Medal /></el-icon>
              </div>
              <div class="stat-value">{{ dashboardStats?.leaderboards?.total || 0 }}</div>
              <div class="stat-change" :class="getChangeClass(dashboardStats?.leaderboards?.change)">
                <span>{{ getChangeText(dashboardStats?.leaderboards?.change) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 图表区域 -->
    <div class="charts-section">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>用户增长趋势</span>
                <el-button type="text" @click="refreshUserTrend">刷新</el-button>
              </div>
            </template>
            <div class="chart-container" ref="userTrendChart" style="height: 300px;"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>平台分布</span>
                <el-button type="text" @click="refreshPlatformStats">刷新</el-button>
              </div>
            </template>
            <div class="chart-container" ref="platformChart" style="height: 300px;"></div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 热门应用和最近活动 -->
    <div class="activity-section">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>热门应用 TOP 10</span>
                <el-button type="text" @click="refreshTopApps">刷新</el-button>
              </div>
            </template>
            <div class="top-apps-list">
              <div 
                v-for="(app, index) in topApps" 
                :key="app.appId" 
                class="app-item"
                @click="viewAppDetail(app)">
                <div class="app-rank">{{ index + 1 }}</div>
                <div class="app-info">
                  <div class="app-name">{{ app.appName }}</div>
                  <div class="app-stats">
                    <span>{{ app.userCount }} 用户</span>
                    <span>{{ app.dailyActive }} 日活</span>
                  </div>
                </div>
                <div class="app-platform">
                  <el-tag size="small" :type="getPlatformType(app.platform)">
                    {{ getPlatformText(app.platform) }}
                  </el-tag>
                </div>
              </div>
              <div v-if="topApps.length === 0" class="empty-state">
                <el-empty description="暂无数据"></el-empty>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>最近活动</span>
                <el-button type="text" @click="refreshRecentActivity">刷新</el-button>
              </div>
            </template>
            <div class="activity-list">
              <el-timeline>
                <el-timeline-item
                  v-for="activity in recentActivities"
                  :key="activity.id"
                  :timestamp="activity.timestamp"
                  :type="getActivityType(activity.type)">
                  <div class="activity-content">
                    <div class="activity-title">{{ activity.title }}</div>
                    <div class="activity-desc">{{ activity.description }}</div>
                  </div>
                </el-timeline-item>
              </el-timeline>
              <div v-if="recentActivities.length === 0" class="empty-state">
                <el-empty description="暂无活动记录"></el-empty>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <el-card>
        <template #header>
          <span>快速操作</span>
        </template>
        <div class="action-buttons">
          <el-button type="primary" @click="goToAppManagement">
            <el-icon><Platform /></el-icon>
            创建应用
          </el-button>
          <el-button type="success" @click="goToUserManagement">
            <el-icon><User /></el-icon>
            用户管理
          </el-button>
          <el-button type="warning" @click="goToLeaderboardManagement">
            <el-icon><Medal /></el-icon>
            排行榜管理
          </el-button>
          <el-button @click="exportData">
            <el-icon><Download /></el-icon>
            导出数据
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Platform, User, TrendCharts, Medal, Download } from '@element-plus/icons-vue'
import { statsAPI } from '../services/api.js'
import * as echarts from 'echarts'

export default {
  name: 'Dashboard',
  components: {
    Platform,
    User,
    TrendCharts,
    Medal,
    Download
  },
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const timeRange = ref('week')
    const dashboardStats = ref(null)
    const topApps = ref([])
    const recentActivities = ref([])
    
    // 图表实例
    const userTrendChart = ref(null)
    const platformChart = ref(null)
    let userTrendChartInstance = null
    let platformChartInstance = null
    
    // 获取仪表板统计数据
    const getDashboardStats = async () => {
      try {
        const result = await statsAPI.getDashboardStats()
        if (result.code === 0) {
          dashboardStats.value = result.data
        }
      } catch (error) {
        console.error('获取仪表板统计失败:', error)
        ElMessage.error('获取统计数据失败')
      }
    }
    
    // 获取热门应用
    const getTopApps = async () => {
      try {
        const result = await statsAPI.getTopApps?.({ limit: 10 })
        if (result?.code === 0) {
          topApps.value = result.data || []
        }
      } catch (error) {
        console.error('获取热门应用失败:', error)
      }
    }
    
    // 获取最近活动
    const getRecentActivity = async () => {
      try {
        const result = await statsAPI.getRecentActivity?.({ limit: 10 })
        if (result?.code === 0) {
          recentActivities.value = result.data || []
        } else {
          // 模拟数据
          recentActivities.value = [
            {
              id: 1,
              type: 'user_register',
              title: '新用户注册',
              description: '玩家 player123 在应用"跳跃小游戏"中注册',
              timestamp: '2024-01-15 14:30:00'
            },
            {
              id: 2,
              type: 'score_submit',
              title: '分数提交',
              description: '玩家 player456 在排行榜"简单模式"中提交了新分数 1250',
              timestamp: '2024-01-15 14:25:00'
            },
            {
              id: 3,
              type: 'app_create',
              title: '应用创建',
              description: '创建了新应用"消除小游戏"',
              timestamp: '2024-01-15 13:45:00'
            }
          ]
        }
      } catch (error) {
        console.error('获取最近活动失败:', error)
      }
    }
    
    // 初始化用户趋势图表
    const initUserTrendChart = async () => {
      if (!userTrendChart.value) return
      
      await nextTick()
      userTrendChartInstance = echarts.init(userTrendChart.value)
      
      const option = {
        title: {
          text: '用户增长趋势',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'axis'
        },
        xAxis: {
          type: 'category',
          data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月']
        },
        yAxis: {
          type: 'value'
        },
        series: [{
          data: [120, 200, 150, 80, 70, 110, 130],
          type: 'line',
          smooth: true,
          itemStyle: {
            color: '#409eff'
          }
        }]
      }
      
      userTrendChartInstance.setOption(option)
    }
    
    // 初始化平台分布图表
    const initPlatformChart = async () => {
      if (!platformChart.value) return
      
      await nextTick()
      platformChartInstance = echarts.init(platformChart.value)
      
      const option = {
        title: {
          text: '平台分布',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'item'
        },
        series: [{
          type: 'pie',
          radius: '50%',
          data: [
            { value: 40, name: '微信小程序' },
            { value: 30, name: '支付宝小程序' },
            { value: 20, name: '抖音小程序' },
            { value: 10, name: '其他' }
          ],
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }]
      }
      
      platformChartInstance.setOption(option)
    }
    
    // 工具函数
    const getChangeClass = (change) => {
      if (change > 0) return 'change-positive'
      if (change < 0) return 'change-negative'
      return 'change-neutral'
    }
    
    const getChangeText = (change) => {
      if (change > 0) return `+${change}%`
      if (change < 0) return `${change}%`
      return '0%'
    }
    
    const getPlatformText = (platform) => {
      const platforms = {
        'wechat': '微信',
        'alipay': '支付宝',
        'douyin': '抖音'
      }
      return platforms[platform] || platform
    }
    
    const getPlatformType = (platform) => {
      const types = {
        'wechat': 'success',
        'alipay': 'primary',
        'douyin': 'warning'
      }
      return types[platform] || 'info'
    }
    
    const getActivityType = (type) => {
      const types = {
        'user_register': 'success',
        'score_submit': 'primary',
        'app_create': 'warning',
        'error': 'danger'
      }
      return types[type] || 'info'
    }
    
    // 事件处理
    const handleTimeRangeChange = () => {
      refreshData()
    }
    
    const refreshData = async () => {
      loading.value = true
      try {
        await Promise.all([
          getDashboardStats(),
          getTopApps(),
          getRecentActivity()
        ])
      } finally {
        loading.value = false
      }
    }
    
    const refreshUserTrend = () => {
      // 刷新用户趋势图表数据
      if (userTrendChartInstance) {
        // 这里可以重新获取数据并更新图表
        console.log('刷新用户趋势图表')
      }
    }
    
    const refreshPlatformStats = () => {
      // 刷新平台统计图表数据
      if (platformChartInstance) {
        console.log('刷新平台统计图表')
      }
    }
    
    const refreshTopApps = () => {
      getTopApps()
    }
    
    const refreshRecentActivity = () => {
      getRecentActivity()
    }
    
    // 导航操作
    const goToAppManagement = () => {
      router.push('/apps')
    }
    
    const goToUserManagement = () => {
      router.push('/users')
    }
    
    const goToLeaderboardManagement = () => {
      router.push('/leaderboards')
    }
    
    const viewAppDetail = (app) => {
      // 跳转到应用详情或应用管理页面
      router.push(`/apps?appId=${app.appId}`)
    }
    
    const exportData = () => {
      ElMessage.info('导出功能开发中...')
    }
    
    onMounted(async () => {
      await refreshData()
      await nextTick()
      initUserTrendChart()
      initPlatformChart()
      
      // 监听窗口大小变化
      window.addEventListener('resize', () => {
        if (userTrendChartInstance) userTrendChartInstance.resize()
        if (platformChartInstance) platformChartInstance.resize()
      })
    })
    
    return {
      loading,
      timeRange,
      dashboardStats,
      topApps,
      recentActivities,
      userTrendChart,
      platformChart,
      getDashboardStats,
      getChangeClass,
      getChangeText,
      getPlatformText,
      getPlatformType,
      getActivityType,
      handleTimeRangeChange,
      refreshData,
      refreshUserTrend,
      refreshPlatformStats,
      refreshTopApps,
      refreshRecentActivity,
      goToAppManagement,
      goToUserManagement,
      goToLeaderboardManagement,
      viewAppDetail,
      exportData
    }
  }
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.overview-stats {
  margin-bottom: 20px;
}

.stat-card {
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.stat-item {
  padding: 10px 0;
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.stat-title {
  font-size: 14px;
  color: #666;
}

.stat-icon {
  font-size: 24px;
}

.app-icon {
  color: #409eff;
}

.user-icon {
  color: #67c23a;
}

.active-icon {
  color: #e6a23c;
}

.leaderboard-icon {
  color: #f56c6c;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #333;
  margin-bottom: 5px;
}

.stat-change {
  font-size: 12px;
  font-weight: 500;
}

.change-positive {
  color: #67c23a;
}

.change-negative {
  color: #f56c6c;
}

.change-neutral {
  color: #909399;
}

.charts-section {
  margin-bottom: 20px;
}

.activity-section {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  width: 100%;
}

.top-apps-list {
  max-height: 400px;
  overflow-y: auto;
}

.app-item {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background-color 0.3s;
}

.app-item:hover {
  background-color: #f5f7fa;
}

.app-item:last-child {
  border-bottom: none;
}

.app-rank {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #409eff;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  margin-right: 15px;
}

.app-info {
  flex: 1;
}

.app-name {
  font-weight: bold;
  font-size: 14px;
  margin-bottom: 4px;
}

.app-stats {
  font-size: 12px;
  color: #666;
}

.app-stats span {
  margin-right: 15px;
}

.app-platform {
  margin-left: 10px;
}

.activity-list {
  max-height: 400px;
  overflow-y: auto;
}

.activity-content {
  padding-left: 10px;
}

.activity-title {
  font-weight: bold;
  font-size: 14px;
  margin-bottom: 4px;
}

.activity-desc {
  font-size: 12px;
  color: #666;
}

.quick-actions {
  margin-bottom: 20px;
}

.action-buttons {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-state {
  padding: 40px 0;
  text-align: center;
}

@media (max-width: 768px) {
  .dashboard {
    padding: 10px;
  }
  
  .action-buttons {
    flex-direction: column;
  }
  
  .action-buttons .el-button {
    width: 100%;
    justify-content: center;
  }
}
</style>