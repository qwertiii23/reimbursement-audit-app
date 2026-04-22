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
                      @click="openInvoiceDetailDialog(invoice)"
                    >
                      详情
                    </el-button>
                    <el-button
                      type="warning"
                      link
                      size="small"
                      @click="handleUpdateInvoiceImage(invoice.id)"
                      v-if="reimbursement.status === 'pending_submission' || reimbursement.status === 'pending'"
                    >
                      更换图片
                    </el-button>
                  </div>
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
                <div class="timeline-desc" v-if="auditResult && auditResult.rule_pass !== undefined">
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
                <div class="timeline-desc" v-if="auditResult && auditResult.rag_pass !== undefined">
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
                <div class="timeline-desc" v-if="auditStatus?.workflow_status === '人工审核通过'">
                  <el-tag type="success" size="small">通过</el-tag>
                </div>
                <div class="timeline-desc" v-else-if="auditStatus?.workflow_status === '人工审核驳回'">
                  <el-tag type="danger" size="small">驳回</el-tag>
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

            <div v-if="auditResult.rag_results" class="rag-results">
              <el-button type="primary" @click="showRAGDialog = true" style="width: 100%">
                <el-icon><Document /></el-icon>
                查看AI智能审核详情
              </el-button>
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

            <div v-if="isAdmin && reimbursement.status === 'auditing' && auditStatus?.workflow_status === '待人工审核'" class="manual-audit-section">
              <div class="manual-audit-title">人工审核</div>
              <div class="manual-audit-buttons">
                <el-button 
                  type="success" 
                  @click="handleManualApprove" 
                  :loading="manualAuditLoading"
                >
                  <el-icon><CircleCheck /></el-icon>
                  通过
                </el-button>
                <el-button 
                  type="danger" 
                  @click="handleManualReject" 
                  :loading="manualAuditLoading"
                >
                  <el-icon><CircleClose /></el-icon>
                  驳回
                </el-button>
              </div>
            </div>

            <el-button 
              v-if="reimbursement.status === 'auditing' && auditStatus?.workflow_status !== '待人工审核'"
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

    <el-dialog v-model="invoiceDetailDialog" title="发票详情" width="700px">
      <div v-if="currentInvoice" class="invoice-detail-content">
        <div class="invoice-image-preview">
          <el-image 
            :src="getImageUrl(currentInvoice.image_path)" 
            fit="contain"
            style="width: 100%; max-height: 400px;"
          >
            <template #error>
              <div class="image-error">
                <el-icon><Picture /></el-icon>
                <span>图片加载失败</span>
              </div>
            </template>
          </el-image>
        </div>
        
        <el-divider />
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="发票类型">
            {{ currentInvoice.type || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="发票号码">
            {{ currentInvoice.number || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="发票日期">
            {{ currentInvoice.date || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="发票金额">
            <span class="amount-text">¥{{ currentInvoice.amount?.toFixed(2) || '0.00' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="税额">
            ¥{{ currentInvoice.tax_amount?.toFixed(2) || '0.00' }}
          </el-descriptions-item>
          <el-descriptions-item label="价税合计">
            <span class="amount-text">¥{{ currentInvoice.total_amount?.toFixed(2) || '0.00' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="购买方" :span="2">
            {{ currentInvoice.buyer_name || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="销售方" :span="2">
            {{ currentInvoice.seller_name || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="商品名称" :span="2">
            {{ currentInvoice.commodity_name || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="规格型号">
            {{ currentInvoice.specification || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="单位">
            {{ currentInvoice.unit || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="数量">
            {{ currentInvoice.quantity || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="单价">
            ¥{{ currentInvoice.price?.toFixed(2) || '0.00' }}
          </el-descriptions-item>
          <el-descriptions-item label="报销类别" :span="2">
            {{ currentInvoice.category || '未识别' }} / {{ currentInvoice.sub_category || '未识别' }}
          </el-descriptions-item>
          <el-descriptions-item label="OCR状态" :span="2">
            <el-tag :type="currentInvoice.ocr_result ? 'success' : 'info'" size="default">
              {{ currentInvoice.ocr_result ? '已识别' : '待识别' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="invoiceDetailDialog = false">关闭</el-button>
      </template>
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

    <el-dialog v-model="showRAGDialog" title="AI智能审核详情" width="700px" top="5vh">
      <div class="rag-dialog-content" v-if="parsedRAGResult">
        <div class="rag-header">
          <div class="rag-conclusion" :class="parsedRAGResult.conclusion === '驳回' ? 'reject' : 'pass'">
            <el-icon size="20">
              <CircleClose v-if="parsedRAGResult.conclusion === '驳回'" />
              <CircleCheck v-else />
            </el-icon>
            <span>审核结论：{{ parsedRAGResult.conclusion }}</span>
          </div>
          <div class="rag-confidence" v-if="parsedRAGResult.confidence">
            <span class="label">置信度：</span>
            <el-progress 
              :percentage="Math.round(parsedRAGResult.confidence * 100)" 
              :color="parsedRAGResult.confidence > 0.7 ? '#67c23a' : '#e6a23c'"
              :stroke-width="10"
              style="width: 150px"
            />
          </div>
        </div>

        <el-divider />

        <div class="rag-section" v-if="parsedRAGResult.reasoning">
          <div class="rag-section-title">
            <el-icon><Warning /></el-icon>
            审核理由
          </div>
          <div class="rag-reasoning">
            <p>{{ parsedRAGResult.reasoning }}</p>
          </div>
        </div>

        <div class="rag-section" v-else-if="parsedRAGResult.reasons && parsedRAGResult.reasons.length > 0">
          <div class="rag-section-title">
            <el-icon><Warning /></el-icon>
            审核理由
          </div>
          <div class="rag-reasons">
            <div v-for="(reason, index) in parsedRAGResult.reasons" :key="index" class="reason-item">
              <div class="reason-number">{{ index + 1 }}</div>
              <div class="reason-content">
                <div class="reason-title" v-if="reason.title">{{ reason.title }}</div>
                <div class="reason-text">{{ reason.text }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="rag-section" v-if="parsedRAGResult.suggestions && parsedRAGResult.suggestions.length > 0">
          <div class="rag-section-title">
            <el-icon><InfoFilled /></el-icon>
            处理建议
          </div>
          <div class="rag-suggestions">
            <div v-for="(suggestion, index) in parsedRAGResult.suggestions" :key="index" class="suggestion-item">
              <el-icon><Right /></el-icon>
              <span>{{ suggestion }}</span>
            </div>
          </div>
        </div>

        <div class="rag-section" v-if="parsedRAGResult.rawContent">
          <div class="rag-section-title collapsible" @click="toggleRawContent">
            <el-icon><Document /></el-icon>
            完整分析报告
            <el-icon class="toggle-icon" :class="{ expanded: showRawContent }"><ArrowDown /></el-icon>
          </div>
          <div class="rag-raw-content" v-show="showRawContent">
            <pre>{{ parsedRAGResult.rawContent }}</pre>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showRAGDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRejectDialog" title="审核驳回" width="500px">
      <el-form label-width="80px">
        <el-form-item label="驳回原因">
          <el-input
            v-model="rejectReason"
            type="textarea"
            :rows="4"
            placeholder="请输入驳回原因（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRejectDialog = false">取消</el-button>
        <el-button type="danger" @click="confirmReject" :loading="manualAuditLoading">
          确认驳回
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh, Back, Document, CircleCheck, CircleClose, Warning, InfoFilled, Right, ArrowDown } from '@element-plus/icons-vue'
import {
  getReimbursementById,
  uploadInvoice,
  triggerOCR as triggerOCRApi,
  updateInvoiceImage as updateInvoiceImageApi
} from '@/api/reimbursement'
import { startAudit as startAuditApi, getAuditResult, getAuditStatus, getFlowLogs, getFlowLogsByReimbursementId, withdrawAudit, manualAudit } from '@/api/audit'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const reimbursement = ref({})
const invoices = ref([])
const showRAGDialog = ref(false)
const showRawContent = ref(false)
const auditResult = ref(null)
const isAdmin = computed(() => userStore.isAdmin)
const auditStatus = ref(null)
const flowLogs = ref([])
const refreshing = ref(false)
const loadingFlowLogs = ref(false)
const manualAuditLoading = ref(false)
const showRejectDialog = ref(false)
const rejectReason = ref('')

const showUploadDialog = ref(false)
const showImageDialog = ref(false)
const showUpdateImageDialog = ref(false)
const invoiceDetailDialog = ref(false)
const previewImageUrl = ref('')
const currentInvoice = ref(null)

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

const toggleRawContent = () => {
  showRawContent.value = !showRawContent.value
}

const parsedRAGResult = computed(() => {
  if (!auditResult.value?.rag_results) return null
  
  let ragData = auditResult.value.rag_results
  if (typeof ragData === 'string') {
    try {
      ragData = JSON.parse(ragData)
    } catch {
      return { rawContent: ragData }
    }
  }

  const content = ragData.content || ''
  const result = {
    conclusion: ragData.conclusion || '通过',
    confidence: ragData.confidence || 0,
    reasoning: ragData.reasoning || '',
    reasons: [],
    suggestions: ragData.suggestions || [],
    rawContent: content
  }

  if (!ragData.conclusion) {
    if (content.includes('驳回') || content.includes('不通过')) {
      result.conclusion = '驳回'
    }

    const conclusionMatch = content.match(/\*\*审核结论[：:]\*\*\s*([^\n]+)/i)
    if (conclusionMatch) {
      result.conclusion = conclusionMatch[1].trim()
    }
  }

  const reasonSection = content.match(/\*\*审核理由[：:]\*\*([\s\S]*?)(?=\*\*|$)/i)
  if (reasonSection) {
    const reasonText = reasonSection[1]
    const reasonItems = reasonText.split(/\d+\.\s+/).filter(item => item.trim())
    reasonItems.forEach(item => {
      const lines = item.trim().split('\n')
      const titleMatch = lines[0].match(/\*\*([^*]+)\*\*[：:]?\s*(.*)/)
      if (titleMatch) {
        result.reasons.push({
          title: titleMatch[1].trim(),
          text: (titleMatch[2] + lines.slice(1).join(' ')).trim()
        })
      } else if (lines[0].trim()) {
        result.reasons.push({ text: lines[0].trim() })
      }
    })
  }

  const suggestionSection = content.match(/\*\*处理建议[：:]\*\*([\s\S]*?)(?=\*\*|$)/i) ||
                            content.match(/\*\*后续操作建议[：:]\*\*([\s\S]*?)(?=\*\*|$)/i)
  if (suggestionSection) {
    const suggestionText = suggestionSection[1]
    const suggestionItems = suggestionText.split(/\d+\.\s+/).filter(item => item.trim())
    suggestionItems.forEach(item => {
      const cleanItem = item.replace(/\*\*/g, '').trim()
      if (cleanItem) {
        result.suggestions.push(cleanItem)
      }
    })
  }

  if (result.reasons.length === 0) {
    const numberedItems = content.match(/\d+\.\s+\*\*[^*]+\*\*[：:][\s\S]*?(?=\d+\.\s+\*\*|$)/g)
    if (numberedItems) {
      numberedItems.forEach(item => {
        const titleMatch = item.match(/\d+\.\s+\*\*([^*]+)\*\*[：:]/)
        const textMatch = item.replace(/\d+\.\s+\*\*[^*]+\*\*[：:]/, '').trim()
        if (titleMatch) {
          result.reasons.push({
            title: titleMatch[1].trim(),
            text: textMatch.replace(/\n/g, ' ').trim()
          })
        }
      })
    }
  }

  if (result.suggestions.length === 0) {
    const suggestionItems = content.match(/建议[：:][^\n]*/g)
    if (suggestionItems) {
      suggestionItems.forEach(item => {
        result.suggestions.push(item.replace(/建议[：:]/, '').trim())
      })
    }
  }

  return result
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
    console.log('审核状态:', auditStatus.value)
    console.log('workflow_status:', auditStatus.value?.workflow_status)
    console.log('isAdmin:', isAdmin.value)
    console.log('reimbursement.status:', reimbursement.value.status)
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
    const rulePassedStates = ['规则审核通过', 'AI审核中', '审核结论：通过', '审核结论：驳回', '待人工审核', '人工审核通过', '人工审核驳回', '已通过', '已驳回']
    const ruleFailedStates = ['规则审核失败']
    const ruleRunningStates = ['规则审核中']
    
    if (rulePassedStates.includes(status)) return 'success'
    if (ruleFailedStates.includes(status)) return 'error'
    if (ruleRunningStates.includes(status)) return 'process'
    return 'wait'
  }
  
  if (step === 2) {
    const ragPassedStates = ['审核结论：通过', '待人工审核', '人工审核通过', '人工审核驳回', '已通过']
    const ragFailedStates = ['审核结论：驳回', '已驳回']
    const ragRunningStates = ['AI审核中']
    
    if (ragPassedStates.includes(status)) return 'success'
    if (ragFailedStates.includes(status)) return 'error'
    if (ragRunningStates.includes(status)) return 'process'
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
    } else if (log.flow_status === '审核结论：通过' || log.flow_status === '审核结论：驳回') {
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

const openInvoiceDetailDialog = (invoice) => {
  currentInvoice.value = invoice
  invoiceDetailDialog.value = true
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
    ElMessage.success('审核完成')
    await loadReimbursement()
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

const handleManualApprove = async () => {
  if (!auditResult.value?.id) {
    ElMessage.error('审核信息不存在')
    return
  }
  
  try {
    manualAuditLoading.value = true
    await manualAudit(auditResult.value.id, 'pass', '')
    ElMessage.success('审核通过')
    setTimeout(() => loadReimbursement(), 1000)
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '审核失败')
  } finally {
    manualAuditLoading.value = false
  }
}

const handleManualReject = () => {
  if (!auditResult.value?.id) {
    ElMessage.error('审核信息不存在')
    return
  }
  rejectReason.value = ''
  showRejectDialog.value = true
}

const confirmReject = async () => {
  try {
    manualAuditLoading.value = true
    await manualAudit(auditResult.value.id, 'reject', rejectReason.value)
    ElMessage.success('审核驳回')
    showRejectDialog.value = false
    setTimeout(() => loadReimbursement(), 1000)
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '审核失败')
  } finally {
    manualAuditLoading.value = false
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

.rag-results {
  margin-top: 20px;
}

.rag-dialog-content {
  max-height: 70vh;
  overflow-y: auto;
}

.rag-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 15px;
}

.rag-conclusion {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  padding: 10px 20px;
  border-radius: 8px;
}

.rag-conclusion.pass {
  background: #f0f9eb;
  color: #67c23a;
}

.rag-conclusion.reject {
  background: #fef0f0;
  color: #f56c6c;
}

.rag-confidence {
  display: flex;
  align-items: center;
  gap: 10px;
}

.rag-confidence .label {
  font-size: 14px;
  color: #606266;
}

.rag-section {
  margin-top: 20px;
}

.rag-section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

.rag-section-title.collapsible {
  cursor: pointer;
  user-select: none;
}

.rag-section-title.collapsible:hover {
  color: #409eff;
}

.toggle-icon {
  margin-left: auto;
  transition: transform 0.3s;
}

.toggle-icon.expanded {
  transform: rotate(180deg);
}

.rag-reasons {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rag-reasoning {
  background: #fafafa;
  padding: 15px;
  border-radius: 8px;
  border-left: 3px solid #e6a23c;
  line-height: 1.8;
  color: #303133;
}

.rag-reasoning p {
  margin: 0;
  white-space: pre-wrap;
}

.reason-item {
  display: flex;
  gap: 12px;
  background: #fafafa;
  padding: 12px 15px;
  border-radius: 8px;
  border-left: 3px solid #e6a23c;
}

.reason-number {
  width: 24px;
  height: 24px;
  background: #e6a23c;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.reason-content {
  flex: 1;
}

.reason-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.reason-text {
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
}

.rag-suggestions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.suggestion-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 15px;
  background: #f0f9eb;
  border-radius: 8px;
  color: #67c23a;
  font-size: 14px;
  line-height: 1.6;
}

.suggestion-item .el-icon {
  margin-top: 3px;
  flex-shrink: 0;
}

.rag-raw-content {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 15px;
  max-height: 300px;
  overflow-y: auto;
}

.rag-raw-content pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  font-size: 13px;
  line-height: 1.8;
  color: #606266;
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

.manual-audit-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.manual-audit-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 15px;
  text-align: center;
}

.manual-audit-buttons {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.manual-audit-buttons .el-button {
  flex: 1;
  max-width: 120px;
}

.invoice-detail-content {
  padding: 10px 0;
}

.invoice-image-preview {
  margin-bottom: 20px;
  text-align: center;
}

.invoice-image-preview :deep(.el-image) {
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.image-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  height: 300px;
  background: #f5f7fa;
  color: #909399;
  border-radius: 8px;
}

.image-error .el-icon {
  font-size: 48px;
}

.amount-text {
  font-size: 16px;
  font-weight: 700;
  color: #e74c3c;
}
</style>
