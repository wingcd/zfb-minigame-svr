import { authAPI } from '../services/api.js'

// 获取token
export function getToken() {
  return localStorage.getItem('admin_token')
}

// 设置token
export function setToken(token) {
  localStorage.setItem('admin_token', token)
}

// 移除token
export function removeToken() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_info')
}

// 获取管理员信息
export function getAdminInfo() {
  const adminInfo = localStorage.getItem('admin_info')
  return adminInfo ? JSON.parse(adminInfo) : null
}

// 设置管理员信息
export function setAdminInfo(adminInfo) {
  localStorage.setItem('admin_info', JSON.stringify(adminInfo))
}

// 检查是否已登录
export function isLoggedIn() {
  return !!getToken()
}

// 验证token是否有效
export async function verifyToken() {
  const token = getToken()
  if (!token) {
    return false
  }

  try {
    const response = await authAPI.verifyToken({ token })
    if (response.code === 0 && response.data.valid) {
      setAdminInfo(response.data.adminInfo)
      return true
    } else {
      removeToken()
      return false
    }
  } catch (error) {
    removeToken()
    return false
  }
}

// 检查权限
export function hasPermission(permission) {
  const adminInfo = getAdminInfo()
  if (!adminInfo) {
    return false
  }

  // 超级管理员拥有所有权限
  if (adminInfo.role === 'super_admin') {
    return true
  }

  // 检查是否有特定权限
  return adminInfo.permissions && adminInfo.permissions.includes(permission)
}

// 检查是否有任意一个权限
export function hasAnyPermission(permissions) {
  if (!Array.isArray(permissions)) {
    return hasPermission(permissions)
  }

  return permissions.some(permission => hasPermission(permission))
}

// 检查是否有所有权限
export function hasAllPermissions(permissions) {
  if (!Array.isArray(permissions)) {
    return hasPermission(permissions)
  }

  return permissions.every(permission => hasPermission(permission))
}

// 登出
export function logout() {
  removeToken()
  // 跳转到登录页
  window.location.href = '/login'
}

// 权限常量
export const PERMISSIONS = {
  ADMIN_MANAGE: 'admin_manage',        // 管理员管理
  ROLE_MANAGE: 'role_manage',          // 角色管理
  APP_MANAGE: 'app_manage',            // 应用管理
  USER_MANAGE: 'user_manage',          // 用户管理
  LEADERBOARD_MANAGE: 'leaderboard_manage', // 排行榜管理
  MAIL_MANAGE: 'mail_manage',          // 邮件管理
  STATS_VIEW: 'stats_view',            // 统计查看
  SYSTEM_CONFIG: 'system_config'       // 系统配置
} 