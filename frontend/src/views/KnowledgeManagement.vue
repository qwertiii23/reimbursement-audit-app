<template>
  <div class="knowledge-management">
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="文件名称">
          <el-input v-model="filterForm.fileName" placeholder="请输入文件名称" clearable style="width: 200px;" />
        </el-form-item>
        <el-form-item label="文件分类">
          <el-select v-model="filterForm.category" placeholder="请选择文件分类" clearable style="width: 150px;">
            <el-option label="报销政策" value="policy" />
            <el-option label="财务制度" value="finance" />
            <el-option label="操作手册" value="manual" />
            <el-option label="培训资料" value="training" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterForm.status" placeholder="请选择状态" clearable style="width: 120px;">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">
            搜索
          </el-button>
          <el-button v-if="isAdmin" type="success" :icon="Upload" @click="handleUpload" style="margin-left: 10px;">
            上传文件
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="fileList" v-loading="loading" stripe>
        <el-table-column prop="file_name" label="文件名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag :type="getCategoryTagType(row.category)">
              {{ getCategoryText(row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="file_type" label="文件类型" width="100" />
        <el-table-column prop="file_size" label="文件大小" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="uploader_name" label="上传者" width="120" />
        <el-table-column prop="download_count" label="下载次数" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="上传时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">
              查看
            </el-button>
            <el-button type="primary" link size="small" @click="handleDownload(row)">
              下载
            </el-button>
            <el-button v-if="isAdmin" type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button 
              v-if="isAdmin" 
              :type="row.status === 'active' ? 'warning' : 'success'" 
              link 
              size="small" 
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
            <el-button v-if="isAdmin" type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="uploadDialogVisible"
      title="上传文件"
      width="600px"
      @close="handleUploadDialogClose"
    >
      <el-form
        ref="uploadFormRef"
        :model="uploadForm"
        :rules="uploadFormRules"
        label-width="120px"
      >
        <el-form-item label="文件" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-exceed="handleExceed"
            drag
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支持 PDF、Word、Excel 等格式，文件大小不超过 50MB
              </div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="文件分类" prop="category">
          <el-select v-model="uploadForm.category" placeholder="请选择文件分类" style="width: 100%">
            <el-option label="报销政策" value="policy" />
            <el-option label="财务制度" value="finance" />
            <el-option label="操作手册" value="manual" />
            <el-option label="培训资料" value="training" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="uploadForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入文件描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUploadSubmit" :loading="uploading">
          上传
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="editDialogVisible"
      title="编辑文件"
      width="600px"
      @close="handleEditDialogClose"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editFormRules"
        label-width="120px"
      >
        <el-form-item label="文件名称">
          <el-input v-model="editForm.file_name" disabled />
        </el-form-item>
        <el-form-item label="文件分类" prop="category">
          <el-select v-model="editForm.category" placeholder="请选择文件分类" style="width: 100%">
            <el-option label="报销政策" value="policy" />
            <el-option label="财务制度" value="finance" />
            <el-option label="操作手册" value="manual" />
            <el-option label="培训资料" value="training" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="editForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入文件描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="editForm.status" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditSubmit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="viewDialogVisible"
      title="查看文档内容"
      width="80%"
      top="5vh"
      @close="handleViewDialogClose"
    >
      <div class="file-info">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="文件名称">{{ viewForm.file_name }}</el-descriptions-item>
          <el-descriptions-item label="文件分类">
            <el-tag :type="getCategoryTagType(viewForm.category)">
              {{ getCategoryText(viewForm.category) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="文件类型">{{ viewForm.file_type }}</el-descriptions-item>
          <el-descriptions-item label="文件大小">{{ formatFileSize(viewForm.file_size) }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <div class="file-content">
        <el-scrollbar height="60vh">
          <pre class="content-text">{{ viewForm.content }}</pre>
        </el-scrollbar>
      </div>
      <template #footer>
        <el-button @click="viewDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleDownloadFromView">下载</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Upload, UploadFilled } from '@element-plus/icons-vue'
import { getKnowledgeFiles, uploadKnowledgeFile, updateKnowledgeFile, deleteKnowledgeFile, downloadKnowledgeFile, viewKnowledgeFile } from '@/api/knowledge'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isAdmin = computed(() => userStore.isAdmin)

const loading = ref(false)
const fileList = ref([])
const filterForm = reactive({
  fileName: '',
  category: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const uploadDialogVisible = ref(false)
const uploadFormRef = ref(null)
const uploadRef = ref(null)
const uploading = ref(false)
const uploadForm = reactive({
  file: null,
  category: '',
  description: ''
})

const uploadFormRules = {
  category: [{ required: true, message: '请选择文件分类', trigger: 'change' }],
  description: [{ required: true, message: '请输入文件描述', trigger: 'blur' }]
}

const editDialogVisible = ref(false)
const editFormRef = ref(null)
const editForm = reactive({
  id: '',
  file_name: '',
  category: '',
  description: '',
  status: true
})

const editFormRules = {
  category: [{ required: true, message: '请选择文件分类', trigger: 'change' }],
  description: [{ required: true, message: '请输入文件描述', trigger: 'blur' }]
}

const viewDialogVisible = ref(false)
const viewForm = reactive({
  id: '',
  file_name: '',
  file_type: '',
  file_size: 0,
  category: '',
  description: '',
  content: ''
})

const fetchFiles = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      category: filterForm.category,
      status: filterForm.status
    }
    const res = await getKnowledgeFiles(params)
    if (res.code === 200) {
      fileList.value = res.data.list || []
      pagination.total = res.data.total || 0
    } else {
      ElMessage.error(res.message || '获取文件列表失败')
    }
  } catch (error) {
    ElMessage.error('获取文件列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchFiles()
}

const handleUpload = () => {
  uploadDialogVisible.value = true
}

const handleFileChange = (file) => {
  uploadForm.file = file.raw
}

const handleExceed = () => {
  ElMessage.warning('最多只能上传一个文件')
}

const handleUploadSubmit = async () => {
  if (!uploadFormRef.value) return
  await uploadFormRef.value.validate(async (valid) => {
    if (valid) {
      if (!uploadForm.file) {
        ElMessage.warning('请选择要上传的文件')
        return
      }

      uploading.value = true
      try {
        const formData = new FormData()
        formData.append('file', uploadForm.file)
        formData.append('category', uploadForm.category)
        formData.append('description', uploadForm.description)
        formData.append('uploaded_by', localStorage.getItem('userId') || '')
        formData.append('uploader_name', localStorage.getItem('userName') || '')

        const res = await uploadKnowledgeFile(formData)
        if (res.code === 200) {
          ElMessage.success('上传成功')
          uploadDialogVisible.value = false
          fetchFiles()
        } else {
          ElMessage.error(res.message || '上传失败')
        }
      } catch (error) {
        ElMessage.error('上传失败')
      } finally {
        uploading.value = false
      }
    }
  })
}

const handleUploadDialogClose = () => {
  if (uploadFormRef.value) {
    uploadFormRef.value.resetFields()
  }
  uploadForm.file = null
  if (uploadRef.value) {
    uploadRef.value.clearFiles()
  }
}

const handleEdit = (row) => {
  Object.assign(editForm, {
    id: row.id,
    file_name: row.file_name,
    category: row.category,
    description: row.description,
    status: row.status === 'active'
  })
  editDialogVisible.value = true
}

const handleEditSubmit = async () => {
  if (!editFormRef.value) return
  await editFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const data = {
          id: editForm.id,
          category: editForm.category,
          description: editForm.description,
          status: editForm.status ? 'active' : 'inactive'
        }
        const res = await updateKnowledgeFile(editForm.id, data)
        if (res.code === 200) {
          ElMessage.success('更新成功')
          editDialogVisible.value = false
          fetchFiles()
        } else {
          ElMessage.error(res.message || '更新失败')
        }
      } catch (error) {
        ElMessage.error('更新失败')
      }
    }
  })
}

const handleEditDialogClose = () => {
  if (editFormRef.value) {
    editFormRef.value.resetFields()
  }
}

const handleView = async (row) => {
  try {
    const res = await viewKnowledgeFile(row.id)
    if (res.code === 200) {
      Object.assign(viewForm, {
        id: res.data.id,
        file_name: res.data.file_name,
        file_type: res.data.file_type,
        file_size: res.data.file_size,
        category: res.data.category,
        description: res.data.description,
        content: res.data.content
      })
      viewDialogVisible.value = true
    } else {
      ElMessage.error(res.message || '获取文件内容失败')
    }
  } catch (error) {
    ElMessage.error('获取文件内容失败')
  }
}

const handleViewDialogClose = () => {
  viewDialogVisible.value = false
  Object.assign(viewForm, {
    id: '',
    file_name: '',
    file_type: '',
    file_size: 0,
    category: '',
    description: '',
    content: ''
  })
}

const handleDownloadFromView = () => {
  handleDownload(viewForm)
}

const handleToggleStatus = async (row) => {
  const newStatus = row.status === 'active' ? 'inactive' : 'active'
  const actionText = newStatus === 'active' ? '启用' : '禁用'
  
  ElMessageBox.confirm(`确定要${actionText}文件"${row.file_name}"吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await updateKnowledgeFile(row.id, { status: newStatus })
      if (res.code === 200) {
        ElMessage.success(`${actionText}成功`)
        fetchFiles()
      } else {
        ElMessage.error(res.message || `${actionText}失败`)
      }
    } catch (error) {
      ElMessage.error(`${actionText}失败`)
    }
  }).catch(() => {})
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除文件"${row.file_name}"吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteKnowledgeFile(row.id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        fetchFiles()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const handleDownload = async (row) => {
  try {
    const res = await downloadKnowledgeFile(row.id)
    const blob = new Blob([res])
    const url = window.URL.createObjectURL(blob)
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

const handleSizeChange = (val) => {
  pagination.pageSize = val
  pagination.page = 1
  fetchFiles()
}

const handleCurrentChange = (val) => {
  pagination.page = val
  fetchFiles()
}

const getCategoryTagType = (category) => {
  const categoryMap = {
    policy: 'danger',
    finance: 'warning',
    manual: 'success',
    training: 'info',
    other: ''
  }
  return categoryMap[category] || ''
}

const getCategoryText = (category) => {
  const categoryMap = {
    policy: '报销政策',
    finance: '财务制度',
    manual: '操作手册',
    training: '培训资料',
    other: '其他'
  }
  return categoryMap[category] || category
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchFiles()
})
</script>

<style scoped>
.knowledge-management {
  padding: 20px;
}

.filter-card,
.table-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 20px;
}

.filter-card :deep(.el-card__body) {
  padding: 24px;
}

.filter-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.filter-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #2c3e50;
}

.filter-form :deep(.el-form-item) {
  margin-bottom: 0;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table th) {
  background-color: #f5f7fa;
  font-weight: 600;
  color: #2c3e50;
}

:deep(.el-upload-dragger) {
  padding: 40px;
}

:deep(.el-icon--upload) {
  font-size: 67px;
  color: #409eff;
  margin-bottom: 16px;
}

:deep(.el-dialog__body) {
  padding: 24px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

.file-info {
  margin-bottom: 20px;
}

.file-content {
  background-color: #f5f7fa;
  border-radius: 8px;
  padding: 20px;
}

.content-text {
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.8;
  color: #2c3e50;
  margin: 0;
}
</style>
