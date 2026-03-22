<template>
  <div class="feature-container">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="特征名称">
          <el-input v-model="searchForm.name" placeholder="请输入特征名称" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="特征编码">
          <el-input v-model="searchForm.code" placeholder="请输入特征编码" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="特征分类">
          <el-select v-model="searchForm.category" placeholder="请选择分类" clearable style="width: 150px">
            <el-option label="金额" value="amount" />
            <el-option label="发票" value="invoice" />
            <el-option label="报销单" value="reimbursement" />
            <el-option label="用户" value="user" />
            <el-option label="时间" value="time" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.enabled" placeholder="请选择状态" clearable style="width: 150px">
            <el-option label="启用" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>特征列表</span>
          <el-button type="primary" @click="handleCreate">新增特征</el-button>
        </div>
      </template>

      <el-table :data="featureList" v-loading="loading" border stripe>
        <el-table-column prop="id" label="特征ID" width="200" show-overflow-tooltip />
        <el-table-column prop="name" label="特征名称" width="150" />
        <el-table-column prop="code" label="特征编码" width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="特征类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.type)">
              {{ getTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="值类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.value_type === 'list' ? 'warning' : 'info'">
              {{ row.value_type === 'list' ? '列表' : '单值' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="100">
          <template #default="{ row }">
            {{ getCategoryText(row.category) }}
          </template>
        </el-table-column>
        <el-table-column label="值数量" width="80" align="center">
          <template #default="{ row }">
            {{ row.values.length }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handleToggleEnable(row)">
              {{ row.enabled ? '禁用' : '启用' }}
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="70%"
      :close-on-click-modal="false"
    >
      <el-form :model="featureForm" :rules="rules" ref="featureFormRef" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="特征名称" prop="name">
              <el-input v-model="featureForm.name" placeholder="请输入特征名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="特征编码" prop="code">
              <el-input v-model="featureForm.code" placeholder="请输入特征编码" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="特征类型" prop="type">
              <el-select v-model="featureForm.type" placeholder="请选择特征类型" style="width: 100%">
                <el-option label="字符串" value="string" />
                <el-option label="数字" value="number" />
                <el-option label="布尔" value="boolean" />
                <el-option label="日期" value="date" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="值类型" prop="value_type">
              <el-select v-model="featureForm.value_type" placeholder="请选择值类型" style="width: 100%">
                <el-option label="单值" value="single" />
                <el-option label="列表" value="list" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="分类" prop="category">
              <el-select v-model="featureForm.category" placeholder="请选择分类" style="width: 100%">
                <el-option label="金额" value="amount" />
                <el-option label="发票" value="invoice" />
                <el-option label="报销单" value="reimbursement" />
                <el-option label="用户" value="user" />
                <el-option label="时间" value="time" />
                <el-option label="其他" value="other" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="启用状态">
              <el-switch v-model="featureForm.enabled" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="描述">
          <el-input v-model="featureForm.description" type="textarea" :rows="2" placeholder="请输入特征描述" />
        </el-form-item>

        <el-divider content-position="left">特征值配置</el-divider>

        <div v-for="(value, index) in featureForm.values" :key="index" class="value-item">
          <el-row :gutter="10">
            <el-col :span="1">
              <el-button type="danger" circle size="small" @click="removeValue(index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-col>
            <el-col :span="6">
              <el-input v-model="value.value" placeholder="值" />
            </el-col>
            <el-col :span="6">
              <el-input v-model="value.label" placeholder="显示名称" />
            </el-col>
            <el-col :span="4">
              <el-input-number v-model="value.sort_order" :min="0" placeholder="排序" style="width: 100%" />
            </el-col>
            <el-col :span="5">
              <el-switch v-model="value.enabled" active-text="启用" />
            </el-col>
          </el-row>
        </div>

        <el-button type="primary" @click="addValue" style="margin-top: 10px">添加特征值</el-button>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'

const searchForm = reactive({
  name: '',
  code: '',
  category: '',
  enabled: null
})

const featureList = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('新增特征')
const featureFormRef = ref(null)

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const featureForm = reactive({
  id: '',
  name: '',
  code: '',
  description: '',
  type: 'string',
  value_type: 'single',
  category: 'other',
  enabled: true,
  values: []
})

const rules = {
  name: [{ required: true, message: '请输入特征名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入特征编码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择特征类型', trigger: 'change' }],
  value_type: [{ required: true, message: '请选择值类型', trigger: 'change' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }]
}

const fetchFeatures = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.pageSize,
      name: searchForm.name,
      code: searchForm.code,
      category: searchForm.category,
      enabled: searchForm.enabled
    }
    const res = await fetch(`/api/v1/ruleengine/features?page=${params.page}&size=${params.size}&name=${params.name}&code=${params.code}&category=${params.category}&enabled=${params.enabled}`)
    const data = await res.json()
    if (data.code === 200) {
      featureList.value = data.data.features || []
      pagination.total = data.data.total || 0
    } else {
      ElMessage.error(data.message || '获取特征列表失败')
    }
  } catch (error) {
    ElMessage.error('获取特征列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchFeatures()
}

const handleReset = () => {
  searchForm.name = ''
  searchForm.code = ''
  searchForm.category = ''
  searchForm.enabled = null
  pagination.page = 1
  fetchFeatures()
}

const handleCreate = () => {
  dialogTitle.value = '新增特征'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑特征'
  Object.assign(featureForm, {
    id: row.id,
    name: row.name,
    code: row.code,
    description: row.description,
    type: row.type,
    value_type: row.value_type,
    category: row.category,
    enabled: row.enabled,
    values: JSON.parse(JSON.stringify(row.values))
  })
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该特征吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const res = await fetch(`/api/v1/ruleengine/features/${row.id}`, {
      method: 'DELETE'
    })
    const data = await res.json()
    if (data.code === 200) {
      ElMessage.success('删除成功')
      fetchFeatures()
    } else {
      ElMessage.error(data.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleToggleEnable = async (row) => {
  try {
    const action = row.enabled ? 'disable' : 'enable'
    const res = await fetch(`/api/v1/ruleengine/features/${row.id}/${action}`, {
      method: 'PUT'
    })
    const data = await res.json()
    if (data.code === 200) {
      ElMessage.success(`${row.enabled ? '禁用' : '启用'}成功`)
      fetchFeatures()
    } else {
      ElMessage.error(data.message || '操作失败')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const handleSubmit = async () => {
  if (!featureFormRef.value) return

  await featureFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const url = featureForm.id ? `/api/v1/ruleengine/features/${featureForm.id}` : '/api/v1/ruleengine/features'
        const method = featureForm.id ? 'PUT' : 'POST'

        const res = await fetch(url, {
          method,
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(featureForm)
        })
        const data = await res.json()

        if (data.code === 200) {
          ElMessage.success(`${featureForm.id ? '更新' : '创建'}成功`)
          dialogVisible.value = false
          fetchFeatures()
        } else {
          ElMessage.error(data.message || '操作失败')
        }
      } catch (error) {
        ElMessage.error('操作失败')
      }
    }
  })
}

const addValue = () => {
  featureForm.values.push({
    value: '',
    label: '',
    sort_order: featureForm.values.length,
    enabled: true
  })
}

const removeValue = (index) => {
  featureForm.values.splice(index, 1)
}

const resetForm = () => {
  Object.assign(featureForm, {
    id: '',
    name: '',
    code: '',
    description: '',
    type: 'string',
    value_type: 'single',
    category: 'other',
    enabled: true,
    values: []
  })
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  fetchFeatures()
}

const handlePageChange = (page) => {
  pagination.page = page
  fetchFeatures()
}

const getTypeTag = (type) => {
  const map = {
    string: 'info',
    number: 'success',
    boolean: 'warning',
    date: 'primary'
  }
  return map[type] || 'info'
}

const getTypeText = (type) => {
  const map = {
    string: '字符串',
    number: '数字',
    boolean: '布尔',
    date: '日期'
  }
  return map[type] || type
}

const getCategoryText = (category) => {
  const map = {
    amount: '金额',
    invoice: '发票',
    reimbursement: '报销单',
    user: '用户',
    time: '时间',
    other: '其他'
  }
  return map[category] || category
}

onMounted(() => {
  fetchFeatures()
})
</script>

<style scoped>
.feature-container {
  padding: 20px;
}

.search-card {
  margin-bottom: 20px;
}

.search-form {
  margin-bottom: 0;
}

.table-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.value-item {
  margin-bottom: 10px;
  padding: 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.el-pagination {
  display: flex;
  justify-content: flex-end;
}
</style>
