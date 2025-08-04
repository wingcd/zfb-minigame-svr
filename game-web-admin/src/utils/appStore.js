import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { appAPI } from '../services/api.js'

// 全局状态
const appList = ref([])
const selectedAppId = ref('')
const loading = ref(false)

// 获取应用列表
const getAppList = async () => {
  try {
    loading.value = true
    const result = await appAPI.getAll()
    if (result.code === 0) {
      const dataList = result.data?.list
      const validApps = Array.isArray(dataList) ? dataList : []
      
      appList.value = validApps
      
      // 如果没有选择的app且有可用app，选择第一个
      if (!selectedAppId.value && appList.value.length > 0) {
        selectedAppId.value = appList.value[0].appId
      }
    } else {
      appList.value = []
      ElMessage.error(result.msg || '获取应用列表失败')
    }
  } catch (error) {
    console.error('获取应用列表失败:', error)
    appList.value = []
    ElMessage.error('获取应用列表失败')
  } finally {
    loading.value = false
  }
}

// 设置选择的app
const setSelectedAppId = (appId) => {
  selectedAppId.value = appId
}

// 获取当前选择的app信息
const getSelectedApp = () => {
  return appList.value.find(app => app.appId === selectedAppId.value)
}

// 获取app名称
const getAppName = (appId) => {
  const app = appList.value.find(app => app.appId === appId)
  return app ? app.appName : appId
}

// 导出状态和方法
export {
  appList,
  selectedAppId,
  loading,
  getAppList,
  setSelectedAppId,
  getSelectedApp,
  getAppName
} 