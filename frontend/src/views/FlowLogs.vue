<template>
  <div class="flow-logs">
    <div class="page-header">
      <el-button @click="$router.back()" :icon="ArrowLeft">返回</el-button>
      <h1>流程日志</h1>
    </div>

    <el-card class="logs-card">
      <template #header>
        <div class="card-header">
          <span>审核流程记录</span>
          <el-tag v-if="reimbursement" :type="getStatusType(reimbursement.status)">
            {{ getStatusText(reimbursement.status) }}
          </el-tag>
        </div>
      </template>

      <div v-loading="loading">
        <el-timeline v-if="flowLogs.length > 0">
          <el-timeline-item
            v-for="(log, index) in flowLogs"
            :key="log.id"
            :timestamp="formatTime(log.created_at)"
            :type="getTimelineType(log.flow_status)"
            placement="top"
          >
            <div class="log-content">
              <div class="log-header">
                <span class="log-stage">{{ getFlowTypeText(log.flow_type) }}</span>
                <el-tag :type="getStatusTagType(log.flow_status)" size="small">
                  {{ getFlowStatusText(log.flow_status) }}
                </el-tag>
              </div>
              <div class="log-details">
                <div class="detail-row" v-if="log.operator_name">
                  <span class="label">操作人：</span>
                  <span class="value">{{ log.operator_name }}</span>
                </div>
                <div class="detail-row" v-if="log.ip_address">
                  <span class="label">IP地址：</span>
                  <span class="value">{{ log.ip_address }}</span>
                </div>
                <div class="detail-row" v-if="log.action">
                  <span class="label">操作：</span>
                  <span class="value">{{ getActionText(log.action) }}</span>
                </div>
                <div class="detail-row" v-if="log.reason">
                  <span class="label">原因：</span>
                  <span class="value reason">{{ log.reason }}</span>
                </div>
                <div class="detail-row" v-if="log.result">
                  <span class="label">结果：</span>
                  <span class="value result">{{ log.result }}</span>
                </div>
              </div>
            </div>
          </el-timeline-item>
        </el-timeline>

        <el-empty v-else description="暂无流程日志" />
      </div>
    </el-card>

    <el-card class="reimbursement-card" v-if="reimbursement">
      <template #header>
        <span>关联报销单信息</span>
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getFlowLogsByReimbursementId, getFlowLogsByAuditId } from '@/api/audit'
import { getReimbursementById } from '@/api/reimbursement'

const route = useRoute()

const flowLogs = ref([])
const reimbursement = ref(null)
const loading = ref(false)

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

const getFlowTypeText = (flowType) => {
  const typeMap = {
    'submitted': '提交审核',
    'rule_check': '规则校验',
    'rag_analysis': 'RAG分析',
    'manual_audit': '人工审核',
    'completed': '审核完成',
    'rejected': '审核驳回'
  }
  return typeMap[flowType] || flowType
}

const getFlowStatusText = (flowStatus) => {
  const statusMap = {
    'in_progress': '进行中',
    'success': '成功',
    'failed': '失败',
    'pending': '等待中'
  }
  return statusMap[flowStatus] || flowStatus
}

const getStatusTagType = (flowStatus) => {
  const typeMap = {
    'in_progress': 'primary',
    'success': 'success',
    'failed': 'danger',
    'pending': 'warning'
  }
  return typeMap[flowStatus] || 'info'
}

const getActionText = (action) => {
  const actionMap = {
    'submit': '提交',
    'start_audit': '开始审核',
    'rule_check': '规则校验',
    'rag_analysis': 'RAG分析',
    'manual_approve': '人工审核通过',
    'manual_reject': '人工审核驳回',
    'auto_reject': '自动驳回',
    'complete': '完成'
  }
  return actionMap[action] || action
}

const getTimelineType = (flowStatus) => {
  if (flowStatus === 'success') return 'success'
  if (flowStatus === 'failed') return 'danger'
  if (flowStatus === 'in_progress') return 'primary'
  return 'info'
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

const loadFlowLogs = async () => {
  loading.value = true
  try {
    const routeName = route.name
    const id = route.params.id

    let logs = []
    if (routeName === 'AuditFlow') {
      const response = await getFlowLogsByAuditId(id)
      logs = response.data
    } else if (routeName === 'ReimbursementFlow') {
      const response = await getFlowLogsByReimbursementId(id)
      logs = response.data
    }

    flowLogs.value = logs

    if (logs.length > 0 && logs[0].reimbursement_id) {
      loadReimbursement(logs[0].reimbursement_id)
    }
  } catch (error) {
    ElMessage.error('加载流程日志失败')
  } finally {
    loading.value = false
  }
}

const loadReimbursement = async (id) => {
  try {
    const response = await getReimbursementById(id)
    reimbursement.value = response.data
  } catch (error) {
    console.error('加载报销单信息失败:', error)
  }
}

onMounted(() => {
  loadFlowLogs()
})
</script>

<style scoped>
.flow-logs {
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

.logs-card,
.reimbursement-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.log-content {
  padding: 10px 0;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.log-stage {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
}

.log-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-left: 10px;
  border-left: 3px solid #e4e7ed;
}

.detail-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.detail-row .label {
  font-weight: 500;
  color: #606266;
  min-width: 80px;
}

.detail-row .value {
  color: #303133;
  line-height: 1.6;
  flex: 1;
}

.detail-row .value.reason {
  color: #e6a23c;
}

.detail-row .value.result {
  color: #67c23a;
}

.amount {
  font-size: 18px;
  font-weight: 700;
  color: #e74c3c;
}
</style>
