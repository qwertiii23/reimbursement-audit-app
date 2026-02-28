<template>
  <div class="layout-container">
    <el-container>
      <el-aside width="240px" class="sidebar">
        <div class="logo">
          <h2>报销审核系统</h2>
        </div>
        <el-menu
          :default-active="activeMenu"
          :router="true"
          class="sidebar-menu"
          background-color="#2c3e50"
          text-color="#ecf0f1"
          active-text-color="#3498db"
        >
          <el-menu-item index="/dashboard">
            <el-icon><Odometer /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          <el-menu-item index="/reimbursement">
            <el-icon><Document /></el-icon>
            <span>报销单管理</span>
          </el-menu-item>
          <el-menu-item v-if="isAdmin" index="/audit">
            <el-icon><Checked /></el-icon>
            <span>审核管理</span>
          </el-menu-item>
          <el-menu-item v-if="isAdmin" index="/rule-engine">
            <el-icon><Setting /></el-icon>
            <span>规则引擎管理</span>
          </el-menu-item>
          <el-menu-item index="/knowledge">
            <el-icon><Folder /></el-icon>
            <span>知识库</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-container>
        <el-header class="header">
          <div class="header-left">
            <div class="breadcrumb-wrapper">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
                <el-breadcrumb-item v-if="currentRoute">{{ currentRoute }}</el-breadcrumb-item>
              </el-breadcrumb>
            </div>
          </div>
          <div class="header-right">
            <el-dropdown @command="handleCommand">
              <div class="user-info">
                <el-avatar :size="35" :src="userAvatar" />
                <span class="username">{{ userStore.user?.username }}</span>
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                  <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Odometer,
  Document,
  Checked,
  Setting,
  Folder,
  ArrowDown
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)
const currentRoute = computed(() => route.meta?.title || '')
const isAdmin = computed(() => userStore.isAdmin)

const userAvatar = computed(() => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${userStore.user?.username || 'user'}`
})

const handleCommand = async (command) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      userStore.logout()
      ElMessage.success('退出成功')
      router.push('/login')
    } catch {
    }
  } else if (command === 'profile') {
    ElMessage.info('个人信息功能开发中')
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  background: #f5f7fa;
}

.sidebar {
  background: linear-gradient(180deg, #2c3e50 0%, #1a252f 100%);
  height: 100vh;
  overflow: hidden;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo h2 {
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  margin: 0;
  letter-spacing: 0.5px;
}

.sidebar-menu {
  border: none;
  height: calc(100vh - 64px);
  overflow-y: auto;
  padding: 12px 0;
}

.sidebar-menu :deep(.el-menu-item) {
  height: 48px;
  line-height: 48px;
  margin: 4px 12px;
  border-radius: 8px;
  transition: all 0.3s;
  color: rgba(255, 255, 255, 0.85);
}

.sidebar-menu :deep(.el-menu-item:hover) {
  background: rgba(52, 152, 219, 0.2) !important;
  color: #fff !important;
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: #3498db !important;
  color: #fff !important;
  font-weight: 500;
}

.sidebar-menu :deep(.el-menu-item .el-icon) {
  font-size: 18px;
  margin-right: 8px;
}

.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 32px;
  height: 64px;
}

.header-left {
  flex: 1;
}

.breadcrumb-wrapper {
  display: inline-block;
  background: #f5f7fa;
  padding: 8px 16px;
  border-radius: 8px;
}

.header-left :deep(.el-breadcrumb) {
  font-size: 14px;
}

.header-left :deep(.el-breadcrumb__item) {
  color: #909399;
}

.header-left :deep(.el-breadcrumb__inner) {
  color: #909399;
  font-weight: 400;
}

.header-left :deep(.el-breadcrumb__inner:hover) {
  color: #409eff;
}

.header-left :deep(.el-breadcrumb__separator) {
  color: #c0c4cc;
  margin: 0 8px;
}

.header-left :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #606266;
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 24px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 24px;
  transition: all 0.3s;
  background: #f5f7fa;
}

.user-info:hover {
  background: #e8e8e8;
}

.username {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.main-content {
  padding: 24px;
  overflow-y: auto;
  height: calc(100vh - 64px);
}
</style>
