<template>
  <div class="rule-management">
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="规则名称">
          <el-input v-model="filterForm.name" placeholder="请输入规则名称" clearable style="width: 200px;" />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="filterForm.type" placeholder="请选择规则类型" clearable style="width: 150px;">
            <el-option label="金额审核" value="amount_validation" />
            <el-option label="发票审核" value="invoice_validation" />
            <el-option label="时间审核" value="time_validation" />
            <el-option label="合规审核" value="compliance" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterForm.enabled" placeholder="请选择状态" clearable style="width: 120px;">
            <el-option label="启用" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">
            搜索
          </el-button>
          <el-button :icon="Refresh" @click="handleReset" style="margin-left: 10px;">
            重置
          </el-button>
          <el-button v-if="isAdmin" type="success" :icon="Plus" @click="handleCreate" style="margin-left: 10px;">
            新建规则
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="ruleList" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.id" placement="top">
              <span class="id-text">{{ row.id.substring(0, 12) }}...</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="规则名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">
              {{ getTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="100" />
        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column v-if="isAdmin" label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="warning" link size="small" @click="handleTest(row)">
              测试
            </el-button>
            <el-button
              :type="row.enabled ? 'warning' : 'success'"
              link
              size="small"
              @click="handleToggleEnable(row)"
            >
              {{ row.enabled ? '禁用' : '启用' }}
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
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

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="800px"
      @close="handleDialogClose"
    >
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="ruleFormRules"
        label-width="120px"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="规则类型" prop="type">
          <el-select v-model="ruleForm.type" placeholder="请选择规则类型" style="width: 100%">
            <el-option label="金额审核" value="amount" />
            <el-option label="发票审核" value="invoice" />
            <el-option label="时间审核" value="time" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-input-number v-model="ruleForm.priority" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="规则表达式" prop="rule_content">
          <el-input
            v-model="ruleForm.rule_content"
            type="textarea"
            :rows="6"
            placeholder="请输入规则表达式，例如：amount > 1000"
          />
        </el-form-item>
        <el-form-item label="错误提示" prop="error_message">
          <el-input
            v-model="ruleForm.error_message"
            type="textarea"
            :rows="3"
            placeholder="请输入规则不通过时的错误提示"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="ruleForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入规则描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="enabled">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="testDialogVisible"
      title="测试规则"
      width="600px"
      @close="handleTestDialogClose"
    >
      <el-form label-width="120px">
        <el-form-item label="规则名称">
          <el-input v-model="currentRule.name" disabled />
        </el-form-item>
        <el-form-item label="测试数据">
          <el-input
            v-model="testData"
            type="textarea"
            :rows="6"
            placeholder='请输入测试数据，JSON格式，例如：{"amount": 1500, "invoice_count": 3}'
          />
        </el-form-item>
        <el-form-item v-if="testResult" label="测试结果">
          <el-alert
            :type="testResult.passed ? 'success' : 'error'"
            :title="testResult.passed ? '规则通过' : '规则不通过'"
            :description="testResult.message"
            :closable="false"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleRunTest">运行测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { getRules, createRule, updateRule, deleteRule, enableRule, disableRule, testRule } from '@/api/rule'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isAdmin = computed(() => userStore.isAdmin)

const loading = ref(false)
const ruleList = ref([])
const filterForm = reactive({
  name: '',
  type: '',
  enabled: null
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const dialogVisible = ref(false)
const dialogTitle = ref('')
const ruleFormRef = ref(null)
const ruleForm = reactive({
  id: '',
  name: '',
  type: '',
  priority: 1,
  rule_content: '',
  error_message: '',
  description: '',
  enabled: true
})

const ruleFormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
  priority: [{ required: true, message: '请输入优先级', trigger: 'blur' }],
  rule_content: [{ required: true, message: '请输入规则表达式', trigger: 'blur' }],
  error_message: [{ required: true, message: '请输入错误提示', trigger: 'blur' }]
}

const testDialogVisible = ref(false)
const currentRule = ref({})
const testData = ref('')
const testResult = ref(null)

const fetchRules = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.pageSize,
      name: filterForm.name,
      type: filterForm.type,
      status: filterForm.enabled === true ? 'enabled' : filterForm.enabled === false ? 'disabled' : ''
    }
    const res = await getRules(params)
    if (res.code === 200) {
      ruleList.value = res.data.rules || []
      pagination.total = res.data.total || 0
    } else {
      ElMessage.error(res.message || '获取规则列表失败')
    }
  } catch (error) {
    ElMessage.error('获取规则列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchRules()
}

const handleReset = () => {
  filterForm.name = ''
  filterForm.type = ''
  filterForm.enabled = null
  pagination.page = 1
  fetchRules()
}

const handleCreate = () => {
  dialogTitle.value = '新建规则'
  Object.assign(ruleForm, {
    id: '',
    name: '',
    type: '',
    priority: 1,
    rule_content: '',
    error_message: '',
    description: '',
    enabled: true
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑规则'
  Object.assign(ruleForm, {
    id: row.id,
    name: row.name,
    type: row.type,
    priority: row.priority,
    rule_content: row.rule_content,
    error_message: row.error_message,
    description: row.description,
    enabled: row.enabled
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!ruleFormRef.value) return
  await ruleFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const data = { ...ruleForm }
        if (data.id) {
          await updateRule(data.id, data)
          ElMessage.success('更新规则成功')
        } else {
          await createRule(data)
          ElMessage.success('创建规则成功')
        }
        dialogVisible.value = false
        fetchRules()
      } catch (error) {
        ElMessage.error(error.message || '操作失败')
      }
    }
  })
}

const handleDialogClose = () => {
  if (ruleFormRef.value) {
    ruleFormRef.value.resetFields()
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除规则"${row.name}"吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteRule(row.id)
      ElMessage.success('删除规则成功')
      fetchRules()
    } catch (error) {
      ElMessage.error('删除规则失败')
    }
  }).catch(() => {})
}

const handleToggleEnable = async (row) => {
  try {
    if (row.enabled) {
      await disableRule(row.id)
      ElMessage.success('禁用规则成功')
    } else {
      await enableRule(row.id)
      ElMessage.success('启用规则成功')
    }
    fetchRules()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const handleTest = (row) => {
  currentRule.value = row
  testData.value = ''
  testResult.value = null
  testDialogVisible.value = true
}

const handleRunTest = async () => {
  try {
    const data = JSON.parse(testData.value)
    const res = await testRule(currentRule.value.id, data)
    if (res.code === 200) {
      testResult.value = {
        passed: res.data.passed,
        message: res.data.message || '测试完成'
      }
    } else {
      ElMessage.error(res.message || '测试失败')
    }
  } catch (error) {
    if (error instanceof SyntaxError) {
      ElMessage.error('测试数据格式错误，请输入有效的JSON')
    } else {
      ElMessage.error('测试失败')
    }
  }
}

const handleTestDialogClose = () => {
  testData.value = ''
  testResult.value = null
}

const handleSizeChange = (val) => {
  pagination.pageSize = val
  pagination.page = 1
  fetchRules()
}

const handleCurrentChange = (val) => {
  pagination.page = val
  fetchRules()
}

const getTypeTagType = (type) => {
  const typeMap = {
    'amount': 'warning',
    'amount_validation': 'warning',
    'invoice': 'success',
    'invoice_validation': 'success',
    'time': 'info',
    'time_validation': 'info',
    'compliance': 'danger',
    'custom': ''
  }
  return typeMap[type] || ''
}

const getTypeText = (type) => {
  const typeMap = {
    'amount': '金额审核',
    'amount_validation': '金额审核',
    'invoice': '发票审核',
    'invoice_validation': '发票审核',
    'time': '时间审核',
    'time_validation': '时间审核',
    'compliance': '合规审核',
    'custom': '自定义'
  }
  return typeMap[type] || type
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchRules()
})
</script>

<style scoped>
.rule-management {
  padding: 20px;
}

.filter-card,
.table-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 20px;
}

.filter-card :deep(.el-card__body) {
  padding: 24px;
}

.filter-form :deep(.el-form-item) {
  margin-bottom: 0;
}

.filter-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #2c3e50;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table th) {
  background-color: #f5f7fa;
  font-weight: 600;
  color: #2c3e50;
}

:deep(.el-dialog__body) {
  padding: 24px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

.id-text {
  cursor: pointer;
  color: #409eff;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.id-text:hover {
  text-decoration: underline;
}
</style>
