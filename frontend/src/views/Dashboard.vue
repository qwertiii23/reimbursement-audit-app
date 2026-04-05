<template>
  <div class="dashboard">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <div class="welcome-badge">
            <el-icon><Sunny /></el-icon>
            <span>工作台</span>
          </div>
          <h1 class="greeting">欢迎回来，{{ userStore.user?.username }}</h1>
          <p class="subtitle">这是您的报销管理概览</p>
        </div>

        <div class="header-right">
          <div class="action-group">
            <button class="btn-primary" @click="$router.push('/reimbursement')">
              <el-icon><Plus /></el-icon>
              <span>新建报销单</span>
            </button>
            <button
              v-if="isAdmin"
              class="btn-success" @click="$router.push('/audit')"
            >
              <el-icon><Checked /></el-icon>
              <span>待审核</span>
              <span class="badge">{{ stats.pending || 0 }}</span>
            </button>
            <button
              v-if="isAdmin"
              class="btn-outline" @click="$router.push('/rule-engine')"
            >
              <el-icon><Setting /></el-icon>
              <span>规则引擎</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <div class="stats-grid">
        <div
          v-for="(stat, index) in statCards"
          :key="stat.key"
          class="stat-card"
          :class="`stat-${stat.key}`"
          :style="{ animationDelay: `${index * 0.1}s` }"
        >
          <div class="card-header">
            <div class="icon-wrapper" :class="`icon-${stat.key}`">
              <el-icon><component :is="stat.icon" /></el-icon>
            </div>
          </div>

          <div class="card-body">
            <div class="stat-value">{{ stats[stat.key] }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>

          <div class="card-footer">
            <div class="trend-indicator" :class="stat.trend > 0 ? 'up' : 'down'">
              <el-icon v-if="stat.trend > 0"><Top /></el-icon>
              <el-icon v-else><Bottom /></el-icon>
              <span>{{ Math.abs(stat.trend) }}%</span>
            </div>
            <span class="trend-text">较上月</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 内容区域 -->
    <div class="content-grid">
      <!-- 近期报销单表格 -->
      <div class="table-card">
        <div class="table-header">
          <div class="header-left-section">
            <h3 class="section-title">
              <el-icon><Document /></el-icon>
              <span>近期报销单</span>
            </h3>
            <p class="section-subtitle">最近提交的报销记录</p>
          </div>
          <button class="view-all-btn" @click="$router.push('/reimbursement')">
            <span>查看全部</span>
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>

        <div class="table-container">
          <el-table
            :data="recentReimbursements"
            style="width: 100%"
            class="custom-table"
            row-class-name="table-row"
          >
            <el-table-column prop="id" label="单号" width="200">
              <template #default="{ row }">
                <div class="id-cell">
                  <span class="id-text">#{{ row.id.substring(0, 8) }}...</span>
                  <el-button type="primary" link size="small" @click="copyId(row.id)">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
              </template>
            </el-table-column>

            <el-table-column prop="title" label="标题" min-width="220">
              <template #default="{ row }">
                <span class="title-text">{{ row.title }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="total_amount" label="金额" width="150">
              <template #default="{ row }">
                <div class="amount-cell">
                  <span class="currency">¥</span>
                  <span class="amount-value">{{ formatAmount(row.total_amount) }}</span>
                </div>
              </template>
            </el-table-column>

            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag
                  :type="getStatusType(row.status)"
                  size="default"
                  effect="light"
                  round
                >
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>

            <el-table-column prop="created_at" label="创建时间" width="180">
              <template #default="{ row }">
                <span class="time-text">{{ formatDate(row.created_at) }}</span>
              </template>
            </el-table-column>

            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <button class="detail-btn" @click="viewDetail(row.id)">
                  <span>详情</span>
                  <el-icon><View /></el-icon>
                </button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="table-footer" v-if="recentReimbursements.length > 0">
          <span class="footer-info">显示 {{ recentReimbursements.length }} 条记录</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import {
  getReimbursementsByUser
} from '@/api/reimbursement'
import {
  Document,
  Clock,
  CircleCheck,
  CircleClose,
  Plus,
  Checked,
  Setting,
  ArrowRight,
  Top,
  Bottom,
  Sunny,
  CopyDocument,
  View
} from '@element-plus/icons-vue'

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

const statCards = [
  { key: 'total', label: '总报销单', icon: Document, trend: 12.5 },
  { key: 'pending', label: '待审核', icon: Clock, trend: 8.2 },
  { key: 'approved', label: '已通过', icon: CircleCheck, trend: 15.3 },
  { key: 'rejected', label: '已驳回', icon: CircleClose, trend: -5.1 }
]

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
    'auditing': 'warning',
    'approved': 'success',
    'rejected': 'danger'
  }
  return typeMap[status] || 'info'
}

const formatAmount = (amount) => {
  if (!amount) return '0.00'
  return Number(amount).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const copyId = async (id) => {
  try {
    await navigator.clipboard.writeText(id)
    ElMessage.success('ID 已复制到剪贴板')
  } catch (err) {
    ElMessage.error('复制失败')
  }
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
      pending: recentReimbursements.value.filter(r =>
        r.status === 'pending' || r.status === 'auditing'
      ).length,
      approved: recentReimbursements.value.filter(r =>
        r.status === 'approved'
      ).length,
      rejected: recentReimbursements.value.filter(r =>
        r.status === 'rejected'
      ).length
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
/* 仪表盘容器 */
.dashboard {
  position: relative;
  animation: fadeInUp var(--transition-normal);
}

/* 页面头部 */
.page-header {
  margin-bottom: var(--space-xl);
}

.header-content {
  background: var(--surface);
  border-radius: var(--radius-xl);
  padding: var(--space-xl) var(--space-xl);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-xl);
  position: relative;
  overflow: hidden;
}

.header-content::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--primary-color), var(--secondary-color), var(--accent-color));
}

.header-left {
  flex: 1;
}

.welcome-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 14px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.08));
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: var(--radius-full);
  font-size: 11px;
  font-weight: 600;
  color: var(--primary-color);
  margin-bottom: var(--space-md);
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.welcome-badge .el-icon {
  font-size: 13px;
}

.greeting {
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 var(--space-xs) 0 !important;
  letter-spacing: -0.02em;
  line-height: 1.3;
}

.subtitle {
  font-size: 15px;
  color: var(--text-secondary);
  margin: 0;
  font-weight: 400;
}

.header-right {
  flex-shrink: 0;
}

.action-group {
  display: flex;
  gap: var(--space-sm);
  align-items: center;
}

/* 按钮样式 */
.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-normal);
  box-shadow: var(--shadow-md), 0 2px 8px var(--primary-glow);
  font-family: inherit;
  letter-spacing: 0.02em;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg), var(--shadow-glow);
  background: linear-gradient(135deg, #7c94f4, #8b63b5);
}

.btn-success {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
  border: 1px solid rgba(34, 197, 94, 0.25);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  font-family: inherit;
}

.btn-success:hover {
  background: rgba(34, 197, 94, 0.18);
  border-color: rgba(34, 197, 94, 0.4);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md), 0 0 12px rgba(34, 197, 94, 0.25);
}

.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: #22c55e;
  color: white;
  font-size: 10px;
  font-weight: 700;
  border-radius: var(--radius-full);
  margin-left: 4px;
}

.btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: var(--surface);
  color: var(--text-secondary);
  border: 2px solid var(--border);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  font-family: inherit;
}

.btn-outline:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
  transform: translateY(-2px);
}

/* 统计卡片区域 */
.stats-section {
  margin-bottom: var(--space-xl);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--space-lg);
}

.stat-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  transition: all var(--transition-normal);
  animation: fadeInUp 0.5s ease-out both;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.stat-card:hover {
  transform: translateY(-6px);
  box-shadow: var(--shadow-lg);
  border-color: transparent;
}

.stat-card:hover::before {
  opacity: 1;
}

.stat-total {
  --card-accent: linear-gradient(90deg, #667eea, #764ba2);
}

.stat-total::before {
  background: var(--card-accent);
}

.stat-total:hover {
  box-shadow: 0 10px 30px rgba(102, 126, 234, 0.15);
}

.stat-pending {
  --card-accent: linear-gradient(90deg, #f59e0b, #f97316);
}

.stat-pending::before {
  background: var(--card-accent);
}

.stat-pending:hover {
  box-shadow: 0 10px 30px rgba(245, 158, 11, 0.15);
}

.stat-approved {
  --card-accent: linear-gradient(90deg, #22c55e, #16a34a);
}

.stat-approved::before {
  background: var(--card-accent);
}

.stat-approved:hover {
  box-shadow: 0 10px 30px rgba(34, 197, 94, 0.15);
}

.stat-rejected {
  --card-accent: linear-gradient(90deg, #ef4444, #dc2626);
}

.stat-rejected::before {
  background: var(--card-accent);
}

.stat-rejected:hover {
  box-shadow: 0 10px 30px rgba(239, 68, 68, 0.15);
}

.card-header {
  margin-bottom: var(--space-md);
}

.icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  transition: all var(--transition-normal);
}

.icon-total {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.12), rgba(118, 75, 162, 0.08));
  color: var(--primary-color);
}

.icon-pending {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.12), rgba(249, 115, 22, 0.08));
  color: #f59e0b;
}

.icon-approved {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.12), rgba(22, 163, 74, 0.08));
  color: #22c55e;
}

.icon-rejected {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.12), rgba(220, 38, 38, 0.08));
  color: #ef4444;
}

.stat-card:hover .icon-wrapper {
  transform: scale(1.1) rotate(5deg);
}

.card-body {
  margin-bottom: var(--space-md);
}

.stat-value {
  font-family: var(--font-display);
  font-size: 40px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  letter-spacing: -0.03em;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 500;
}

.card-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: var(--space-md);
  border-top: 1px solid var(--divider);
}

.trend-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
}

.trend-indicator.up {
  color: #16a34a;
  background: rgba(22, 163, 74, 0.1);
}

.trend-indicator.down {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.1);
}

.trend-indicator .el-icon {
  font-size: 13px;
}

.trend-text {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 400;
}

/* 内容网格 */
.content-grid {
  display: grid;
  gap: var(--space-xl);
}

/* 表格卡片 */
.table-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-normal);
  animation: fadeInUp 0.6s ease-out 0.3s both;
}

.table-card:hover {
  box-shadow: var(--shadow-lg);
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-lg) var(--space-xl);
  border-bottom: 1px solid var(--divider);
  background: var(--background);
}

.header-left-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 !important;
  letter-spacing: -0.01em;
}

.section-title .el-icon {
  color: var(--primary-color);
  font-size: 22px;
}

.section-subtitle {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 0;
  font-weight: 400;
}

.view-all-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  background: var(--surface);
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  font-family: inherit;
}

.view-all-btn:hover {
  background: var(--primary-light);
  box-shadow: var(--shadow-md), 0 0 12px var(--primary-glow);
  transform: translateX(4px);
}

.view-all-btn .el-icon {
  font-size: 14px;
  transition: transform var(--transition-fast);
}

.view-all-btn:hover .el-icon {
  transform: translateX(4px);
}

.table-container {
  overflow-x: auto;
  padding: 0 var(--space-md);
}

/* 自定义表格样式 */
.custom-table {
  --el-table-bg-color: transparent !important;
  --el-table-tr-bg-color: transparent !important;
  --el-table-header-bg-color: var(--background) !important;
  --el-table-row-hover-bg-color: var(--primary-light) !important;
  --el-table-border-color: var(--divider) !important;
  --el-table-text-color: var(--text-secondary) !important;
  --el-table-header-text-color: var(--text-primary) !important;
}

:deep(.el-table__header th) {
  font-weight: 600 !important;
  font-size: 13px !important;
  text-transform: uppercase !important;
  letter-spacing: 0.05em !important;
  border-bottom: 2px solid var(--border) !important;
  background: var(--background) !important;
}

:deep(.el-table__row) {
  transition: all var(--transition-fast) !important;
}

:deep(.el-table__row:hover > td) {
  background: var(--primary-light) !important;
}

:deep(.el-table__cell) {
  border-bottom: 1px solid var(--divider) !important;
  padding: 14px 12px !important;
}

/* ID 单元格 */
.id-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.id-text {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: 13px;
  color: var(--text-secondary);
  letter-spacing: -0.01em;
}

/* 标题单元格 */
.title-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  display: block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 金额单元格 */
.amount-cell {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.currency {
  font-size: 13px;
  color: var(--text-tertiary);
  font-weight: 400;
}

.amount-value {
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.time-text {
  font-size: 13px;
  color: var(--text-tertiary);
  font-weight: 400;
}

.detail-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: var(--primary-light);
  color: var(--primary-color);
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: inherit;
}

.detail-btn:hover {
  background: rgba(102, 126, 234, 0.2);
  border-color: var(--primary-color);
  box-shadow: var(--shadow-sm), 0 0 12px var(--primary-glow);
}

.detail-btn .el-icon {
  font-size: 12px;
}

/* 表格底部 */
.table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-xl);
  border-top: 1px solid var(--divider);
  background: var(--background);
}

.footer-info {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 400;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .header-content {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-lg);
  }

  .action-group {
    width: 100%;
    flex-wrap: wrap;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .dashboard {
    padding: var(--space-md);
  }

  .header-content {
    padding: var(--space-lg);
  }

  .greeting {
    font-size: 26px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .stat-value {
    font-size: 32px;
  }

  .table-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-md);
    padding: var(--space-md);
  }

  .btn-primary,
  .btn-success,
  .btn-outline {
    flex: 1;
    min-width: 130px;
    justify-content: center;
  }
}
</style>
