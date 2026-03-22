<template>
  <div class="dashboard">
    <div class="page-header">
      <div class="header-left">
        <h1>仪表盘</h1>
        <p>欢迎回来，{{ userStore.user?.username }}</p>
      </div>
      <div class="header-right">
        <div class="quick-actions">
          <el-button type="primary" size="small" @click="$router.push('/reimbursement')">
            <el-icon><Plus /></el-icon>
            新建报销单
          </el-button>
          <el-button type="success" size="small" @click="$router.push('/audit')" v-if="isAdmin">
            <el-icon><Checked /></el-icon>
            待审核
          </el-button>
          <el-button type="info" size="small" @click="$router.push('/rule-engine')" v-if="isAdmin">
            <el-icon><Setting /></el-icon>
            规则引擎管理
          </el-button>
        </div>
      </div>
    </div>

    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon total">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总报销单</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon pending">
            <el-icon><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending }}</div>
            <div class="stat-label">待审核</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon approved">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.approved }}</div>
            <div class="stat-label">已通过</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon rejected">
            <el-icon><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.rejected }}</div>
            <div class="stat-label">已驳回</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="content-row">
      <el-col :xs="24" :lg="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">近期报销单</span>
              <el-button type="primary" link @click="$router.push('/reimbursement')">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table :data="recentReimbursements" style="width: 100%" stripe>
            <el-table-column prop="id" label="ID" width="200">
              <template #default="{ row }">
                <el-tooltip :content="row.id" placement="top">
                  <span class="id-text">{{ row.id.substring(0, 12) }}...</span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column prop="title" label="标题" min-width="200" />
            <el-table-column prop="total_amount" label="金额" width="150">
              <template #default="{ row }">
                <span class="amount">¥{{ row.total_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="viewDetail(row.id)">
                  查看
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getReimbursementsByUser } from '@/api/reimbursement'

const router = useRouter()
const userStore = useUserStore()

const isAdmin = computed(() => userStore.isAdmin)
const recentReimbursements = ref([])

const stats = ref({
  total: 0,
  pending: 0,
  approved: 0,
  rejected: 0
})

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

const viewDetail = (id) => {
  router.push(`/reimbursement/${id}`)
}

const loadReimbursements = async () => {
  try {
    const response = await getReimbursementsByUser(userStore.user.id, 1, 10)
    recentReimbursements.value = response.data?.list || []

    stats.value = {
      total: response.data?.total || 0,
      pending: recentReimbursements.value.filter(r => r.status === 'pending' || r.status === 'auditing').length,
      approved: recentReimbursements.value.filter(r => r.status === 'approved').length,
      rejected: recentReimbursements.value.filter(r => r.status === 'rejected').length
    }
  } catch (error) {
    console.error('加载报销单失败:', error)
  }
}

onMounted(() => {
  loadReimbursements()
})
</script>

<style scoped>
.dashboard {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.3);
}

.header-left h1 {
  font-size: 32px;
  font-weight: 700;
  color: #fff;
  margin: 0 0 8px 0;
}

.header-left p {
  font-size: 15px;
  color: rgba(255, 255, 255, 0.85);
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
}

.quick-actions {
  display: flex;
  gap: 12px;
}

.quick-actions .el-button {
  border-radius: 8px;
  padding: 8px 16px;
  font-weight: 500;
}

.stats-row {
  margin-bottom: 24px;
}

.stat-card {
  background: #fff;
  border-radius: 16px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  border: 1px solid #e8e8e8;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  flex-shrink: 0;
}

.stat-icon.total {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.stat-icon.pending {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: #fff;
}

.stat-icon.approved {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: #fff;
}

.stat-icon.rejected {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  color: #fff;
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 36px;
  font-weight: 700;
  color: #2c3e50;
  line-height: 1;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 14px;
  color: #7f8c8d;
  font-weight: 500;
}

.content-row {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
}

.chart-card {
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e8;
}

.chart-card :deep(.el-card__header) {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.chart-card :deep(.el-card__body) {
  padding: 0;
}

.amount {
  font-weight: 600;
  color: #2c3e50;
}

.id-text {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #606266;
}

@media (max-width: 768px) {
  .dashboard {
    padding: 16px;
  }

  .page-header {
    flex-direction: column;
    gap: 20px;
    padding: 20px;
  }

  .page-header h1 {
    font-size: 24px;
  }

  .quick-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .quick-actions .el-button {
    flex: 1;
    min-width: 120px;
  }

  .stat-card {
    padding: 20px;
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    font-size: 28px;
  }

  .stat-value {
    font-size: 28px;
  }
}
</style>
