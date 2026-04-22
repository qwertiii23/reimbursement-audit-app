<template>
  <div class="rule-engine-container">
    <div class="header">
      <h2>规则引擎管理</h2>
      <el-button v-if="isAdmin" type="primary" @click="handleAddRule">
        <el-icon><Plus /></el-icon>
        新增规则
      </el-button>
    </div>

    <el-card class="rule-list-card">
      <template #header>
        <div class="card-header">
          <span>规则列表</span>
          <div class="header-actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索规则名称"
              style="width: 200px; margin-right: 10px"
              clearable
              @clear="handleSearch"
              @keyup.enter="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-select v-model="statusFilter" placeholder="状态筛选" style="width: 120px" clearable @change="handleSearch">
              <el-option label="全部" value="" />
              <el-option label="已上线" value="enabled" />
              <el-option label="已下线" value="disabled" />
            </el-select>
          </div>
        </div>
      </template>

      <el-table :data="ruleList" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" show-overflow-tooltip />
        <el-table-column prop="name" label="规则名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="description" label="规则描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="priority" label="优先级" width="80" align="center" />
        <el-table-column label="条件数量" width="100" align="center">
          <template #default="{ row }">
            {{ row.conditions?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="决策类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getDecisionTypeTag(row.decision?.type)">
              {{ getDecisionTypeText(row.decision?.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '已上线' : '已下线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleViewRule(row)">
              查看
            </el-button>
            <el-button v-if="isAdmin" link type="primary" size="small" @click="handleEditRule(row)">
              编辑
            </el-button>
            <el-button 
              v-if="isAdmin"
              link 
              :type="row.enabled ? 'warning' : 'success'" 
              size="small" 
              @click="handleToggleStatus(row)"
            >
              {{ row.enabled ? '下线' : '上线' }}
            </el-button>
            <el-popconfirm v-if="isAdmin" title="确定要删除此规则吗？" @confirm="handleDeleteRule(row.id)">
              <template #reference>
                <el-button link type="danger" size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="900px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="formRules"
        label-width="120px"
        label-position="left"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="规则名称" prop="name">
              <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-input-number
                v-model="ruleForm.priority"
                :min="0"
                :max="100"
                placeholder="数字越大优先级越高"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="规则描述" prop="description">
          <el-input
            v-model="ruleForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入规则描述"
          />
        </el-form-item>

        <el-form-item label="规则状态" prop="enabled">
          <el-switch v-model="ruleForm.enabled" active-text="上线" inactive-text="下线" />
        </el-form-item>

        <el-divider content-position="left">规则条件</el-divider>

        <div class="conditions-container">
          <div v-if="!ruleForm.conditions || ruleForm.conditions.length === 0" class="empty-conditions">
            <el-empty description="暂无条件，请添加条件" />
          </div>
          <div v-else class="conditions-tree">
            <ConditionGroup
              :conditions="ruleForm.conditions"
              :features="features"
              @update="handleConditionsUpdate"
              @add-condition="handleAddCondition"
              @add-group="handleAddGroup"
              @remove="handleRemoveCondition"
            />
          </div>
          <el-button type="primary" link @click="handleAddCondition" style="margin-top: 10px">
            <el-icon><Plus /></el-icon>
            添加条件
          </el-button>
        </div>

        <el-divider content-position="left">决策配置</el-divider>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="决策类型" prop="decision.type">
              <el-select v-model="ruleForm.decision.type" placeholder="请选择决策类型" style="width: 100%">
                <el-option label="通过" value="approve" />
                <el-option label="拒绝" value="reject" />
                <el-option label="标记" value="mark" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="决策原因" prop="decision.reason">
              <el-input
                v-model="ruleForm.decision.reason"
                placeholder="请输入决策原因"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveRule" :loading="saving">
            保存
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="viewDialogVisible"
      title="规则详情"
      width="800px"
      :close-on-click-modal="false"
      @close="handleViewDialogClose"
    >
      <div v-if="currentRule" class="rule-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="规则ID">
            {{ currentRule.id }}
          </el-descriptions-item>
          <el-descriptions-item label="规则名称">
            {{ currentRule.name }}
          </el-descriptions-item>
          <el-descriptions-item label="规则描述" :span="2">
            {{ currentRule.description }}
          </el-descriptions-item>
          <el-descriptions-item label="优先级">
            {{ currentRule.priority }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentRule.enabled ? 'success' : 'info'">
              {{ currentRule.enabled ? '已上线' : '已下线' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="决策类型">
            <el-tag :type="getDecisionTypeTag(currentRule.decision?.type)">
              {{ getDecisionTypeText(currentRule.decision?.type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="决策原因" :span="2">
            {{ currentRule.decision?.reason }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider>规则条件</el-divider>
        <div class="conditions-view-container">
          <el-empty v-if="!currentRule.conditions || currentRule.conditions.length === 0" description="暂无条件" />
          <div v-else class="conditions-table">
            <template v-for="(condition, index) in currentRule.conditions" :key="condition.id">
              <div v-if="condition.type === 'group'" class="condition-group-view">
                <div class="group-header-view">
                  <span class="group-label-view">条件组</span>
                </div>
                <div class="group-conditions-view">
                  <div
                    v-for="(subCondition, subIndex) in condition.conditions"
                    :key="subCondition.id"
                    class="condition-row-view"
                  >
                    <div class="condition-logic-view" v-if="subIndex > 0">
                      <span class="logic-text">{{ subCondition.logic_op ? subCondition.logic_op.toUpperCase() : (condition.logic_op ? condition.logic_op.toUpperCase() : 'AND') }}</span>
                    </div>
                    <div class="condition-content-view">
                      <span class="feature-text">{{ getFeatureName(subCondition.feature_id) }}</span>
                      <span class="operator-text">{{ getOperatorText(subCondition.operator) }}</span>
                      <span class="value-text">{{ getFeatureValueLabel(subCondition.feature_id, subCondition.value) }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="condition-row-view">
                <div class="condition-logic-view" v-if="index > 0">
                  <span class="logic-text">{{ condition.logic_op ? condition.logic_op.toUpperCase() : 'AND' }}</span>
                </div>
                <div class="condition-content-view">
                  <span class="feature-text">{{ getFeatureName(condition.feature_id) }}</span>
                  <span class="operator-text">{{ getOperatorText(condition.operator) }}</span>
                  <span class="value-text">{{ getFeatureValueLabel(condition.feature_id, condition.value) }}</span>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Search } from '@element-plus/icons-vue'
import request from '@/utils/request'
import ConditionGroup from '@/components/ConditionGroup.vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isAdmin = computed(() => userStore.isAdmin)

const loading = ref(false)
const saving = ref(false)
const searchKeyword = ref('')
const statusFilter = ref('')
const dialogVisible = ref(false)
const viewDialogVisible = ref(false)
const dialogTitle = ref('新增规则')
const ruleFormRef = ref(null)
const currentRule = ref(null)

const ruleList = ref([])
const features = ref([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const ruleForm = reactive({
  id: '',
  name: '',
  description: '',
  priority: 50,
  enabled: true,
  conditions: [],
  decision: {
    type: 'reject',
    reason: ''
  }
})

const formRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入规则描述', trigger: 'blur' }
  ],
  priority: [
    { required: true, message: '请输入优先级', trigger: 'blur' }
  ],
  'decision.type': [
    { required: true, message: '请选择决策类型', trigger: 'change' }
  ],
  'decision.reason': [
    { required: true, message: '请输入决策原因', trigger: 'blur' }
  ]
}

const fetchRules = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.pageSize,
      name: searchKeyword.value,
      enabled: statusFilter.value === 'enabled' ? true : statusFilter.value === 'disabled' ? false : undefined
    }
    const res = await request.get('/engine/rules', { params })
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

const fetchFeatures = async () => {
  try {
    const res = await request.get('/engine/features', {
      params: { page: 1, size: 100 }
    })
    if (res.code === 200) {
      features.value = res.data.features || []
      console.log('获取到的特征列表:', features.value)
    }
  } catch (error) {
    console.error('获取特征列表失败', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchRules()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchRules()
}

const handlePageChange = (page) => {
  pagination.page = page
  fetchRules()
}

const handleAddRule = () => {
  dialogTitle.value = '新增规则'
  resetForm()
  dialogVisible.value = true
}

const handleEditRule = async (rule) => {
  dialogTitle.value = '编辑规则'
  
  if (features.value.length === 0) {
    await fetchFeatures()
  }
  
  ruleForm.id = rule.id
  ruleForm.name = rule.name
  ruleForm.description = rule.description
  ruleForm.priority = rule.priority
  ruleForm.enabled = rule.enabled
  
  const conditions = rule.conditions || []
  if (conditions.length > 0) {
    ruleForm.conditions = conditions.map(cond => ({
      id: cond.id || '',
      type: 'condition',
      featureId: cond.feature_id || cond.featureId,
      operator: cond.operator,
      value: cond.value,
      logicOp: cond.logic_op || cond.logicOp || 'and'
    }))
  } else {
    ruleForm.conditions = []
  }
  
  ruleForm.decision = rule.decision ? JSON.parse(JSON.stringify(rule.decision)) : { type: 'reject', reason: '' }
  dialogVisible.value = true
}

const handleViewRule = async (rule) => {
  try {
    const res = await request.get(`/engine/rules/${rule.id}`)
    if (res.code === 200) {
      currentRule.value = JSON.parse(JSON.stringify(res.data.rule))
      viewDialogVisible.value = true
    } else {
      ElMessage.error(res.message || '获取规则详情失败')
    }
  } catch (error) {
    ElMessage.error('获取规则详情失败')
  }
}

const handleViewDialogClose = () => {
  currentRule.value = null
  viewDialogVisible.value = false
}

const handleToggleStatus = async (rule) => {
  try {
    const action = rule.enabled ? '下线' : '上线'
    await ElMessageBox.confirm(`确定要${action}此规则吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.put(`/engine/rules/${rule.id}/toggle`)
    ElMessage.success(`${action}成功`)
    fetchRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const handleDeleteRule = async (id) => {
  try {
    await request.delete(`/engine/rules/${id}`)
    ElMessage.success('删除成功')
    fetchRules()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const handleAddCondition = () => {
  ruleForm.conditions.push({
    id: '',
    type: 'condition',
    featureId: '',
    operator: 'eq',
    value: '',
    logicOp: ruleForm.conditions.length > 0 ? 'and' : 'and'
  })
}

const handleAddGroup = () => {
  ruleForm.conditions.push({
    id: '',
    type: 'group',
    logicOp: ruleForm.conditions.length > 0 ? 'and' : 'and',
    conditions: []
  })
}

const handleRemoveCondition = (index) => {
  ruleForm.conditions.splice(index, 1)
}

const handleConditionsUpdate = (newConditions) => {
  ruleForm.conditions = newConditions
}

const getFeatureValues = (featureId) => {
  const feature = features.value.find(f => f.id === featureId)
  return feature && feature.values ? feature.values : []
}

const getFeatureType = (featureId) => {
  const feature = features.value.find(f => f.id === featureId)
  return feature ? feature.type : 'string'
}

const flattenConditions = (conditions, parentLogicOp = 'and') => {
  let result = []
  
  conditions.forEach((item, index) => {
    if (item.type === 'group') {
      if (item.conditions && item.conditions.length > 0) {
        const groupConditions = flattenConditions(
          item.conditions,
          item.logicOp
        )
        result = result.concat(groupConditions)
      }
    } else {
      result.push({
        featureId: item.featureId,
        operator: item.operator,
        value: item.value,
        logicOp: index === 0 ? parentLogicOp : item.logicOp,
        sortOrder: index
      })
    }
  })
  
  return result
}

const handleSaveRule = async () => {
  if (!ruleFormRef.value) return
  
  try {
    await ruleFormRef.value.validate()
  } catch (error) {
    return
  }

  if (!ruleForm.conditions || ruleForm.conditions.length === 0) {
    ElMessage.warning('请至少添加一个条件')
    return
  }

  saving.value = true
  try {
    const flattenedConditions = flattenConditions(ruleForm.conditions)
    
    const data = {
      name: ruleForm.name,
      description: ruleForm.description,
      priority: ruleForm.priority,
      enabled: ruleForm.enabled,
      conditions: flattenedConditions,
      decision: {
        type: ruleForm.decision.type,
        reason: ruleForm.decision.reason
      }
    }

    if (ruleForm.id) {
      await request.put(`/engine/rules/${ruleForm.id}`, data)
      ElMessage.success('更新成功')
    } else {
      await request.post('/engine/rules', data)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    fetchRules()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleDialogClose = () => {
  resetForm()
}

const resetForm = () => {
  ruleForm.id = ''
  ruleForm.name = ''
  ruleForm.description = ''
  ruleForm.priority = 50
  ruleForm.enabled = true
  ruleForm.conditions = []
  ruleForm.decision = {
    type: 'reject',
    reason: ''
  }
  if (ruleFormRef.value) {
    ruleFormRef.value.clearValidate()
  }
}

const getDecisionTypeText = (type) => {
  const typeMap = {
    'approve': '通过',
    'reject': '拒绝',
    'mark': '标记'
  }
  return typeMap[type] || type
}

const getDecisionTypeTag = (type) => {
  const typeMap = {
    'approve': 'success',
    'reject': 'danger',
    'mark': 'warning'
  }
  return typeMap[type] || 'info'
}

const getFeatureName = (featureId) => {
  const feature = features.value.find(f => f.id === featureId)
  return feature ? feature.name : featureId
}

const getOperatorText = (operator) => {
  const operatorMap = {
    'eq': '等于',
    'ne': '不等于',
    'gt': '大于',
    'gte': '大于等于',
    'lt': '小于',
    'lte': '小于等于',
    'contains': '包含',
    'not_contains': '不包含'
  }
  return operatorMap[operator] || operator
}

const getFeatureValueLabel = (featureId, value) => {
  const feature = features.value.find(f => f.id === featureId)
  if (!feature || !feature.values || feature.values.length === 0) {
    return value
  }
  
  const featureValue = feature.values.find(v => v.value === value)
  return featureValue ? featureValue.label : value
}

onMounted(() => {
  fetchFeatures()
  fetchRules()
})
</script>

<style scoped>
.rule-engine-container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.rule-list-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.conditions-container {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 20px;
}

.conditions-tree {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.empty-conditions {
  text-align: center;
  padding: 20px;
}

.condition-item {
  padding: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.condition-item:last-child {
  border-bottom: none;
}

.condition-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}

.condition-detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: white;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  min-height: 40px;
}

.condition-detail-logic {
  display: flex;
  align-items: center;
  padding-right: 12px;
}

.condition-detail-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.rule-detail {
  padding: 10px;
}

.conditions-view-container {
  padding: 0;
}

.conditions-table {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.condition-row-view {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
  transition: background 0.2s;
}

.condition-row-view:last-child {
  border-bottom: none;
}

.condition-row-view:hover {
  background: #f5f7fa;
}

.condition-logic-view {
  display: flex;
  align-items: center;
  padding-right: 20px;
  margin-right: 10px;
}

.logic-text {
  display: inline-block;
  padding: 4px 12px;
  background: #e6f7ff;
  color: #1890ff;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 1px;
}

.condition-content-view {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.feature-text {
  color: #303133;
  font-size: 14px;
  font-weight: 500;
}

.operator-text {
  color: #606266;
  font-size: 14px;
  padding: 0 8px;
}

.value-text {
  color: #409eff;
  font-size: 14px;
  font-weight: 500;
  background: #ecf5ff;
  padding: 2px 8px;
  border-radius: 3px;
}

.condition-group-view {
  margin-bottom: 16px;
  padding: 16px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 4px;
}

.group-header-view {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #dee2e6;
}

.group-label-view {
  font-weight: 600;
  color: #495057;
  font-size: 13px;
}

.group-conditions-view {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding-left: 16px;
}

:deep(.el-pagination) {
  justify-content: flex-end;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-dialog__body) {
  max-height: 600px;
  overflow-y: auto;
}
</style>
