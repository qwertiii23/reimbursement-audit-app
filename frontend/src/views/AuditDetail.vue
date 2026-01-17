<template>
  <div class="audit-detail">
    <div class="page-header">
      <el-button @click="$router.back()" :icon="ArrowLeft">返回</el-button>
      <h1>审核详情</h1>
    </div>

    <el-row :gutter="20">
      <el-col :xs="24" :lg="16">
        <el-card class="reimbursement-card">
          <template #header>
            <div class="card-header">
              <span>报销单信息</span>
              <el-tag :type="getStatusType(reimbursement.status)">
                {{ getStatusText(reimbursement.status) }}
              </el-tag>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="报销单ID">
              {{ reimbursement.id }}
            </el-descriptions-item>
            <el-descriptions-item label="标题">
              {{ reimbursement.title }}
            </el-descriptions-item>
            <el-descriptions-item label="金额">
              <span class="amount">¥{{ reimbursement.amount?.toFixed(2) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="报销日期">
              {{ reimbursement.expense_date }}
            </el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">
              {{ reimbursement.description }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card class="invoices-card">
          <template #header>
            <span>发票列表</span>
          </template>
          <el-row :gutter="15">
            <el-col :xs="24" :sm="12" :md="8" v-for="invoice in invoices" :key="invoice.id">
              <div class="invoice-item">
                <div class="invoice-image" @click="previewImage(invoice.image_url)">
                  <el-image :src="invoice.image_url" fit="cover" lazy>
                    <template #error>
                      <div class="image-error">
                        <el-icon><Picture /></el-icon>
                      </div>
                    </template>
                  </el-image>
                </div>
                <div class="invoice-info">
                  <div class="invoice-category">{{ invoice.category }}</div>
                  <div class="invoice-amount">¥{{ invoice.amount?.toFixed(2) }}</div>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card class="audit-result-card">
          <template #header>
            <span>审核结果</span>
          </template>
          <div v-if="auditResult" class="audit-result">
            <div class="audit-status" :class="auditResult.final_pass ? 'pass' : 'fail'">
              <el-icon>
                <CircleCheck v-if="auditResult.final_pass" />
                <CircleClose v-else />
              </el-icon>
              <span>{{ auditResult.final_pass ? '审核通过' : '审核驳回' }}</span>
            </div>
            <div class="audit-reason">
              <div class="label">审核原因：</div>
              <div class="content">{{ auditResult.reason }}</div>
            </div>
            <div class="audit-details">
              <div class="detail-item">
                <span class="label">规则校验：</span>
                <el-tag :type="auditResult.rule_pass ? 'success' : 'danger'">
                  {{ auditResult.rule_pass ? '通过' : '失败' }}
                </el-tag>
              </div>
              <div class="detail-item">
                <span class="label">RAG分析：</span>
                <el-tag :type="auditResult.rag_pass ? 'success' : 'danger'">
                  {{ auditResult.rag_pass ? '通过' : '失败' }}
                </el-tag>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无审核结果" />
        </el-card>

        <el-card class="actions-card" v-if="auditResult?.final_pass && reimbursement.status === 'pending'">
          <template #header>
            <span>人工审核</span>
          </template>
          <div class="actions">
            <el-button type="success" @click="handleAudit('pass')" style="width: 100%">
              <el-icon><CircleCheck /></el-icon>
              审核通过
            </el-button>
            <el-button type="danger" @click="handleAudit('reject')" style="width: 100%; margin-top: 10px">
              <el-icon><CircleClose /></el-icon>
              审核驳回
            </el-button>
            <el-button
              type="info"
              @click="viewFlowLogs"
              style="width: 100%; margin-top: 10px"
            >
              <el-icon><List /></el-icon>
              查看流程日志
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="showImageDialog" title="图片预览" width="60%">
      <el-image :src="previewImageUrl" fit="contain" style="width: 100%" />
    </el-dialog>

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
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getReimbursementById } from '@/api/reimbursement'
import { getAuditResult, manualAudit } from '@/api/audit'

const route = useRoute()
const router = useRouter()

const reimbursement = ref({})
const invoices = ref([])
const auditResult = ref(null)

const showImageDialog = ref(false)
const showAuditDialog = ref(false)
const previewImageUrl = ref('')
const auditAction = ref('pass')
const auditing = ref(false)

const auditForm = reactive({
  reason: ''
})

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'warning',
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

const loadAuditDetail = async () => {
  try {
    const auditId = route.params.id
    const [auditResponse, reimbursementResponse] = await Promise.all([
      getAuditResult(auditId),
      getReimbursementById(auditResponse?.data?.reimbursement_id)
    ])

    auditResult.value = auditResponse.data
    reimbursement.value = reimbursementResponse.data
    invoices.value = reimbursementResponse.data.invoices || []
  } catch (error) {
    ElMessage.error('加载审核详情失败')
  }
}

const previewImage = (url) => {
  previewImageUrl.value = url
  showImageDialog.value = true
}

const handleAudit = (action) => {
  auditAction.value = action
  auditForm.reason = ''
  showAuditDialog.value = true
}

const confirmAudit = async () => {
  auditing.value = true
  try {
    await manualAudit(route.params.id, auditAction.value, auditForm.reason)
    ElMessage.success('审核成功')
    showAuditDialog.value = false
    loadAuditDetail()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '审核失败')
  } finally {
    auditing.value = false
  }
}

const viewFlowLogs = () => {
  router.push(`/audit/${route.params.id}/flow`)
}

onMounted(() => {
  loadAuditDetail()
})
</script>

<style scoped>
.audit-detail {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: #2c3e50;
  margin: 0;
}

.reimbursement-card,
.invoices-card,
.audit-result-card,
.actions-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.amount {
  font-size: 20px;
  font-weight: 700;
  color: #e74c3c;
}

.invoice-item {
  background: #f8f9fa;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 15px;
  transition: all 0.3s;
}

.invoice-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.invoice-image {
  height: 180px;
  cursor: pointer;
  overflow: hidden;
}

.invoice-image :deep(.el-image) {
  width: 100%;
  height: 100%;
}

.image-error {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #ccc;
  font-size: 40px;
}

.invoice-info {
  padding: 12px;
}

.invoice-category {
  font-size: 14px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 5px;
}

.invoice-amount {
  font-size: 16px;
  font-weight: 700;
  color: #e74c3c;
}

.audit-result {
  padding: 10px 0;
}

.audit-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 18px;
  font-weight: 600;
}

.audit-status.pass {
  background: #d4edda;
  color: #155724;
}

.audit-status.fail {
  background: #f8d7da;
  color: #721c24;
}

.audit-reason {
  margin-bottom: 20px;
}

.audit-reason .label {
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 5px;
}

.audit-reason .content {
  color: #666;
  line-height: 1.6;
}

.audit-details {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 6px;
}

.detail-item .label {
  font-weight: 500;
  color: #2c3e50;
}

.actions {
  display: flex;
  flex-direction: column;
}
</style>
