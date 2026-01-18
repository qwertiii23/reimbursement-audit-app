<template>
  <div class="rules">
    <div class="page-header">
      <div class="header-left">
        <h1>规则管理</h1>
        <p>管理报销审核规则</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新建规则
        </el-button>
      </div>
    </div>

    <el-card class="rules-card">
      <el-table :data="rules" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="规则名称" width="200" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column prop="type" label="规则类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getRuleTypeColor(row.type)">
              {{ getRuleTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="toggleRule(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="editRule(row)">
              编辑
            </el-button>
            <el-button type="danger" link @click="handleDeleteRule(row.id)">
              删除
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
        @size-change="loadRules"
        @current-change="loadRules"
        class="pagination"
      />
    </el-card>

    <el-dialog
      v-model="showCreateDialog"
      :title="editingRule ? '编辑规则' : '新建规则'"
      width="600px"
      @close="resetForm"
    >
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="ruleForm.type" placeholder="请选择规则类型">
            <el-option label="金额限制" value="amount_limit" />
            <el-option label="日期限制" value="date_limit" />
            <el-option label="类别限制" value="category_limit" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="ruleForm.category" placeholder="请输入分类" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="ruleForm.priority" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="ruleForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入规则描述"
          />
        </el-form-item>
        <el-form-item label="规则配置">
          <el-input
            v-model="ruleForm.definition"
            type="textarea"
            :rows="5"
            placeholder='请输入JSON配置，例如：{"max_amount": 1000}'
          />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveRule" :loading="saving">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getRules, createRule, updateRule, deleteRule as deleteRuleApi, enableRule, disableRule } from '@/api/rules'

const rules = ref([])
const loading = ref(false)
const saving = ref(false)
const showCreateDialog = ref(false)
const editingRule = ref(null)

const ruleForm = reactive({
  id: '',
  name: '',
  type: '',
  category: '',
  priority: 1,
  description: '',
  definition: '',
  enabled: true
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const getRuleTypeText = (type) => {
  const typeMap = {
    'amount_limit': '金额限制',
    'date_limit': '日期限制',
    'category_limit': '类别限制',
    'custom': '自定义'
  }
  return typeMap[type] || type
}

const getRuleTypeColor = (type) => {
  const colorMap = {
    'amount_limit': 'primary',
    'date_limit': 'success',
    'category_limit': 'warning',
    'custom': 'info'
  }
  return colorMap[type] || 'info'
}

const loadRules = async () => {
  loading.value = true
  try {
    const response = await getRules({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    rules.value = response.data?.rules || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '加载规则失败')
  } finally {
    loading.value = false
  }
}

const editRule = (rule) => {
  editingRule.value = rule
  ruleForm.id = rule.id
  ruleForm.name = rule.name
  ruleForm.type = rule.type
  ruleForm.category = rule.category
  ruleForm.priority = rule.priority
  ruleForm.description = rule.description
  ruleForm.definition = rule.definition
  ruleForm.enabled = rule.enabled
  showCreateDialog.value = true
}

const saveRule = async () => {
  if (!ruleForm.name || !ruleForm.type) {
    ElMessage.warning('请填写完整信息')
    return
  }

  try {
    JSON.parse(ruleForm.definition)
  } catch (error) {
    ElMessage.warning('规则配置必须是有效的JSON格式')
    return
  }

  saving.value = true
  try {
    if (editingRule.value) {
      await updateRule(ruleForm.id, ruleForm)
      ElMessage.success('更新成功')
    } else {
      await createRule(ruleForm)
      ElMessage.success('创建成功')
    }
    showCreateDialog.value = false
    resetForm()
    loadRules()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const toggleRule = async (rule) => {
  try {
    if (rule.enabled) {
      await enableRule(rule.id)
      ElMessage.success('规则已启用')
    } else {
      await disableRule(rule.id)
      ElMessage.success('规则已禁用')
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '操作失败')
    rule.enabled = !rule.enabled
  }
}

const handleDeleteRule = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteRule(id)
    ElMessage.success('删除成功')
    loadRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

const resetForm = () => {
  editingRule.value = null
  ruleForm.id = ''
  ruleForm.name = ''
  ruleForm.type = ''
  ruleForm.category = ''
  ruleForm.priority = 1
  ruleForm.description = ''
  ruleForm.definition = ''
  ruleForm.enabled = true
}

onMounted(() => {
  loadRules()
})
</script>

<style scoped>
.rules {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
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

.rules-card {
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e8;
}

.rules-card :deep(.el-card__body) {
  padding: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .rules {
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
}
</style>
