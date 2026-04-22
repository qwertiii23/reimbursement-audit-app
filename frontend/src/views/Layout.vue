<template>
  <div class="layout-container">
    <el-container>
      <!-- 侧边栏 - 白色简洁 -->
      <el-aside width="260px" class="sidebar">
        <div class="sidebar-content">
          <!-- Logo区域 -->
          <div class="logo-section">
            <div class="logo-wrapper">
              <div class="logo-icon-wrapper">
                <svg viewBox="0 0 40 40" class="logo-svg">
                  <defs>
                    <linearGradient id="sidebarGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                      <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
                      <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
                    </linearGradient>
                  </defs>
                  <rect x="4" y="4" width="32" height="32" rx="8" fill="none" stroke="url(#sidebarGradient)" stroke-width="2.5"/>
                  <path d="M12 16 L20 24 L28 12" fill="none" stroke="url(#sidebarGradient)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <div class="logo-text">
                <h2 class="logo-title">报销审核</h2>
                <p class="logo-subtitle">智能管理系统</p>
              </div>
            </div>
          </div>

          <!-- 导航菜单 -->
          <nav class="nav-menu">
            <div
              v-for="(item, index) in menuItems"
              :key="item.path"
              class="menu-item"
              :class="{ 'active': activeMenu === item.path, 'hidden': !item.visible }"
              @click="$router.push(item.path)"
            >
              <div class="menu-icon">
                <el-icon><component :is="item.icon" /></el-icon>
              </div>
              <span class="menu-label">{{ item.label }}</span>
              <div v-if="activeMenu === item.path" class="active-dot"></div>
            </div>
          </nav>

          <!-- 底部版本信息 -->
          <div class="sidebar-footer">
            <div class="version-badge">v2.0</div>
          </div>
        </div>
      </el-aside>

      <!-- 主内容区 -->
      <el-container class="main-container">
        <!-- 顶部导航栏 -->
        <el-header class="header">
          <div class="header-left">
            <div class="breadcrumb-wrapper">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item :to="{ path: '/' }">
                  <el-icon><HomeFilled /></el-icon>
                  <span>首页</span>
                </el-breadcrumb-item>
                <el-breadcrumb-item v-if="currentRoute">{{ currentRoute }}</el-breadcrumb-item>
              </el-breadcrumb>
            </div>
          </div>

          <div class="header-right">
            <!-- 快捷操作按钮组 -->
            <div class="action-buttons">
              <button class="action-btn" title="通知">
                <el-icon><Bell /></el-icon>
                <span class="notification-badge">3</span>
              </button>
              <button class="action-btn" title="设置">
                <el-icon><Setting /></el-icon>
              </button>
            </div>

            <!-- 用户信息下拉 -->
            <el-dropdown @command="handleCommand" trigger="click">
              <div class="user-profile">
                <div class="avatar-wrapper">
                  <el-avatar :size="36" :src="userAvatar" class="user-avatar" />
                  <div class="status-indicator online"></div>
                </div>
                <span class="username">{{ userStore.user?.username }}</span>
                <el-icon class="dropdown-arrow"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu class="dropdown-custom">
                  <div class="dropdown-header">
                    <el-avatar :size="44" :src="userAvatar" />
                    <div class="header-info">
                      <div class="header-name">{{ userStore.user?.username }}</div>
                      <div class="header-email">{{ userStore.user?.email || 'user@example.com' }}</div>
                    </div>
                  </div>
                  <el-dropdown-item command="profile">
                    <el-icon><User /></el-icon>
                    <span>个人信息</span>
                  </el-dropdown-item>
                  <el-dropdown-item command="settings">
                    <el-icon><Setting /></el-icon>
                    <span>账户设置</span>
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" divided class="logout-item">
                    <el-icon><SwitchButton /></el-icon>
                    <span>退出登录</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <!-- 主内容 -->
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
  ArrowDown,
  Bell,
  HomeFilled,
  User,
  SwitchButton
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)
const currentRoute = computed(() => route.meta?.title || '')
const isAdmin = computed(() => userStore.isAdmin)

const userAvatar = computed(() => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${userStore.user?.username || 'user'}&backgroundColor=transparent`
})

const menuItems = computed(() => [
  { path: '/dashboard', label: '仪表盘', icon: Odometer, visible: true },
  { path: '/reimbursement', label: '报销单管理', icon: Document, visible: true },
  { path: '/audit', label: '审核管理', icon: Checked, visible: isAdmin.value },
  { path: '/rule-engine', label: '规则引擎管理', icon: Setting, visible: isAdmin.value },
  { path: '/knowledge', label: '知识库', icon: Folder, visible: true }
])

const handleCommand = async (command) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？\n退出后需要重新验证身份', '安全退出', {
        confirmButtonText: '确认退出',
        cancelButtonText: '取消',
        type: 'warning',
        customClass: 'logout-dialog'
      })
      userStore.logout()
      ElMessage.success('已安全退出')
      router.push('/login')
    } catch {
      // 用户取消
    }
  } else if (command === 'profile') {
    ElMessage.info('个人信息功能开发中')
  } else if (command === 'settings') {
    ElMessage.info('账户设置功能开发中')
  }
}
</script>

<style scoped>
/* 布局容器 */
.layout-container {
  height: 100vh;
  background: var(--background);
}

/* 侧边栏 - 白色简洁 */
.sidebar {
  background: var(--surface);
  height: 100vh;
  overflow: hidden;
  border-right: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
  position: relative;
  z-index: 10;
}

.sidebar-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Logo 区域 */
.logo-section {
  padding: 24px 20px;
  border-bottom: 1px solid var(--divider);
  background: var(--surface);
}

.logo-wrapper {
  display: flex;
  align-items: center;
  gap: 14px;
}

.logo-icon-wrapper {
  width: 42px;
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-svg {
  width: 40px;
  height: 40px;
}

.logo-text {
  flex: 1;
}

.logo-title {
  font-family: var(--font-display);
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.01em;
  line-height: 1.3;
}

.logo-subtitle {
  font-size: 11px;
  color: var(--text-tertiary);
  margin: 2px 0 0 0;
  font-weight: 400;
  letter-spacing: 0.03em;
}

/* 导航菜单 */
.nav-menu {
  flex: 1;
  padding: 16px 12px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-menu::-webkit-scrollbar {
  width: 4px;
}

.nav-menu::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 2px;
}

.menu-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 16px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-normal);
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 14px;
}

.menu-item.hidden {
  display: none;
}

.menu-item:hover {
  background: var(--primary-light);
  color: var(--primary-color);
  transform: translateX(4px);
}

.menu-item.active {
  background: linear-gradient(135deg, var(--primary-light), rgba(118, 75, 162, 0.08));
  color: var(--primary-color);
  font-weight: 600;
  box-shadow: var(--shadow-sm);
}

.menu-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 19px;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.menu-item:hover .menu-icon,
.menu-item.active .menu-icon {
  background: rgba(102, 126, 234, 0.15);
  transform: scale(1.05);
}

.menu-label {
  flex: 1;
  white-space: nowrap;
}

.active-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--primary-color);
  box-shadow: 0 0 8px var(--primary-glow);
  flex-shrink: 0;
}

/* 侧边栏底部 */
.sidebar-footer {
  padding: 20px;
  border-top: 1px solid var(--divider);
  text-align: center;
}

.version-badge {
  display: inline-block;
  padding: 6px 16px;
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  font-size: 11px;
  color: var(--text-tertiary);
  font-weight: 600;
  letter-spacing: 0.05em;
}

/* 主内容区容器 */
.main-container {
  background: transparent;
}

/* 顶部导航栏 - 白色简洁 */
.header {
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 32px;
  height: 68px;
  box-shadow: var(--shadow-sm);
  position: relative;
  z-index: 5;
}

.header-left {
  flex: 1;
}

.breadcrumb-wrapper {
  display: inline-flex;
  padding: 10px 18px;
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.breadcrumb-wrapper:hover {
  border-color: var(--primary-color);
  box-shadow: var(--shadow-sm);
}

.header-left :deep(.el-breadcrumb) {
  font-size: 13px;
}

.header-left :deep(.el-breadcrumb__item) {
  display: flex;
  align-items: center;
  gap: 6px;
}

.header-left :deep(.el-breadcrumb__inner) {
  color: var(--text-secondary);
  font-weight: 400;
  transition: color var(--transition-fast);
  display: flex;
  align-items: center;
  gap: 6px;
}

.header-left :deep(.el-breadcrumb__inner:hover) {
  color: var(--primary-color);
}

.header-left :deep(.el-breadcrumb__inner .el-icon) {
  font-size: 14px;
}

.header-left :deep(.el-breadcrumb__separator) {
  color: var(--text-tertiary);
  margin: 0 8px;
}

.header-left :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: var(--text-primary);
  font-weight: 600;
}

/* 右侧操作区 */
.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-sm);
  background: var(--background);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 17px;
  position: relative;
  transition: all var(--transition-normal);
}

.action-btn:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md), 0 0 12px var(--primary-glow);
}

.notification-badge {
  position: absolute;
  top: 5px;
  right: 5px;
  width: 17px;
  height: 17px;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
  font-size: 9px;
  font-weight: 700;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--surface);
  box-shadow: 0 2px 6px rgba(239, 68, 68, 0.35);
}

/* 用户资料区 */
.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 14px;
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: all var(--transition-normal);
}

.user-profile:hover {
  border-color: var(--primary-color);
  box-shadow: var(--shadow-sm), 0 0 12px var(--primary-light);
  transform: translateY(-2px);
}

.avatar-wrapper {
  position: relative;
  flex-shrink: 0;
}

.user-avatar {
  border: 2px solid var(--border);
  transition: all var(--transition-normal);
}

.user-profile:hover .user-avatar {
  border-color: var(--primary-color);
}

.status-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid var(--surface);
}

.status-indicator.online {
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.5);
}

.username {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  max-width: 110px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dropdown-arrow {
  color: var(--text-tertiary);
  font-size: 13px;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.user-profile:hover .dropdown-arrow {
  color: var(--primary-color);
  transform: translateY(2px);
}

/* 下拉菜单样式覆盖 */
:deep(.dropdown-custom) {
  background: var(--surface) !important;
  border: 1px solid var(--border) !important;
  border-radius: var(--radius-lg) !important;
  box-shadow: var(--shadow-xl) !important;
  padding: 12px !important;
  min-width: 240px !important;
  margin-top: 8px !important;
}

:deep(.dropdown-header) {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 8px 14px;
  margin-bottom: 6px;
  border-bottom: 1px solid var(--divider);
}

:deep(.header-info) {
  flex: 1;
}

:deep(.header-name) {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 3px;
}

:deep(.header-email) {
  font-size: 12px;
  color: var(--text-tertiary);
}

:deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 400;
  transition: all var(--transition-fast);
  margin-bottom: 2px;
}

:deep(.el-dropdown-menu__item:hover) {
  background: var(--primary-light) !important;
  color: var(--primary-color);
}

:deep(.el-dropdown-menu__item .el-icon) {
  font-size: 15px;
}

:deep(.logout-item) {
  color: #ef4444 !important;
}

:deep(.logout-item:hover) {
  background: rgba(239, 68, 68, 0.08) !important;
  color: #dc2626 !important;
}

/* 主内容区 */
.main-content {
  padding: var(--space-lg);
  overflow-y: auto;
  height: calc(100vh - 68px);
  background: var(--background);
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    height: 100vh;
    z-index: 100;
    transform: translateX(-100%);
    transition: transform var(--transition-slow);
  }

  .sidebar.open {
    transform: translateX(0);
  }
}

@media (max-width: 768px) {
  .header {
    padding: 0 16px;
  }

  .main-content {
    padding: var(--space-md);
  }

  .username {
    display: none;
  }

  .action-buttons {
    gap: 6px;
  }

  .action-btn {
    width: 34px;
    height: 34px;
  }
}
</style>
