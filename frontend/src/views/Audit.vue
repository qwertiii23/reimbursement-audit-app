<template>
  <div class="audit">
    <div class="page-header">
      <h1>审核管理</h1>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="审核状态">
          <el-select v-model="filters.status" placeholder="全部状态" clearable>
            <el-option label="待人工审核" value="pending_manual" />
            <el-option label="审核通过" value="approved" />
            <el-option label="审核驳回" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadAudits">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="audits" v-loading="loading" stripe>
        <el-table-column prop="id" label="审核ID" width="200" />
        <el-table-column prop="reimbursement_id" label="报销单ID" width="200" />
        <el-table-column prop="reimbursement_title" label="报销单标题" min-width="150" />
        <el-table-column prop="amount" label="金额" width="120">
          <template #default="{ row }">
            ¥{{ row.amount?.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
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
              v-if="row.status === 'pending_manual'"
              type="success"
              link
              size="small"
              @click="handleAudit(row.id, 'pass')"
            >
              通过
            </el-button>
            <el-button
              v-if="row.status === 'pending_manual'"
              type="danger"
              link
              size="small"
              @click="handleAudit(row.id, 'reject')"
            >
              驳回
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
        @size-change="loadAudits"
        @current-change="loadAudits"
        class="pagination"
      />
    </el-card>

    <el-dialog
      v-model="showAuditDialog"
      :title="auditAction === 'pass' ? '审核通过' : '审核驳回'"
      width="500px"
    >
      <el-form :model="auditForm" label-width="80px">
        <el-form-item label="审核意见">
          <el-input
            v-model="auditForm.reason"
            type="textarea"
            :rows="4"
            :placeholder="auditAction === 'pass' ? '请输入通过意见（可选）' : '请输入驳回原因'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAuditDialog = false">取消</el-button>
        <el-button
          :type="auditAction === 'pass' ? 'success' : 'danger'"
          @click="confirmAudit"
          :loading="auditing"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { manualAudit } from '@/api/audit'
import { getReimbursementsByUser } from '@/api/reimbursement'

const router = useRouter()

const loading = ref(false)
const auditing = ref(false)
const showAuditDialog = ref(false)
const auditAction = ref('pass')

const audits = ref([])

const filters = reactive({
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const auditForm = reactive({
  reason: ''
})

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'warning',
    'pending_manual': 'primary',
    'approved': 'success',
    'rejected': 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    'pending': '审核中',
    'pending_manual': '待人工审核',
    'approved': '已通过',
    'rejected': '已驳回'
  }
  return textMap[status] || status
}

const loadAudits = async () => {
  loading.value = true
  try {
    const response = await getReimbursementsByUser('all', pagination.page, pagination.pageSize)
    const items = response.data?.items || []
    audits.value = items.filter(item => item.audit_id).map(item => ({
      id: item.audit_id,
      reimbursement_id: item.id,
      reimbursement_title: item.title,
      amount: item.amount,
      status: item.status,
      created_at: item.created_at
    }))
    pagination.total = audits.value.length
  } catch (error) {
    ElMessage.error('加载审核列表失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.status = ''
  pagination.page = 1
  loadAudits()
}

const viewDetail = (id) => {
  router.push(`/audit/${id}`)
}

const handleAudit = async (id, action) => {
  auditAction.value = action
  auditForm.reason = ''
  showAuditDialog.value = true
}

const confirmAudit = async () => {
  const auditId = audits.value.find(a => a.status === 'pending_manual')?.id
  if (!auditId) return

  auditing.value = true
  try {
    await manualAudit(auditId, auditAction.value, auditForm.reason)
    ElMessage.success('审核成功')
    showAuditDialog.value = false
    loadAudits()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '审核失败')
  } finally {
    auditing.value = false
  }
}

onMounted(() => {
  loadAudits()
})
</script>

<style scoped>
.audit {
  padding: 20px;
}

.page-header {
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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
