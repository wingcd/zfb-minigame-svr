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
                <el-button link @click="refreshUserTrend">刷新</el-button>
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
                <el-button link @click="refreshPlatformStats">刷新</el-button>
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
                <el-button link @click="refreshTopApps">刷新</el-button>
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
                <el-button link @click="refreshRecentActivity">刷新</el-button>
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
                    <div class="activity-title">{{ activity.action }}</div>
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
        } else if (result.code === 401 || result.code === 4001) {
          // Token无效或已过期
          ElMessage.error('登录已过期，请重新登录')
        } else {
          ElMessage.error(result.msg || '获取统计数据失败')
        }
      } catch (error) {
        console.error('获取仪表板统计失败:', error)
        
        // 检查是否是认证相关错误
        if (error.response?.status === 401) {
          ElMessage.error('登录已过期，请重新登录')
        } else if (error.message?.includes('token') || error.message?.includes('认证')) {
          ElMessage.error('认证失败，请重新登录')
        } else {
          ElMessage.error('获取统计数据失败')
        }
      }
    }
    
    // 获取热门应用
    const getTopApps = async () => {
      try {
        const result = await statsAPI.getTopApps?.({ limit: 10 })
        if (result?.code === 0) {
          // 确保数据是数组格式，防止迭代错误
          const dataList = result.data
          topApps.value = Array.isArray(dataList) ? dataList : []
        } else if (result?.code === 401 || result?.code === 4001) {
          // Token无效，不显示错误，让主要的统计接口处理
          return
        }
      } catch (error) {
        console.error('获取热门应用失败:', error)
        // 如果是认证错误，不显示错误消息，让主接口处理
        if (error.response?.status !== 401) {
          console.warn('获取热门应用数据失败，使用默认数据')
        }
      }
    }
    
    // 获取最近活动
    const getRecentActivity = async () => {
      try {
        const result = await statsAPI.getRecentActivity?.({ limit: 10 })
        if (result?.code === 0) {
          // 确保数据是数组格式，防止迭代错误
          const dataList = result.data
          recentActivities.value = Array.isArray(dataList) ? dataList : []
        } else if (result?.code === 401 || result?.code === 4001) {
          // Token无效，不显示错误，让主要的统计接口处理
          return
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
        // 如果是认证错误，不显示错误消息，让主接口处理
        if (error.response?.status !== 401) {
          console.warn('获取最近活动数据失败，使用默认数据')
        }
      }
    }
    
    // 初始化用户趋势图表
    const initUserTrendChart = async () => {
      if (!userTrendChart.value) return
      
      await nextTick()
      userTrendChartInstance = echarts.init(userTrendChart.value)
      
      // 获取真实的用户增长数据
      await loadUserGrowthData()
    }
    
    // 加载用户增长数据
    const loadUserGrowthData = async () => {
      try {
        const days = timeRange.value === 'today' ? 1 : (timeRange.value === 'week' ? 7 : 30)
        const result = await statsAPI.userGrowth({ days })
        
        console.log('用户增长API返回数据:', result.data)
        console.log('用户增长API返回数据详细:', JSON.stringify(result.data, null, 2))
        
        // 检查返回数据的结构
        let growthData = null
        if (Array.isArray(result.data)) {
          // 直接返回数组的情况
          growthData = result.data
          console.log('API直接返回数组数据')
        } else if (result.data && result.data.code === 0 && Array.isArray(result.data.data)) {
          // 返回包含code和data字段的对象的情况
          growthData = result.data.data
          console.log('API返回标准格式数据')
        }
        
        if (growthData && Array.isArray(growthData) && growthData.length > 0) {
          console.log('处理前的用户增长数据:', growthData)
          
          const dates = growthData.map(item => {
            const date = new Date(item.date)
            return `${date.getMonth() + 1}/${date.getDate()}`
          })
          const values = growthData.map(item => item.totalUsers)
          
          console.log('处理后的日期数据:', dates)
          console.log('处理后的用户数据:', values)
          
          const option = {
            title: {
              text: '用户增长趋势',
              textStyle: { fontSize: 14 }
            },
            tooltip: {
              trigger: 'axis',
              formatter: function(params) {
                const data = params[0]
                const originalData = growthData[data.dataIndex]
                return `${data.axisValueLabel}<br/>
                        累计用户: ${originalData.totalUsers}<br/>
                        新增用户: ${originalData.newUsers}`
              }
            },
            xAxis: {
              type: 'category',
              data: dates
            },
            yAxis: {
              type: 'value',
              name: '用户数'
            },
            series: [{
              name: '累计用户',
              data: values,
              type: 'line',
              smooth: true,
              itemStyle: {
                color: '#409eff'
              },
              areaStyle: {
                color: {
                  type: 'linear',
                  x: 0, y: 0, x2: 0, y2: 1,
                  colorStops: [{
                    offset: 0, color: 'rgba(64, 158, 255, 0.3)'
                  }, {
                    offset: 1, color: 'rgba(64, 158, 255, 0.1)'
                  }]
                }
              }
            }]
          }
          
          console.log('使用真实数据渲染用户增长图表')
          userTrendChartInstance?.setOption(option)
        } else {
          // 如果获取数据失败，使用默认数据
          console.log('API数据无效，使用默认数据')
          console.log('growthData:', growthData)
          console.log('Array.isArray(result.data):', Array.isArray(result.data))
          console.log('result.data长度:', Array.isArray(result.data) ? result.data.length : 'N/A')
          loadDefaultUserTrendChart()
        }
      } catch (error) {
        console.error('获取用户增长数据失败:', error)
        loadDefaultUserTrendChart()
      }
    }
    
    // 加载默认用户趋势图表
    const loadDefaultUserTrendChart = () => {
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
      
      userTrendChartInstance?.setOption(option)
    }
    
    // 初始化平台分布图表
    const initPlatformChart = async () => {
      if (!platformChart.value) {
        console.error('平台分布图表DOM元素未找到')
        return
      }
      
      console.log('开始初始化平台分布图表...')
      await nextTick()
      platformChartInstance = echarts.init(platformChart.value)
      console.log('平台分布图表实例已创建:', platformChartInstance)
      
      // 获取真实的平台分布数据
      await loadPlatformDistributionData()
    }
    
    // 加载平台分布数据
    const loadPlatformDistributionData = async () => {
      try {
        const result = await statsAPI.getPlatformDistribution({})
        console.log('平台分布API返回数据:', result.data)
        console.log('平台分布API返回数据详细:', JSON.stringify(result.data, null, 2))
        
        console.log('检查result结构:', {
          hasData: !!result.data,
          dataType: typeof result.data,
          isArray: Array.isArray(result.data),
          hasCode: result.data && 'code' in result.data,
          code: result.data && result.data.code,
          hasNestedData: result.data && result.data.data,
          nestedDataType: result.data && typeof result.data.data,
          nestedDataLength: result.data && result.data.data && result.data.data.length
        })
        
        // 处理可能的数据结构
        let distributionData = null
        if (result.data && result.data.code === 0 && result.data.data && result.data.data.length > 0) {
          distributionData = result.data.data
        } else if (Array.isArray(result.data) && result.data.length > 0) {
          distributionData = result.data
        }
        
        console.log('最终使用的distributionData:', distributionData)
        
        if (distributionData && distributionData.length > 0) {
          
          // 为不同平台设置不同的颜色
          const platformColors = {
            'wechat': '#1AAD19',   // 微信绿
            'alipay': '#1677FF',   // 支付宝蓝
            'douyin': '#FE2C55',   // 抖音红
            'baidu': '#2932E1',    // 百度蓝
            'ios': '#007AFF',      // iOS蓝
            'android': '#3DDC84'   // Android绿
          }
          
          const chartData = distributionData.map((item, index) => ({
            value: item.value,
            name: item.name,
            itemStyle: {
              color: platformColors[item.platform] || `hsl(${index * 60}, 70%, 50%)`
            }
          }))
          
          console.log('处理后的图表数据:', chartData)
          console.log('处理后的图表数据详细:', JSON.stringify(chartData, null, 2))
          
          const option = {
            title: {
              text: '平台分布',
              textStyle: { fontSize: 14 }
            },
            tooltip: {
              trigger: 'item',
              formatter: '{a} <br/>{b}: {c} ({d}%)'
            },
            legend: {
              orient: 'horizontal',
              bottom: '10px',
              left: 'center',
              itemGap: 20,
              textStyle: {
                fontSize: 12
              },
              data: distributionData.map(item => item.name)
            },
            series: [{
              name: '平台分布',
              type: 'pie',
              radius: ['35%', '60%'],
              center: ['50%', '45%'],
              avoidLabelOverlap: false,
              data: chartData,
              emphasis: {
                itemStyle: {
                  shadowBlur: 10,
                  shadowOffsetX: 0,
                  shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
              },
              label: {
                show: false,
                position: 'center'
              },
              labelLine: {
                show: false
              }
            }]
          }
          
          console.log('ECharts配置选项:', option)
          platformChartInstance?.setOption(option)
          console.log('图表实例状态:', platformChartInstance ? '已初始化' : '未初始化')
        } else {
          // 如果获取数据失败，使用默认数据
          loadDefaultPlatformChart()
        }
      } catch (error) {
        console.error('获取平台分布数据失败:', error)
        loadDefaultPlatformChart()
      }
    }
    
    // 加载默认平台分布图表
    const loadDefaultPlatformChart = () => {
      const option = {
        title: {
          text: '平台分布',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'item',
          formatter: '{a} <br/>{b}: {c} ({d}%)'
        },
        series: [{
          name: '平台分布',
          type: 'pie',
          radius: '50%',
          data: [
            { value: 40, name: '微信小程序', itemStyle: { color: '#1AAD19' } },
            { value: 30, name: '支付宝小程序', itemStyle: { color: '#1677FF' } },
            { value: 20, name: '抖音小程序', itemStyle: { color: '#FE2C55' } },
            { value: 10, name: '其他', itemStyle: { color: '#909399' } }
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
      
      platformChartInstance?.setOption(option)
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
    const handleTimeRangeChange = async () => {
      await refreshData()
      // 时间范围变化时也要重新加载图表数据
      if (userTrendChartInstance) {
        await loadUserGrowthData()
      }
      // 平台分布不随时间变化，不需要重新加载
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
    
    const refreshUserTrend = async () => {
      // 刷新用户趋势图表数据
      if (userTrendChartInstance) {
        await loadUserGrowthData()
        ElMessage.success('用户增长趋势数据已刷新')
      }
    }
    
    const refreshPlatformStats = async () => {
      // 刷新平台统计图表数据
      if (platformChartInstance) {
        await loadPlatformDistributionData()
        ElMessage.success('平台分布数据已刷新')
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
      exportData,
      loadUserGrowthData,
      loadPlatformDistributionData
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