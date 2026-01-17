<template>
  <div class="reimbursement-edit">
    <div class="page-header">
      <el-button @click="$router.back()" :icon="ArrowLeft">返回</el-button>
      <h1>编辑报销单</h1>
    </div>

    <el-card class="form-card">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        v-loading="loading"
      >
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入报销单标题" />
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number
            v-model="form.amount"
            :min="0"
            :precision="2"
            :step="100"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="报销日期" prop="expense_date">
          <el-date-picker
            v-model="form.expense_date"
            type="date"
            placeholder="选择报销日期"
            style="width: 100%"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="6"
            placeholder="请输入报销描述"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            保存
          </el-button>
          <el-button @click="$router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getReimbursementById, updateReimbursement } from '@/api/reimbursement'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const submitting = ref(false)
const formRef = ref(null)

const form = reactive({
  title: '',
  amount: 0,
  expense_date: '',
  description: ''
})

const rules = {
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' }
  ],
  amount: [
    { required: true, message: '请输入金额', trigger: 'blur' }
  ],
  expense_date: [
    { required: true, message: '请选择报销日期', trigger: 'change' }
  ],
  description: [
    { required: true, message: '请输入描述', trigger: 'blur' }
  ]
}

const loadReimbursement = async () => {
  loading.value = true
  try {
    const response = await getReimbursementById(route.params.id)
    form.title = response.data.title
    form.amount = response.data.amount
    form.expense_date = response.data.expense_date
    form.description = response.data.description
  } catch (error) {
    ElMessage.error('加载报销单详情失败')
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        await updateReimbursement(route.params.id, form)
        ElMessage.success('保存成功')
        router.back()
      } catch (error) {
        ElMessage.error(error.response?.data?.message || '保存失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

onMounted(() => {
  loadReimbursement()
})
</script>

<style scoped>
.reimbursement-edit {
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

.form-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  max-width: 800px;
}
</style>
