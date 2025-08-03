<template>
  <div>
    <div class="page-header">
      <h1>排行榜管理</h1>
      <el-button type="primary" @click="showAddDialog = true">添加排行榜</el-button>
    </div>

    <el-table :data="leaderboards" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="排行榜名称" />
      <el-table-column prop="appName" label="所属应用" />
      <el-table-column prop="type" label="类型" />
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button size="small" @click="viewLeaderboard(scope.row)">查看</el-button>
          <el-button size="small" type="danger" @click="deleteLeaderboard(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :model-value="showAddDialog" @update:model-value="showAddDialog = $event" title="添加排行榜" width="500px">
      <el-form :model="leaderboardForm" label-width="100px">
        <el-form-item label="排行榜名称">
          <el-input v-model="leaderboardForm.name" />
        </el-form-item>
        <el-form-item label="所属应用">
          <el-select v-model="leaderboardForm.appId" style="width: 100%">
            <el-option
              v-for="app in apps"
              :key="app.id"
              :label="app.name"
              :value="app.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="leaderboardForm.type" style="width: 100%">
            <el-option label="最高分" value="highest" />
            <el-option label="累计" value="cumulative" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addLeaderboard">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { appAPI, leaderboardAPI } from '../services/api.js'

const leaderboards = ref([])
const apps = ref([])
const loading = ref(false)
const showAddDialog = ref(false)
const leaderboardForm = ref({
  name: '',
  appId: '',
  type: 'highest'
})

const loadLeaderboards = async () => {
  loading.value = true
  try {
    // 使用封装的API服务获取排行榜列表
    const data = await leaderboardAPI.getAllLeaderboards?.() || []
    leaderboards.value = data
  } catch (error) {
    ElMessage.error('获取排行榜列表失败')
  } finally {
    loading.value = false
  }
}

const loadApps = async () => {
  try {
    const data = await appAPI.getAllApps()
    apps.value = data
  } catch (error) {
    console.error('获取应用列表失败:', error)
  }
}

const addLeaderboard = async () => {
  try {
    await leaderboardAPI.initLeaderboard(leaderboardForm.value)
    ElMessage.success('添加成功')
    showAddDialog.value = false
    leaderboardForm.value = { name: '', appId: '', type: 'highest' }
    loadLeaderboards()
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const viewLeaderboard = (leaderboard) => {
  console.log('查看排行榜:', leaderboard)
}

const deleteLeaderboard = async (leaderboard) => {
  try {
    // 这里需要添加删除排行榜的API调用
    ElMessage.success('删除成功')
    loadLeaderboards()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadLeaderboards()
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