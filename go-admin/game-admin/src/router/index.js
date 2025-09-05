import { createRouter, createWebHistory } from 'vue-router'
import { verifyToken, isLoggedIn } from '../utils/auth.js'
import { ElMessage } from 'element-plus'

// 懒加载组件
const Dashboard = () => import('../views/Dashboard.vue')
const AppManagement = () => import('../views/AppManagement.vue')
const LeaderboardManagement = () => import('../views/LeaderboardManagement.vue')
const CounterManagement = () => import('../views/CounterManagement.vue')
const UserManagement = () => import('../views/UserManagement.vue')
const AdminManagement = () => import('../views/AdminManagement.vue')
const RoleManagement = () => import('../views/RoleManagement.vue')
const MailManagement = () => import('../views/MailManagement.vue')
const MailTest = () => import('../views/MailTest.vue')
const GameConfigManagement = () => import('../views/GameConfigManagement.vue')

const Login = () => import('../views/Login.vue')

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
      permissions: ['leaderboard_manage'],
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
      const { hasAnyPermission } = await import('../utils/auth.js')
      if (!hasAnyPermission(to.meta.permissions)) {
        ElMessage.error('权限不足，无法访问该页面')
        next('/')
        return
      }
    }
  }

  next()
})

export default router