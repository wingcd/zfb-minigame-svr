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
        <el-text type="info" size="small">
          默认账户：admin / 123456
        </el-text>
        <br>
        <el-button
          type="text"
          size="small"
          @click="handleInitSystem"
          :loading="initLoading"
        >
          {{ initLoading ? '初始化中...' : '初始化系统' }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authAPI, adminAPI } from '../services/api.js'
import { setToken, setAdminInfo } from '../utils/auth.js'

const router = useRouter()
const loginFormRef = ref()
const loading = ref(false)
const initLoading = ref(false)

// 登录表单数据
const loginForm = reactive({
  username: '',
  password: ''
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
    
    const response = await adminAPI.initAdmin({ force: false })
    
    if (response.code === 0) {
      ElMessage.success('系统初始化成功！默认管理员账户已创建')
      loginForm.username = 'admin'
      loginForm.password = '123456'
    } else if (response.code === 4003) {
      ElMessage.info(response.msg)
    } else {
      ElMessage.error(response.msg || '初始化失败')
    }
  } catch (error) {
    console.error('初始化错误:', error)
    ElMessage.error('初始化失败')
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
  border-radius: 8px;
}

.login-footer {
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #eee;
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