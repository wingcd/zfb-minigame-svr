<template>
  <div class="counter-management">
    <div class="page-header">
      <h1>计数器管理</h1>
      <div class="header-actions">
        <el-select v-model="filters.appId" placeholder="选择应用" @change="loadCounters" style="width: 200px; margin-right: 10px;">
          <el-option
            v-for="app in apps"
            :key="app.appId"
            :label="app.appName"
            :value="app.appId"
          />
        </el-select>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          新建计数器
        </el-button>
        <el-button @click="loadCounters">刷新</el-button>
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
      </el-form>
    </div>

    <!-- 计数器配置管理 -->
    <div class="counter-config-section">
      <h2>计数器配置</h2>
      <el-table 
        :data="counters" 
        v-loading="loading"
        style="width: 100%"
      >
        <el-table-column prop="key" label="计数器Key" width="200" />
        <el-table-column prop="description" label="描述" width="200" show-overflow-tooltip />
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
        <el-form-item label="应用" prop="appId">
          <el-select v-model="form.appId" placeholder="选择应用" :disabled="isEdit" style="width: 100%">
            <el-option
              v-for="app in apps"
              :key="app.appId"
              :label="app.appName"
              :value="app.appId"
            />
          </el-select>
        </el-form-item>
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
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { counterAPI, appAPI } from '@/services/api'

export default {
  name: 'CounterManagement',
  components: {
    Plus
  },
  setup() {
    const loading = ref(false)
    const saving = ref(false)
    const dialogVisible = ref(false)
    const isEdit = ref(false)
    const formRef = ref(null)
    
    const counters = ref([])
    const apps = ref([])
    
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
      description: '',
      resetType: 'permanent',
      resetValue: null,
      value: null
    })
    
    const rules = {
      appId: [{ required: true, message: '请选择应用', trigger: 'change' }],
      key: [{ required: true, message: '请输入计数器Key', trigger: 'blur' }],
      resetType: [{ required: true, message: '请选择重置类型', trigger: 'change' }],
      resetValue: [
        { 
          required: computed(() => form.resetType === 'custom'), 
          message: '请输入重置间隔小时数', 
          trigger: 'blur' 
        }
      ]
    }
    
    // 加载应用列表
    const loadApps = async () => {
      try {
        const response = await appAPI.getAll()
        if (response.code === 0) {
          apps.value = response.data.list || []
          if (apps.value.length > 0 && !filters.appId) {
            filters.appId = apps.value[0].appId
            await loadCounters()
          }
        }
      } catch (error) {
        console.error('加载应用列表失败:', error)
      }
    }
    
    // 加载计数器列表
    const loadCounters = async () => {
      if (!filters.appId) {
        ElMessage.warning('请先选择应用')
        return
      }
      
      loading.value = true
      try {
        const params = {
          appId: filters.appId,
          page: pagination.page,
          pageSize: pagination.pageSize
        }
        
        if (filters.key) params.key = filters.key
        if (filters.resetType) params.resetType = filters.resetType
        
        const response = await counterAPI.getList(params)
        if (response.code === 0) {
          counters.value = response.data.list || []
          pagination.total = response.data.total || 0
        } else {
          ElMessage.error(response.msg || '加载计数器列表失败')
        }
      } catch (error) {
        console.error('加载计数器列表失败:', error)
        ElMessage.error('加载计数器列表失败')
      } finally {
        loading.value = false
      }
    }
    
    // 显示创建对话框
    const showCreateDialog = () => {
      if (!filters.appId) {
        ElMessage.warning('请先选择应用')
        return
      }
      
      isEdit.value = false
      Object.assign(form, {
        appId: filters.appId,
        key: '',
        description: '',
        resetType: 'permanent',
        resetValue: null,
        value: null
      })
      dialogVisible.value = true
    }
    
    // 编辑计数器
    const editCounter = (counter) => {
      isEdit.value = true
      Object.assign(form, {
        _id: counter._id,
        appId: counter.appId || filters.appId,
        key: counter.key,
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
    
    // 删除计数器
    const deleteCounter = async (counter) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除计数器 "${counter.key}" 吗？此操作不可恢复。`,
          '确认删除',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        const response = await counterAPI.delete({
          appId: filters.appId,
          key: counter.key
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
    
    onMounted(() => {
      loadApps()
    })
    
    return {
      loading,
      saving,
      dialogVisible,
      isEdit,
      formRef,
      counters,
      apps,
      filters,
      pagination,
      form,
      rules,
      loadCounters,
      showCreateDialog,
      editCounter,
      saveCounter,
      deleteCounter,
      getResetTypeTagType,
      getResetTypeLabel,
      handleResetTypeChange,
      formatDate
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
</style> 