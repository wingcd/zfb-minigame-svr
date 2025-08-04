<template>
  <div class="leaderboard-management">
    <div class="page-header">
      <h1>排行榜管理</h1>
      <div class="header-actions">
        <el-select v-model="selectedAppId" placeholder="选择应用" @change="handleAppChange" style="width: 200px; margin-right: 10px;">
          <template v-if="appList && appList.length > 0">
            <el-option
              v-for="app in appList"
              :key="app.appId || app.id || Math.random()"
              :label="app.appName || '未命名应用'"
              :value="app.appId">
            </el-option>
          </template>
        </el-select>
        <el-button type="primary" @click="showCreateDialog">创建排行榜</el-button>
        <el-button @click="refreshData">刷新</el-button>
      </div>
    </div>

    <!-- 排行榜配置管理 -->
    <div class="leaderboard-config-section">
      <h2>排行榜配置</h2>
      <el-table :data="leaderboardConfigs" style="width: 100%" v-loading="configLoading">
        <el-table-column prop="leaderboardType" label="排行榜类型" width="150">
        </el-table-column>
        <el-table-column prop="name" label="排行榜名称" width="200">
        </el-table-column>
        <el-table-column label="分数类型" width="140">
          <template #default="scope">
            <el-tag :type="scope.row.scoreType === 'higher_better' ? 'success' : 'info'">
              {{ scope.row.scoreType === 'higher_better' ? '分数越高越好' : '分数越低越好' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" width="150" show-overflow-tooltip>
        </el-table-column>
        <el-table-column label="重置类型" width="120">
          <template #default="scope">
            <el-tag :type="getResetTypeTagType(scope.row.resetType)">
              {{ getResetTypeText(scope.row.resetType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="下次重置" width="160" show-overflow-tooltip>
          <template #default="scope">
            <span v-if="scope.row.resetType === 'permanent'">永不重置</span>
            <span v-else>{{ scope.row.resetTime || '未设置' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="maxRank" label="最大排名数" width="100">
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160">
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'danger'">
              {{ scope.row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button type="text" @click="editConfig(scope.row)">编辑</el-button>
            <el-button type="text" @click="viewLeaderboard(scope.row)">查看排行</el-button>
            <el-button type="text" class="danger" @click="deleteConfig(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 排行榜数据查看 -->
    <div class="leaderboard-data-section" v-if="selectedLeaderboard">
      <div class="section-header">
        <h2>{{ selectedLeaderboard.name }} - 排行数据</h2>
        <div class="header-actions">
          <el-select v-model="rankParams.count" @change="loadLeaderboardData" style="width: 120px; margin-right: 10px;">
            <el-option label="前10名" :value="10"></el-option>
            <el-option label="前20名" :value="20"></el-option>
            <el-option label="前50名" :value="50"></el-option>
            <el-option label="前100名" :value="100"></el-option>
          </el-select>
          <el-button @click="loadLeaderboardData">刷新数据</el-button>
        </div>
      </div>

      <!-- 搜索特定玩家 -->
      <div class="search-player">
        <el-form :model="playerSearchForm" :inline="true">
          <el-form-item label="搜索玩家:">
            <el-input v-model="playerSearchForm.playerId" placeholder="输入玩家ID" @keyup.enter="searchPlayerScore"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="searchPlayerScore">搜索</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 排行榜统计 -->
      <div class="leaderboard-stats" v-if="leaderboardStats">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ leaderboardStats.totalPlayers }}</div>
                <div class="stat-label">参与玩家</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ leaderboardStats.highestScore }}</div>
                <div class="stat-label">最高分数</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ leaderboardStats.averageScore }}</div>
                <div class="stat-label">平均分数</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ leaderboardStats.todaySubmissions }}</div>
                <div class="stat-label">今日提交</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 排行榜数据表格 -->
      <el-table :data="leaderboardData" style="width: 100%" v-loading="dataLoading">
        <el-table-column label="排名" width="80">
          <template #default="{ $index }">
            <div class="rank-badge" :class="getRankClass($index + 1)">
              {{ $index + 1 }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="openId" label="玩家ID" width="120">
        </el-table-column>
        <el-table-column label="玩家信息" width="200">
          <template #default="scope">
            <div v-if="scope.row.userInfo" class="player-info">
              <el-avatar v-if="scope.row.userInfo.avatarUrl" :src="scope.row.userInfo.avatarUrl" size="small"></el-avatar>
              <span class="player-name">{{ scope.row.userInfo.nickName || '未设置' }}</span>
            </div>
            <span v-else>无信息</span>
          </template>
        </el-table-column>
        <el-table-column prop="score" label="分数" width="120" sortable>
          <template #default="scope">
            <div class="score-display">
              <span class="score-number">{{ scope.row.score }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="gmtCreate" label="首次记录" width="160">
        </el-table-column>
        <el-table-column prop="gmtModify" label="最后更新" width="160">
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-button type="text" @click="editScore(scope.row)">编辑分数</el-button>
            <el-button type="text" class="danger" @click="deleteScore(scope.row)">删除记录</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建/编辑排行榜配置对话框 -->
    <el-dialog 
      v-model="configDialog.visible" 
      :title="configDialog.isEdit ? '编辑排行榜配置' : '创建排行榜配置'" 
      width="500px">
      <el-form :model="configDialog.form" :rules="configRules" ref="configFormRef" label-width="120px">
        <el-form-item label="排行榜类型" prop="leaderboardType">
          <el-input v-model="configDialog.form.leaderboardType" placeholder="如: easy, hard, daily"></el-input>
        </el-form-item>
        <el-form-item label="排行榜名称" prop="name">
          <el-input v-model="configDialog.form.name" placeholder="排行榜显示名称"></el-input>
        </el-form-item>
        <el-form-item label="分数类型" prop="scoreType">
          <el-select v-model="configDialog.form.scoreType" style="width: 100%">
            <el-option label="分数越高越好(降序)" value="higher_better"></el-option>
            <el-option label="分数越低越好(升序)" value="lower_better"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="排行榜描述" prop="description">
          <el-input v-model="configDialog.form.description" placeholder="排行榜描述信息" type="textarea" :rows="2"></el-input>
        </el-form-item>
        <el-form-item label="最大排名数" prop="maxRank">
          <el-input-number v-model="configDialog.form.maxRank" :min="10" :max="10000" style="width: 100%"></el-input-number>
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-input v-model="configDialog.form.category" placeholder="排行榜分类"></el-input>
        </el-form-item>
        <el-form-item label="重置类型" prop="resetType">
          <el-select v-model="configDialog.form.resetType" style="width: 100%" @change="handleResetTypeChange">
            <el-option label="永久保存" value="permanent"></el-option>
            <el-option label="每日重置" value="daily"></el-option>
            <el-option label="每周重置" value="weekly"></el-option>
            <el-option label="每月重置" value="monthly"></el-option>
            <el-option label="自定义间隔" value="custom"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item 
          v-if="configDialog.form.resetType === 'custom'" 
          label="重置间隔(小时)" 
          prop="resetValue">
          <el-input-number 
            v-model="configDialog.form.resetValue" 
            :min="1" 
            :max="8760" 
            style="width: 100%"
            placeholder="请输入重置间隔小时数">
          </el-input-number>
        </el-form-item>
        <el-form-item label="状态" prop="enabled">
          <el-switch v-model="configDialog.form.enabled"></el-switch>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="configDialog.visible = false">取消</el-button>
          <el-button type="primary" @click="saveConfig">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 编辑分数对话框 -->
    <el-dialog v-model="scoreDialog.visible" title="编辑分数" width="400px">
      <el-form :model="scoreDialog.form" label-width="100px">
        <el-form-item label="玩家ID">
          <el-input v-model="scoreDialog.form.playerId" disabled></el-input>
        </el-form-item>
        <el-form-item label="当前分数">
          <el-input-number v-model="scoreDialog.form.score" :min="0" style="width: 100%"></el-input-number>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="scoreDialog.visible = false">取消</el-button>
          <el-button type="primary" @click="saveScore">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { leaderboardAPI, appAPI, statsAPI } from '../services/api.js'

export default {
  name: 'LeaderboardManagement',
  setup() {
    const configLoading = ref(false)
    const dataLoading = ref(false)
    const appList = ref([])
    const selectedAppId = ref('')
    const leaderboardConfigs = ref([])
    const selectedLeaderboard = ref(null)
    const leaderboardData = ref([])
    const leaderboardStats = ref(null)
    
    const rankParams = reactive({
      startRank: 0,
      count: 20
    })
    
    const playerSearchForm = reactive({
      playerId: ''
    })
    
    const configDialog = reactive({
      visible: false,
      isEdit: false,
      form: {
        leaderboardType: '',
        name: '',
        description: '',
        scoreType: 'higher_better',
        maxRank: 100,
        category: 'default',
        resetType: 'permanent',
        resetValue: 24,
        enabled: true
      }
    })
    
    const scoreDialog = reactive({
      visible: false,
      form: {
        playerId: '',
        score: 0
      },
      originalData: null
    })
    
    const configRules = {
      leaderboardType: [
        { required: true, message: '请输入排行榜类型', trigger: 'blur' }
      ],
      name: [
        { required: true, message: '请输入排行榜名称', trigger: 'blur' }
      ],
      scoreType: [
        { required: true, message: '请选择分数类型', trigger: 'change' }
      ],
      maxRank: [
        { required: true, message: '请输入最大排名数', trigger: 'blur' },
        { type: 'number', min: 10, max: 10000, message: '最大排名数必须在10-10000之间', trigger: 'blur' }
      ]
    }
    
    // 获取应用列表
    const getAppList = async () => {
      try {
        const result = await appAPI.getAll()
        if (result.code === 0) {
          // 过滤掉无效的应用数据，确保每个应用都有有效的appId和appName
          const dataList = result.data?.list
          const validApps = Array.isArray(dataList) ? dataList : []
          
          // 确保数组是响应式的，并且在设置前清空之前的数据
          appList.value = []
          // 使用 nextTick 确保 DOM 更新
          await new Promise(resolve => setTimeout(resolve, 0))
          appList.value = validApps
          
          if (appList.value.length > 0) {
            selectedAppId.value = appList.value[0].appId
            await loadLeaderboardConfigs()
          }
        } else {
          appList.value = []
          ElMessage.error(result.msg || '获取应用列表失败')
        }
      } catch (error) {
        console.error('获取应用列表失败:', error)
        appList.value = []
        ElMessage.error('获取应用列表失败')
      }
    }
    
    // 加载排行榜配置
    const loadLeaderboardConfigs = async () => {
      if (!selectedAppId.value) return
      
      configLoading.value = true
      try {
        const result = await leaderboardAPI.getAll({
          appId: selectedAppId.value
        })
        
        if (result.code === 0) {
          // 确保数据是数组格式，防止迭代错误
          const dataList = result.data.list;
          leaderboardConfigs.value = Array.isArray(dataList) ? dataList : []
        } else {
          leaderboardConfigs.value = []
          ElMessage.error(result.msg || '获取排行榜配置失败')
        }
      } catch (error) {
        console.error('获取排行榜配置失败:', error)
        leaderboardConfigs.value = []
        ElMessage.error('获取排行榜配置失败')
      } finally {
        configLoading.value = false
      }
    }
    
    // 查看排行榜数据
    const viewLeaderboard = async (config) => {
      selectedLeaderboard.value = config
      await loadLeaderboardData()
      await loadLeaderboardStats()
    }
    
    // 加载排行榜数据
    const loadLeaderboardData = async () => {
      if (!selectedLeaderboard.value) return
      
      dataLoading.value = true
      try {
        const result = await leaderboardAPI.getData({
          appId: selectedAppId.value,
          leaderboardId: selectedLeaderboard.value.leaderboardId,
          leaderboardType: selectedLeaderboard.value.leaderboardType,
          offset: rankParams.startRank,
          limit: rankParams.count,
          includeUserInfo: true
        })
        
        if (result.code === 0) {
          // 确保数据是数组格式，防止迭代错误
          const dataList = result.data?.scores
          leaderboardData.value = Array.isArray(dataList) ? dataList : []
        } else {
          leaderboardData.value = []
          ElMessage.error(result.msg || '获取排行榜数据失败')
        }
      } catch (error) {
        console.error('获取排行榜数据失败:', error)
        leaderboardData.value = []
        ElMessage.error('获取排行榜数据失败')
      } finally {
        dataLoading.value = false
      }
    }
    
    // 加载排行榜统计
    const loadLeaderboardStats = async () => {
      if (!selectedAppId.value) return
      
      try {
        const result = await statsAPI.leaderboardStats(selectedAppId.value)
        if (result.code === 0) {
          leaderboardStats.value = result.data
        }
      } catch (error) {
        console.error('获取排行榜统计失败:', error)
      }
    }
    
    // 搜索玩家分数
    const searchPlayerScore = async () => {
      if (!playerSearchForm.playerId || !selectedLeaderboard.value) {
        ElMessage.warning('请输入玩家ID')
        return
      }
      
      try {
        const result = await leaderboardAPI.queryScore({
          appId: selectedAppId.value,
          playerId: playerSearchForm.playerId,
          leaderboardId: selectedLeaderboard.value.leaderboardId,
          leaderboardType: selectedLeaderboard.value.leaderboardType,
        })
        
        if (result.code === 0) {
          // 显示搜索结果
          leaderboardData.value = [result.data]
          ElMessage.success('查询成功')
        } else {
          ElMessage.error(result.msg || '查询失败')
        }
      } catch (error) {
        console.error('搜索玩家分数失败:', error)
        ElMessage.error('搜索失败')
      }
    }
    
    // 显示创建对话框
    const showCreateDialog = () => {
      configDialog.isEdit = false
      configDialog.form = {
        leaderboardId: '',
        leaderboardType: '',
        name: '',
        description: '',
        scoreType: 'higher_better',
        maxRank: 100,
        category: 'default',
        resetType: 'permanent',
        resetValue: 24,
        enabled: true
      }
      configDialog.visible = true
    }
    
    // 编辑配置
    const editConfig = (config) => {
      configDialog.isEdit = true
      configDialog.form = { ...config }
      configDialog.visible = true
    }
    
    // 保存配置
    const saveConfig = async () => {
      try {
        const apiCall = configDialog.isEdit 
          ? leaderboardAPI.update 
          : leaderboardAPI.create
        
        const data = {
          appId: selectedAppId.value,
          ...configDialog.form
        }
        
        const result = await apiCall(data)
        
        if (result.code === 0) {
          ElMessage.success(configDialog.isEdit ? '更新成功' : '创建成功')
          configDialog.visible = false
          await loadLeaderboardConfigs()
        } else {
          ElMessage.error(result.msg || '操作失败')
        }
      } catch (error) {
        console.error('保存配置失败:', error)
        ElMessage.error('操作失败')
      }
    }
    
    // 删除配置
    const deleteConfig = async (config) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除排行榜 "${config.name}" 吗？此操作将同时删除所有相关数据！`, 
          '危险操作', 
          { type: 'warning' }
        )
        
        const result = await leaderboardAPI.delete({
          appId: selectedAppId.value,
          leaderboardType: config.leaderboardType
        })
        
        if (result.code === 0) {
          ElMessage.success('删除成功')
          await loadLeaderboardConfigs()
          if (selectedLeaderboard.value?.leaderboardId === config.leaderboardId) {
            selectedLeaderboard.value = null
            leaderboardData.value = []
          }
        } else {
          ElMessage.error(result.msg || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除配置失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 编辑分数
    const editScore = (scoreData) => {
      scoreDialog.form = {
        playerId: scoreData.openId,
        score: scoreData.score
      }
      scoreDialog.originalData = scoreData
      scoreDialog.visible = true
    }
    
    // 保存分数
    const saveScore = async () => {
      try {
        const result = await leaderboardAPI.commitScore({
          appId: selectedAppId.value,
          playerId: scoreDialog.form.playerId,
          leaderboardId: selectedLeaderboard.value.leaderboardId,
          leaderboardType: selectedLeaderboard.value.leaderboardType,
          score: scoreDialog.form.score,
          playerInfo: scoreDialog.originalData.userInfo || {}
        })
        
        if (result.code === 0) {
          ElMessage.success('更新成功')
          scoreDialog.visible = false
          await loadLeaderboardData()
        } else {
          ElMessage.error(result.msg || '更新失败')
        }
      } catch (error) {
        console.error('更新分数失败:', error)
        ElMessage.error('更新失败')
      }
    }
    
    // 删除分数记录
    const deleteScore = async (scoreData) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除玩家 ${scoreData.openId} 的分数记录吗？`, 
          '确认删除', 
          { type: 'warning' }
        )
        
        const result = await leaderboardAPI.deleteScore({
          appId: selectedAppId.value,
          playerId: scoreData.openId,
          leaderboardId: selectedLeaderboard.value.leaderboardId,
          leaderboardType: selectedLeaderboard.value.leaderboardType,
        })
        
        if (result.code === 0) {
          ElMessage.success('删除成功')
          await loadLeaderboardData()
        } else {
          ElMessage.error(result.msg || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除分数失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 工具函数
    const getRankClass = (rank) => {
      if (rank === 1) return 'rank-gold'
      if (rank === 2) return 'rank-silver'
      if (rank === 3) return 'rank-bronze'
      return 'rank-normal'
    }
    
    const getResetTypeText = (resetType) => {
      const typeMap = {
        'permanent': '永久保存',
        'daily': '每日重置',
        'weekly': '每周重置',
        'monthly': '每月重置',
        'custom': '自定义间隔'
      }
      return typeMap[resetType] || '未知'
    }
    
    const getResetTypeTagType = (resetType) => {
      const typeMap = {
        'permanent': 'info',
        'daily': 'success',
        'weekly': 'warning',
        'monthly': 'danger',
        'custom': 'primary'
      }
      return typeMap[resetType] || 'info'
    }
    
    const handleResetTypeChange = () => {
      if (configDialog.form.resetType !== 'custom') {
        configDialog.form.resetValue = null
      } else {
        configDialog.form.resetValue = 24
      }
    }
    
    // 事件处理
    const handleAppChange = () => {
      selectedLeaderboard.value = null
      leaderboardData.value = []
      loadLeaderboardConfigs()
    }
    
    const refreshData = () => {
      loadLeaderboardConfigs()
      if (selectedLeaderboard.value) {
        loadLeaderboardData()
        loadLeaderboardStats()
      }
    }
    
    onMounted(() => {
      getAppList()
    })
    
    return {
      configLoading,
      dataLoading,
      appList,
      selectedAppId,
      leaderboardConfigs,
      selectedLeaderboard,
      leaderboardData,
      leaderboardStats,
      rankParams,
      playerSearchForm,
      configDialog,
      scoreDialog,
      configRules,
      viewLeaderboard,
      loadLeaderboardData,
      searchPlayerScore,
      showCreateDialog,
      editConfig,
      saveConfig,
      deleteConfig,
      editScore,
      saveScore,
      deleteScore,
      getRankClass,
      getResetTypeText,
      getResetTypeTagType,
      handleResetTypeChange,
      handleAppChange,
      refreshData
    }
  }
}
</script>

<style scoped>
.leaderboard-management {
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
}

.leaderboard-config-section {
  margin-bottom: 30px;
}

.leaderboard-config-section h2 {
  color: #333;
  margin-bottom: 15px;
}

.leaderboard-data-section {
  border-top: 1px solid #e4e7ed;
  padding-top: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h2 {
  margin: 0;
  color: #333;
}

.search-player {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.leaderboard-stats {
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

.rank-badge {
  display: inline-block;
  width: 30px;
  height: 30px;
  line-height: 30px;
  text-align: center;
  border-radius: 50%;
  font-weight: bold;
  color: white;
}

.rank-gold {
  background: linear-gradient(45deg, #ffd700, #ffed4e);
  color: #333;
}

.rank-silver {
  background: linear-gradient(45deg, #c0c0c0, #e8e8e8);
  color: #333;
}

.rank-bronze {
  background: linear-gradient(45deg, #cd7f32, #daa520);
  color: white;
}

.rank-normal {
  background: #909399;
}

.player-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.player-name {
  font-weight: 500;
}

.score-display {
  font-weight: bold;
  color: #e6a23c;
}

.score-number {
  font-size: 16px;
}

.el-button.danger {
  color: #f56c6c;
}
</style>