<template>
  <div class="mail-management">
    <div class="page-header">
      <h1>邮件管理</h1>
      <div class="header-actions">
        <el-button 
          type="primary" 
          @click="showCreateDialog"
          v-if="hasPermission(PERMISSIONS.MAIL_MANAGE)"
        >
          创建邮件
        </el-button>
        <el-button @click="refreshMails">刷新</el-button>
        <!-- <el-button 
          type="warning" 
          @click="initMailSystem"
          v-if="hasPermission(PERMISSIONS.MAIL_MANAGE) && !mailSystemInitialized"
        >
          初始化邮件系统
        </el-button> -->
      </div>
    </div>

    <!-- 搜索筛选 -->
    <div class="search-section">
      <el-form :model="searchForm" :inline="true">
        <el-form-item label="邮件标题:">
          <el-input v-model="searchForm.title" placeholder="输入邮件标题" @keyup.enter="searchMails" style="width: 180px"></el-input>
        </el-form-item>
        <el-form-item label="游戏:">
          <el-select v-model="searchForm.appId" placeholder="选择游戏" clearable style="width: 180px">
            <el-option 
              v-for="app in appList" 
              :key="app.appId" 
              :label="app.appName" 
              :value="app.appId">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="状态:">
          <el-select v-model="searchForm.status" placeholder="选择状态" clearable style="width: 150px">
            <el-option label="待发布" value="pending"></el-option>
            <el-option label="定时发布" value="scheduled"></el-option>
            <el-option label="已发布" value="active"></el-option>
            <el-option label="已过期" value="expired"></el-option>
            <el-option label="草稿" value="draft"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="类型:">
          <el-select v-model="searchForm.type" placeholder="选择类型" clearable style="width: 150px">
            <el-option label="系统邮件" value="system"></el-option>
            <el-option label="公告邮件" value="notice"></el-option>
            <el-option label="奖励邮件" value="reward"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchMails">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 邮件统计 -->
    <div class="stats-section" v-if="mailStats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ mailStats.mailStats?.total || 0 }}</div>
              <div class="stat-label">邮件总数</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ mailStats.mailStats?.active || 0 }}</div>
              <div class="stat-label">已发布</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ mailStats.mailStats?.draft || 0 }}</div>
              <div class="stat-label">草稿</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <div class="stat-item">
              <div class="stat-value">{{ mailStats.interactionStats?.readRate || 0 }}%</div>
              <div class="stat-label">平均阅读率</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 批量操作 -->
    <div class="batch-actions" v-if="selectedMails.length > 0">
      <div class="batch-info">
        已选择 {{ selectedMails.length }} 个邮件
      </div>
      <div class="batch-buttons">
        <el-button 
          type="primary" 
          @click="batchPublish" 
          v-if="hasPermission(PERMISSIONS.MAIL_MANAGE)"
          :disabled="!canBatchPublish"
        >
          批量发布
        </el-button>
        <el-button 
          type="warning" 
          @click="batchExpire" 
          v-if="hasPermission(PERMISSIONS.MAIL_MANAGE)"
          :disabled="!canBatchExpire"
        >
          批量下线
        </el-button>
        <el-button 
          type="danger" 
          @click="batchDelete" 
          v-if="hasPermission(PERMISSIONS.MAIL_MANAGE)"
          :disabled="!canBatchDelete"
        >
          批量删除
        </el-button>
        <el-button @click="clearSelection">取消选择</el-button>
      </div>
    </div>

    <!-- 邮件表格 -->
    <el-table 
      :data="mailList" 
      style="width: 100%" 
      v-loading="loading"
      @sort-change="handleSortChange"
      @selection-change="handleSelectionChange"
      ref="mailTableRef">
      <el-table-column type="selection" width="55"></el-table-column>
      <el-table-column prop="title" label="邮件标题" width="200" sortable="custom">
        <template #default="scope">
          <div class="mail-info">
            <div class="mail-title">{{ scope.row.title }}</div>
            <div class="mail-id">ID: {{ scope.row.mailId }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="100">
        <template #default="scope">
          <el-tag :type="getTypeColor(scope.row.type)">
            {{ getTypeText(scope.row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="targetType" label="目标类型" width="120">
        <template #default="scope">
          <el-tag type="info">
            {{ getTargetTypeText(scope.row.targetType) }}
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
      <el-table-column prop="createTime" label="创建时间" width="160" sortable="custom">
      </el-table-column>
      <el-table-column prop="publishTime" label="发布时间" width="160">
        <template #default="scope">
          <span v-if="scope.row.publishTime">{{ scope.row.publishTime }}</span>
          <span v-else class="text-muted">立即发布</span>
        </template>
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
      <el-table-column label="操作" width="320" fixed="right">
        <template #default="scope">
          <el-button link @click="viewMailDetail(scope.row)">详情</el-button>
          <el-button link @click="viewMailStats(scope.row)" v-if="scope.row.status === 'active'">统计</el-button>
          <el-button link @click="copyMail(scope.row)" v-if="hasPermission(PERMISSIONS.MAIL_MANAGE)">复制</el-button>
          <el-button 
            link 
            @click="editMail(scope.row)" 
            v-if="(scope.row.status === 'pending' || scope.row.status === 'scheduled') && hasPermission(PERMISSIONS.MAIL_MANAGE)"
          >
            编辑
          </el-button>
          <el-button 
            link 
            class="success"
            @click="publishMail(scope.row)" 
            v-if="(scope.row.status === 'pending' || scope.row.status === 'scheduled') && hasPermission(PERMISSIONS.MAIL_MANAGE)"
          >
            发布
          </el-button>
          <el-button 
            link 
            class="warning"
            @click="expireMail(scope.row)" 
            v-if="scope.row.status === 'active' && hasPermission(PERMISSIONS.MAIL_MANAGE)"
          >
            下线
          </el-button>
          <el-button 
            link 
            class="primary"
            @click="republishMail(scope.row)" 
            v-if="scope.row.status === 'expired' && hasPermission(PERMISSIONS.MAIL_MANAGE)"
          >
            重新发布
          </el-button>
          <el-button 
            link 
            class="danger" 
            @click="deleteMail(scope.row)" 
            v-if="hasPermission(PERMISSIONS.MAIL_MANAGE)"
          >
            删除
          </el-button>
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

    <!-- 创建/编辑邮件对话框 -->
    <el-dialog 
      v-model="mailDialog.visible" 
      :title="mailDialog.isEdit ? '编辑邮件' : '创建邮件'" 
      width="800px">
      <el-form :model="mailDialog.form" :rules="mailRules" ref="mailFormRef" label-width="120px">
        <!-- 游戏信息显示，不可选择 -->
        <el-form-item label="游戏">
          <span class="selected-app-info">{{ getAppName(selectedAppId) }}</span>
        </el-form-item>
        <el-form-item label="邮件标题" prop="title">
          <el-input v-model="mailDialog.form.title" placeholder="请输入邮件标题" style="width: 400px"></el-input>
        </el-form-item>
        <el-form-item label="邮件类型" prop="type">
          <el-radio-group v-model="mailDialog.form.type">
            <el-radio label="system">系统邮件</el-radio>
            <el-radio label="notice">公告邮件</el-radio>
            <el-radio label="reward">奖励邮件</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="目标类型" prop="targetType">
          <el-radio-group v-model="mailDialog.form.targetType" @change="onTargetTypeChange">
            <el-radio label="all">全部玩家</el-radio>
            <el-radio label="user">指定玩家</el-radio>
            <el-radio label="level">等级范围</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="目标用户" v-if="mailDialog.form.targetType === 'user'" prop="targetUsers">
          <el-input 
            v-model="targetUsersText" 
            type="textarea" 
            :rows="3" 
            placeholder="请输入用户playerId，一行一个">
          </el-input>
          <div class="form-tip">每行输入一个用户playerId</div>
        </el-form-item>
        <el-form-item label="等级范围" v-if="mailDialog.form.targetType === 'level'">
          <el-input-number v-model="mailDialog.form.minLevel" :min="0" placeholder="最小等级" style="width: 120px"></el-input-number>
          <span style="margin: 0 10px;">至</span>
          <el-input-number v-model="mailDialog.form.maxLevel" :min="0" placeholder="最大等级" style="width: 120px"></el-input-number>
        </el-form-item>
        <el-form-item label="邮件内容" prop="content">
          <el-input 
            v-model="mailDialog.form.content" 
            type="textarea" 
            :rows="6" 
            placeholder="请输入邮件内容">
          </el-input>
        </el-form-item>
        <el-form-item label="奖励设置">
          <div class="rewards-section">
            <div v-for="(reward, index) in mailDialog.form.rewards" :key="index" class="reward-item">
              <el-input v-model="reward.type" placeholder="奖励类型" style="width: 120px"></el-input>
              <el-input-number v-model="reward.amount" :min="1" placeholder="数量" style="width: 100px; margin: 0 10px"></el-input-number>
              <el-input v-model="reward.name" placeholder="奖励名称" style="width: 150px"></el-input>
              <el-button link class="danger" @click="removeReward(index)">删除</el-button>
            </div>
            <el-button link @click="addReward">添加奖励</el-button>
          </div>
        </el-form-item>
        <el-form-item label="发布时间">
          <el-date-picker
            v-model="mailDialog.form.publishTime"
            type="datetime"
            placeholder="选择发布时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px">
          </el-date-picker>
          <div class="form-tip">不设置则立即发布</div>
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker
            v-model="mailDialog.form.expireTime"
            type="datetime"
            placeholder="选择过期时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px">
          </el-date-picker>
          <div class="form-tip">不设置则永不过期</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="mailDialog.visible = false" :disabled="saving">取消</el-button>
          <el-button @click="saveMail(false)" :loading="saving" :disabled="saving">保存草稿</el-button>
          <el-button type="primary" @click="saveMail(true)" :loading="saving" :disabled="saving">保存并发布</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 邮件详情对话框 -->
    <el-dialog v-model="detailDialog.visible" title="邮件详情" width="800px">
      <div v-if="detailDialog.mail">
        <el-descriptions border :column="2">
          <el-descriptions-item label="邮件标题">{{ detailDialog.mail.title }}</el-descriptions-item>
          <el-descriptions-item label="邮件ID">{{ detailDialog.mail.id }}</el-descriptions-item>
          <el-descriptions-item label="游戏">{{ getAppName(detailDialog.mail.appId) }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ getTypeText(detailDialog.mail.type) }}</el-descriptions-item>
          <el-descriptions-item label="目标类型">{{ getTargetTypeText(detailDialog.mail.targetType) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusColor(detailDialog.mail.status)">
              {{ getStatusText(detailDialog.mail.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detailDialog.mail.createTime }}</el-descriptions-item>
          <el-descriptions-item label="发布时间">{{ detailDialog.mail.publishTime || '立即发布' }}</el-descriptions-item>
          <el-descriptions-item label="过期时间">{{ detailDialog.mail.expireTime || '永不过期' }}</el-descriptions-item>
          <el-descriptions-item label="内容" :span="2">{{ detailDialog.mail.content }}</el-descriptions-item>
        </el-descriptions>
        
        <div v-if="detailDialog.mail.rewards && detailDialog.mail.rewards.length > 0" style="margin-top: 20px;">
          <h4>奖励列表</h4>
          <el-table :data="detailDialog.mail.rewards" style="width: 100%">
            <el-table-column prop="type" label="类型" width="120"></el-table-column>
            <el-table-column prop="name" label="名称" width="150"></el-table-column>
            <el-table-column prop="amount" label="数量" width="100"></el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>

    <!-- 邮件统计对话框 -->
    <el-dialog v-model="statsDialog.visible" title="邮件统计" width="800px">
      <div v-if="statsDialog.stats">
        <el-row :gutter="20" style="margin-bottom: 20px;">
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ statsDialog.stats.totalUsers }}</div>
                <div class="stat-label">交互用户</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ statsDialog.stats.readRate }}%</div>
                <div class="stat-label">阅读率</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ statsDialog.stats.receiveRate }}%</div>
                <div class="stat-label">领取率</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div class="stat-item">
                <div class="stat-value">{{ statsDialog.stats.deleteRate }}%</div>
                <div class="stat-label">删除率</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="stat-detail">
              <div class="stat-number">{{ statsDialog.stats.readUsers }}</div>
              <div class="stat-text">已阅读用户</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-detail">
              <div class="stat-number">{{ statsDialog.stats.receivedUsers }}</div>
              <div class="stat-text">已领取用户</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-detail">
              <div class="stat-number">{{ statsDialog.stats.deletedUsers }}</div>
              <div class="stat-text">已删除用户</div>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mailAPI, appAPI } from '../services/api.js'
import { hasPermission, PERMISSIONS, getAdminInfo } from '../utils/auth.js'
import { appList, getAppName, selectedAppId } from '../utils/appStore.js'

export default {
  name: 'MailManagement',
  setup() {
    const loading = ref(false)
    const saving = ref(false) // 添加保存状态
    const mailList = ref([])
    const mailStats = ref(null)
    const mailSystemInitialized = ref(false) // 邮件系统初始化状态，默认为未知，需要检测
    const selectedMails = ref([]) // 选中的邮件列表
    const mailTableRef = ref(null) // 表格引用
    
    const searchForm = reactive({
      title: '',
      appId: '',
      status: '',
      type: ''
    })
    
    const pagination = reactive({
      current: 1,
      pageSize: 20,
      total: 0
    })
    
    const mailDialog = reactive({
      visible: false,
      isEdit: false,
      form: {
        appId: '', // 这个会在 showCreateDialog 中设置为当前选择的应用ID
        title: '',
        content: '',
        type: 'system',
        targetType: 'all',
        targetUsers: [],
        minLevel: 1,
        maxLevel: 999,
        rewards: [],
        publishTime: '',
        expireTime: ''
      }
    })
    
    const detailDialog = reactive({
      visible: false,
      mail: null
    })
    
    const statsDialog = reactive({
      visible: false,
      stats: null
    })
    
    const targetUsersText = ref('')
    
    // 批量操作的计算属性
    const canBatchPublish = computed(() => {
      return selectedMails.value.some(mail => mail.status === 'pending' || mail.status === 'scheduled')
    })
    
    const canBatchExpire = computed(() => {
      return selectedMails.value.some(mail => mail.status === 'active')
    })
    
    const canBatchDelete = computed(() => {
      return selectedMails.value.some(mail => mail.status === 'pending' || mail.status === 'scheduled' || mail.status === 'expired')
    })
    
    const mailRules = {
      title: [
        { required: true, message: '请输入邮件标题', trigger: 'blur' }
      ],
      content: [
        { required: true, message: '请输入邮件内容', trigger: 'blur' }
      ],
      type: [
        { required: true, message: '请选择邮件类型', trigger: 'change' }
      ],
      targetType: [
        { required: true, message: '请选择目标类型', trigger: 'change' }
      ]
    }
    
    
    // 获取邮件列表
    const getMailList = async () => {
      loading.value = true
      try {
        const params = {
          page: pagination.current,
          pageSize: pagination.pageSize,
          ...searchForm,
          // 如果没有选择特定应用，则查询当前选择的应用
          appId: searchForm.appId || selectedAppId.value
        }
        
        const result = await mailAPI.getAll(params)
        if (result.code === 0) {
          mailList.value = result.data?.list || []
          pagination.total = result.data?.total || 0
          // 如果成功获取邮件列表，说明邮件系统已经初始化
          mailSystemInitialized.value = true
        } else {
          mailList.value = []
          // 检查是否是因为系统未初始化导致的错误
          if (result.msg && result.msg.includes('集合') && result.msg.includes('不存在')) {
            mailSystemInitialized.value = false
            ElMessage.warning('邮件系统尚未初始化，请先初始化邮件系统')
          } else {
            ElMessage.error(result.msg || '获取邮件列表失败')
          }
        }
      } catch (error) {
        console.error('获取邮件列表失败:', error)
        mailList.value = []
        // 如果是网络错误或其他错误，也可能表示系统未初始化
        mailSystemInitialized.value = false
        ElMessage.error('获取邮件列表失败，请检查邮件系统是否已初始化')
      } finally {
        loading.value = false
      }
    }
    
    // 获取邮件统计
    const getMailStats = async () => {
      try {
        const result = await mailAPI.getStats({
          appId: selectedAppId.value
        })
        if (result.code === 0) {
          mailStats.value = result.data
          // 如果成功获取统计，说明邮件系统已经初始化
          mailSystemInitialized.value = true
        } else {
          // 如果获取统计失败，可能是系统未初始化
          if (result.msg && result.msg.includes('集合')) {
            mailSystemInitialized.value = false
          }
        }
      } catch (error) {
        console.error('获取邮件统计失败:', error)
        // 如果获取统计失败，可能是系统未初始化
        mailSystemInitialized.value = false
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
    
    const getTargetTypeText = (targetType) => {
      const types = {
        'all': '全部玩家',
        'user': '指定玩家',
        'level': '等级范围'
      }
      return types[targetType] || targetType
    }
    
    const getStatusText = (status) => {
      const statuses = {
        'pending': '待发布',
        'scheduled': '定时发布',
        'active': '已发布',
        'expired': '已过期',
        'draft': '草稿' // 保留兼容性
      }
      return statuses[status] || status
    }
    
    const getStatusColor = (status) => {
      const colors = {
        'pending': 'warning',
        'scheduled': 'info',
        'active': 'success',
        'expired': 'danger',
        'draft': 'info' // 保留兼容性
      }
      return colors[status] || 'info'
    }
    
    // 事件处理
    const showCreateDialog = () => {
      mailDialog.isEdit = false
      mailDialog.form = {
        appId: selectedAppId.value, // 使用全局选择的应用ID
        title: '',
        content: '',
        type: 'system',
        targetType: 'all',
        targetUsers: [],
        minLevel: 1,
        maxLevel: 999,
        rewards: [],
        publishTime: '',
        expireTime: ''
      }
      targetUsersText.value = ''
      mailDialog.visible = true
    }
    
    const editMail = (mail) => {
      mailDialog.isEdit = true
      mailDialog.form = { 
        ...mail,
        appId: selectedAppId.value // 使用当前选择的应用ID，而不是邮件原有的应用ID
      }
      if (mail.targetUsers && mail.targetUsers.length > 0) {
        targetUsersText.value = mail.targetUsers.join('\n')
      }
      mailDialog.visible = true
    }
    
    const onTargetTypeChange = () => {
      if (mailDialog.form.targetType !== 'user') {
        targetUsersText.value = ''
        mailDialog.form.targetUsers = []
      }
    }
    
    const addReward = () => {
      mailDialog.form.rewards.push({
        type: '',
        name: '',
        amount: 1
      })
    }
    
    const removeReward = (index) => {
      mailDialog.form.rewards.splice(index, 1)
    }
    
    const saveMail = async (publish = false) => {
      if (saving.value) return // 防止重复提交
      
      try {
        saving.value = true
        
        // 处理目标用户
        if (mailDialog.form.targetType === 'user' && targetUsersText.value) {
          mailDialog.form.targetUsers = targetUsersText.value.split('\n').filter(id => id.trim())
        }
        
        const data = { ...mailDialog.form }
        data.createBy = getAdminInfo().id
        
        let result
        if (mailDialog.isEdit) {
          if (publish) {
            // 编辑时如果需要发布，先更新再发布
            result = await mailAPI.update(data)
            if (result.code === 0) {
              // 更新成功后，如果是草稿状态，则发布
              if (data.status === 'draft' || data.status === 'pending') {
                const publishResult = await mailAPI.publish({
                  id: data.id,
                  appId: selectedAppId.value
                })
                if (publishResult.code === 0) {
                  ElMessage.success('更新并发布成功')
                } else {
                  ElMessage.error(publishResult.msg || '发布失败')
                  return
                }
              } else {
                ElMessage.success('更新成功')
              }
            }
          } else {
            result = await mailAPI.update(data)
            if (result.code === 0) {
              ElMessage.success('更新成功')
            }
          }
        } else if (publish) {
          // 如果需要立即发布，直接调用发布API而不是先创建再发布
          const publishData = {
            appId: selectedAppId.value,
            title: data.title,
            content: data.content,
            type: data.type,
            targetType: data.targetType,
            targets: data.targets,
            rewards: data.rewards ? data.rewards.map(r => `${r.type}:${r.amount}:${r.name}`).join(',') : '',
            expireDays: data.expireDays || 7
          }
          
          result = await mailAPI.publish(publishData)
          if (result.code === 0) {
            ElMessage.success('邮件发布成功')
          }
        } else {
          result = await mailAPI.create(data)
          if (result.code === 0) {
            ElMessage.success('创建成功')
          }
        }
        
        if (result.code === 0) {
          mailDialog.visible = false
          await getMailList()
          await getMailStats()
        } else {
          ElMessage.error(result.msg || '操作失败')
        }
      } catch (error) {
        console.error('保存邮件失败:', error)
        ElMessage.error('操作失败')
      } finally {
        saving.value = false
      }
    }
    
    const publishMail = async (mail) => {
      try {
        const title = mail.title || '该邮件'
        await ElMessageBox.confirm(`确定要发布邮件 "${title}" 吗？发布后不可修改内容。`, '确认发布')
        
        // 准备发布数据，根据新的API要求
        const publishData = {
          appId: selectedAppId.value,
          title: mail.title,
          content: mail.content,
          rewards: mail.rewards ? mail.rewards.map(r => `${r.type}:${r.amount}:${r.name}`).join(',') : '',
          expireDays: 7 // 默认7天过期
        }
        
        const result = await mailAPI.publish(publishData)
        
        if (result.code === 0) {
          ElMessage.success('邮件发布成功')
          await getMailList()
          await getMailStats()
        } else {
          ElMessage.error(result.msg || '发布失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('发布邮件失败:', error)
          ElMessage.error('发布失败')
        }
      }
    }
    
    const expireMail = async (mail) => {
      try {
        await ElMessageBox.confirm(`确定要下线邮件 "${mail.title}" 吗？`, '确认下线')
        
        const result = await mailAPI.update({
          mailId: mail.id,
          status: 'expired'
        })
        
        if (result.code === 0) {
          ElMessage.success('邮件已下线')
          await getMailList()
          await getMailStats()
        } else {
          ElMessage.error(result.msg || '下线失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('下线邮件失败:', error)
          ElMessage.error('下线失败')
        }
      }
    }
    
    const deleteMail = async (mail) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除邮件 "${mail.title}" 吗？此操作不可恢复！`, 
          '危险操作', 
          { type: 'warning' }
        )
        
        const result = await mailAPI.delete({
          appId: selectedAppId.value,
          mailId: mail.id
        })
        
        if (result.code === 0) {
          ElMessage.success('删除成功')
          await getMailList()
          await getMailStats()
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
    
    const copyMail = (mail) => {
      mailDialog.isEdit = false
      mailDialog.form = {
        appId: selectedAppId.value, // 使用当前选择的应用ID
        title: `${mail.title} - 副本`,
        content: mail.content,
        type: mail.type,
        targetType: mail.targetType,
        targetUsers: mail.targetUsers ? [...mail.targetUsers] : [],
        minLevel: mail.minLevel || 1,
        maxLevel: mail.maxLevel || 999,
        rewards: mail.rewards ? JSON.parse(JSON.stringify(mail.rewards)) : [],
        publishTime: '', // 清空发布时间
        expireTime: ''   // 清空过期时间
      }
      
      if (mail.targetUsers && mail.targetUsers.length > 0) {
        targetUsersText.value = mail.targetUsers.join('\n')
      } else {
        targetUsersText.value = ''
      }
      
      mailDialog.visible = true
      ElMessage.success('邮件已复制，请修改后保存')
    }
    
    const republishMail = async (mail) => {
      try {
        await ElMessageBox.confirm(
          `确定要重新发布邮件 "${mail.title}" 吗？重新发布后邮件状态将变为已发布。`, 
          '确认重新发布'
        )
        
        // 更新邮件状态为已发布，并重置发布时间
        const result = await mailAPI.update({
          mailId: mail.id,
          status: 'active',
          publishTime: new Date().toISOString().replace('T', ' ').substring(0, 19),
          // 重新设置过期时间（7天后）
          expireTime: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().replace('T', ' ').substring(0, 19)
        })
        
        if (result.code === 0) {
          ElMessage.success('邮件重新发布成功')
          await getMailList()
          await getMailStats()
        } else {
          ElMessage.error(result.msg || '重新发布失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('重新发布邮件失败:', error)
          ElMessage.error('重新发布失败')
        }
      }
    }
    
    const viewMailDetail = (mail) => {
      detailDialog.mail = mail
      detailDialog.visible = true
    }
    
    const viewMailStats = async (mail) => {
      try {
        const result = await mailAPI.getStats({ mailId: mail.id })
        if (result.code === 0) {
          statsDialog.stats = result.data.stats
          statsDialog.visible = true
        } else {
          ElMessage.error(result.msg || '获取统计失败')
        }
      } catch (error) {
        console.error('获取邮件统计失败:', error)
        ElMessage.error('获取统计失败')
      }
    }
    
    const refreshMails = () => {
      getMailList()
      getMailStats()
    }
    
    const initMailSystem = async () => {
      try {
        await ElMessageBox.confirm(
          '确认要初始化邮件系统吗？这会创建必要的数据库集合。',
          '初始化邮件系统',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        const result = await mailAPI.initSystem()
        if (result.code === 0) {
          ElMessage.success('邮件系统初始化成功')
          mailSystemInitialized.value = true // 更新初始化状态
          refreshMails()
        } else if (result.code === 4003) {
          // 系统已经初始化
          ElMessage.info('邮件系统已经初始化')
          mailSystemInitialized.value = true
          refreshMails()
        } else {
          ElMessage.error(result.msg || '初始化失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('初始化邮件系统失败:', error)
          ElMessage.error('初始化失败')
        }
      }
    }
    
    const searchMails = () => {
      pagination.current = 1
      getMailList()
    }
    
    const resetSearch = () => {
      searchForm.title = ''
      searchForm.appId = ''
      searchForm.status = ''
      searchForm.type = ''
      pagination.current = 1
      getMailList()
    }
    
    const handleSortChange = ({ column, prop, order }) => {
      // 实现排序逻辑
      getMailList()
    }
    
    const handleSizeChange = (val) => {
      pagination.pageSize = val
      pagination.current = 1
      getMailList()
    }
    
    const handleCurrentChange = (val) => {
      pagination.current = val
      getMailList()
    }
    
    // 批量操作相关方法
    const handleSelectionChange = (selection) => {
      selectedMails.value = selection
    }
    
    const clearSelection = () => {
      mailTableRef.value?.clearSelection()
      selectedMails.value = []
    }
    
    const batchPublish = async () => {
      const draftMails = selectedMails.value.filter(mail => 
        mail.status === 'pending' || mail.status === 'scheduled'
      )
      if (draftMails.length === 0) {
        ElMessage.warning('没有可发布的邮件')
        return
      }
      
      try {
        await ElMessageBox.confirm(
          `确定要批量发布 ${draftMails.length} 个邮件吗？发布后不可修改内容。`,
          '批量发布确认'
        )
        
        let successCount = 0
        let failCount = 0
        
        for (const mail of draftMails) {
          try {
            const publishData = {
              appId: selectedAppId.value,
              title: mail.title,
              content: mail.content,
              rewards: mail.rewards ? mail.rewards.map(r => `${r.type}:${r.amount}:${r.name}`).join(',') : '',
              expireDays: 7
            }
            const result = await mailAPI.publish(publishData)
            if (result.code === 0) {
              successCount++
            } else {
              failCount++
            }
          } catch (error) {
            failCount++
          }
        }
        
        if (successCount > 0) {
          ElMessage.success(`成功发布 ${successCount} 个邮件${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
          clearSelection()
          await getMailList()
          await getMailStats()
        } else {
          ElMessage.error('批量发布失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('批量发布失败:', error)
          ElMessage.error('批量发布失败')
        }
      }
    }
    
    const batchExpire = async () => {
      const activeMails = selectedMails.value.filter(mail => mail.status === 'active')
      if (activeMails.length === 0) {
        ElMessage.warning('没有可下线的邮件')
        return
      }
      
      try {
        await ElMessageBox.confirm(
          `确定要批量下线 ${activeMails.length} 个邮件吗？`,
          '批量下线确认'
        )
        
        let successCount = 0
        let failCount = 0
        
        for (const mail of activeMails) {
          try {
            const result = await mailAPI.update({
              mailId: mail.id,
              status: 'expired'
            })
            if (result.code === 0) {
              successCount++
            } else {
              failCount++
            }
          } catch (error) {
            failCount++
          }
        }
        
        if (successCount > 0) {
          ElMessage.success(`成功下线 ${successCount} 个邮件${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
          clearSelection()
          await getMailList()
          await getMailStats()
        } else {
          ElMessage.error('批量下线失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('批量下线失败:', error)
          ElMessage.error('批量下线失败')
        }
      }
    }
    
    const batchDelete = async () => {
      const deletableMails = selectedMails.value.filter(mail => 
        mail.status === 'pending' || mail.status === 'scheduled' || mail.status === 'expired'
      )
      if (deletableMails.length === 0) {
        ElMessage.warning('没有可删除的邮件')
        return
      }
      
      try {
        await ElMessageBox.confirm(
          `确定要批量删除 ${deletableMails.length} 个邮件吗？此操作不可恢复！`,
          '危险操作',
          { type: 'warning' }
        )
        
        let successCount = 0
        let failCount = 0
        
        for (const mail of deletableMails) {
          try {
            const result = await mailAPI.delete({
              appId: selectedAppId.value,
              mailId: mail.id
            })
            if (result.code === 0) {
              successCount++
            } else {
              failCount++
            }
          } catch (error) {
            failCount++
          }
        }
        
        if (successCount > 0) {
          ElMessage.success(`成功删除 ${successCount} 个邮件${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
          clearSelection()
          await getMailList()
          await getMailStats()
        } else {
          ElMessage.error('批量删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('批量删除失败:', error)
          ElMessage.error('批量删除失败')
        }
      }
    }
    
    onMounted(() => {
      getMailList()
      getMailStats()
    })
    
    return {
      loading,
      saving,
      mailList,
      mailStats,
      mailSystemInitialized,
      selectedMails,
      mailTableRef,
      appList,
      searchForm,
      pagination,
      mailDialog,
      detailDialog,
      statsDialog,
      mailRules,
      targetUsersText,
      canBatchPublish,
      canBatchExpire,
      canBatchDelete,
      getAppName,
      getTypeText,
      getTypeColor,
      getTargetTypeText,
      getStatusText,
      getStatusColor,
      showCreateDialog,
      editMail,
      onTargetTypeChange,
      addReward,
      removeReward,
      saveMail,
      publishMail,
      expireMail,
      deleteMail,
      copyMail,
      republishMail,
      viewMailDetail,
      viewMailStats,
      refreshMails,
      initMailSystem,
      searchMails,
      resetSearch,
      handleSortChange,
      handleSizeChange,
      handleCurrentChange,
      handleSelectionChange,
      clearSelection,
      batchPublish,
      batchExpire,
      batchDelete,
      hasPermission,
      PERMISSIONS,
      selectedAppId
    }
  }
}
</script>

<style scoped>
.mail-management {
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

.mail-info {
  display: flex;
  flex-direction: column;
}

.mail-title {
  font-weight: bold;
  font-size: 14px;
  margin-bottom: 4px;
}

.mail-id {
  font-size: 12px;
  color: #999;
}

.text-muted {
  color: #999;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.rewards-section {
  border: 1px solid #ddd;
  padding: 15px;
  border-radius: 4px;
  background: #fafafa;
}

.reward-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.stat-detail {
  text-align: center;
  padding: 15px;
}

.stat-number {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 5px;
}

.stat-text {
  font-size: 14px;
  color: #666;
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

.el-button.primary {
  color: #409eff;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.selected-app-info {
  font-weight: bold;
  color: #409eff;
  padding: 8px 12px;
  background: #f0f9ff;
  border: 1px solid #d1ecf1;
  border-radius: 4px;
  display: inline-block;
}

.batch-actions {
  background: #f0f9ff;
  border: 1px solid #d1ecf1;
  border-radius: 8px;
  padding: 15px 20px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.batch-info {
  font-size: 14px;
  color: #409eff;
  font-weight: bold;
}

.batch-buttons {
  display: flex;
  gap: 10px;
}
</style> 