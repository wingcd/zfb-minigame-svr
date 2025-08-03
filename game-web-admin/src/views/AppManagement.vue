<template>
  <div>
    <div class="page-header">
      <h1>应用管理</h1>
      <el-button type="primary" @click="showAddDialog = true">添加应用</el-button>
    </div>

    <el-table :data="apps" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="应用名称" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="status" label="状态">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
            {{ scope.row.status === 'active' ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button size="small" @click="editApp(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteApp(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAddDialog" title="添加应用" width="500px">
      <el-form :model="appForm" label-width="80px">
        <el-form-item label="应用名称">
          <el-input v-model="appForm.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="appForm.description" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addApp">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { appAPI } from '../services/api.js'

const apps = ref([])
const loading = ref(false)
const showAddDialog = ref(false)
const appForm = ref({
  name: '',
  description: ''
})

const loadApps = async () => {
  loading.value = true
  try {
    // 使用封装的API服务获取应用列表
    const data = await appAPI.getAllApps()
    apps.value = data
  } catch (error) {
    ElMessage.error('获取应用列表失败')
  } finally {
    loading.value = false
  }
}

const addApp = async () => {
  try {
    await appAPI.createApp(appForm.value)
    ElMessage.success('添加成功')
    showAddDialog.value = false
    appForm.value = { name: '', description: '' }
    loadApps()
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const editApp = async (app) => {
  try {
    await appAPI.updateApp(app.id, app)
    ElMessage.success('更新成功')
    loadApps()
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const deleteApp = async (app) => {
  try {
    await appAPI.deleteApp(app.id)
    ElMessage.success('删除成功')
    loadApps()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadApps()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
</style>