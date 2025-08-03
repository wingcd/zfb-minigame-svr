<template>
  <div>
    <h1>仪表盘</h1>
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <h3>总应用数</h3>
            <p class="stat-number">{{ stats.totalApps }}</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <h3>总排行榜</h3>
            <p class="stat-number">{{ stats.totalLeaderboards }}</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <h3>总用户数</h3>
            <p class="stat-number">{{ stats.totalUsers }}</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <h3>今日活跃</h3>
            <p class="stat-number">{{ stats.todayActive }}</p>

          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { appAPI, leaderboardAPI } from '../services/api.js'

const stats = ref({
  totalApps: 0,
  totalLeaderboards: 0,
  totalUsers: 0,
  todayActive: 0
})

const loadStats = async () => {
  try {
    // 使用封装的API服务获取统计数据
    const [appsData, leaderboardsData] = await Promise.all([
      appAPI.getAllApps(),
      leaderboardAPI.getAllLeaderboards?.() || Promise.resolve([])
    ])
    
    stats.value = {
      totalApps: appsData.length || 0,
      totalLeaderboards: leaderboardsData.length || 0,
      totalUsers: 1568, // 需要从用户API获取
      todayActive: 342  // 需要从用户API获取
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.stat-card {
  text-align: center;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin: 10px 0;
}
</style>