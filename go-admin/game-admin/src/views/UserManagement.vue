<template>
  <div class="user-management">
    <div class="page-header">
      <h1>用户管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="refreshUsers">刷新</el-button>
      </div>
    </div>

    <!-- 搜索筛选 -->
    <div class="search-section">
      <el-form :model="searchForm" :inline="true">
        <el-form-item label="玩家ID:">
          <el-input v-model="searchForm.playerId" placeholder="输入玩家ID" @keyup.enter="searchUsers"></el-input>
        </el-form-item>
        <el-form-item label="OpenID:">
          <el-input v-model="searchForm.openId" placeholder="输入OpenID" @keyup.enter="searchUsers"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchUsers">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 用户统计 -->
    <div class="stats-section" v-if="userStats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ userStats.total }}</div>
              <div class="stat-label">总用户数</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ userStats.newToday }}</div>
              <div class="stat-label">今日新增</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ userStats.activeToday }}</div>
              <div class="stat-label">今日活跃</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ userStats.banned }}</div>
              <div class="stat-label">封禁用户</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 用户表格 -->
    <el-table 
      :data="userList" 
      style="width: 100%" 
      v-loading="loading"
      @sort-change="handleSortChange">
      <el-table-column prop="playerId" label="玩家ID" width="120" sortable="custom">
      </el-table-column>
      <el-table-column prop="openId" label="OpenID" width="180" show-overflow-tooltip>
      </el-table-column>
      <el-table-column label="玩家信息" width="200">
        <template #default="scope">
          <div v-if="scope.row.userInfo">
            <div class="player-name">{{ scope.row.userInfo.nickName || '未设置' }}</div>
            <div class="player-avatar" v-if="scope.row.userInfo.avatarUrl">
              <el-avatar :src="scope.row.userInfo.avatarUrl" size="small"></el-avatar>
            </div>
          </div>
          <span v-else>无信息</span>
        </template>
      </el-table-column>
      <el-table-column label="游戏数据" width="120">
        <template #default="scope">
          <el-button link @click="viewUserData(scope.row)">
            查看数据
          </el-button>
        </template>
      </el-table-column>
      <el-table-column prop="CreatedAt" label="注册时间" width="160" sortable="custom">
      </el-table-column>
      <el-table-column prop="gmtModify" label="最后登录" width="160" sortable="custom">
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.banned ? 'danger' : 'success'">
            {{ scope.row.banned ? '已封禁' : '正常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="scope">
          <el-button link @click="editUserData(scope.row)">编辑数据</el-button>
          <el-button 
            link 
            :class="scope.row.banned ? 'success' : 'danger'"
            @click="toggleUserBan(scope.row)">
            {{ scope.row.banned ? '解封' : '封禁' }}
          </el-button>
          <el-button link class="danger" @click="deleteUser(scope.row)">删除</el-button>
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

    <!-- 用户数据查看对话框 -->
    <el-dialog 
      v-model="userDataDialog.visible" 
      title="用户游戏数据" 
      width="60%"
      :before-close="closeUserDataDialog">
      <div v-if="userDataDialog.data">
        <div class="user-info">
          <h3>基本信息</h3>
          <p>玩家ID: {{ userDataDialog.user.playerId }}</p>
          <p>OpenID: {{ userDataDialog.user.openId }}</p>
          <p>注册时间: {{ userDataDialog.user.CreatedAt }}</p>
        </div>
        <div class="game-data">
          <h3>游戏数据</h3>
          <el-input
            v-model="userDataDialog.dataStr"
            type="textarea"
            :rows="15"
            placeholder="游戏数据 (JSON格式)">
          </el-input>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="closeUserDataDialog">取消</el-button>
          <el-button type="primary" @click="saveUserData">保存数据</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userAPI, appAPI, statsAPI } from '../services/api.js'
import { selectedAppId, getAppName } from '../utils/appStore.js'

export default {
  name: 'UserManagement',
  setup() {
    const loading = ref(false)
    const userList = ref([])
    const userStats = ref(null)
    
    const searchForm = reactive({
      playerId: '',
      openId: ''
    })
    
    const pagination = reactive({
      current: 1,
      pageSize: 20,
      total: 0
    })
    
    const userDataDialog = reactive({
      visible: false,
      user: null,
      data: null,
      dataStr: ''
    })
    
    
    // 获取用户列表
    const getUserList = async () => {
      if (!selectedAppId.value) {
        ElMessage.warning('请先选择应用')
        return
      }
      
      loading.value = true
      try {
        const params = {
          appId: selectedAppId.value,
          page: pagination.current,
          pageSize: pagination.pageSize,
          ...searchForm
        }
        
        const result = await userAPI.getAll(params)
        if (result.code === 0) {
          // 确保数据是数组格式，防止迭代错误
          const dataList = result.data?.list
          userList.value = Array.isArray(dataList) ? dataList : []
          pagination.total = result.data?.total || 0
        } else {
          userList.value = []
          ElMessage.error(result.msg || '获取用户列表失败')
        }
      } catch (error) {
        console.error('获取用户列表失败:', error)
        userList.value = []
        ElMessage.error('获取用户列表失败')
      } finally {
        loading.value = false
      }
    }
    
    // 获取用户统计
    const getUserStats = async () => {
      if (!selectedAppId.value) return
      
      try {
        const result = await userAPI.getStats(selectedAppId.value)
        if (result.code === 0) {
          userStats.value = result.data
        }
      } catch (error) {
        console.error('获取用户统计失败:', error)
      }
    }
    
    // 查看用户数据
    const viewUserData = async (user) => {
      try {
        const result = await userAPI.getDetail({
          appId: selectedAppId.value,
          playerId: user.playerId,
          openId: user.openId,
        })
        
        if (result.code === 0) {
          userDataDialog.user = user
          userDataDialog.data = result.data
          userDataDialog.dataStr = JSON.stringify(result.data, null, 2)
          userDataDialog.visible = true
        } else {
          ElMessage.error(result.msg || '获取用户数据失败')
        }
      } catch (error) {
        console.error('获取用户数据失败:', error)
        ElMessage.error('获取用户数据失败')
      }
    }
    
    // 编辑用户数据
    const editUserData = (user) => {
      viewUserData(user)
    }
    
    // 保存用户数据
    const saveUserData = async () => {
      try {
        // 验证JSON格式
        let data
        try {
          data = JSON.parse(userDataDialog.dataStr)
        } catch (e) {
          ElMessage.error('数据格式错误，请输入有效的JSON')
          return
        }
        
        const result = await userAPI.setDetail({
          appId: selectedAppId.value,
          openId: userDataDialog.user.openId,
          userData: JSON.stringify(data)  // 后端期望的是 userData 字段，不是 data
        })
        
        if (result.code === 0) {
          ElMessage.success('保存成功')
          userDataDialog.visible = false
        } else {
          ElMessage.error(result.msg || '保存失败')
        }
      } catch (error) {
        console.error('保存用户数据失败:', error)
        ElMessage.error('保存失败')
      }
    }
    
    // 切换用户封禁状态
    const toggleUserBan = async (user) => {
      const action = user.banned ? '解封' : '封禁'
      try {
        await ElMessageBox.confirm(`确定要${action}用户 ${user.playerId} 吗？`, '确认操作')
        
        const apiCall = user.banned ? userAPI.unban : userAPI.ban
        const result = await apiCall({
          appId: selectedAppId.value,
          playerId: user.playerId
        })
        
        if (result.code === 0) {
          ElMessage.success(`${action}成功`)
          await getUserList()
        } else {
          ElMessage.error(result.msg || `${action}失败`)
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error(`${action}用户失败:`, error)
          ElMessage.error(`${action}失败`)
        }
      }
    }
    
    // 删除用户
    const deleteUser = async (user) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除用户 ${user.playerId} 吗？此操作不可恢复！`, 
          '危险操作', 
          { type: 'warning' }
        )
        
        const result = await userAPI.delete({
          appId: selectedAppId.value,
          playerId: user.playerId,
          force: true,
        })
        
        if (result.code === 0) {
          ElMessage.success('删除成功')
          await getUserList()
        } else {
          ElMessage.error(result.msg || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除用户失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 监听全局app选择变化
    watch(selectedAppId, () => {
      if (selectedAppId.value) {
        pagination.current = 1
        getUserList()
        getUserStats()
      }
    }, { immediate: true })
    
    const refreshUsers = () => {
      getUserList()
      getUserStats()
    }
    
    const searchUsers = () => {
      pagination.current = 1
      getUserList()
    }
    
    const resetSearch = () => {
      searchForm.playerId = ''
      searchForm.openId = ''
      pagination.current = 1
      getUserList()
    }
    
    const handleSortChange = ({ column, prop, order }) => {
      // 实现排序逻辑
      getUserList()
    }
    
    const handleSizeChange = (val) => {
      pagination.pageSize = val
      pagination.current = 1
      getUserList()
    }
    
    const handleCurrentChange = (val) => {
      pagination.current = val
      getUserList()
    }
    
    const closeUserDataDialog = () => {
      userDataDialog.visible = false
      userDataDialog.user = null
      userDataDialog.data = null
      userDataDialog.dataStr = ''
    }
    
    onMounted(() => {
      // 组件挂载时如果已有选择的app，则获取数据
      if (selectedAppId.value) {
        getUserList()
        getUserStats()
      }
    })
    
    return {
      loading,
      userList,
      userStats,
      searchForm,
      pagination,
      userDataDialog,
      getUserList,
      viewUserData,
      editUserData,
      saveUserData,
      toggleUserBan,
      deleteUser,
      refreshUsers,
      searchUsers,
      resetSearch,
      handleSortChange,
      handleSizeChange,
      handleCurrentChange,
      closeUserDataDialog,
      getAppName
    }
  }
}
</script>

<style scoped>
.user-management {
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

.player-name {
  font-weight: bold;
  margin-bottom: 5px;
}

.player-avatar {
  display: flex;
  align-items: center;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.user-info {
  margin-bottom: 20px;
  padding: 15px;
  background: #f9f9f9;
  border-radius: 5px;
}

.user-info h3 {
  margin-top: 0;
  color: #333;
}

.game-data h3 {
  color: #333;
  margin-bottom: 10px;
}

.el-button.danger {
  color: #f56c6c;
}

.el-button.success {
  color: #67c23a;
}
</style>