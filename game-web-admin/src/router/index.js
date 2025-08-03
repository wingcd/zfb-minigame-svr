import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import AppManagement from '../views/AppManagement.vue'
import LeaderboardManagement from '../views/LeaderboardManagement.vue'
import UserManagement from '../views/UserManagement.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/apps',
    name: 'AppManagement',
    component: AppManagement
  },
  {
    path: '/leaderboards',
    name: 'LeaderboardManagement',
    component: LeaderboardManagement
  },
  {
    path: '/users',
    name: 'UserManagement',
    component: UserManagement
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router