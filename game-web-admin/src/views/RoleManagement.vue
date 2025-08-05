<template>
  <div class="role-management">
    <div class="page-header">
      <h2>角色管理</h2>
      <el-button 
        type="primary" 
        @click="showCreateDialog = true"
        v-if="hasPermission(PERMISSIONS.ROLE_MANAGE)"
      >
        <el-icon><Plus /></el-icon>
        添加角色
      </el-button>
    </div>

    <!-- 搜索表单 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="角色名称">
          <el-input 
            v-model="searchForm.roleName" 
            placeholder="请输入角色名称"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 角色列表 -->
    <el-card class="table-card">
      <el-table 
        :data="roleList" 
        v-loading="loading"
        stripe
      >
        <el-table-column prop="roleCode" label="角色代码" width="120" />
        <el-table-column prop="roleName" label="角色名称" width="150" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="权限" width="300">
          <template #default="{ row }">
            <el-tag 
              v-for="permission in row.permissions" 
              :key="permission"
              size="small"
              style="margin-right: 5px; margin-bottom: 5px;"
            >
              {{ getPermissionName(permission) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="adminCount" label="管理员数量" width="100" />
        <el-table-column prop="createTime" label="创建时间" width="160" />
        <el-table-column 
          label="操作" 
          width="150"
          v-if="hasPermission(PERMISSIONS.ROLE_MANAGE)"
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
              type="danger" 
              size="small" 
              @click="handleDelete(row)"
              :disabled="row.roleCode === 'super_admin'"
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

    <!-- 创建/编辑角色对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingRole ? '编辑角色' : '添加角色'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="roleRules"
        label-width="80px"
      >
        <el-form-item label="角色代码" prop="roleCode">
          <el-input 
            v-model="roleForm.roleCode" 
            :disabled="editingRole"
            placeholder="请输入角色代码（英文）"
          />
        </el-form-item>
        <el-form-item label="角色名称" prop="roleName">
          <el-input 
            v-model="roleForm.roleName" 
            placeholder="请输入角色名称"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="roleForm.description" 
            type="textarea"
            placeholder="请输入角色描述"
            rows="3"
          />
        </el-form-item>
        <el-form-item label="权限" prop="permissions">
          <el-checkbox-group v-model="roleForm.permissions">
            <el-checkbox 
              v-for="permission in allPermissions" 
              :key="permission.code"
              :label="permission.code"
            >
              {{ permission.name }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number 
            v-model="roleForm.sort" 
            :min="1" 
            :max="100"
          />
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
import { roleAPI } from '../services/api.js'
import { hasPermission, PERMISSIONS } from '../utils/auth.js'

const loading = ref(false)
const submitLoading = ref(false)
const showCreateDialog = ref(false)
const editingRole = ref(null)
const roleFormRef = ref()
const roleList = ref([])

// 所有权限列表
const allPermissions = [
  { code: 'admin_manage', name: '管理员管理' },
  { code: 'role_manage', name: '角色管理' },
  { code: 'app_manage', name: '应用管理' },
  { code: 'user_manage', name: '用户管理' },
  { code: 'counter_manage', name: '计数器管理'},
  { code: 'leaderboard_manage', name: '排行榜管理' },
  { code: 'mail_manage', name: '邮件管理' },
  { code: 'stats_view', name: '统计查看' },
  { code: 'system_config', name: '系统配置' },
]

// 搜索表单
const searchForm = reactive({
  roleName: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 角色表单
const roleForm = reactive({
  roleCode: '',
  roleName: '',
  description: '',
  permissions: [],
  sort: 1
})

// 表单验证规则
const roleRules = {
  roleCode: [
    { required: true, message: '请输入角色代码', trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: '角色代码只能包含字母、数字和下划线，且以字母或下划线开头', trigger: 'blur' }
  ],
  roleName: [
    { required: true, message: '请输入角色名称', trigger: 'blur' }
  ],
  permissions: [
    { required: true, message: '请选择至少一个权限', trigger: 'change' }
  ]
}

// 获取角色列表
const fetchRoleList = async () => {
  try {
    loading.value = true
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...searchForm
    }
    
    const response = await roleAPI.getList(params)
    
    if (response.code === 0) {
      roleList.value = response.data.list
      pagination.total = response.data.total
    } else {
      ElMessage.error(response.msg || '获取角色列表失败')
    }
  } catch (error) {
    console.error('获取角色列表错误:', error)
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchRoleList()
}

// 重置搜索
const handleReset = () => {
  searchForm.roleName = ''
  pagination.page = 1
  fetchRoleList()
}

// 分页大小改变
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchRoleList()
}

// 当前页改变
const handleCurrentChange = (page) => {
  pagination.page = page
  fetchRoleList()
}

// 编辑角色
const handleEdit = (role) => {
  editingRole.value = role
  Object.assign(roleForm, {
    roleCode: role.roleCode,
    roleName: role.roleName,
    description: role.description,
    permissions: [...(role.permissions || [])],
    sort: role.sort || 1
  })
  showCreateDialog.value = true
}

// 重置表单
const resetForm = () => {
  editingRole.value = null
  Object.assign(roleForm, {
    roleCode: '',
    roleName: '',
    description: '',
    permissions: [],
    sort: 1
  })
  roleFormRef.value?.resetFields()
}

// 提交表单
const handleSubmit = async () => {
  try {
    await roleFormRef.value.validate()
    
    submitLoading.value = true
    
    let response
    if (editingRole.value) {
      response = await roleAPI.update(roleForm)
    } else {
      response = await roleAPI.create(roleForm)
    }
    
    if (response.code === 0) {
      ElMessage.success(editingRole.value ? '更新成功' : '创建成功')
      showCreateDialog.value = false
      fetchRoleList()
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

// 删除角色
const handleDelete = async (role) => {
  try {
    if (role.adminCount > 0) {
      ElMessage.warning(`该角色下还有 ${role.adminCount} 个管理员，无法删除`)
      return
    }
    
    await ElMessageBox.confirm(
      `确定要删除角色 "${role.roleName}" 吗？此操作不可恢复！`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await roleAPI.delete(role.roleCode)
    
    if (response.code === 0) {
      ElMessage.success('删除成功')
      fetchRoleList()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除角色错误:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 获取权限名称
const getPermissionName = (permissionCode) => {
  const permission = allPermissions.find(p => p.code === permissionCode)
  return permission ? permission.name : permissionCode
}

onMounted(() => {
  fetchRoleList()
})
</script>

<style scoped>
.role-management {
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

:deep(.el-checkbox-group) {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

:deep(.el-checkbox) {
  margin-right: 0;
}
</style> 