import { createRouter, createWebHistory } from 'vue-router'
import { verifyToken, isLoggedIn } from '../utils/auth.js'
import { ElMessage } from 'element-plus'

// 直接导入组件 - 临时修复ChunkLoadError
import Dashboard from '../views/Dashboard.vue'
import AppManagement from '../views/AppManagement.vue'
import LeaderboardManagement from '../views/LeaderboardManagement.vue'
import CounterManagement from '../views/CounterManagement.vue'
import UserManagement from '../views/UserManagement.vue'
import AdminManagement from '../views/AdminManagement.vue'
import RoleManagement from '../views/RoleManagement.vue'
import MailManagement from '../views/MailManagement.vue'
import MailTest from '../views/MailTest.vue'
import GameConfigManagement from '../views/GameConfigManagement.vue'
import Login from '../views/Login.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { 
      title: '登录',
      requiresAuth: false,
      hideInMenu: true
    }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard,
    meta: { 
      title: '仪表板',
      requiresAuth: true,
      icon: 'DataBoard',
      permissions: ['stats_view']
    }
  },
  // 游戏管理相关页面
  {
    path: '/apps',
    name: 'AppManagement',
    component: AppManagement,
    meta: { 
      title: '应用管理',
      requiresAuth: true,
      icon: 'Grid',
      permissions: ['app_manage'],
      group: 'game',
      groupTitle: '游戏管理',
      groupIcon: 'Grid'
    }
  },
  {
    path: '/leaderboards',
    name: 'LeaderboardManagement',
    component: LeaderboardManagement,
    meta: { 
      title: '排行榜管理',
      requiresAuth: true,
      icon: 'Trophy',
      permissions: ['leaderboard_manage'],
      group: 'game',
      groupTitle: '游戏管理',
      groupIcon: 'Grid'
    }
  },
  {
    path: '/counters',
    name: 'CounterManagement',
    component: CounterManagement,
    meta: { 
      title: '计数器管理',
      requiresAuth: true,
      icon: 'Timer',
      permissions: ['counter_manage'],
      group: 'game',
      groupTitle: '游戏管理',
      groupIcon: 'Grid'
    }
  },
  {
    path: '/users',
    name: 'UserManagement',
    component: UserManagement,
    meta: { 
      title: '用户管理',
      requiresAuth: true,
      icon: 'UserFilled',
      permissions: ['user_manage'],
      group: 'game',
      groupTitle: '游戏管理',
      groupIcon: 'Grid'
    }
  },
  {
    path: '/game-config',
    name: 'GameConfigManagement',
    component: GameConfigManagement,
    meta: { 
      title: '游戏配置',
      requiresAuth: true,
      icon: 'Setting',
      permissions: ['app_manage'],
      group: 'game',
      groupTitle: '游戏管理',
      groupIcon: 'Grid'
    }
  },
  {
    path: '/mails',
    name: 'MailManagement',
    component: MailManagement,
    meta: { 
      title: '邮件管理',
      requiresAuth: true,
      icon: 'Message',
      permissions: ['mail_manage'],
      group: 'game',
      groupTitle: '游戏管理',
      groupIcon: 'Grid'
    }
  },
  {
    path: '/mail-test',
    name: 'MailTest',
    component: MailTest,
    meta: { 
      title: '邮件测试',
      requiresAuth: true,
      icon: 'Monitor',
      permissions: ['mail_manage'],
      group: 'game',
      groupTitle: '游戏管理',
      groupIcon: 'Grid'
    }
  },
  // 系统管理相关页面
  {
    path: '/admins',
    name: 'AdminManagement',
    component: AdminManagement,
    meta: { 
      title: '管理员管理',
      requiresAuth: true,
      icon: 'Avatar',
      permissions: ['admin_manage'],
      group: 'system',
      groupTitle: '系统管理',
      groupIcon: 'Setting'
    }
  },
  {
    path: '/roles',
    name: 'RoleManagement',
    component: RoleManagement,
    meta: { 
      title: '角色管理',
      requiresAuth: true,
      icon: 'Key',
      permissions: ['role_manage'],
      group: 'system',
      groupTitle: '系统管理',
      groupIcon: 'Setting'
    }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - 小游戏管理后台`
  }

  // 如果是登录页面，检查是否已登录
  if (to.path === '/login') {
    if (isLoggedIn()) {
      // 验证token是否有效
      const isValid = await verifyToken()
      if (isValid) {
        next('/')
        return
      }
    }
    next()
    return
  }

  // 检查是否需要登录
  if (to.meta.requiresAuth) {
    if (!isLoggedIn()) {
      ElMessage.warning('请先登录')
      next('/login')
      return
    }

    // 验证token是否有效
    const isValid = await verifyToken()
    if (!isValid) {
      ElMessage.warning('登录已过期，请重新登录')
      next('/login')
      return
    }

    // 检查权限（如果路由定义了权限要求）
    if (to.meta.permissions) {
      const { hasAnyPermission, getAdminInfo } = await import('../utils/auth.js')
      
      // 先检查管理员信息是否存在
      const adminInfo = getAdminInfo()
      if (!adminInfo) {
        console.warn('管理员信息不存在，重新登录')
        ElMessage.warning('登录信息异常，请重新登录')
        next('/login')
        return
      }
      
      // 临时调试信息
      console.log('当前用户信息:', adminInfo)
      console.log('需要的权限:', to.meta.permissions)
      console.log('用户权限:', adminInfo.permissions)
      console.log('用户角色:', adminInfo.role)
      
      // 超级管理员跳过权限检查
      if (adminInfo.role === 'super_admin') {
        console.log('超级管理员，跳过权限检查')
        next()
        return
      }
      
      if (!hasAnyPermission(to.meta.permissions)) {
        // 防止无限重定向：记录重定向次数
        const redirectCount = (to.query._redirectCount ? parseInt(to.query._redirectCount) : 0) + 1
        
        if (redirectCount > 2) {
          // 重定向次数过多，可能是权限配置问题
          ElMessage.error('权限配置异常，请联系管理员')
          console.error('权限检查失败，用户信息:', adminInfo, '需要权限:', to.meta.permissions)
          next('/login')
          return
        }
        
        ElMessage.error(`权限不足，无法访问该页面。需要权限: ${to.meta.permissions.join(', ')}`)
        
        if (to.path !== '/') {
          // 重定向到首页，并携带重定向计数
          next({ path: '/', query: { _redirectCount: redirectCount.toString() } })
        } else {
          // 如果已经在首页还没权限，说明权限配置有问题
          ElMessage.error('没有可访问的页面，请联系管理员')
          next('/login')
        }
        return
      }
    }
  }

  next()
})

export default router