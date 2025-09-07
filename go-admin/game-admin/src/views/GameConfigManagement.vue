<template>
  <div class="game-config-management">
    <div class="page-header">
      <h1>游戏配置管理</h1>
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
              <div class="stats-number">{{ configStats.versionCount || 0 }}</div>
              <div class="stats-label">版本数量</div>
            </div>
          </el-card>
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">{{ configStats.globalConfigs || 0 }}</div>
              <div class="stats-label">全局配置</div>
            </div>
          </el-card>
        </div>
        
        <!-- 配置类型分布 -->
        <div class="type-distribution">
          <h3>配置类型分布</h3>
          <div class="type-items">
            <div v-for="(count, type) in (configStats.typeStats || {})" :key="type" class="type-item">
              <el-tag :type="getTypeTagType(type)">{{ type }}</el-tag>
              <span class="count">{{ count }}个</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 筛选区域 -->
      <div class="filters-section">
        <el-form :inline="true" class="filter-form">
          <el-form-item label="版本筛选">
            <el-select v-model="filterVersion" @change="loadConfigs" placeholder="全部版本" clearable style="width: 150px;">
              <el-option label="全局配置" value=""></el-option>
              <el-option
                v-for="version in versions"
                :key="version"
                :label="version"
                :value="version">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="配置键名">
            <el-input 
              v-model="filterConfigKey" 
              @input="loadConfigs"
              placeholder="搜索配置键名" 
              clearable 
              style="width: 200px;">
            </el-input>
          </el-form-item>
          <el-form-item label="配置类型">
            <el-select v-model="filterType" @change="loadConfigs" placeholder="全部类型" clearable style="width: 150px;">
              <el-option label="字符串" value="string"></el-option>
              <el-option label="数字" value="number"></el-option>
              <el-option label="布尔值" value="boolean"></el-option>
              <el-option label="对象" value="object"></el-option>
              <el-option label="数组" value="array"></el-option>
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
        <h2>配置列表</h2>
        
        <el-table 
          :data="configList" 
          v-loading="loading"
          style="width: 100%"
          @sort-change="handleSortChange">
          <el-table-column prop="configKey" label="配置键名" width="180" sortable="custom">
            <template #default="scope">
              <el-tag>{{ scope.row.configKey }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="configValue" label="配置值" width="200">
            <template #default="scope">
              <div class="config-value">
                <span v-if="scope.row.configType === 'string'" class="value-string">{{ scope.row.configValue }}</span>
                <span v-else-if="scope.row.configType === 'number'" class="value-number">{{ scope.row.configValue }}</span>
                <span v-else-if="scope.row.configType === 'boolean'" class="value-boolean">
                  <el-tag :type="scope.row.configValue ? 'success' : 'danger'">
                    {{ scope.row.configValue ? 'true' : 'false' }}
                  </el-tag>
                </span>
                <span v-else class="value-object">{{ JSON.stringify(scope.row.configValue) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="configType" label="类型" width="100">
            <template #default="scope">
              <el-tag size="small" :type="getTypeTagType(scope.row.configType)">
                {{ scope.row.configType }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="version" label="版本" width="120">
            <template #default="scope">
              <el-tag v-if="scope.row.version" type="info">{{ scope.row.version }}</el-tag>
              <el-tag v-else type="warning">全局</el-tag>
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
          <el-table-column label="操作" width="160" fixed="right" align="center">
            <template #default="scope">
              <el-button link @click="openEditDialog(scope.row)">编辑</el-button>
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
      :title="configDialog.isEdit ? '编辑配置' : '创建配置'" 
      width="600px"
      @close="resetConfigDialog">
      <el-form :model="configDialog.form" :rules="configRules" ref="configFormRef" label-width="120px">
        <el-form-item label="配置键名" prop="configKey">
          <el-input 
            v-model="configDialog.form.configKey" 
            placeholder="如: max_level, game_name"
            :disabled="configDialog.isEdit">
          </el-input>
        </el-form-item>
        <el-form-item label="配置类型" prop="configType">
          <el-select v-model="configDialog.form.configType" @change="handleTypeChange" style="width: 100%">
            <el-option label="字符串 (string)" value="string"></el-option>
            <el-option label="数字 (number)" value="number"></el-option>
            <el-option label="布尔值 (boolean)" value="boolean"></el-option>
            <el-option label="对象 (object)" value="object"></el-option>
            <el-option label="数组 (array)" value="array"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="配置值" prop="configValue">
          <el-input 
            v-if="configDialog.form.configType === 'string'"
            v-model="configDialog.form.configValue" 
            placeholder="输入字符串值">
          </el-input>
          <el-input-number 
            v-else-if="configDialog.form.configType === 'number'"
            v-model="configDialog.form.configValue" 
            style="width: 100%">
          </el-input-number>
          <el-switch 
            v-else-if="configDialog.form.configType === 'boolean'"
            v-model="configDialog.form.configValue">
          </el-switch>
          <el-input 
            v-else
            v-model="configDialog.form.configValueJson" 
            type="textarea"
            :rows="4"
            placeholder="输入JSON格式的值">
          </el-input>
        </el-form-item>
        <el-form-item label="游戏版本" prop="version">
          <el-input 
            v-model="configDialog.form.version" 
            placeholder="如: 1.0.0 (留空为全局配置)">
          </el-input>
        </el-form-item>
        <el-form-item label="配置描述" prop="description">
          <el-input 
            v-model="configDialog.form.description" 
            type="textarea"
            :rows="2"
            placeholder="配置的用途说明">
          </el-input>
        </el-form-item>
        <el-form-item label="是否激活" prop="isActive">
          <el-switch v-model="configDialog.form.isActive"></el-switch>
          <span style="margin-left: 10px; color: #909399; font-size: 12px;">
            关闭后客户端将无法获取此配置
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
  </div>
</template>

<script>
import { ref, reactive, onMounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { gameConfigAPI } from '../services/api.js'
import { selectedAppId, getAppName } from '../utils/appStore.js'

export default {
  name: 'GameConfigManagement',
  components: {
    Plus
  },
  setup() {
    const loading = ref(false)
    const configList = ref([])
    const versions = ref([])
    const configStats = ref(null)
    const filterVersion = ref(null)
    const filterConfigKey = ref('')
    const filterType = ref('')
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
        configKey: '',
        configValue: '',
        configValueJson: '',
        configType: 'string',
        version: '',
        description: '',
        isActive: true
      }
    })
    
    const configRules = {
      configKey: [
        { required: true, message: '请输入配置键名', trigger: 'blur' },
        { min: 2, max: 50, message: '配置键名长度在 2 到 50 个字符', trigger: 'blur' }
      ],
      configType: [
        { required: true, message: '请选择配置类型', trigger: 'change' }
      ],
      configValue: [
        { required: true, message: '请输入配置值', trigger: 'blur' }
      ]
    }
    
    const configFormRef = ref()
    
    // 加载配置列表
    const loadConfigs = async () => {
      if (!selectedAppId.value) {
        ElMessage.warning('请先选择应用')
        return
      }
      
      loading.value = true
      try {
        const params = {
          appId: selectedAppId.value,
          page: pagination.page,
          pageSize: pagination.pageSize
        }
        
        if (filterVersion.value !== null) {
          params.version = filterVersion.value
        }
        
        if (filterConfigKey.value) {
          params.configKey = filterConfigKey.value
        }
        
        if (filterType.value) {
          params.configType = filterType.value
        }
        
        if (filterStatus.value) {
          params.isActive = filterStatus.value === 'true'
        }
        
        const response = await gameConfigAPI.getList(params)
        if (response.code === 0) {
          configList.value = response.data.list
          pagination.total = response.data.total
          versions.value = response.data.versions
          
          // 计算统计数据
          calculateStats(response.data.list || [])
        } else {
          ElMessage.error(response.msg || '加载配置列表失败')
        }
      } catch (error) {
        ElMessage.error('加载配置列表失败')
      } finally {
        loading.value = false
      }
    }
    
    // 获取类型标签颜色
    const getTypeTagType = (type) => {
      const typeMap = {
        'string': 'primary',
        'number': 'success',
        'boolean': 'warning',
        'object': 'info',
        'array': 'danger'
      }
      return typeMap[type] || 'primary'
    }
    
    // 格式化日期
    const formatDate = (dateStr) => {
      if (!dateStr) return ''
      return new Date(dateStr).toLocaleString('zh-CN')
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
        globalConfigs: configList.filter(config => !config.version).length,
        versionCount: new Set(configList.filter(config => config.version).map(config => config.version)).size,
        typeStats: {}
      }
      
      // 统计各类型的数量
      configList.forEach(config => {
        const type = config.configType || 'unknown'
        stats.typeStats[type] = (stats.typeStats[type] || 0) + 1
      })
      
      configStats.value = stats
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
        configKey: config.configKey,
        configValue: config.configValue,
        configValueJson: ['object', 'array'].includes(config.configType) 
          ? JSON.stringify(config.configValue, null, 2) 
          : '',
        configType: config.configType,
        version: config.version || '',
        description: config.description || '',
        isActive: config.isActive
      }
    }
    
    // 重置配置表单
    const resetConfigForm = () => {
      configDialog.form = {
        configKey: '',
        configValue: '',
        configValueJson: '',
        configType: 'string',
        version: '',
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
    
    // 处理类型变更
    const handleTypeChange = (type) => {
      if (type === 'boolean') {
        configDialog.form.configValue = false
      } else if (type === 'number') {
        configDialog.form.configValue = 0
      } else if (type === 'string') {
        configDialog.form.configValue = ''
      } else {
        configDialog.form.configValue = null
        configDialog.form.configValueJson = type === 'object' ? '{}' : '[]'
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
        let configValue = configDialog.form.configValue
        
        // 处理JSON类型的值
        if (['object', 'array'].includes(configDialog.form.configType)) {
          try {
            configValue = JSON.parse(configDialog.form.configValueJson)
          } catch (error) {
            ElMessage.error('JSON格式不正确')
            configDialog.loading = false
            return
          }
        }
        
        const params = {
          appId: selectedAppId.value,
          id: configDialog.form.id,
          configKey: configDialog.form.configKey,
          configValue: configValue,
          configType: configDialog.form.configType,
          description: configDialog.form.description,
          isActive: configDialog.form.isActive
        }
        
        if (configDialog.form.version) {
          params.version = configDialog.form.version
        }
        
        let response
        if (configDialog.isEdit) {
          params.appId = selectedAppId.value
          params.configKey = configDialog.form.configKey
          response = await gameConfigAPI.update(params)
        } else {
          response = await gameConfigAPI.create(params)
        }
        
        if (response.code === 0) {
          ElMessage.success(configDialog.isEdit ? '更新成功' : '创建成功')
          configDialog.visible = false
          loadConfigs()
        } else {
          ElMessage.error(response.msg || '操作失败')
        }
      } catch (error) {
        ElMessage.error('操作失败')
      } finally {
        configDialog.loading = false
      }
    }
    
    // 切换配置状态
    const toggleConfigStatus = async (config) => {
      config.updating = true
      try {
        const response = await gameConfigAPI.update({
          appId: selectedAppId.value,
          id: config.id,
          isActive: config.isActive
        })
        
        if (response.code === 0) {
          ElMessage.success('状态更新成功')
        } else {
          config.isActive = !config.isActive // 回滚状态
          ElMessage.error(response.msg || '状态更新失败')
        }
      } catch (error) {
        config.isActive = !config.isActive // 回滚状态
        ElMessage.error('状态更新失败')
      } finally {
        config.updating = false
      }
    }
    
    // 删除配置
    const deleteConfig = async (config) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除配置 "${config.configKey}" 吗？此操作不可恢复。`,
          '确认删除',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          }
        )
        
        const response = await gameConfigAPI.delete({
          appId: selectedAppId.value,
          configKey: config.configKey
        })
        if (response.code === 0) {
          ElMessage.success('删除成功')
          loadConfigs()
        } else {
          ElMessage.error(response.msg || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 处理排序变更
    const handleSortChange = ({ column, prop, order }) => {
      // TODO: 实现排序功能
      console.log('排序变更:', { column, prop, order })
    }
    
    // 监听全局app选择变化
    watch(selectedAppId, () => {
      if (selectedAppId.value) {
        loadConfigs()
      }
    }, { immediate: true })
    
    onMounted(() => {
      // 直接加载配置，selectedAppId从appStore获取
      if (selectedAppId.value) {
        loadConfigs()
      }
    })
    
    return {
      loading,
      configList,
      versions,
      configStats,
      selectedAppId,
      filterVersion,
      filterConfigKey,
      filterType,
      filterStatus,
      pagination,
      configDialog,
      configRules,
      configFormRef,
      loadConfigs,
      getTypeTagType,
      formatDate,
      openCreateDialog,
      openEditDialog,
      resetConfigDialog,
      handleTypeChange,
      saveConfig,
      toggleConfigStatus,
      deleteConfig,
      handleSortChange
    }
  }
}
</script>

<style scoped>
.game-config-management {
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

.config-value {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.value-string {
  color: #67c23a;
}

.value-number {
  color: #409eff;
  font-weight: bold;
}

.value-object {
  color: #909399;
  font-family: monospace;
  font-size: 12px;
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

:deep(.el-table .el-table__cell) {
  padding: 8px 0;
}

:deep(.el-form-item) {
  margin-bottom: 18px;
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

.type-distribution {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
}

.type-distribution h3 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 16px;
}

.type-items {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.type-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-item .count {
  font-size: 14px;
  color: #666;
}
</style> 