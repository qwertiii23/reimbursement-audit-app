<template>
  <div class="reimbursement">
    <div class="page-header">
      <h1>报销单管理</h1>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        新建报销单
      </el-button>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="标题">
          <el-input v-model="filters.title" placeholder="请输入标题" clearable style="width: 200px;" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="请选择状态" clearable style="width: 150px;">
            <el-option label="待提交" value="pending_submission" />
            <el-option label="待审核" value="pending" />
            <el-option label="审核中" value="auditing" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item label="创建时间">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 300px;"
            @change="handleDateChange"
          />
          <el-button type="primary" @click="loadReimbursements" style="margin-left: 10px;">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetFilters" style="margin-left: 10px;">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="reimbursements" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.id" placement="top">
              <span class="id-text">{{ row.id.substring(0, 12) }}...</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="150" />
        <el-table-column prop="total_amount" label="金额" width="120">
          <template #default="{ row }">
            ¥{{ row.total_amount?.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="expense_date" label="报销日期" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row.id)">
              查看
            </el-button>
            <el-button
              v-if="row.status === 'pending_submission'"
              type="primary"
              link
              size="small"
              @click="editReimbursement(row.id)"
            >
              编辑
            </el-button>
            <el-button
              v-if="row.status === 'pending_submission'"
              type="success"
              link
              size="small"
              @click="startAudit(row.id)"
            >
              提交审核
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadReimbursements"
        @current-change="loadReimbursements"
        class="pagination"
      />
    </el-card>

    <el-dialog
      v-model="showCreateDialog"
      title="新建报销单"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="100px"
      >
        <el-form-item label="标题" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入报销单标题" />
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number
            v-model="createForm.amount"
            :min="0"
            :precision="2"
            :step="100"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="报销类别" prop="category">
          <el-select v-model="createForm.category" placeholder="请选择报销类别" style="width: 100%">
            <el-option label="差旅费" value="差旅费" />
            <el-option label="交通费" value="交通费" />
            <el-option label="住宿费" value="住宿费" />
            <el-option label="餐饮费" value="餐饮费" />
            <el-option label="办公用品" value="办公用品" />
            <el-option label="招待费" value="招待费" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="报销事由" prop="reason">
          <el-input
            v-model="createForm.reason"
            type="textarea"
            :rows="2"
            placeholder="请输入报销事由"
          />
        </el-form-item>
        <el-form-item label="报销日期" prop="expense_date">
          <el-date-picker
            v-model="createForm.expense_date"
            type="date"
            placeholder="选择报销日期"
            style="width: 100%"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入报销描述"
          />
        </el-form-item>
        <el-form-item label="上传发票">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :limit="10"
            :file-list="fileList"
            accept="image/*"
            list-type="picture-card"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getReimbursementsByUser, getAllReimbursements, uploadReimbursement, uploadInvoice } from '@/api/reimbursement'
import { startAudit as startAuditApi } from '@/api/audit'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const creating = ref(false)
const showCreateDialog = ref(false)
const createFormRef = ref(null)
const uploadRef = ref(null)
const fileList = ref([])
const currentReimbursementId = ref('')

const reimbursements = ref([])

const filters = reactive({
  title: '',
  status: '',
  dateRange: []
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const createForm = reactive({
  title: '',
  amount: 0,
  category: '',
  reason: '',
  expense_date: '',
  description: ''
})

const createRules = {
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' }
  ],
  amount: [
    { required: true, message: '请输入金额', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择报销类别', trigger: 'change' }
  ],
  reason: [
    { required: true, message: '请输入报销事由', trigger: 'blur' }
  ],
  expense_date: [
    { required: true, message: '请选择报销日期', trigger: 'change' }
  ],
  description: [
    { required: true, message: '请输入描述', trigger: 'blur' }
  ]
}

const getStatusType = (status) => {
  const typeMap = {
    'pending_submission': 'info',
    'pending': 'warning',
    'auditing': 'primary',
    'approved': 'success',
    'rejected': 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    'pending_submission': '待提交',
    'pending': '待审核',
    'auditing': '审核中',
    'approved': '已通过',
    'rejected': '已驳回'
  }
  return textMap[status] || status
}

const loadReimbursements = async () => {
  loading.value = true
  try {
    const params = {}
    
    if (filters.title) {
      params.title = filters.title
    }
    if (filters.status) {
      params.status = filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }

    let response
    if (userStore.user.role === 'admin') {
      response = await getAllReimbursements(
        pagination.page,
        pagination.pageSize,
        params
      )
    } else {
      response = await getReimbursementsByUser(
        userStore.user.id,
        pagination.page,
        pagination.pageSize,
        params
      )
    }
    reimbursements.value = response.data?.list || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('加载报销单失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.title = ''
  filters.status = ''
  filters.dateRange = []
  pagination.page = 1
  loadReimbursements()
}

const handleFilterClear = () => {
  loadReimbursements()
}

const handleDateChange = () => {
  loadReimbursements()
}

const handleFileChange = (file) => {
  fileList.value.push(file)
}

const handleFileRemove = (file) => {
  const index = fileList.value.findIndex(item => item.uid === file.uid)
  if (index > -1) {
    fileList.value.splice(index, 1)
  }
}

const handleCreate = async () => {
  if (!createFormRef.value) return

  await createFormRef.value.validate(async (valid) => {
    if (valid) {
      creating.value = true
      try {
        const formData = new FormData()
        formData.append('user_id', userStore.user.id)
        formData.append('user_name', userStore.user.real_name || userStore.user.username || '未知用户')
        formData.append('total_amount', createForm.amount.toString())
        formData.append('category', createForm.category)
        formData.append('title', createForm.title)
        formData.append('reason', createForm.reason)
        formData.append('expense_date', createForm.expense_date)
        formData.append('description', createForm.description)

        const response = await uploadReimbursement(formData)
        currentReimbursementId.value = response.data.reimbursement_id

        if (fileList.value.length > 0) {
          for (const file of fileList.value) {
            await uploadInvoice(currentReimbursementId.value, createForm.category, file.raw)
          }
        }

        ElMessage.success('创建成功')
        showCreateDialog.value = false
        resetCreateForm()
        loadReimbursements()
      } catch (error) {
        ElMessage.error(error.response?.data?.message || '创建失败')
      } finally {
        creating.value = false
      }
    }
  })
}

const resetCreateForm = () => {
  createForm.title = ''
  createForm.amount = 0
  createForm.category = ''
  createForm.reason = ''
  createForm.expense_date = ''
  createForm.description = ''
  fileList.value = []
  currentReimbursementId.value = ''
  createFormRef.value?.resetFields()
}

const viewDetail = (id) => {
  router.push(`/reimbursement/${id}`)
}

const editReimbursement = (id) => {
  router.push(`/reimbursement/${id}/edit`)
}

const startAudit = async (id) => {
  try {
    await startAuditApi({ reimbursement_id: id })
    ElMessage.success('已提交审核')
    loadReimbursements()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '提交审核失败')
  }
}

onMounted(() => {
  loadReimbursements()
})
</script>

<style scoped>
.reimbursement {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: #2c3e50;
  margin: 0;
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

.filter-card :deep(.el-form-item) {
  margin-bottom: 16px;
}

.filter-card :deep(.el-form-item__label) {
  font-weight: 500;
  color: #2c3e50;
}

.filter-card :deep(.el-input__wrapper),
.filter-card :deep(.el-select) {
  width: 100%;
}

.filter-card :deep(.el-button) {
  padding: 10px 20px;
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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
