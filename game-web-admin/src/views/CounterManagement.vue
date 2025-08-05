<template>
  <div class="counter-management">
    <div class="page-header">
      <h1>计数器管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          新建计数器
        </el-button>
        <el-button @click="loadCounters">刷新</el-button>
      </div>
    </div>

    <!-- 统计概览 -->
    <div class="stats-section" v-if="stats">
      <div class="stats-cards">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-number">{{ stats.totalCounters || 0 }}</div>
            <div class="stats-label">计数器总数</div>
          </div>
        </el-card>
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-number">{{ stats.totalLocations || 0 }}</div>
            <div class="stats-label">点位总数</div>
          </div>
        </el-card>
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-number">{{ formatNumber(stats.totalValue || 0) }}</div>
            <div class="stats-label">总计数值</div>
          </div>
        </el-card>
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-number">{{ stats.summary?.averageLocationsPerCounter || 0 }}</div>
            <div class="stats-label">平均点位数</div>
          </div>
        </el-card>
      </div>
      
      <!-- 重置类型分布 -->
      <div class="reset-type-distribution">
        <h3>重置类型分布</h3>
        <div class="reset-type-items">
          <div v-for="(count, type) in (stats.resetTypeStats || {})" :key="type" class="reset-type-item">
            <el-tag :type="getResetTypeTagType(type)">{{ getResetTypeLabel(type) }}</el-tag>
            <span class="count">{{ count }}个</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选条件 -->
    <div class="filters-section">
      <el-form :model="filters" inline>
        <el-form-item label="计数器Key">
          <el-input 
            v-model="filters.key" 
            placeholder="输入计数器Key搜索" 
            clearable
            @change="loadCounters"
            style="width: 200px;"
          />
        </el-form-item>
        <el-form-item label="重置类型">
          <el-select v-model="filters.resetType" placeholder="选择重置类型" clearable @change="loadCounters" style="width: 150px;">
            <el-option label="永久" value="permanent" />
            <el-option label="每日" value="daily" />
            <el-option label="每周" value="weekly" />
            <el-option label="每月" value="monthly" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="显示模式">
          <el-select v-model="displayMode" placeholder="选择显示模式" @change="loadCounters" style="width: 150px;">
            <el-option label="按Key分组" value="grouped" />
            <el-option label="显示所有记录" value="list" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button @click="loadStats" :loading="statsLoading">
            <el-icon><Refresh /></el-icon>
            刷新统计
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 计数器配置管理 -->
    <div class="counter-config-section">
      <h2>计数器配置</h2>
      
      <!-- 分组显示模式 -->
      <el-table 
        v-if="displayMode === 'grouped'"
        :data="counters" 
        v-loading="loading"
        style="width: 100%"
        :expand-row-keys="expandedRows"
        row-key="key"
        :default-expand-all="false"
      >
        <el-table-column type="expand">
          <template #default="props">
            <div class="location-details">
              <h4>点位详情 ({{ props.row.key }})</h4>
              <el-table :data="props.row.locations || []" size="small">
                <el-table-column prop="location" label="点位" width="150" />
                <el-table-column prop="value" label="当前值" width="120" align="center">
                  <template #default="scope">
                    <span class="value-number">{{ scope.row.value }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="重置时间" width="180" align="center">
                  <template #default="scope">
                    {{ scope.row.resetTime ? formatDate(scope.row.resetTime) : '永不重置' }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="120" align="center">
                  <template #default="scope">
                    <el-button type="text" size="small" @click="editLocationCounter(props.row, scope.row)">
                      编辑
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="key" label="计数器Key" width="200" />
        <el-table-column prop="description" label="描述" width="200" show-overflow-tooltip />
        <el-table-column label="总值/点位数" width="150" align="center">
          <template #default="scope">
            <div class="summary-info">
              <div class="total-value">总值: {{ scope.row.totalValue }}</div>
              <div class="location-count">点位: {{ scope.row.locationCount }}个</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="重置类型" width="120" align="center">
          <template #default="scope">
            <el-tag :type="getResetTypeTagType(scope.row.resetType)">
              {{ getResetTypeLabel(scope.row.resetType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="重置间隔" width="120" align="center">
          <template #default="scope">
            {{ scope.row.resetType === 'custom' ? `${scope.row.resetValue}小时` : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="gmtCreate" label="创建时间" width="180" align="center">
          <template #default="scope">
            {{ formatDate(scope.row.gmtCreate) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="scope">
            <el-button type="text" class="danger" @click="deleteCounterAllLocations(scope.row)">删除全部</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 列表显示模式 -->
      <el-table 
        v-else
        :data="counters" 
        v-loading="loading"
        style="width: 100%"
      >
        <el-table-column prop="key" label="计数器Key" width="180" />
        <el-table-column prop="location" label="点位" width="120" align="center">
          <template #default="scope">
            <el-tag size="small" type="info">{{ scope.row.location }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" width="180" show-overflow-tooltip />
        <el-table-column prop="value" label="当前值" width="120" align="center">
          <template #default="scope">
            <div class="value-display">
              <span class="value-number">{{ scope.row.value || 0 }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="重置类型" width="120" align="center">
          <template #default="scope">
            <el-tag :type="getResetTypeTagType(scope.row.resetType)">
              {{ getResetTypeLabel(scope.row.resetType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="重置间隔" width="120" align="center">
          <template #default="scope">
            {{ scope.row.resetType === 'custom' ? `${scope.row.resetValue}小时` : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="下次重置时间" width="180" align="center">
          <template #default="scope">
            {{ scope.row.resetTime ? formatDate(scope.row.resetTime) : '永不重置' }}
          </template>
        </el-table-column>
        <el-table-column prop="gmtCreate" label="创建时间" width="180" align="center">
          <template #default="scope">
            {{ formatDate(scope.row.gmtCreate) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="scope">
            <el-button type="text" @click="editCounter(scope.row)">编辑</el-button>
            <el-button type="text" class="danger" @click="deleteCounter(scope.row)">删除</el-button>
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
          @size-change="loadCounters"
          @current-change="loadCounters"
        />
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="isEdit ? '编辑计数器' : '新建计数器'"
      width="500px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="计数器Key" prop="key">
          <el-input v-model="form.key" placeholder="输入计数器Key" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" placeholder="输入计数器描述" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="重置类型" prop="resetType">
          <el-select v-model="form.resetType" placeholder="选择重置类型" style="width: 100%" @change="handleResetTypeChange">
            <el-option label="永久保存" value="permanent" />
            <el-option label="每日重置" value="daily" />
            <el-option label="每周重置" value="weekly" />
            <el-option label="每月重置" value="monthly" />
            <el-option label="自定义间隔" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item 
          v-if="form.resetType === 'custom'" 
          label="重置间隔(小时)" 
          prop="resetValue"
        >
          <el-input-number 
            v-model="form.resetValue" 
            :min="1" 
            :max="8760"
            style="width: 100%"
            placeholder="输入重置间隔小时数"
          />
        </el-form-item>
        <el-form-item v-if="isEdit" label="当前值" prop="value">
          <el-input-number 
            v-model="form.value" 
            :min="0"
            style="width: 100%"
            placeholder="设置计数器值"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveCounter" :loading="saving">
            {{ isEdit ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { counterAPI, appAPI } from '@/services/api'
import { selectedAppId, appList } from '@/utils/appStore.js'

export default {
  name: 'CounterManagement',
  components: {
    Plus,
    Refresh
  },
  setup() {
    const loading = ref(false)
    const saving = ref(false)
    const statsLoading = ref(false)
    const dialogVisible = ref(false)
    const isEdit = ref(false)
    const formRef = ref(null)
    
    const counters = ref([])
    const expandedRows = ref([])
    const displayMode = ref('grouped')
    const stats = ref(null)
    
    const filters = reactive({
      appId: '',
      key: '',
      resetType: ''
    })
    
    const pagination = reactive({
      page: 1,
      pageSize: 20,
      total: 0
    })
    
    const form = reactive({
      appId: '',
      key: '',
      location: 'default',
      description: '',
      resetType: 'permanent',
      resetValue: null,
      value: null
    })
    
    const rules = reactive({
      appId: [{ required: true, message: '请选择应用', trigger: 'change' }],
      key: [{ required: true, message: '请输入计数器Key', trigger: 'blur' }],
      resetType: [{ required: true, message: '请选择重置类型', trigger: 'change' }],
      resetValue: [
        { 
          required: false, // 将在验证时动态检查
          message: '请输入重置间隔小时数', 
          trigger: 'blur' 
        }
      ]
    })
    
    // 加载计数器列表
    const loadCounters = async () => {
      if (!selectedAppId.value) {
        ElMessage.warning('请先选择应用')
        return
      }
      
      loading.value = true
      try {
        const params = {
          appId: selectedAppId.value,
          page: pagination.page,
          pageSize: pagination.pageSize,
          groupByKey: displayMode.value === 'grouped'
        }
        
        if (filters.key) params.key = filters.key
        if (filters.resetType) params.resetType = filters.resetType
        
        const response = await counterAPI.getList(params)
        if (response.code === 0) {
          // 确保数据是数组格式
          const list = response.data?.list || response.data || []
          const safeList = Array.isArray(list) ? list : []
          
          // 确保每个项的 locations 也是数组（用于分组模式）
          counters.value = safeList.map(item => ({
            ...item,
            key: item.key || item._id || `counter_${Date.now()}_${Math.random()}`, // 确保有唯一key
            locations: Array.isArray(item.locations) ? item.locations : []
          }))
          pagination.total = response.data?.total || 0
        } else {
          ElMessage.error(response.msg || '加载计数器列表失败')
          counters.value = [] // 确保失败时也是空数组
        }
      } catch (error) {
        console.error('加载计数器列表失败:', error)
        ElMessage.error('加载计数器列表失败')
        counters.value = [] // 确保异常时也是空数组
      } finally {
        loading.value = false
      }
    }
    
    // 加载统计信息
    const loadStats = async () => {
      if (!selectedAppId.value) {
        return
      }
      
      statsLoading.value = true
      try {
        const response = await counterAPI.getAllStats({ appId: selectedAppId.value })
        if (response.code === 0) {
          stats.value = response.data || null
        } else {
          console.error('加载统计信息失败:', response.msg)
          stats.value = null
        }
      } catch (error) {
        console.error('加载统计信息失败:', error)
        stats.value = null
      } finally {
        statsLoading.value = false
      }
    }
    
    // 显示创建对话框
    const showCreateDialog = () => {
      if (!selectedAppId.value) {
        ElMessage.warning('请先选择应用')
        return
      }
      
      isEdit.value = false
      Object.assign(form, {
        appId: selectedAppId.value,
        key: '',
        location: 'default',
        description: '',
        resetType: 'permanent',
        resetValue: null,
        value: null
      })
      dialogVisible.value = true
    }
    
    // 编辑特定点位的计数器
    const editLocationCounter = (counter, location) => {
      isEdit.value = true
      Object.assign(form, {
        appId: selectedAppId.value,
        key: counter.key,
        location: location.location,
        description: counter.description,
        resetType: counter.resetType,
        resetValue: counter.resetValue,
        value: location.value
      })
      dialogVisible.value = true
    }
    
    // 编辑计数器（列表模式）
    const editCounter = (counter) => {
      isEdit.value = true
      Object.assign(form, {
        _id: counter._id,
        appId: counter.appId || selectedAppId.value,
        key: counter.key,
        location: counter.location || 'default',
        description: counter.description || '',
        resetType: counter.resetType,
        resetValue: counter.resetValue,
        value: counter.value
      })
      dialogVisible.value = true
    }
    
    // 保存计数器
    const saveCounter = async () => {
      if (!formRef.value) return
      
      try {
        // 自定义验证逻辑
        if (form.resetType === 'custom' && (!form.resetValue || form.resetValue <= 0)) {
          ElMessage.error('请输入有效的重置间隔小时数')
          return
        }
        
        await formRef.value.validate()
        saving.value = true
        
        const data = { ...form }
        if (data.resetType !== 'custom') {
          data.resetValue = null
        }
        
        let response
        if (isEdit.value) {
          response = await counterAPI.update(data)
        } else {
          response = await counterAPI.create(data)
        }
        
        if (response.code === 0) {
          ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
          dialogVisible.value = false
          loadCounters()
        } else {
          ElMessage.error(response.msg || (isEdit.value ? '更新失败' : '创建失败'))
        }
      } catch (error) {
        console.error('保存计数器失败:', error)
      } finally {
        saving.value = false
      }
    }
    
    // 删除计数器（单个点位）
    const deleteCounter = async (counter) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除计数器 "${counter.key}" 的点位 "${counter.location}" 吗？此操作不可恢复。`,
          '确认删除',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        const response = await counterAPI.delete({
          appId: selectedAppId.value,
          key: counter.key,
          location: counter.location
        })
        
        if (response.code === 0) {
          ElMessage.success('删除成功')
          loadCounters()
        } else {
          ElMessage.error(response.msg || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除计数器失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 删除计数器的所有点位
    const deleteCounterAllLocations = async (counter) => {
      try {
        if (!counter?.locations || !Array.isArray(counter.locations)) {
          ElMessage.error('计数器数据异常')
          return
        }
        
        await ElMessageBox.confirm(
          `确定要删除计数器 "${counter.key}" 的所有点位吗？这将删除该计数器的全部数据，此操作不可恢复。`,
          '确认删除全部',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        // 删除所有点位
        const deletePromises = counter.locations.map(location => 
          counterAPI.delete({
            appId: selectedAppId.value,
            key: counter.key,
            location: location.location
          })
        )
        
        const results = await Promise.all(deletePromises)
        const failedCount = results.filter(r => r.code !== 0).length
        
        if (failedCount === 0) {
          ElMessage.success('删除成功')
        } else {
          ElMessage.warning(`部分删除失败，失败数量：${failedCount}`)
        }
        
        loadCounters()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除计数器失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 获取重置类型标签类型
    const getResetTypeTagType = (resetType) => {
      const types = {
        permanent: 'info',
        daily: 'success',
        weekly: 'warning',
        monthly: 'danger',
        custom: 'primary'
      }
      return types[resetType] || 'info'
    }
    
    // 获取重置类型标签文本
    const getResetTypeLabel = (resetType) => {
      const labels = {
        permanent: '永久',
        daily: '每日',
        weekly: '每周',
        monthly: '每月',
        custom: '自定义'
      }
      return labels[resetType] || resetType
    }
    
    // 处理重置类型变化
    const handleResetTypeChange = () => {
      if (form.resetType !== 'custom') {
        form.resetValue = null
      } else {
        form.resetValue = 24
      }
    }
    
    // 格式化日期
    const formatDate = (dateStr) => {
      if (!dateStr) return ''
      return new Date(dateStr).toLocaleString('zh-CN')
    }
    
    // 格式化数字
    const formatNumber = (num) => {
      if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M'
      } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K'
      }
      return num.toString()
    }
    
    // 监听全局app选择变化
    watch(selectedAppId, () => {
      if (selectedAppId.value) {
        filters.appId = selectedAppId.value
        loadCounters()
        loadStats()
      }
    })
    
    onMounted(() => {
      // 初始化时检查是否已有选中的应用
      if (selectedAppId.value) {
        filters.appId = selectedAppId.value
        loadCounters()
        loadStats()
      }
    })
    
    return {
      loading,
      saving,
      dialogVisible,
      isEdit,
      formRef,
      counters,
      expandedRows,
      displayMode,
      apps: appList, // Use global app list
      filters,
      pagination,
      form,
      rules,
      loadCounters,
      showCreateDialog,
      editLocationCounter,
      editCounter,
      saveCounter,
      deleteCounter,
      deleteCounterAllLocations,
      getResetTypeTagType,
      getResetTypeLabel,
      handleResetTypeChange,
      formatDate,
      formatNumber,
      loadStats,
      statsLoading,
      stats
    }
  }
}
</script>

<style scoped>
.counter-management {
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

.filters-section {
  background: #f5f7fa;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.counter-config-section {
  margin-bottom: 30px;
}

.counter-config-section h2 {
  color: #333;
  margin-bottom: 15px;
}

.value-display {
  font-weight: bold;
  color: #e6a23c;
}

.value-number {
  font-size: 16px;
}

.summary-info {
  text-align: center;
}

.total-value {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 4px;
}

.location-count {
  font-size: 12px;
  color: #909399;
}

.location-details {
  padding: 20px;
  background: #fafafa;
  border-radius: 4px;
}

.location-details h4 {
  margin: 0 0 15px 0;
  color: #333;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
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

.reset-type-distribution {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
}

.reset-type-distribution h3 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 16px;
}

.reset-type-items {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.reset-type-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.reset-type-item .count {
  font-size: 14px;
  color: #666;
}
</style> 