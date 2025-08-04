<template>
  <div class="app-management">
    <div class="page-header">
      <h1>应用管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog">创建应用</el-button>
        <el-button @click="refreshApps">刷新</el-button>
      </div>
    </div>

    <!-- 搜索筛选 -->
    <div class="search-section">
      <el-form :model="searchForm" :inline="true">
        <el-form-item label="应用名称:">
          <el-input v-model="searchForm.appName" placeholder="输入应用名称" @keyup.enter="searchApps" style="width: 180px"></el-input>
        </el-form-item>
        <el-form-item label="应用ID:">
          <el-input v-model="searchForm.appId" placeholder="输入应用ID" @keyup.enter="searchApps" style="width: 180px"></el-input>
        </el-form-item>
        <el-form-item label="平台:">
          <el-select v-model="searchForm.platform" placeholder="选择平台" clearable style="width: 180px">
            <el-option label="微信小程序" value="wechat"></el-option>
            <el-option label="支付宝小程序" value="alipay"></el-option>
            <el-option label="抖音小程序" value="douyin"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchApps">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 应用统计 -->
    <div class="stats-section" v-if="appStats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ appStats.total }}</div>
              <div class="stat-label">应用总数</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ appStats.active }}</div>
              <div class="stat-label">活跃应用</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ appStats.newThisMonth }}</div>
              <div class="stat-label">本月新增</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ appStats.totalUsers }}</div>
              <div class="stat-label">总用户数</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 应用表格 -->
    <el-table 
      :data="appList" 
      style="width: 100%" 
      v-loading="loading"
      @sort-change="handleSortChange">
      <el-table-column prop="appName" label="应用名称" width="200" sortable="custom">
        <template #default="scope">
          <div class="app-info">
            <div class="app-name">{{ scope.row.appName }}</div>
            <div class="app-id">ID: {{ scope.row.appId }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="platform" label="平台" width="120" sortable="custom">
        <template #default="scope">
          <el-tag :type="getPlatformType(scope.row.platform)">
            {{ getPlatformText(scope.row.platform) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="channelAppId" label="渠道应用ID" width="180" show-overflow-tooltip>
      </el-table-column>
      <el-table-column label="用户统计" width="150">
        <template #default="scope">
          <div class="user-stats">
            <div>总用户: {{ scope.row.userCount || 0 }}</div>
            <div>今日活跃: {{ scope.row.dailyActive || 0 }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="排行榜数量" width="120">
        <template #default="scope">
          <el-tag type="info">{{ scope.row.leaderboardCount || 0 }} 个</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="160" sortable="custom">
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
            {{ scope.row.status === 'active' ? '正常' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250" fixed="right">
        <template #default="scope">
          <el-button type="text" @click="viewAppDetail(scope.row)">详情</el-button>
          <el-button type="text" @click="editApp(scope.row)">编辑</el-button>
          <el-button 
            type="text" 
            :class="scope.row.status === 'active' ? 'warning' : 'success'"
            @click="toggleAppStatus(scope.row)">
            {{ scope.row.status === 'active' ? '停用' : '启用' }}
          </el-button>
          <el-button type="text" class="danger" @click="deleteApp(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange">
      </el-pagination>
    </div>

    <!-- 创建/编辑应用对话框 -->
    <el-dialog 
      v-model="appDialog.visible" 
      :title="appDialog.isEdit ? '编辑应用' : '创建应用'" 
      width="600px">
      <el-form :model="appDialog.form" :rules="appRules" ref="appFormRef" label-width="120px">
        <el-form-item label="应用名称" prop="appName">
          <el-input v-model="appDialog.form.appName" placeholder="请输入应用名称"></el-input>
        </el-form-item>
        <el-form-item label="平台" prop="platform" style="width: 180px">
          <el-select v-model="appDialog.form.platform" placeholder="选择平台">
            <el-option label="微信小程序" value="wechat"></el-option>
            <el-option label="支付宝小程序" value="alipay"></el-option>
            <el-option label="抖音小程序" value="douyin"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="渠道应用ID" prop="channelAppId">
          <el-input v-model="appDialog.form.channelAppId" placeholder="请输入渠道应用ID"></el-input>
        </el-form-item>
        <el-form-item label="渠道应用密钥" prop="channelAppKey">
          <el-input v-model="appDialog.form.channelAppKey" type="textarea" :rows="3" placeholder="请输入渠道应用密钥"></el-input>
        </el-form-item>
        <el-form-item label="应用描述">
          <el-input v-model="appDialog.form.description" type="textarea" :rows="3" placeholder="请输入应用描述"></el-input>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch 
            v-model="appDialog.form.status" 
            active-value="active" 
            inactive-value="inactive"
            active-text="启用" 
            inactive-text="停用">
          </el-switch>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="appDialog.visible = false">取消</el-button>
          <el-button type="primary" @click="saveApp">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 应用详情对话框 -->
    <el-dialog v-model="detailDialog.visible" title="应用详情" width="800px">
      <div v-if="detailDialog.app">
        <el-tabs v-model="detailDialog.activeTab">
          <el-tab-pane label="基本信息" name="basic">
            <div class="detail-content">
              <el-descriptions border :column="2">
                <el-descriptions-item label="应用名称">{{ detailDialog.app.appName }}</el-descriptions-item>
                <el-descriptions-item label="应用ID">{{ detailDialog.app.appId }}</el-descriptions-item>
                <el-descriptions-item label="平台">{{ getPlatformText(detailDialog.app.platform) }}</el-descriptions-item>
                <el-descriptions-item label="渠道应用ID">{{ detailDialog.app.channelAppId }}</el-descriptions-item>
                <el-descriptions-item label="创建时间">{{ detailDialog.app.createTime }}</el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag :type="detailDialog.app.status === 'active' ? 'success' : 'danger'">
                    {{ detailDialog.app.status === 'active' ? '正常' : '停用' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="描述" :span="2">{{ detailDialog.app.description || '无' }}</el-descriptions-item>
              </el-descriptions>
            </div>
          </el-tab-pane>
          <el-tab-pane label="用户统计" name="users">
            <div class="stats-content">
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-card>
                    <div class="stat-item">
                      <div class="stat-value">{{ detailDialog.stats?.totalUsers || 0 }}</div>
                      <div class="stat-label">总用户数</div>
                    </div>
                  </el-card>
                </el-col>
                <el-col :span="8">
                  <el-card>
                    <div class="stat-item">
                      <div class="stat-value">{{ detailDialog.stats?.dailyActive || 0 }}</div>
                      <div class="stat-label">今日活跃</div>
                    </div>
                  </el-card>
                </el-col>
                <el-col :span="8">
                  <el-card>
                    <div class="stat-item">
                      <div class="stat-value">{{ detailDialog.stats?.newToday || 0 }}</div>
                      <div class="stat-label">今日新增</div>
                    </div>
                  </el-card>
                </el-col>
              </el-row>
            </div>
          </el-tab-pane>
          <el-tab-pane label="排行榜" name="leaderboards">
            <div class="leaderboards-content">
              <el-table :data="detailDialog.leaderboards" style="width: 100%">
                <el-table-column prop="leaderboardType" label="排行榜类型" width="150"></el-table-column>
                <el-table-column prop="name" label="名称" width="200"></el-table-column>
                <el-table-column label="排序方式" width="120">
                  <template #default="scope">
                    <el-tag :type="scope.row.sort === 1 ? 'success' : 'info'">
                      {{ scope.row.sort === 1 ? '降序' : '升序' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="参与人数" width="120">
                  <template #default="scope">
                    {{ scope.row.participantCount || 0 }}
                  </template>
                </el-table-column>
                <el-table-column prop="createTime" label="创建时间"></el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { appAPI, statsAPI } from '../services/api.js'

export default {
  name: 'AppManagement',
  setup() {
    const loading = ref(false)
    const appList = ref([])
    const appStats = ref(null)
    
    const searchForm = reactive({
      appName: '',
      appId: '',
      platform: ''
    })
    
    const pagination = reactive({
      current: 1,
      pageSize: 20,
      total: 0
    })
    
    const appDialog = reactive({
      visible: false,
      isEdit: false,
      form: {
        appName: '',
        platform: '',
        channelAppId: '',
        channelAppKey: '',
        description: '',
        status: 'active'
      }
    })
    
    const detailDialog = reactive({
      visible: false,
      activeTab: 'basic',
      app: null,
      stats: null,
      leaderboards: []
    })
    
    const appRules = {
      appName: [
        { required: true, message: '请输入应用名称', trigger: 'blur' }
      ],
      platform: [
        { required: true, message: '请选择平台', trigger: 'change' }
      ],
      channelAppId: [
        { required: true, message: '请输入渠道应用ID', trigger: 'blur' }
      ],
      channelAppKey: [
        { required: true, message: '请输入渠道应用密钥', trigger: 'blur' }
      ]
    }
    
    // 获取应用列表
    const getAppList = async () => {
      loading.value = true
      try {
        const params = {
          page: pagination.current,
          pageSize: pagination.pageSize,
          ...searchForm
        }
        
        const result = await appAPI.getAll(params)
        if (result.code === 0) {
          // 确保数据是数组格式，防止迭代错误
          const dataList = result.data?.list
          appList.value = Array.isArray(dataList) ? dataList : []
          pagination.total = result.data?.total || 0
        } else {
          appList.value = []
          ElMessage.error(result.msg || '获取应用列表失败')
        }
      } catch (error) {
        console.error('获取应用列表失败:', error)
        appList.value = []
        ElMessage.error('获取应用列表失败')
      } finally {
        loading.value = false
      }
    }
    
    // 获取应用统计
    const getAppStats = async () => {
      try {
        const result = await statsAPI.getDashboardStats()
        if (result.code === 0) {
          appStats.value = result.data?.apps
        }
      } catch (error) {
        console.error('获取应用统计失败:', error)
      }
    }
    
    // 显示创建对话框
    const showCreateDialog = () => {
      appDialog.isEdit = false
      appDialog.form = {
        appName: '',
        platform: '',
        channelAppId: '',
        channelAppKey: '',
        description: '',
        status: 'active'
      }
      appDialog.visible = true
    }
    
    // 编辑应用
    const editApp = (app) => {
      appDialog.isEdit = true
      appDialog.form = { ...app }
      appDialog.visible = true
    }
    
    // 保存应用
    const saveApp = async () => {
      try {
        const apiCall = appDialog.isEdit ? appAPI.update : appAPI.init
        const data = { ...appDialog.form }
        
        // 如果是创建，使用initApp接口的参数格式
        if (!appDialog.isEdit) {
          data.appId = data.channelAppId
          data.appKey = data.channelAppKey
        }
        
        const result = await apiCall(data)
        
        if (result.code === 0) {
          ElMessage.success(appDialog.isEdit ? '更新成功' : '创建成功')
          appDialog.visible = false
          await getAppList()
          await getAppStats()
        } else {
          ElMessage.error(result.msg || '操作失败')
        }
      } catch (error) {
        console.error('保存应用失败:', error)
        ElMessage.error('操作失败')
      }
    }
    
    // 切换应用状态
    const toggleAppStatus = async (app) => {
      const action = app.status === 'active' ? '停用' : '启用'
      try {
        await ElMessageBox.confirm(`确定要${action}应用 "${app.appName}" 吗？`, '确认操作')
        
        const result = await appAPI.update({
          ...app,
          status: app.status === 'active' ? 'inactive' : 'active'
        })
        
        if (result.code === 0) {
          ElMessage.success(`${action}成功`)
          await getAppList()
        } else {
          ElMessage.error(result.msg || `${action}失败`)
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error(`${action}应用失败:`, error)
          ElMessage.error(`${action}失败`)
        }
      }
    }
    
    // 删除应用
    const deleteApp = async (app) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除应用 "${app.appName}" 吗？此操作将删除所有相关数据，不可恢复！`, 
          '危险操作', 
          { type: 'warning' }
        )
        
        const result = await appAPI.delete({
          appId: app.appId
        })
        
        if (result.code === 0) {
          ElMessage.success('删除成功')
          await getAppList()
          await getAppStats()
        } else {
          ElMessage.error(result.msg || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除应用失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 查看应用详情
    const viewAppDetail = async (app) => {
      detailDialog.app = app
      detailDialog.activeTab = 'basic'
      detailDialog.visible = true
      
      // 加载详细统计数据
      try {
        const [statsResult, leaderboardResult] = await Promise.all([
          statsAPI.getAppStats(app.appId),
          appAPI.query({ appId: app.appId })
        ])
        
        if (statsResult.code === 0) {
          detailDialog.stats = statsResult.data
        }
        
        if (leaderboardResult.code === 0) {
          detailDialog.leaderboards = leaderboardResult.data?.leaderBoardList || []
        }
      } catch (error) {
        console.error('加载应用详情失败:', error)
      }
    }
    
    // 工具函数
    const getPlatformText = (platform) => {
      const platforms = {
        'wechat': '微信小程序',
        'alipay': '支付宝小程序',
        'douyin': '抖音小程序'
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
    
    // 事件处理
    const refreshApps = () => {
      getAppList()
      getAppStats()
    }
    
    const searchApps = () => {
      pagination.current = 1
      getAppList()
    }
    
    const resetSearch = () => {
      searchForm.appName = ''
      searchForm.appId = ''
      searchForm.platform = ''
      pagination.current = 1
      getAppList()
    }
    
    const handleSortChange = ({ column, prop, order }) => {
      // 实现排序逻辑
      getAppList()
    }
    
    const handleSizeChange = (val) => {
      pagination.pageSize = val
      pagination.current = 1
      getAppList()
    }
    
    const handleCurrentChange = (val) => {
      pagination.current = val
      getAppList()
    }
    
    onMounted(() => {
      getAppList()
      getAppStats()
    })
    
    return {
      loading,
      appList,
      appStats,
      searchForm,
      pagination,
      appDialog,
      detailDialog,
      appRules,
      getAppList,
      showCreateDialog,
      editApp,
      saveApp,
      toggleAppStatus,
      deleteApp,
      viewAppDetail,
      getPlatformText,
      getPlatformType,
      refreshApps,
      searchApps,
      resetSearch,
      handleSortChange,
      handleSizeChange,
      handleCurrentChange
    }
  }
}
</script>

<style scoped>
.app-management {
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

.search-section {
  background: #f5f5f5;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.stats-section {
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

.app-info {
  display: flex;
  flex-direction: column;
}

.app-name {
  font-weight: bold;
  font-size: 14px;
  margin-bottom: 4px;
}

.app-id {
  font-size: 12px;
  color: #999;
}

.user-stats {
  font-size: 12px;
  line-height: 1.5;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.detail-content {
  padding: 20px 0;
}

.stats-content {
  padding: 20px 0;
}

.leaderboards-content {
  padding: 20px 0;
}

.el-button.danger {
  color: #f56c6c;
}

.el-button.warning {
  color: #e6a23c;
}

.el-button.success {
  color: #67c23a;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>