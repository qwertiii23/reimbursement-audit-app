<template>
  <div class="knowledge">
    <div class="page-header">
      <div class="header-left">
        <h1>知识库</h1>
        <p>管理和查看报销相关的文档和资料</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showUploadDialog = true" v-if="isAdmin">
          <el-icon><Upload /></el-icon>
          上传文件
        </el-button>
      </div>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="分类">
          <el-select v-model="filters.category" placeholder="全部分类" clearable @change="loadFiles">
            <el-option label="政策文件" value="policy" />
            <el-option label="报销指南" value="guide" />
            <el-option label="发票模板" value="template" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadFiles">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="files-card">
      <el-table :data="files" v-loading="loading" stripe>
        <el-table-column prop="file_name" label="文件名" min-width="200">
          <template #default="{ row }">
            <div class="file-name">
              <el-icon><Document /></el-icon>
              <span>{{ row.file_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag :type="getCategoryType(row.category)" size="small">
              {{ getCategoryText(row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="file_size" label="文件大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="uploader_name" label="上传者" width="120" />
        <el-table-column prop="download_count" label="下载次数" width="100" />
        <el-table-column prop="created_at" label="上传时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="downloadFile(row)">
              <el-icon><Download /></el-icon>
              下载
            </el-button>
            <el-button
              type="primary"
              link
              size="small"
              @click="editFile(row)"
              v-if="isAdmin"
            >
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button
              type="danger"
              link
              size="small"
              @click="deleteFile(row)"
              v-if="isAdmin || row.uploaded_by === userStore.user?.id"
            >
              <el-icon><Delete /></el-icon>
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
        @size-change="loadFiles"
        @current-change="loadFiles"
        class="pagination"
      />
    </el-card>

    <el-dialog
      v-model="showUploadDialog"
      title="上传文件"
      width="600px"
      @close="resetUploadForm"
    >
      <el-form :model="uploadForm" label-width="100px">
        <el-form-item label="文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :on-change="handleFileChange"
            :limit="1"
            :on-exceed="handleExceed"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持上传 PDF、Word、Excel、图片等格式文件
              </div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="uploadForm.category" placeholder="请选择分类">
            <el-option label="政策文件" value="policy" />
            <el-option label="报销指南" value="guide" />
            <el-option label="发票模板" value="template" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="uploadForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入文件描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpload" :loading="uploading">
          上传
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showEditDialog"
      title="编辑文件"
      width="600px"
      @close="resetEditForm"
    >
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="文件名">
          <el-input v-model="editForm.file_name" placeholder="请输入文件名" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="editForm.category" placeholder="请选择分类">
            <el-option label="政策文件" value="policy" />
            <el-option label="报销指南" value="guide" />
            <el-option label="发票模板" value="template" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="editForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入文件描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleEdit" :loading="editing">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Upload, Download, Edit, Delete } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import {
  getKnowledgeFiles,
  uploadKnowledgeFile,
  updateKnowledgeFile,
  deleteKnowledgeFile,
  downloadKnowledgeFile
} from '@/api/knowledge'

const userStore = useUserStore()
const isAdmin = computed(() => userStore.isAdmin)

const files = ref([])
const loading = ref(false)
const uploading = ref(false)
const editing = ref(false)
const showUploadDialog = ref(false)
const showEditDialog = ref(false)

const filters = reactive({
  category: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const uploadForm = reactive({
  file: null,
  file_name: '',
  category: '',
  description: ''
})

const editForm = reactive({
  id: '',
  file_name: '',
  category: '',
  description: ''
})

const getCategoryText = (category) => {
  const categoryMap = {
    'policy': '政策文件',
    'guide': '报销指南',
    'template': '发票模板',
    'other': '其他'
  }
  return categoryMap[category] || category
}

const getCategoryType = (category) => {
  const typeMap = {
    'policy': 'primary',
    'guide': 'success',
    'template': 'warning',
    'other': 'info'
  }
  return typeMap[category] || 'info'
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const loadFiles = async () => {
  loading.value = true
  try {
    const response = await getKnowledgeFiles({
      category: filters.category,
      page: pagination.page,
      page_size: pagination.pageSize
    })
    files.value = response.data?.list || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '加载文件列表失败')
  } finally {
    loading.value = false
  }
}

const handleFileChange = (file) => {
  uploadForm.file = file.raw
  uploadForm.file_name = file.name
}

const handleExceed = () => {
  ElMessage.warning('只能上传一个文件')
}

const handleUpload = async () => {
  if (!uploadForm.file) {
    ElMessage.warning('请选择文件')
    return
  }
  if (!uploadForm.category) {
    ElMessage.warning('请选择分类')
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', uploadForm.file)
    formData.append('file_name', uploadForm.file_name)
    formData.append('file_type', uploadForm.file.type)
    formData.append('file_size', uploadForm.file.size)
    formData.append('category', uploadForm.category)
    formData.append('description', uploadForm.description)
    formData.append('uploaded_by', userStore.user?.id)
    formData.append('uploader_name', userStore.user?.username)

    await uploadKnowledgeFile(formData)
    ElMessage.success('上传成功')
    showUploadDialog.value = false
    resetUploadForm()
    loadFiles()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

const editFile = (row) => {
  editForm.id = row.id
  editForm.file_name = row.file_name
  editForm.category = row.category
  editForm.description = row.description
  showEditDialog.value = true
}

const handleEdit = async () => {
  if (!editForm.file_name) {
    ElMessage.warning('请输入文件名')
    return
  }

  editing.value = true
  try {
    await updateKnowledgeFile(editForm.id, {
      file_name: editForm.file_name,
      category: editForm.category,
      description: editForm.description
    })
    ElMessage.success('更新成功')
    showEditDialog.value = false
    resetEditForm()
    loadFiles()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '更新失败')
  } finally {
    editing.value = false
  }
}

const deleteFile = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该文件吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteKnowledgeFile(row.id)
    ElMessage.success('删除成功')
    loadFiles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

const downloadFile = async (row) => {
  try {
    const response = await downloadKnowledgeFile(row.id)
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.download = row.file_name
    link.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const resetUploadForm = () => {
  uploadForm.file = null
  uploadForm.file_name = ''
  uploadForm.category = ''
  uploadForm.description = ''
}

const resetEditForm = () => {
  editForm.id = ''
  editForm.file_name = ''
  editForm.category = ''
  editForm.description = ''
}

const resetFilters = () => {
  filters.category = ''
  loadFiles()
}

onMounted(() => {
  loadFiles()
})
</script>

<style scoped>
.knowledge {
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

.filter-card,
.files-card {
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e8;
  margin-bottom: 20px;
}

.filter-card :deep(.el-card__body),
.files-card :deep(.el-card__body) {
  padding: 20px;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-name .el-icon {
  color: #409eff;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.el-upload__tip {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

@media (max-width: 768px) {
  .knowledge {
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
