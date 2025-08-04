<template>
  <div id="app">
    <!-- 登录页面 -->
    <router-view v-if="$route.path === '/login'" />
    
    <!-- 主界面布局 -->
    <el-container v-else class="layout-container">
      <!-- 侧边栏 -->
      <el-aside width="200px" class="sidebar">
        <div class="logo">
          <h3>小游戏管理后台</h3>
        </div>
        
        <el-menu
          :default-active="$route.path"
          router
          background-color="#001529"
          text-color="#fff"
          active-text-color="#1890ff"
        >
          <el-menu-item
            v-for="route in visibleRoutes"
            :key="route.path"
            :index="route.path"
          >
            <el-icon v-if="route.meta.icon">
              <component :is="route.meta.icon" />
            </el-icon>
            <span>{{ route.meta.title }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <!-- 头部 -->
        <el-header class="header">
          <div class="header-left">
            <h4>{{ currentPageTitle }}</h4>
          </div>
          
          <div class="header-right">
            <el-dropdown @command="handleCommand">
              <span class="user-info">
                <el-icon><Avatar /></el-icon>
                {{ adminInfo?.nickname || adminInfo?.username || '管理员' }}
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                  <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
                  <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <!-- 内容区 -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showChangePasswordDialog"
      title="修改密码"
      width="400px"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="80px"
      >
        <el-form-item label="原密码" prop="oldPassword">
          <el-input 
            v-model="passwordForm.oldPassword" 
            type="password"
            placeholder="请输入原密码"
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input 
            v-model="passwordForm.newPassword" 
            type="password"
            placeholder="请输入新密码"
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="passwordForm.confirmPassword" 
            type="password"
            placeholder="请再次输入新密码"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showChangePasswordDialog = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleChangePassword"
          :loading="passwordLoading"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Avatar, 
  ArrowDown,
  DataBoard,
  Grid,
  Trophy,
  Timer,
  UserFilled,
  Key
} from '@element-plus/icons-vue'
import { getAdminInfo, hasAnyPermission, logout, verifyToken } from './utils/auth.js'
import { adminAPI } from './services/api.js'

const router = useRouter()
const route = useRoute()

const adminInfo = ref(null)
const showChangePasswordDialog = ref(false)
const passwordLoading = ref(false)
const passwordFormRef = ref()

// 修改密码表单
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 密码表单验证规则
const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 当前页面标题
const currentPageTitle = computed(() => {
  return route.meta?.title || '小游戏管理后台'
})

// 可见的路由菜单
const visibleRoutes = computed(() => {
  return router.getRoutes().filter(route => {
    // 过滤掉不需要在菜单中显示的路由
    if (route.meta?.hideInMenu) {
      return false
    }
    
    // 检查权限
    if (route.meta?.permissions) {
      return hasAnyPermission(route.meta.permissions)
    }
    
    return route.path !== '/login' && route.path !== '/:pathMatch(.*)*'
  })
})

// 处理用户下拉菜单命令
const handleCommand = (command) => {
  switch (command) {
    case 'profile':
      ElMessage.info('个人资料功能开发中...')
      break
    case 'changePassword':
      showChangePasswordDialog.value = true
      break
    case 'logout':
      handleLogout()
      break
  }
}

// 修改密码
const handleChangePassword = async () => {
  try {
    await passwordFormRef.value.validate()
    
    passwordLoading.value = true
    
    const response = await adminAPI.changePassword(passwordForm)
    
    if (response.code === 0) {
      ElMessage.success('密码修改成功，请重新登录')
      showChangePasswordDialog.value = false
      logout()
    } else {
      ElMessage.error(response.msg || '密码修改失败')
    }
  } catch (error) {
    console.error('修改密码错误:', error)
    ElMessage.error('密码修改失败')
  } finally {
    passwordLoading.value = false
  }
}

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要退出登录吗？',
      '退出确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    logout()
  } catch (error) {
    // 用户取消退出
  }
}

// 初始化用户信息
const initUserInfo = async () => {
  try {
    // 验证登录状态
    const isValid = await verifyToken()
    if (isValid) {
      adminInfo.value = getAdminInfo()
    } else if (route.path !== '/login') {
      router.push('/login')
    }
  } catch (error) {
    console.error('验证登录状态错误:', error)
    if (route.path !== '/login') {
      router.push('/login')
    }
  }
}

onMounted(() => {
  initUserInfo()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  height: 100vh;
}

.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #001529;
  overflow: hidden;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #002140;
  margin-bottom: 1px;
}

.logo h3 {
  color: #fff;
  font-size: 16px;
  margin: 0;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.header-left h4 {
  margin: 0;
  color: #333;
  font-size: 18px;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: #f5f5f5;
}

.user-info .el-icon {
  margin-right: 8px;
}

.main-content {
  background-color: #f0f2f5;
  padding: 0;
  overflow-y: auto;
}

/* Element Plus 自定义样式 */
.el-menu {
  border-right: none;
}

.el-menu-item {
  height: 48px;
  line-height: 48px;
}

.el-menu-item .el-icon {
  margin-right: 8px;
}

.el-aside {
  transition: width 0.3s;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .layout-container {
    flex-direction: column;
  }
  
  .sidebar {
    width: 100% !important;
    height: auto;
  }
  
  .logo h3 {
    font-size: 14px;
  }
}
</style>