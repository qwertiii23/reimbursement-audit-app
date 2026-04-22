<template>
  <div class="audit">
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="审核状态">
          <el-select v-model="filters.workflowStatus" placeholder="全部状态" clearable style="width: 150px;">
            <el-option label="待人工审核" value="待人工审核" />
            <el-option label="人工审核通过" value="人工审核通过" />
            <el-option label="人工审核驳回" value="人工审核驳回" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadAudits">查询</el-button>
          <el-button @click="resetFilters" style="margin-left: 10px;">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="audits" v-loading="loading" stripe>
        <el-table-column prop="reimbursement_title" label="报销单标题" min-width="150" />
        <el-table-column prop="amount" label="金额" width="120">
          <template #default="{ row }">
            ¥{{ row.amount?.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="workflow_status" label="审核状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.workflow_status)">
              {{ row.workflow_status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ai_conclusion" label="AI审核结论" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.ai_conclusion === '通过'" type="success">通过</el-tag>
            <el-tag v-else-if="row.ai_conclusion === '驳回'" type="danger">驳回</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row.reimbursement_id)">
              查看详情
            </el-button>
            <template v-if="canManualAudit && row.workflow_status === '待人工审核'">
              <el-button
                type="success"
                link
                size="small"
                @click="handleAudit(row.audit_id, 'pass')"
              >
                通过
              </el-button>
              <el-button
                type="danger"
                link
                size="small"
                @click="handleAudit(row.audit_id, 'reject')"
              >
                驳回
              </el-button>
            </template>
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
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { manualAudit, getAuditByReimbursementId } from '@/api/audit'
import { getReimbursementsByUser } from '@/api/reimbursement'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const canManualAudit = computed(() => userStore.canManualAudit)

const loading = ref(false)
const auditing = ref(false)
const showAuditDialog = ref(false)
const auditAction = ref('pass')
const currentAuditId = ref('')

const audits = ref([])

const filters = reactive({
  workflowStatus: '待人工审核'
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
    '待人工审核': 'primary',
    '人工审核通过': 'success',
    '人工审核驳回': 'danger'
  }
  return typeMap[status] || 'info'
}

const loadAudits = async () => {
  loading.value = true
  try {
    const response = await getReimbursementsByUser('all', pagination.page, pagination.pageSize)
    const items = response.data?.list || []
    
    const auditPromises = items
      .filter(item => item.audit_id)
      .map(async item => {
        try {
          const auditRes = await getAuditByReimbursementId(item.id)
          const auditData = auditRes.data
          return {
            audit_id: item.audit_id,
            reimbursement_id: item.id,
            reimbursement_title: item.title,
            amount: item.amount,
            workflow_status: auditData?.workflow_status || '',
            ai_conclusion: auditData?.rag_pass ? '通过' : (auditData?.rag_pass === false ? '驳回' : ''),
            created_at: item.created_at
          }
        } catch (e) {
          return null
        }
      })
    
    const auditResults = (await Promise.all(auditPromises)).filter(Boolean)
    
    let filtered = auditResults
    if (filters.workflowStatus) {
      filtered = auditResults.filter(a => a.workflow_status === filters.workflowStatus)
    }
    
    audits.value = filtered
    pagination.total = filtered.length
  } catch (error) {
    ElMessage.error('加载审核列表失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.workflowStatus = '待人工审核'
  pagination.page = 1
  loadAudits()
}

const viewDetail = (reimbursementId) => {
  router.push(`/reimbursement/${reimbursementId}`)
}

const handleAudit = (auditId, action) => {
  currentAuditId.value = auditId
  auditAction.value = action
  auditForm.reason = ''
  showAuditDialog.value = true
}

const confirmAudit = async () => {
  if (auditAction.value === 'reject' && !auditForm.reason.trim()) {
    ElMessage.warning('请输入驳回原因')
    return
  }

  auditing.value = true
  try {
    await manualAudit(currentAuditId.value, auditAction.value, auditForm.reason)
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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
