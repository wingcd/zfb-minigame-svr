<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>小游戏管理后台</h1>
        <p>请登录您的管理员账户</p>
      </div>
      
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        size="large"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            prefix-icon="User"
            clearable
          />
        </el-form-item>
        
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        
        <el-form-item>
          <el-checkbox v-model="loginForm.rememberMe">
            记住我（30天免登录）
          </el-checkbox>
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            class="login-button"
            :loading="loading"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
      </el-form>
      
      <div class="login-footer">
        <div class="init-section" v-if="showInitButton">
          <el-divider>系统初始化</el-divider>
          <el-alert 
            title="系统未初始化" 
            type="warning" 
            description="检测到系统尚未初始化，请点击下方按钮进行系统初始化"
            show-icon
            :closable="false"
            style="margin-bottom: 15px;"
          />
          <el-button
            type="warning"
            class="init-button"
            :loading="initLoading"
            @click="handleInitSystem"
            block
          >
            {{ initLoading ? '初始化中...' : '初始化系统' }}
          </el-button>
          <div class="init-tip">
            <el-text size="small" type="info">
              初始化将创建默认管理员账户：admin/123456
            </el-text>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authAPI, adminAPI, initSystemInstall, checkInstallStatus } from '../services/api.js'
import { setToken, setAdminInfo } from '../utils/auth.js'

const router = useRouter()
const loginFormRef = ref()
const loading = ref(false)
const initLoading = ref(false)
const showInitButton = ref(false)

// 检查系统初始化状态
const checkSystemStatus = async () => {
  try {
    const response = await checkInstallStatus()
    
    // 如果系统未初始化，显示初始化按钮
    if (response.code === 404 || response.code === 4001) {
      showInitButton.value = true
    } else if (response.code === 200 || response.code === 0) {
      showInitButton.value = false
    }
  } catch (error) {
    // 如果接口不存在或返回错误，假设系统未初始化
    if (error.response?.status === 404) {
      showInitButton.value = true
    }
  }
}

// 页面加载时检查系统状态
onMounted(() => {
  checkSystemStatus()
})

// 登录表单数据
const loginForm = reactive({
  username: '',
  password: '',
  rememberMe: false
})

// 表单验证规则
const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ]
}

// 处理登录
const handleLogin = async () => {
  try {
    await loginFormRef.value.validate()
    
    loading.value = true
    
    const response = await authAPI.login(loginForm)
    
    if (response.code === 0) {
      // 保存token和用户信息
      setToken(response.data.token)
      setAdminInfo(response.data.adminInfo)
      
      ElMessage.success('登录成功')
      
      // 跳转到首页
      router.push('/')
    } else {
      ElMessage.error(response.msg || '登录失败')
    }
  } catch (error) {
    console.error('登录错误:', error)
    ElMessage.error('登录失败，请检查用户名和密码')
  } finally {
    loading.value = false
  }
}

// 初始化系统
const handleInitSystem = async () => {
  try {
    initLoading.value = true
    
    // 调用新的Golang后端安装接口
    const response = await initSystemInstall({ 
      adminUsername: 'admin',
      adminPassword: '123456',
      force: false 
    })
    
    // 适配新的返回格式 - Golang后端使用标准HTTP状态码
    if (response.code === 200 || response.code === 0) {
      ElMessage.success('系统初始化成功！默认管理员账户已创建')
      
      // 隐藏初始化按钮
      showInitButton.value = false
      
      // 从返回数据中获取默认凭据
      const defaultCredentials = response.data?.defaultCredentials
      if (defaultCredentials) {
        loginForm.username = defaultCredentials.username || 'admin'
        loginForm.password = defaultCredentials.password || '123456'
        
        // 显示安全提醒
        if (defaultCredentials.warning) {
          ElMessage.warning(defaultCredentials.warning)
        }
      } else {
        // 兼容旧格式
        loginForm.username = 'admin'
        loginForm.password = '123456'
      }
      
      // 显示初始化结果信息
      const data = response.data
      if (data) {
        const details = []
        if (data.createdCollections) details.push(`创建了${data.createdCollections}个数据表`)
        if (data.createdRoles) details.push(`创建了${data.createdRoles}个角色`)
        if (data.createdAdmins) details.push(`创建了${data.createdAdmins}个管理员`)
        
        if (details.length > 0) {
          ElMessage.info(details.join('，'))
        }
      }
      
    } else if (response.code === 409 || response.code === 4003) {
      // 系统已经初始化 - 使用409状态码表示冲突
      showInitButton.value = false
      ElMessage.info(response.message || response.msg || '系统已经初始化')
    } else {
      ElMessage.error(response.message || response.msg || '初始化失败')
    }
  } catch (error) {
    console.error('初始化错误:', error)
    
    // 处理网络错误或其他异常
    if (error.response) {
      const { status, data } = error.response
      if (status === 409) {
        ElMessage.info('系统已经初始化')
      } else if (status === 500) {
        ElMessage.error('服务器内部错误，请检查数据库连接')
      } else {
        ElMessage.error(data?.message || '初始化失败')
      }
    } else {
      ElMessage.error('网络连接失败，请检查服务器状态')
    }
  } finally {
    initLoading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-box {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  padding: 40px;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  color: #333;
  margin: 0 0 10px 0;
  font-size: 28px;
  font-weight: 600;
}

.login-header p {
  color: #666;
  margin: 0;
  font-size: 14px;
}

.login-form {
  margin-bottom: 20px;
}

.login-button {
  width: 100%;
  height: 45px;
  font-size: 16px;
  font-weight: 600;
}

.init-section {
  margin-top: 20px;
}

.init-button {
  width: 100%;
  height: 40px;
  margin-bottom: 10px;
}

.init-tip {
  text-align: center;
  margin-top: 8px;
}

.login-footer {
  margin-top: 20px;
}

:deep(.el-input__inner) {
  height: 45px;
  line-height: 45px;
  border-radius: 8px;
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}
</style> 