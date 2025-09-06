<template>
  <div class="admin-management">
    <div class="page-header">
      <h2>管理员管理</h2>
      <el-button 
        type="primary" 
        @click="showCreateDialog = true"
        v-if="hasPermission(PERMISSIONS.ADMIN_MANAGE)"
      >
        <el-icon><Plus /></el-icon>
        添加管理员
      </el-button>
    </div>

    <!-- 搜索表单 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="用户名">
          <el-input 
            v-model="searchForm.username" 
            placeholder="请输入用户名"
            clearable
          />
        </el-form-item>
        <el-form-item label="角色">
          <el-select 
            v-model="searchForm.role" 
            placeholder="请选择角色"
            clearable
          >
            <el-option 
              v-for="role in roleOptions" 
              :key="role.roleCode"
              :label="role.roleName" 
              :value="role.roleCode"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select 
            v-model="searchForm.status" 
            placeholder="请选择状态"
            clearable
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 管理员列表 -->
    <el-card class="table-card">
      <el-table 
        :data="adminList" 
        v-loading="loading"
        stripe
      >
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)">
              {{ row.roleInfo?.roleName || row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" width="180" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastLoginTime" label="最后登录" width="160" />
        <el-table-column prop="createTime" label="创建时间" width="160" />
        <el-table-column 
          label="操作" 
          width="180"
          v-if="hasPermission(PERMISSIONS.ADMIN_MANAGE)"
        >
          <template #default="{ row }">
            <el-button 
              type="primary" 
              size="small" 
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button 
              type="warning" 
              size="small" 
              @click="handleResetPassword(row)"
            >
              重置密码
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="handleDelete(row)"
              :disabled="row.username === 'admin'"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 创建/编辑管理员对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingAdmin ? '编辑管理员' : '添加管理员'"
      width="500px"
      @close="resetForm"
    >
      <el-form
        ref="adminFormRef"
        :model="adminForm"
        :rules="adminRules"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input 
            v-model="adminForm.username" 
            :disabled="editingAdmin"
            placeholder="请输入用户名"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!editingAdmin">
          <el-input 
            v-model="adminForm.password" 
            type="password"
            placeholder="请输入密码"
          />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input 
            v-model="adminForm.nickname" 
            placeholder="请输入昵称"
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select 
            v-model="adminForm.role" 
            placeholder="请选择角色"
            style="width: 100%"
          >
            <el-option 
              v-for="role in roleOptions" 
              :key="role.roleCode"
              :label="role.roleName" 
              :value="role.roleCode"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input 
            v-model="adminForm.email" 
            placeholder="请输入邮箱"
          />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input 
            v-model="adminForm.phone" 
            placeholder="请输入手机号"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="editingAdmin">
          <el-radio-group v-model="adminForm.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleSubmit"
          :loading="submitLoading"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminAPI, roleAPI } from '../services/api.js'
import { hasPermission, PERMISSIONS } from '../utils/auth.js'

const loading = ref(false)
const submitLoading = ref(false)
const showCreateDialog = ref(false)
const editingAdmin = ref(null)
const adminFormRef = ref()
const adminList = ref([])
const roleOptions = ref([])

// 搜索表单
const searchForm = reactive({
  username: '',
  role: '',
  status: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 管理员表单
const adminForm = reactive({
  username: '',
  password: '',
  nickname: '',
  role: '',
  email: '',
  phone: '',
  status: 1
})

// 表单验证规则
const adminRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ]
}

// 获取管理员列表
const fetchAdminList = async () => {
  try {
    loading.value = true
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...searchForm
    }
    
    const response = await adminAPI.getList(params)
    
    if (response.code === 0) {
      adminList.value = response.data.list
      pagination.total = response.data.total
    } else {
      ElMessage.error(response.msg || '获取管理员列表失败')
    }
  } catch (error) {
    console.error('获取管理员列表错误:', error)
    ElMessage.error('获取管理员列表失败')
  } finally {
    loading.value = false
  }
}

// 获取角色列表
const fetchRoleList = async () => {
  try {
    const response = await roleAPI.getAll()
    
    if (response.code === 0) {
      roleOptions.value = response.data.list || []
    }
  } catch (error) {
    console.error('获取角色列表错误:', error)
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchAdminList()
}

// 重置搜索
const handleReset = () => {
  Object.assign(searchForm, {
    username: '',
    role: '',
    status: ''
  })
  pagination.page = 1
  fetchAdminList()
}

// 分页大小改变
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchAdminList()
}

// 当前页改变
const handleCurrentChange = (page) => {
  pagination.page = page
  fetchAdminList()
}

// 编辑管理员
const handleEdit = (admin) => {
  editingAdmin.value = admin
  Object.assign(adminForm, {
    username: admin.username,
    nickname: admin.nickname,
    role: admin.role,
    email: admin.email,
    phone: admin.phone,
    status: admin.status
  })
  showCreateDialog.value = true
}

// 重置表单
const resetForm = () => {
  editingAdmin.value = null
  Object.assign(adminForm, {
    username: '',
    password: '',
    nickname: '',
    role: '',
    email: '',
    phone: '',
    status: 1
  })
  adminFormRef.value?.resetFields()
}

// 提交表单
const handleSubmit = async () => {
  try {
    await adminFormRef.value.validate()
    
    submitLoading.value = true
    
    let response
    if (editingAdmin.value) {
      response = await adminAPI.update({
        id: editingAdmin.value._id,
        ...adminForm
      })
    } else {
      response = await adminAPI.create(adminForm)
    }
    
    if (response.code === 0) {
      ElMessage.success(editingAdmin.value ? '更新成功' : '创建成功')
      showCreateDialog.value = false
      fetchAdminList()
    } else {
      ElMessage.error(response.msg || '操作失败')
    }
  } catch (error) {
    console.error('提交表单错误:', error)
    ElMessage.error('操作失败')
  } finally {
    submitLoading.value = false
  }
}

// 重置密码
const handleResetPassword = async (admin) => {
  try {
    const { value: newPassword } = await ElMessageBox.prompt(
      '请输入新密码（至少6位）',
      '重置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputPattern: /.{6,}/,
        inputErrorMessage: '密码至少6位'
      }
    )
    
    const response = await adminAPI.resetPassword({
      id: admin._id,
      newPassword
    })
    
    if (response.code === 0) {
      ElMessage.success('密码重置成功')
    } else {
      ElMessage.error(response.msg || '密码重置失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重置密码错误:', error)
      ElMessage.error('密码重置失败')
    }
  }
}

// 删除管理员
const handleDelete = async (admin) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除管理员 "${admin.username}" 吗？此操作不可恢复！`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await adminAPI.delete(admin._id)
    
    if (response.code === 0) {
      ElMessage.success('删除成功')
      fetchAdminList()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除管理员错误:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 获取角色标签类型
const getRoleTagType = (role) => {
  const typeMap = {
    'super_admin': 'danger',
    'admin': 'warning',
    'operator': 'primary',
    'viewer': 'info'
  }
  return typeMap[role] || 'default'
}

onMounted(() => {
  fetchAdminList()
  fetchRoleList()
})
</script>

<style scoped>
.admin-management {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #333;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style> 