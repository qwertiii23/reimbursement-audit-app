<template>
  <div class="reimbursement-detail">
    <div class="page-header">
      <el-button @click="$router.back()" :icon="ArrowLeft">返回</el-button>
      <h1>报销单详情</h1>
    </div>

    <el-row :gutter="20">
      <el-col :xs="24" :lg="16">
        <el-card class="detail-card">
          <template #header>
            <div class="card-header">
              <span>基本信息</span>
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
              <span class="amount">¥{{ reimbursement.total_amount?.toFixed(2) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="报销日期">
              {{ reimbursement.expense_date }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间" :span="2">
              {{ reimbursement.created_at }}
            </el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">
              {{ reimbursement.description }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card class="invoices-card">
          <template #header>
            <div class="card-header">
              <span>发票列表</span>
              <el-button
                type="primary"
                size="small"
                @click="showUploadDialog = true"
                v-if="reimbursement.status === 'pending'"
              >
                <el-icon><Upload /></el-icon>
                上传发票
              </el-button>
            </div>
          </template>
          <div v-if="invoices.length === 0" class="empty-state">
            <el-empty description="暂无发票" />
          </div>
          <el-row :gutter="15" v-else>
            <el-col :xs="24" :sm="12" :md="8" v-for="invoice in invoices" :key="invoice.id">
              <div class="invoice-item">
                <div class="invoice-image" @click="previewImage(getImageUrl(invoice.image_path))">
                  <el-image :src="getImageUrl(invoice.image_path)" fit="cover" lazy>
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
                  <div class="invoice-actions">
                    <el-button
                      type="primary"
                      link
                      size="small"
                      @click="toggleInvoiceDetail(invoice.id)"
                    >
                      {{ expandedInvoices[invoice.id] ? '收起' : '详情' }}
                    </el-button>
                    <el-button
                      type="primary"
                      link
                      size="small"
                      @click="triggerOCR(invoice.id)"
                      v-if="!invoice.ocr_result"
                    >
                      OCR识别
                    </el-button>
                    <el-button
                      type="warning"
                      link
                      size="small"
                      @click="handleUpdateInvoiceImage(invoice.id)"
                      v-if="reimbursement.status === 'pending'"
                    >
                      更换图片
                    </el-button>
                  </div>
                </div>
                <div class="invoice-ocr-detail" v-if="expandedInvoices[invoice.id] && invoice.ocr_result">
                  <el-descriptions :column="2" border size="small">
                    <el-descriptions-item label="发票类型">{{ invoice.type }}</el-descriptions-item>
                    <el-descriptions-item label="发票号码">{{ invoice.number }}</el-descriptions-item>
                    <el-descriptions-item label="发票日期">{{ invoice.date }}</el-descriptions-item>
                    <el-descriptions-item label="发票金额">¥{{ invoice.amount?.toFixed(2) }}</el-descriptions-item>
                    <el-descriptions-item label="税额">¥{{ invoice.tax_amount?.toFixed(2) }}</el-descriptions-item>
                    <el-descriptions-item label="购买方">{{ invoice.buyer_name }}</el-descriptions-item>
                    <el-descriptions-item label="销售方">{{ invoice.seller_name }}</el-descriptions-item>
                    <el-descriptions-item label="商品名称">{{ invoice.commodity_name }}</el-descriptions-item>
                    <el-descriptions-item label="数量">{{ invoice.quantity }}</el-descriptions-item>
                    <el-descriptions-item label="单价">¥{{ invoice.price?.toFixed(2) }}</el-descriptions-item>
                    <el-descriptions-item label="类别" :span="2">{{ invoice.category }} / {{ invoice.sub_category }}</el-descriptions-item>
                  </el-descriptions>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card class="audit-card" v-if="auditResult || auditStatus || flowLogs.length > 0">
          <template #header>
            <div class="card-header">
              <span>审核进度</span>
              <el-button
                type="primary"
                size="small"
                @click="refreshAuditStatus"
                :loading="refreshing"
              >
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <el-timeline class="audit-timeline">
            <el-timeline-item
              :type="getTimelineStatus(0)"
              :timestamp="getTimelineTime(0)"
            >
              <div class="timeline-content">
                <div class="timeline-title">已提交</div>
                <div class="timeline-desc">报销单已提交审核</div>
              </div>
            </el-timeline-item>
            
            <el-timeline-item
              :type="getTimelineStatus(1)"
              :timestamp="getTimelineTime(1)"
            >
              <div class="timeline-content">
                <div class="timeline-title">规则审核</div>
                <div class="timeline-desc" v-if="auditResult">
                  <el-tag :type="auditResult.rule_pass ? 'success' : 'danger'" size="small">
                    {{ auditResult.rule_pass ? '通过' : '失败' }}
                  </el-tag>
                </div>
              </div>
            </el-timeline-item>
            
            <el-timeline-item
              :type="getTimelineStatus(2)"
              :timestamp="getTimelineTime(2)"
              v-if="auditResult && auditResult.rule_pass"
            >
              <div class="timeline-content">
                <div class="timeline-title">AI审核</div>
                <div class="timeline-desc">
                  <el-tag :type="auditResult.rag_pass ? 'success' : 'danger'" size="small">
                    {{ auditResult.rag_pass ? '通过' : '失败' }}
                  </el-tag>
                </div>
              </div>
            </el-timeline-item>
            
            <el-timeline-item
              :type="getTimelineStatus(3)"
              :timestamp="getTimelineTime(3)"
              v-if="auditResult && auditResult.rule_pass && auditResult.rag_pass"
            >
              <div class="timeline-content">
                <div class="timeline-title">人工审核</div>
                <div class="timeline-desc" v-if="auditStatus?.workflow_status && (auditStatus.workflow_status === '人工审核通过' || auditStatus.workflow_status === '人工审核驳回')">
                  <el-tag :type="auditStatus.workflow_status === '人工审核通过' ? 'success' : 'danger'" size="small">
                    {{ auditStatus.workflow_status === '人工审核通过' ? '通过' : '驳回' }}
                  </el-tag>
                </div>
              </div>
            </el-timeline-item>
            
            <el-timeline-item
              type="warning"
              :timestamp="getTimelineTime(4)"
              v-if="hasWithdrawnLog()"
            >
              <div class="timeline-content">
                <div class="timeline-title">已撤回</div>
                <div class="timeline-desc">
                  <el-tag type="warning" size="small">
                    用户撤回审核
                  </el-tag>
                </div>
              </div>
            </el-timeline-item>
          </el-timeline>
          
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
            
            <div v-if="auditResult.rule_results && auditResult.rule_results.filter(r => !r.passed).length > 0" class="rule-results">
              <div class="section-title">规则校验详情</div>
              <div v-for="result in auditResult.rule_results.filter(r => !r.passed)" :key="result.rule_id" class="rule-result-item">
                <div class="rule-header">
                  <el-tag type="danger" size="small">
                    失败
                  </el-tag>
                  <span class="rule-name">{{ result.rule_name }}</span>
                </div>
                <div class="rule-message">{{ result.message }}</div>
              </div>
            </div>
          </div>
        </el-card>

        <el-card class="actions-card">
          <template #header>
            <div class="card-header">
              <span>操作</span>
            </div>
          </template>
          <div class="actions">
            <el-button 
              v-if="reimbursement.status === 'pending_submission' || reimbursement.status === 'pending'"
              type="primary" 
              @click="startAudit" 
              style="width: 100%"
            >
              <el-icon><Checked /></el-icon>
              提交审核
            </el-button>
            <el-button 
              v-if="reimbursement.status === 'pending_submission' || reimbursement.status === 'pending'"
              @click="editReimbursement" 
              style="width: 100%; margin-top: 10px"
            >
              <el-icon><Edit /></el-icon>
              编辑报销单
            </el-button>
            <el-button 
              v-if="reimbursement.status === 'auditing'"
              type="warning"
              @click="handleWithdrawAudit" 
              style="width: 100%; margin-top: 10px"
            >
              <el-icon><Back /></el-icon>
              撤回报销单
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="showUploadDialog" title="上传发票" width="500px">
      <el-form :model="uploadForm" label-width="80px">
        <el-form-item label="发票类型">
          <el-select v-model="uploadForm.category" placeholder="请选择发票类型">
            <el-option label="交通费" value="transport" />
            <el-option label="餐饮费" value="food" />
            <el-option label="住宿费" value="accommodation" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="上传图片">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            accept="image/*"
          >
            <el-button type="primary">选择图片</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpload" :loading="uploading">
          上传
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showImageDialog" title="图片预览" width="60%">
      <el-image :src="previewImageUrl" fit="contain" style="width: 100%" />
    </el-dialog>

    <el-dialog v-model="showUpdateImageDialog" title="更换图片" width="500px">
      <el-upload
        ref="updateImageRef"
        :auto-upload="false"
        :limit="1"
        :on-change="handleUpdateFileChange"
        accept="image/*"
      >
        <el-button type="primary">选择新图片</el-button>
      </el-upload>
      <template #footer>
        <el-button @click="showUpdateImageDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateImage" :loading="updating">
          更新
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh, Back } from '@element-plus/icons-vue'
import {
  getReimbursementById,
  uploadInvoice,
  triggerOCR as triggerOCRApi,
  updateInvoiceImage as updateInvoiceImageApi
} from '@/api/reimbursement'
import { startAudit as startAuditApi, getAuditResult, getAuditStatus, getFlowLogs, getFlowLogsByReimbursementId, withdrawAudit } from '@/api/audit'

const route = useRoute()
const router = useRouter()

const reimbursement = ref({})
const invoices = ref([])
const auditResult = ref(null)
const auditStatus = ref(null)
const flowLogs = ref([])
const expandedInvoices = ref({})
const refreshing = ref(false)
const loadingFlowLogs = ref(false)

const showUploadDialog = ref(false)
const showImageDialog = ref(false)
const showUpdateImageDialog = ref(false)
const previewImageUrl = ref('')

const uploading = ref(false)
const updating = ref(false)
const uploadRef = ref(null)
const updateImageRef = ref(null)

const uploadForm = reactive({
  category: '',
  file: null
})

const updateForm = reactive({
  invoiceId: '',
  file: null
})

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

const loadReimbursement = async () => {
  try {
    const response = await getReimbursementById(route.params.id)
    reimbursement.value = response.data
    invoices.value = response.data.invoices || []
    
    console.log('发票数据:', invoices.value)
    console.log('第一张发票图片路径:', invoices.value[0]?.image_path)
    console.log('第一张发票完整URL:', getImageUrl(invoices.value[0]?.image_path))

    if (response.data.audit_id) {
      loadAuditStatus(response.data.audit_id)
      loadAuditResult(response.data.audit_id)
      loadFlowLogs(response.data.audit_id)
    } else {
      loadFlowLogsByReimbursementId(response.data.id)
    }
  } catch (error) {
    ElMessage.error('加载报销单详情失败')
  }
}

const loadFlowLogs = async (auditId) => {
  try {
    loadingFlowLogs.value = true
    const response = await getFlowLogs(auditId)
    flowLogs.value = response.data || []
  } catch (error) {
    ElMessage.error('加载流程日志失败')
  } finally {
    loadingFlowLogs.value = false
  }
}

const loadFlowLogsByReimbursementId = async (reimbursementId) => {
  try {
    loadingFlowLogs.value = true
    const response = await getFlowLogsByReimbursementId(reimbursementId)
    flowLogs.value = response.data || []
  } catch (error) {
    console.error('加载流程日志失败', error)
  } finally {
    loadingFlowLogs.value = false
  }
}

const viewFlowLogs = () => {
  if (reimbursement.value.audit_id) {
    loadFlowLogs(reimbursement.value.audit_id)
  } else {
    loadFlowLogsByReimbursementId(reimbursement.value.id)
  }
}

const loadAuditStatus = async (auditId) => {
  try {
    const response = await getAuditStatus(auditId)
    auditStatus.value = response.data
  } catch (error) {
    console.error('加载审核状态失败:', error)
  }
}

const loadAuditResult = async (auditId) => {
  try {
    const response = await getAuditResult(auditId)
    auditResult.value = response.data
  } catch (error) {
    console.error('加载审核结果失败:', error)
  }
}

const refreshAuditStatus = async () => {
  if (!reimbursement.value.audit_id) return
  refreshing.value = true
  try {
    await loadAuditStatus(reimbursement.value.audit_id)
    if (auditStatus.value?.status === 'completed' || auditStatus.value?.status === 'failed') {
      await loadAuditResult(reimbursement.value.audit_id)
    }
    ElMessage.success('刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败')
  } finally {
    refreshing.value = false
  }
}

const getTimelineStatus = (step) => {
  if (!auditStatus.value) return 'process'
  const status = auditStatus.value.workflow_status
  
  if (step === 0) {
    return 'success'
  }
  
  if (step === 1) {
    if (status === '规则审核通过') return 'success'
    if (status === '规则审核失败') return 'error'
    if (status === '规则审核中') return 'process'
    return 'wait'
  }
  
  if (step === 2) {
    if (status === 'AI审核通过') return 'success'
    if (status === 'AI审核失败') return 'error'
    if (status === 'AI审核中') return 'process'
    return 'wait'
  }
  
  if (step === 3) {
    if (status === '人工审核通过') return 'success'
    if (status === '人工审核驳回') return 'error'
    if (status === '待人工审核') return 'process'
    return 'wait'
  }
  
  return 'wait'
}

const getTimelineTime = (step) => {
  if (step === 0) {
    return reimbursement.value.created_at || ''
  }
  
  if (!flowLogs.value || flowLogs.value.length === 0) return ''
  
  const logMap = {}
  flowLogs.value.forEach(log => {
    if (log.flow_status === '智能审核开始') {
      logMap['start'] = log.created_at
    } else if (log.flow_status === '规则审核通过' || log.flow_status === '规则审核失败') {
      logMap['rule'] = log.created_at
    } else if (log.flow_status === 'AI审核通过' || log.flow_status === 'AI审核失败') {
      logMap['rag'] = log.created_at
    } else if (log.flow_status === '人工审核通过' || log.flow_status === '人工审核驳回') {
      logMap['manual'] = log.created_at
    } else if (log.flow_status === '已撤回') {
      logMap['withdrawn'] = log.created_at
    }
  })
  
  if (step === 1) return logMap['rule'] || ''
  if (step === 2) return logMap['rag'] || ''
  if (step === 3) return logMap['manual'] || ''
  if (step === 4) return logMap['withdrawn'] || ''
  
  return ''
}

const hasWithdrawnLog = () => {
  if (!flowLogs.value || flowLogs.value.length === 0) return false
  return flowLogs.value.some(log => log.flow_status === '已撤回')
}

const getWorkflowStatusType = (status) => {
  const typeMap = {
    '已提交': 'info',
    '规则审核中': 'warning',
    '规则审核通过': 'success',
    '规则审核失败': 'danger',
    'AI审核中': 'warning',
    'AI审核通过': 'success',
    'AI审核失败': 'danger',
    '待人工审核': 'warning',
    '人工审核通过': 'success',
    '人工审核驳回': 'danger',
    '已通过': 'success',
    '已驳回': 'danger'
  }
  return typeMap[status] || 'info'
}

const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const handleFileChange = (file) => {
  uploadForm.file = file.raw
}

const handleUpload = async () => {
  if (!uploadForm.category || !uploadForm.file) {
    ElMessage.warning('请填写完整信息')
    return
  }

  uploading.value = true
  try {
    await uploadInvoice(reimbursement.value.id, uploadForm.category, uploadForm.file)
    ElMessage.success('上传成功')
    showUploadDialog.value = false
    uploadForm.category = ''
    uploadForm.file = null
    uploadRef.value?.clearFiles()
    loadReimbursement()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

const triggerOCR = async (invoiceId) => {
  try {
    await triggerOCRApi(invoiceId)
    ElMessage.success('OCR识别已触发，请稍后查看结果')
    setTimeout(() => loadReimbursement(), 2000)
  } catch (error) {
    ElMessage.error('触发OCR识别失败')
  }
}

const previewImage = (url) => {
  previewImageUrl.value = url
  showImageDialog.value = true
}

const getImageUrl = (imagePath) => {
  if (!imagePath) return ''
  return `http://127.0.0.1:8080/api/v1/files/${imagePath}`
}

const toggleInvoiceDetail = (invoiceId) => {
  expandedInvoices.value[invoiceId] = !expandedInvoices.value[invoiceId]
}

const handleUpdateInvoiceImage = async (invoiceId) => {
  updateForm.invoiceId = invoiceId
  updateForm.file = null
  showUpdateImageDialog.value = true
}

const handleUpdateFileChange = (file) => {
  updateForm.file = file.raw
}

const handleUpdateImage = async () => {
  if (!updateForm.file) {
    ElMessage.warning('请选择图片')
    return
  }

  updating.value = true
  try {
    await updateInvoiceImage(updateForm.invoiceId, updateForm.file)
    ElMessage.success('更新成功')
    showUpdateImageDialog.value = false
    updateForm.file = null
    updateImageRef.value?.clearFiles()
    loadReimbursement()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '更新失败')
  } finally {
    updating.value = false
  }
}

const startAudit = async () => {
  try {
    await startAuditApi({ reimbursement_id: reimbursement.value.id })
    ElMessage.success('已提交审核，正在自动进行规则审核和AI审核')
    setTimeout(() => loadReimbursement(), 2000)
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '提交审核失败')
  }
}

const editReimbursement = () => {
  router.push(`/reimbursement/${reimbursement.value.id}/edit`)
}

const handleWithdrawAudit = async () => {
  try {
    await withdrawAudit(reimbursement.value.id)
    ElMessage.success('撤回成功')
    setTimeout(() => loadReimbursement(), 1000)
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '撤回失败')
  }
}

onMounted(() => {
  loadReimbursement()
})
</script>

<style scoped>
.reimbursement-detail {
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

.detail-card,
.invoices-card,
.audit-card,
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

.empty-state {
  padding: 40px 0;
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
  margin-bottom: 8px;
}

.invoice-ocr-detail {
  padding: 12px;
  background: #fff;
  border-top: 1px solid #e4e7ed;
}

.invoice-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.audit-progress {
  margin-bottom: 20px;
}

.audit-result {
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
}

.audit-status {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 15px;
}

.audit-status.pass {
  color: #67c23a;
}

.audit-status.fail {
  color: #f56c6c;
}

.audit-reason {
  margin-bottom: 20px;
}

.audit-reason .label {
  font-weight: 600;
  margin-bottom: 5px;
}

.audit-reason .content {
  color: #606266;
  line-height: 1.6;
}

.audit-details {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.audit-details .detail-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.audit-details .label {
  font-weight: 600;
}

.audit-timeline {
  margin: 20px 0;
}

.timeline-content {
  padding: 10px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #409eff;
}

.timeline-title {
  font-size: 16px;
  font-weight: 700;
  color: #2c3e50;
  margin-bottom: 8px;
}

.timeline-desc {
  color: #606266;
  line-height: 1.6;
}

.workflow-status {
  margin-top: 20px;
  text-align: center;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #2c3e50;
  margin: 20px 0 15px 0;
  padding-bottom: 10px;
  border-bottom: 2px solid #e4e7ed;
}

.rule-results {
  margin-top: 20px;
}

.rule-result-item {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 10px;
  border-left: 4px solid #e4e7ed;
}

.rule-result-item:last-child {
  margin-bottom: 0;
}

.rule-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.rule-name {
  font-weight: 600;
  color: #2c3e50;
}

.rule-message {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 8px;
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
  color: #606266;
  line-height: 1.6;
}

.actions {
  display: flex;
  flex-direction: column;
}
</style>
