<template>
  <div class="rules">
    <div class="page-header">
      <h1>规则管理</h1>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        新建规则
      </el-button>
    </div>

    <el-card class="rules-card">
      <el-table :data="rules" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="规则名称" width="200" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="rule_type" label="规则类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getRuleTypeColor(row.rule_type)">
              {{ getRuleTypeText(row.rule_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="toggleRule(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="primary" link @click="editRule(row)">
              编辑
            </el-button>
            <el-button type="danger" link @click="deleteRule(row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="showCreateDialog"
      :title="editingRule ? '编辑规则' : '新建规则'"
      width="600px"
    >
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="ruleForm.rule_type" placeholder="请选择规则类型">
            <el-option label="金额限制" value="amount_limit" />
            <el-option label="日期限制" value="date_limit" />
            <el-option label="类别限制" value="category_limit" />
            <el-option label="自定义" value="custom" />
          </el-select>
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
            v-model="ruleForm.config"
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

const rules = ref([])
const loading = ref(false)
const saving = ref(false)
const showCreateDialog = ref(false)
const editingRule = ref(null)

const ruleForm = reactive({
  name: '',
  rule_type: '',
  description: '',
  config: '',
  enabled: true
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
    ElMessage.info('规则管理功能开发中...')
  } catch (error) {
    ElMessage.error('加载规则失败')
  } finally {
    loading.value = false
  }
}

const editRule = (rule) => {
  editingRule.value = rule
  ruleForm.name = rule.name
  ruleForm.rule_type = rule.rule_type
  ruleForm.description = rule.description
  ruleForm.config = rule.config
  ruleForm.enabled = rule.enabled
  showCreateDialog.value = true
}

const saveRule = async () => {
  if (!ruleForm.name || !ruleForm.rule_type) {
    ElMessage.warning('请填写完整信息')
    return
  }

  try {
    JSON.parse(ruleForm.config)
  } catch (error) {
    ElMessage.warning('规则配置必须是有效的JSON格式')
    return
  }

  saving.value = true
  try {
    ElMessage.success(editingRule.value ? '更新成功' : '创建成功')
    showCreateDialog.value = false
    resetForm()
    loadRules()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const toggleRule = async (rule) => {
  try {
    ElMessage.success(rule.enabled ? '规则已启用' : '规则已禁用')
  } catch (error) {
    ElMessage.error('操作失败')
    rule.enabled = !rule.enabled
  }
}

const deleteRule = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    ElMessage.success('删除成功')
    loadRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const resetForm = () => {
  editingRule.value = null
  ruleForm.name = ''
  ruleForm.rule_type = ''
  ruleForm.description = ''
  ruleForm.config = ''
  ruleForm.enabled = true
}

onMounted(() => {
  loadRules()
})
</script>

<style scoped>
.rules {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: #2c3e50;
  margin: 0;
}

.rules-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}
</style>
