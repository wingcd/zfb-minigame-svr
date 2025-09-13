<template>
  <div class="yalla-config-management">
    <div class="page-header">
      <h1>Yalla配置管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          创建配置
        </el-button>
        <el-button @click="loadConfigs">刷新</el-button>
      </div>
    </div>

    <div v-if="selectedAppId" class="content">
      <!-- 统计概览 -->
      <div class="stats-section" v-if="configStats">
        <div class="stats-cards">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">{{ configStats.totalConfigs || 0 }}</div>
              <div class="stats-label">配置总数</div>
            </div>
          </el-card>
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">{{ configStats.activeConfigs || 0 }}</div>
              <div class="stats-label">启用配置</div>
            </div>
          </el-card>
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">{{ configStats.sandboxConfigs || 0 }}</div>
              <div class="stats-label">沙盒配置</div>
            </div>
          </el-card>
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">{{ configStats.productionConfigs || 0 }}</div>
              <div class="stats-label">生产配置</div>
            </div>
          </el-card>
        </div>
      </div>

      <!-- 筛选区域 -->
      <div class="filters-section">
        <el-form :inline="true" class="filter-form">
          <el-form-item label="应用筛选">
            <el-input 
              v-model="filterAppId" 
              @input="loadConfigs"
              placeholder="搜索应用ID" 
              clearable 
              style="width: 200px;">
            </el-input>
          </el-form-item>
          <el-form-item label="环境">
            <el-select v-model="filterEnvironment" @change="loadConfigs" placeholder="全部环境" clearable style="width: 150px;">
              <el-option label="沙盒环境" value="sandbox"></el-option>
              <el-option label="生产环境" value="production"></el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="filterStatus" @change="loadConfigs" placeholder="全部状态" clearable style="width: 120px;">
              <el-option label="启用" value="true"></el-option>
              <el-option label="禁用" value="false"></el-option>
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <!-- 配置管理区域 -->
      <div class="config-section">
        <h2>Yalla配置列表</h2>
        
        <el-table 
          :data="configList" 
          v-loading="loading"
          style="width: 100%"
          @sort-change="handleSortChange">
          <el-table-column prop="appId" label="应用ID" width="200" sortable="custom">
            <template #default="scope">
              <el-tag>{{ scope.row.appId }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="appGameId" label="游戏ID" width="200">
            <template #default="scope">
              <el-tag>{{ scope.row.appGameId }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="secretKey" label="密钥" width="150">
            <template #default="scope">
              <div class="sensitive-field">
                <span v-if="!scope.row.showSecretKey">{{ maskSensitiveData(scope.row.secretKey) }}</span>
                <span v-else>{{ scope.row.secretKey }}</span>
                <el-button 
                  link 
                  size="small" 
                  @click="toggleShowSensitive(scope.row, 'showSecretKey')"
                  style="margin-left: 5px;">
                  <el-icon><View v-if="!scope.row.showSecretKey" /><Hide v-else /></el-icon>
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="baseUrl" label="基础URL" width="200" show-overflow-tooltip>
          </el-table-column>
          <el-table-column prop="pushUrl" label="推送URL" width="200" show-overflow-tooltip>
          </el-table-column>
          <el-table-column prop="environment" label="环境" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.environment === 'production' ? 'danger' : 'warning'">
                {{ scope.row.environment === 'production' ? '生产' : '沙盒' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip>
          </el-table-column>
          <el-table-column prop="isActive" label="状态" width="80" align="center">
            <template #default="scope">
              <el-switch 
                v-model="scope.row.isActive" 
                @change="toggleConfigStatus(scope.row)"
                :loading="scope.row.updating">
              </el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="createTime" label="创建时间" width="180" sortable="custom" align="center">
            <template #default="scope">
              {{ formatDate(scope.row.createTime) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right" align="center">
            <template #default="scope">
              <el-button link @click="openEditDialog(scope.row)">编辑</el-button>
              <el-button link @click="testConnection(scope.row)">测试连接</el-button>
              <el-button link class="danger" @click="deleteConfig(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <div class="pagination">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="pagination.total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadConfigs"
            @current-change="loadConfigs">
          </el-pagination>
        </div>
      </div>
    </div>

    <!-- 创建/编辑配置对话框 -->
    <el-dialog 
      v-model="configDialog.visible" 
      :title="configDialog.isEdit ? '编辑Yalla配置' : '创建Yalla配置'" 
      width="700px"
      @close="resetConfigDialog">
      <el-form :model="configDialog.form" :rules="configRules" ref="configFormRef" label-width="120px">
        <el-form-item label="应用ID" prop="appId">
          <el-input 
            v-model="configDialog.form.appId" 
            placeholder="输入Yalla应用ID"
            :disabled="configDialog.isEdit">
          </el-input>
        </el-form-item>
        <el-form-item label="游戏ID" prop="appGameId">
          <el-input 
            v-model="configDialog.form.appGameId" 
            placeholder="输入游戏ID">
          </el-input>
        </el-form-item>
        <el-form-item label="密钥" prop="secretKey">
          <el-input 
            v-model="configDialog.form.secretKey" 
            placeholder="输入Yalla密钥"
            show-password>
          </el-input>
        </el-form-item>
        <el-form-item label="基础URL" prop="baseUrl">
          <el-input 
            v-model="configDialog.form.baseUrl" 
            placeholder="如: https://sdkapi.yallagame.com">
          </el-input>
        </el-form-item>
        <el-form-item label="推送URL" prop="pushUrl">
          <el-input 
            v-model="configDialog.form.pushUrl" 
            placeholder="如: https://sdklogapi.yallagame.com">
          </el-input>
        </el-form-item>
        <el-form-item label="环境" prop="environment">
          <el-select v-model="configDialog.form.environment" style="width: 100%">
            <el-option label="沙盒环境 (sandbox)" value="sandbox"></el-option>
            <el-option label="生产环境 (production)" value="production"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="超时时间(秒)" prop="timeout">
          <el-input-number 
            v-model="configDialog.form.timeout" 
            :min="1"
            :max="300"
            style="width: 100%">
          </el-input-number>
        </el-form-item>
        <el-form-item label="重试次数" prop="retryCount">
          <el-input-number 
            v-model="configDialog.form.retryCount" 
            :min="0"
            :max="10"
            style="width: 100%">
          </el-input-number>
        </el-form-item>
        <el-form-item label="配置描述" prop="description">
          <el-input 
            v-model="configDialog.form.description" 
            type="textarea"
            :rows="3"
            placeholder="配置的用途说明">
          </el-input>
        </el-form-item>
        <el-form-item label="是否启用" prop="isActive">
          <el-switch v-model="configDialog.form.isActive"></el-switch>
          <span style="margin-left: 10px; color: #909399; font-size: 12px;">
            关闭后将无法使用此Yalla配置
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="configDialog.visible = false">取消</el-button>
          <el-button type="primary" @click="saveConfig" :loading="configDialog.loading">
            {{ configDialog.isEdit ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 连接测试对话框 -->
    <el-dialog 
      v-model="testDialog.visible" 
      title="Yalla连接测试" 
      width="600px">
      <div class="test-content">
        <div class="test-info">
          <p><strong>应用ID:</strong> {{ testDialog.config?.appId }}</p>
          <p><strong>环境:</strong> {{ testDialog.config?.environment === 'production' ? '生产环境' : '沙盒环境' }}</p>
          <p><strong>基础URL:</strong> {{ testDialog.config?.baseUrl }}</p>
          <p><strong>推送URL:</strong> {{ testDialog.config?.pushUrl }}</p>
        </div>
        
        <div class="test-results" v-if="testDialog.results.length > 0">
          <h4>测试结果:</h4>
          <div v-for="(result, index) in testDialog.results" :key="index" class="test-result-item">
            <div class="result-header">
              <el-tag :type="result.success ? 'success' : 'danger'">
                {{ result.testName }}
              </el-tag>
              <span class="result-time">{{ result.responseTime }}ms</span>
            </div>
            <div class="result-message">{{ result.message }}</div>
          </div>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="testDialog.visible = false">关闭</el-button>
          <el-button type="primary" @click="runConnectionTest" :loading="testDialog.loading">
            重新测试
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, View, Hide } from '@element-plus/icons-vue'
import { yallaConfigAPI } from '../services/api.js'
import { selectedAppId } from '../utils/appStore.js'

export default {
  name: 'YallaConfigManagement',
  components: {
    Plus,
    View,
    Hide
  },
  setup() {
    const loading = ref(false)
    const configList = ref([])
    const configStats = ref(null)
    const filterAppId = ref('')
    const filterEnvironment = ref('')
    const filterStatus = ref('')
    
    const pagination = reactive({
      page: 1,
      pageSize: 20,
      total: 0
    })
    
    const configDialog = reactive({
      visible: false,
      isEdit: false,
      loading: false,
      form: {
        appId: '',
        appGameId: '',
        secretKey: '',
        baseUrl: 'https://sdkapitest.yallagame.com',
        pushUrl: 'https://sdklogapitest.yallagame.com',
        environment: 'sandbox',
        timeout: 30,
        retryCount: 3,
        description: '',
        isActive: true
      }
    })

    const testDialog = reactive({
      visible: false,
      loading: false,
      config: null,
      results: []
    })
    
    const configRules = {
      appId: [
        { required: true, message: '请输入应用ID', trigger: 'blur' },
        { min: 2, max: 50, message: '应用ID长度在 2 到 50 个字符', trigger: 'blur' }
      ],
      appGameId: [
        { required: true, message: '请输入游戏ID', trigger: 'blur' }
      ],
      secretKey: [
        { required: true, message: '请输入密钥', trigger: 'blur' }
      ],
      baseUrl: [
        { required: true, message: '请输入基础URL', trigger: 'blur' },
        { 
          pattern: /^https?:\/\/.+/, 
          message: '请输入有效的URL地址', 
          trigger: 'blur' 
        }
      ],
      pushUrl: [
        { required: true, message: '请输入推送URL', trigger: 'blur' },
        { 
          pattern: /^https?:\/\/.+/, 
          message: '请输入有效的URL地址', 
          trigger: 'blur' 
        }
      ],
      environment: [
        { required: true, message: '请选择环境', trigger: 'change' }
      ]
    }
    
    const configFormRef = ref()
    
    // 加载配置列表
    const loadConfigs = async () => {
      loading.value = true
      try {
        const params = {
          page: pagination.page,
          pageSize: pagination.pageSize
        }
        
        if (filterAppId.value) {
          params.appId = filterAppId.value
        }
        
        if (filterEnvironment.value) {
          params.environment = filterEnvironment.value
        }
        
        if (filterStatus.value) {
          params.isActive = filterStatus.value === 'true'
        }
        
        const response = await yallaConfigAPI.getList(params)
        if (response.code === 0) {
          configList.value = response.data.list || []
          pagination.total = response.data.total || 0
          
          // 计算统计数据
          calculateStats(configList.value)
        } else {
          ElMessage.error(response.message || '加载配置列表失败')
        }
      } catch (error) {
        console.error('加载配置失败:', error)
        ElMessage.error('加载配置列表失败')
      } finally {
        loading.value = false
      }
    }
    
    // 计算统计数据
    const calculateStats = (configList) => {
      if (!Array.isArray(configList)) {
        configStats.value = null
        return
      }
      
      const stats = {
        totalConfigs: configList.length,
        activeConfigs: configList.filter(config => config.isActive).length,
        sandboxConfigs: configList.filter(config => config.environment === 'sandbox').length,
        productionConfigs: configList.filter(config => config.environment === 'production').length
      }
      
      configStats.value = stats
    }
    
    // 格式化日期
    const formatDate = (dateStr) => {
      if (!dateStr) return ''
      return new Date(dateStr).toLocaleString('zh-CN')
    }
    
    // 掩码敏感数据
    const maskSensitiveData = (data) => {
      if (!data || data.length <= 8) return '***'
      return data.substring(0, 4) + '***' + data.substring(data.length - 4)
    }
    
    // 切换敏感字段显示
    const toggleShowSensitive = (row, field) => {
      row[field] = !row[field]
    }
    
    // 打开创建对话框
    const openCreateDialog = () => {
      configDialog.visible = true
      configDialog.isEdit = false
      resetConfigForm()
    }
    
    // 打开编辑对话框
    const openEditDialog = (config) => {
      configDialog.visible = true
      configDialog.isEdit = true
      configDialog.form = {
        id: config.id,
        appId: config.appId,
        appGameId: config.appGameId,
        secretKey: config.secretKey,
        baseUrl: config.baseUrl,
        pushUrl: config.pushUrl,
        environment: config.environment,
        timeout: config.timeout || 30,
        retryCount: config.retryCount || 3,
        description: config.description || '',
        isActive: config.isActive
      }
    }
    
    // 重置配置表单
    const resetConfigForm = () => {
      configDialog.form = {
        appId: '',
        appGameId: '',
        secretKey: '',
        baseUrl: 'https://sdkapitest.yallagame.com',
        pushUrl: 'https://sdklogapitest.yallagame.com',
        environment: 'sandbox',
        timeout: 30,
        retryCount: 3,
        description: '',
        isActive: true
      }
    }
    
    // 重置配置对话框
    const resetConfigDialog = () => {
      resetConfigForm()
      if (configFormRef.value) {
        configFormRef.value.resetFields()
      }
    }
    
    // 保存配置
    const saveConfig = async () => {
      if (!configFormRef.value) return
      
      try {
        await configFormRef.value.validate()
      } catch (error) {
        return
      }
      
      configDialog.loading = true
      
      try {
        const params = { ...configDialog.form }
        
        let response
        if (configDialog.isEdit) {
          response = await yallaConfigAPI.update(params)
        } else {
          response = await yallaConfigAPI.create(params)
        }
        
        if (response.code === 0) {
          ElMessage.success(configDialog.isEdit ? '更新成功' : '创建成功')
          configDialog.visible = false
          loadConfigs()
        } else {
          ElMessage.error(response.message || '操作失败')
        }
      } catch (error) {
        console.error('保存配置失败:', error)
        ElMessage.error('操作失败')
      } finally {
        configDialog.loading = false
      }
    }
    
    // 切换配置状态
    const toggleConfigStatus = async (config) => {
      config.updating = true
      try {
        const response = await yallaConfigAPI.update({
          id: config.id,
          appId: config.appId,
          isActive: config.isActive
        })
        
        if (response.code === 0) {
          ElMessage.success('状态更新成功')
        } else {
          config.isActive = !config.isActive // 回滚状态
          ElMessage.error(response.message || '状态更新失败')
        }
      } catch (error) {
        config.isActive = !config.isActive // 回滚状态
        console.error('状态更新失败:', error)
        ElMessage.error('状态更新失败')
      } finally {
        config.updating = false
      }
    }
    
    // 删除配置
    const deleteConfig = async (config) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除应用 "${config.appId}" 的Yalla配置吗？此操作不可恢复。`,
          '确认删除',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          }
        )
        
        const response = await yallaConfigAPI.delete({
          appId: config.appId
        })
        if (response.code === 0) {
          ElMessage.success('删除成功')
          loadConfigs()
        } else {
          ElMessage.error(response.message || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除配置失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 测试连接
    const testConnection = async (config) => {
      testDialog.config = config
      testDialog.results = []
      testDialog.visible = true
      
      await runConnectionTest()
    }
    
    // 运行连接测试
    const runConnectionTest = async () => {
      if (!testDialog.config) return
      
      testDialog.loading = true
      testDialog.results = []
      
      try {
        const response = await yallaConfigAPI.testConnection({
          appId: testDialog.config.appId
        })
        
        if (response.code === 0) {
          testDialog.results = response.data.testResults || []
        } else {
          testDialog.results = [{
            testName: '连接测试',
            success: false,
            message: response.message || '测试失败',
            responseTime: 0
          }]
        }
      } catch (error) {
        console.error('连接测试失败:', error)
        testDialog.results = [{
          testName: '连接测试',
          success: false,
          message: '网络错误或服务不可用',
          responseTime: 0
        }]
      } finally {
        testDialog.loading = false
      }
    }
    
    // 处理排序变更
    const handleSortChange = ({ column, prop, order }) => {
      // TODO: 实现排序功能
      console.log('排序变更:', { column, prop, order })
    }
    
    onMounted(() => {
      loadConfigs()
    })
    
    return {
      loading,
      configList,
      configStats,
      selectedAppId,
      filterAppId,
      filterEnvironment,
      filterStatus,
      pagination,
      configDialog,
      testDialog,
      configRules,
      configFormRef,
      loadConfigs,
      formatDate,
      maskSensitiveData,
      toggleShowSensitive,
      openCreateDialog,
      openEditDialog,
      resetConfigDialog,
      saveConfig,
      toggleConfigStatus,
      deleteConfig,
      testConnection,
      runConnectionTest,
      handleSortChange
    }
  }
}
</script>

<style scoped>
.yalla-config-management {
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

.content {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.filters-section {
  background: #f5f7fa;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.filter-form {
  margin: 0;
}

.config-section {
  margin-bottom: 30px;
}

.config-section h2 {
  color: #333;
  margin-bottom: 15px;
}

.sensitive-field {
  display: flex;
  align-items: center;
  font-family: monospace;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

.dialog-footer {
  text-align: right;
}

.el-button.danger {
  color: #f56c6c;
}

/* 统计样式 */
.stats-section {
  margin-bottom: 20px;
}

.stats-cards {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.stats-card {
  flex: 1;
  min-width: 200px;
}

.stats-content {
  text-align: center;
  padding: 10px;
}

.stats-number {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 8px;
}

.stats-label {
  font-size: 14px;
  color: #666;
}

/* 测试对话框样式 */
.test-content {
  padding: 10px 0;
}

.test-info {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.test-info p {
  margin: 5px 0;
  color: #333;
}

.test-results h4 {
  margin-bottom: 15px;
  color: #333;
}

.test-result-item {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 10px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.result-time {
  font-size: 12px;
  color: #909399;
}

.result-message {
  font-size: 14px;
  color: #666;
}

:deep(.el-table .el-table__cell) {
  padding: 8px 0;
}

:deep(.el-form-item) {
  margin-bottom: 18px;
}
</style>
