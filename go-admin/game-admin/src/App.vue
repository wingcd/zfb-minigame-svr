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
          <template v-for="item in menuItems" :key="item.key">
            <!-- 分组菜单 -->
            <el-sub-menu v-if="item.isGroup" :index="item.key">
              <template #title>
                <el-icon v-if="item.icon">
                  <component :is="item.icon" />
                </el-icon>
                <span>{{ item.title }}</span>
              </template>
              <el-menu-item
                v-for="child in item.children"
                :key="child.path"
                :index="child.path"
              >
                <el-icon v-if="child.meta?.icon">
                  <component :is="child.meta.icon" />
                </el-icon>
                <span>{{ child.meta?.title }}</span>
              </el-menu-item>
            </el-sub-menu>
            
            <!-- 普通菜单项 -->
            <el-menu-item
              v-else
              :index="item.path"
            >
              <el-icon v-if="item.meta?.icon">
                <component :is="item.meta.icon" />
              </el-icon>
              <span>{{ item.meta?.title }}</span>
            </el-menu-item>
          </template>
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
            <!-- 全局应用选择器 -->
            <el-select 
              v-model="globalSelectedAppId" 
              placeholder="选择应用" 
              @change="handleAppChange" 
              style="width: 200px; margin-right: 15px;"
              :loading="appLoading"
            >
              <template v-if="globalAppList && globalAppList.length > 0">
                <el-option
                  v-for="app in globalAppList"
                  :key="app.appId"
                  :label="app.appName || '未命名应用'"
                  :value="app.appId">
                </el-option>
              </template>
            </el-select>
            
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
import { ref, computed, reactive, onMounted, onUnmounted } from 'vue'
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
import { getAdminInfo, hasAnyPermission, logout, verifyToken, startTokenValidation } from './utils/auth.js'
import { adminAPI } from './services/api.js'
import { appList, selectedAppId, loading, getAppList, setSelectedAppId } from './utils/appStore.js'

const router = useRouter()
const route = useRoute()

const adminInfo = ref(null)
const showChangePasswordDialog = ref(false)
const passwordLoading = ref(false)
const passwordFormRef = ref()

// 全局app状态
const globalAppList = appList
const globalSelectedAppId = selectedAppId
const appLoading = loading

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
    
    // 过滤掉登录页和404页
    if (route.path === '/login' || route.path === '/:pathMatch(.*)*') {
      return false
    }
    
    // 如果是分组路由，检查是否有子路由有权限访问
    if (route.meta?.isGroup && route.children) {
      const hasVisibleChildren = route.children.some(child => {
        // 检查子路由权限
        if (child.meta?.permissions) {
          return hasAnyPermission(child.meta.permissions)
        }
        return !child.meta?.hideInMenu
      })
      return hasVisibleChildren
    }
    
    // 检查普通路由权限
    if (route.meta?.permissions) {
      return hasAnyPermission(route.meta.permissions)
    }
    
    return true
  })
})

// 生成菜单项结构
const menuItems = computed(() => {
  const items = []
  const groups = new Map()
  // 处理所有可见路由（已经过权限过滤）
  visibleRoutes.value.forEach(route => {
    if (route.meta?.group) {
      // 有分组的路由
      const groupKey = route.meta.group
      if (!groups.has(groupKey)) {
        groups.set(groupKey, {
          key: groupKey,
          title: route.meta.groupTitle,
          icon: route.meta.groupIcon,
          isGroup: true,
          children: []
        })
      }
      groups.get(groupKey).children.push(route)
    } else {
      // 没有分组的路由，直接添加
      items.push({
        key: route.path,
        path: route.path,
        meta: route.meta,
        isGroup: false
      })
    }
  })
  
  // 将分组添加到菜单项中
  groups.forEach(group => {
    if (group.children.length > 0) {
      items.push(group)
    }
  })
  
  return items
})

// 处理应用选择变化
const handleAppChange = (appId) => {
  setSelectedAppId(appId)
  ElMessage.success('已切换应用')
}

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
    
    const response = await adminAPI.resetPassword({
      newPassword: passwordForm.newPassword,
      id: adminInfo.value.id
    })
    
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

// 存储定时验证的清理函数
let tokenValidationCleanup = null

// 初始化用户信息
const initUserInfo = async () => {
  try {
    // 验证登录状态
    const isValid = await verifyToken()
    if (isValid) {
      adminInfo.value = getAdminInfo()
      
      // 启动定期token验证（每30分钟检查一次）
      if (tokenValidationCleanup) {
        tokenValidationCleanup()
      }
      tokenValidationCleanup = startTokenValidation(30)
      
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
  getAppList()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (tokenValidationCleanup) {
    tokenValidationCleanup()
  }
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