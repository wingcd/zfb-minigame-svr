<template>
  <div class="mail-test">
    <div class="page-header">
      <h1>邮件系统测试</h1>
      <div class="header-actions">
        <el-button @click="refreshUserMails">刷新邮件</el-button>
      </div>
    </div>

    <!-- 测试用户信息 -->
    <div class="test-section">
      <el-card header="测试用户设置">
        <el-form :model="testUser" :inline="true">
          <el-form-item label="openId:">
            <el-input v-model="testUser.openId" placeholder="输入测试用户openId" style="width: 200px"></el-input>
          </el-form-item>
          <el-form-item label="游戏:">
            <el-select v-model="testUser.appId" placeholder="选择游戏" style="width: 180px">
              <el-option 
                v-for="app in appList" 
                :key="app.appId" 
                :label="app.appName" 
                :value="app.appId">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="getUserMails">获取邮件列表</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <!-- 邮件列表 -->
    <div class="mail-list-section">
      <el-card header="用户邮件列表">
        <el-table :data="userMails" style="width: 100%" v-loading="loading">
          <el-table-column prop="title" label="邮件标题" width="200">
          </el-table-column>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="scope">
              <el-tag :type="getTypeColor(scope.row.type)">
                {{ getTypeText(scope.row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="奖励" width="100">
            <template #default="scope">
              <el-tag v-if="scope.row.rewards && scope.row.rewards.length > 0" type="warning">
                {{ scope.row.rewards.length }} 项
              </el-tag>
              <span v-else class="text-muted">无</span>
            </template>
          </el-table-column>
          <el-table-column prop="publishTime" label="发布时间" width="160">
          </el-table-column>
          <el-table-column prop="expireTime" label="过期时间" width="160">
            <template #default="scope">
              <span v-if="scope.row.expireTime">{{ scope.row.expireTime }}</span>
              <span v-else class="text-muted">永不过期</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusColor(scope.row.status)">
                {{ getStatusText(scope.row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="250">
            <template #default="scope">
              <el-button 
                type="text" 
                @click="readMail(scope.row)" 
                v-if="!scope.row.isRead">
                阅读
              </el-button>
              <el-button 
                type="text" 
                class="success"
                @click="receiveMail(scope.row)" 
                v-if="scope.row.isRead && !scope.row.isReceived && scope.row.rewards && scope.row.rewards.length > 0">
                领取奖励
              </el-button>
              <el-button 
                type="text" 
                class="danger"
                @click="deleteMail(scope.row)" 
                v-if="!scope.row.isDeleted">
                删除
              </el-button>
              <span v-if="scope.row.isReceived" class="text-success">已领取</span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 邮件详情对话框 -->
    <el-dialog v-model="detailDialog.visible" title="邮件详情" width="600px">
      <div v-if="detailDialog.mail">
        <h3>{{ detailDialog.mail.title }}</h3>
        <div class="mail-content">
          {{ detailDialog.mail.content }}
        </div>
        
        <div v-if="detailDialog.mail.rewards && detailDialog.mail.rewards.length > 0" class="rewards-section">
          <h4>奖励内容</h4>
          <el-table :data="detailDialog.mail.rewards" style="width: 100%">
            <el-table-column prop="type" label="类型" width="120"></el-table-column>
            <el-table-column prop="name" label="名称" width="150"></el-table-column>
            <el-table-column prop="amount" label="数量" width="100"></el-table-column>
          </el-table>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="detailDialog.visible = false">关闭</el-button>
          <el-button 
            type="primary" 
            @click="receiveFromDetail" 
            v-if="detailDialog.mail && detailDialog.mail.isRead && !detailDialog.mail.isReceived && detailDialog.mail.rewards && detailDialog.mail.rewards.length > 0">
            领取奖励
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mailAPI, appAPI } from '../services/api.js'

export default {
  name: 'MailTest',
  setup() {
    const loading = ref(false)
    const userMails = ref([])
    const appList = ref([])
    
    const testUser = reactive({
      openId: 'test_user_001',
      appId: ''
    })
    
    const detailDialog = reactive({
      visible: false,
      mail: null
    })
    
    // 获取应用列表
    const getAppList = async () => {
      try {
        const result = await appAPI.getAll({ page: 1, pageSize: 1000 })
        if (result.code === 0) {
          appList.value = result.data?.list || []
          if (appList.value.length > 0) {
            testUser.appId = appList.value[0].appId
          }
        }
      } catch (error) {
        console.error('获取应用列表失败:', error)
      }
    }
    
    // 获取用户邮件列表
    const getUserMails = async () => {
      if (!testUser.openId || !testUser.appId) {
        ElMessage.warning('请输入完整的测试用户信息')
        return
      }
      
      loading.value = true
      try {
        const result = await mailAPI.getUserMails({
          openId: testUser.openId,
          appId: testUser.appId,
          page: 1,
          pageSize: 100
        })
        
        if (result.code === 0) {
          userMails.value = result.data?.list || []
          ElMessage.success(`获取到 ${userMails.value.length} 封邮件`)
        } else {
          userMails.value = []
          ElMessage.error(result.msg || '获取邮件失败')
        }
      } catch (error) {
        console.error('获取用户邮件失败:', error)
        userMails.value = []
        ElMessage.error('获取邮件失败')
      } finally {
        loading.value = false
      }
    }
    
    // 阅读邮件
    const readMail = async (mail) => {
      try {
        const result = await mailAPI.updateStatus({
          openId: testUser.openId,
          appId: testUser.appId,
          mailId: mail.mailId,
          action: 'read'
        })
        
        if (result.code === 0) {
          ElMessage.success('邮件已标记为已读')
          
          // 显示邮件详情
          detailDialog.mail = { ...mail, isRead: true }
          detailDialog.visible = true
          
          // 刷新列表
          await getUserMails()
        } else {
          ElMessage.error(result.msg || '操作失败')
        }
      } catch (error) {
        console.error('阅读邮件失败:', error)
        ElMessage.error('操作失败')
      }
    }
    
    // 领取奖励
    const receiveMail = async (mail) => {
      try {
        const result = await mailAPI.updateStatus({
          openId: testUser.openId,
          appId: testUser.appId,
          mailId: mail.mailId,
          action: 'receive'
        })
        
        if (result.code === 0) {
          const rewards = result.data?.rewards || []
          if (rewards.length > 0) {
            const rewardText = rewards.map(r => `${r.name} x${r.amount}`).join(', ')
            ElMessage.success(`奖励领取成功！获得: ${rewardText}`)
          } else {
            ElMessage.success('奖励领取成功！')
          }
          
          // 刷新列表
          await getUserMails()
        } else {
          ElMessage.error(result.msg || '领取失败')
        }
      } catch (error) {
        console.error('领取奖励失败:', error)
        ElMessage.error('领取失败')
      }
    }
    
    // 删除邮件
    const deleteMail = async (mail) => {
      try {
        await ElMessageBox.confirm(`确定要删除邮件 "${mail.title}" 吗？`, '确认删除')
        
        const result = await mailAPI.updateStatus({
          openId: testUser.openId,
          appId: testUser.appId,
          mailId: mail.mailId,
          action: 'delete'
        })
        
        if (result.code === 0) {
          ElMessage.success('邮件已删除')
          // 刷新列表
          await getUserMails()
        } else {
          ElMessage.error(result.msg || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除邮件失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    // 从详情页领取奖励
    const receiveFromDetail = async () => {
      if (detailDialog.mail) {
        await receiveMail(detailDialog.mail)
        detailDialog.visible = false
      }
    }
    
    // 刷新用户邮件
    const refreshUserMails = () => {
      if (testUser.openId && testUser.appId) {
        getUserMails()
      }
    }
    
    // 工具函数
    const getTypeText = (type) => {
      const types = {
        'system': '系统邮件',
        'notice': '公告邮件',
        'reward': '奖励邮件'
      }
      return types[type] || type
    }
    
    const getTypeColor = (type) => {
      const colors = {
        'system': 'primary',
        'notice': 'warning',
        'reward': 'success'
      }
      return colors[type] || 'info'
    }
    
    const getStatusText = (status) => {
      const statuses = {
        'unread': '未读',
        'read': '已读',
        'received': '已领取',
        'deleted': '已删除'
      }
      return statuses[status] || status
    }
    
    const getStatusColor = (status) => {
      const colors = {
        'unread': 'danger',
        'read': 'warning',
        'received': 'success',
        'deleted': 'info'
      }
      return colors[status] || 'info'
    }
    
    onMounted(() => {
      getAppList()
    })
    
    return {
      loading,
      userMails,
      appList,
      testUser,
      detailDialog,
      getAppList,
      getUserMails,
      readMail,
      receiveMail,
      deleteMail,
      receiveFromDetail,
      refreshUserMails,
      getTypeText,
      getTypeColor,
      getStatusText,
      getStatusColor
    }
  }
}
</script>

<style scoped>
.mail-test {
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

.test-section {
  margin-bottom: 20px;
}

.mail-list-section {
  margin-bottom: 20px;
}

.text-muted {
  color: #999;
}

.text-success {
  color: #67c23a;
  font-weight: bold;
}

.mail-content {
  padding: 15px;
  background: #f9f9f9;
  border-radius: 4px;
  margin: 15px 0;
  white-space: pre-wrap;
}

.rewards-section {
  margin-top: 20px;
}

.rewards-section h4 {
  margin-bottom: 10px;
  color: #333;
}

.el-button.danger {
  color: #f56c6c;
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