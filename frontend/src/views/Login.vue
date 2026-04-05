<template>
  <div class="login-container">
    <!-- 左侧品牌展示区 -->
    <div class="brand-panel">
      <!-- 背景装饰 -->
      <div class="bg-decoration">
        <div class="circle circle-1"></div>
        <div class="circle circle-2"></div>
        <div class="circle circle-3"></div>
        <div class="grid-pattern"></div>
        <div class="dot-matrix"></div>
      </div>

      <!-- 品牌内容 -->
      <div class="brand-content">
        <div class="logo-mark">
          <svg viewBox="0 0 100 100" class="logo-svg">
            <defs>
              <linearGradient id="brandGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
                <stop offset="50%" style="stop-color:#764ba2;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#f093fb;stop-opacity:1" />
              </linearGradient>
            </defs>
            <rect x="10" y="10" width="80" height="80" rx="16" fill="none" stroke="url(#brandGradient)" stroke-width="3"/>
            <path d="M30 35 L45 55 L70 28" fill="none" stroke="url(#brandGradient)" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>

        <h1 class="brand-title">
          <span class="title-line">报销审核</span>
          <span class="title-line accent">管理系统</span>
        </h1>

        <p class="brand-description">
          企业级智能财务管理平台<br/>
          <span class="highlight">高效 · 安全 · 智能</span>
        </p>

        <div class="feature-list">
          <div class="feature-item">
            <div class="feature-icon">
              <el-icon><Check /></el-icon>
            </div>
            <span>智能审核流程</span>
          </div>
          <div class="feature-item">
            <div class="feature-icon">
              <el-icon><Check /></el-icon>
            </div>
            <span>实时数据分析</span>
          </div>
          <div class="feature-item">
            <div class="feature-icon">
              <el-icon><Check /></el-icon>
            </div>
            <span>多级权限管理</span>
          </div>
        </div>

        <div class="version-badge">
          <span>v2.0</span>
        </div>
      </div>

      <!-- 底部装饰线 -->
      <div class="bottom-accent"></div>
    </div>

    <!-- 右侧登录表单区 -->
    <div class="form-panel">
      <div class="form-wrapper">
        <!-- 移动端 Logo (仅小屏显示) -->
        <div class="mobile-logo">
          <h2>报销审核系统</h2>
        </div>

        <!-- 表单头部 -->
        <div class="form-header">
          <h2 class="form-title">欢迎登录</h2>
          <p class="form-subtitle">请输入您的账号信息以继续</p>
        </div>

        <!-- 登录表单 -->
        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <label class="input-label">
              <el-icon><User /></el-icon>
              <span>用户名</span>
            </label>
            <el-input
              v-model="loginForm.username"
              placeholder="请输入用户名"
              size="large"
              clearable
              :class="{ 'is-focused': focusState.username }"
              @focus="focusState.username = true"
              @blur="focusState.username = false"
            />
          </el-form-item>

          <el-form-item prop="password">
            <label class="input-label">
              <el-icon><Lock /></el-icon>
              <span>密码</span>
            </label>
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              show-password
              clearable
              :class="{ 'is-focused': focusState.password }"
              @focus="focusState.password = true"
              @blur="focusState.password = false"
            />
          </el-form-item>

          <div class="form-options">
            <el-checkbox v-model="loginForm.remember" class="remember-checkbox">
              记住我
            </el-checkbox>
            <a href="#" class="forgot-link">忘记密码？</a>
          </div>

          <el-form-item>
            <button
              type="button"
              class="login-button"
              :class="{ 'loading': loading }"
              :disabled="loading"
              @click="handleLogin"
            >
              <span v-if="!loading" class="btn-content">
                <span>登 录</span>
                <el-icon><ArrowRight /></el-icon>
              </span>
              <span v-else class="btn-loading">
                <el-icon class="is-loading"><Loading /></el-icon>
                <span>登录中...</span>
              </span>
            </button>
          </el-form-item>
        </el-form>

        <!-- 演示账号 -->
        <div class="demo-section">
          <p class="demo-title">快速体验</p>
          <div class="demo-buttons">
            <button class="demo-btn admin" @click="fillAdminAccount">
              <el-icon><UserFilled /></el-icon>
              <span>管理员</span>
            </button>
            <button class="demo-btn user" @click="fillUserAccount">
              <el-icon><User /></el-icon>
              <span>普通用户</span>
            </button>
          </div>
        </div>

        <!-- 页脚 -->
        <div class="form-footer">
          <p>&copy; 2026 报销审核系统 · 安全可靠</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { User, Lock, Loading, ArrowRight, UserFilled, Check } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

const loginFormRef = ref(null)
const loading = ref(false)
const focusState = reactive({
  username: false,
  password: false
})

const loginForm = reactive({
  username: '',
  password: '',
  remember: false
})

const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return

  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await userStore.login(loginForm.username, loginForm.password)
        ElMessage.success('登录成功')
        router.push('/')
      } catch (error) {
        ElMessage.error(error.response?.data?.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const fillAdminAccount = () => {
  loginForm.username = 'admin'
  loginForm.password = 'admin123'
}

const fillUserAccount = () => {
  loginForm.username = 'user'
  loginForm.password = 'password123'
}
</script>

<style scoped>
/* 容器 - 左右分栏 */
.login-container {
  display: flex;
  min-height: 100vh;
  background: #ffffff;
  position: relative;
  overflow: hidden;
}

/* ========== 左侧品牌展示区 ========== */
.brand-panel {
  flex: 0 0 48%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px;
  overflow: hidden;
}

/* 背景装饰 */
.bg-decoration {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
}

.circle-1 {
  width: 400px;
  height: 400px;
  top: -100px;
  right: -100px;
  animation: floatSlow 15s ease-in-out infinite;
}

.circle-2 {
  width: 300px;
  height: 300px;
  bottom: -80px;
  left: -80px;
  animation: floatSlow 12s ease-in-out infinite reverse;
}

.circle-3 {
  width: 200px;
  height: 200px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: floatSlow 18s ease-in-out infinite;
}

@keyframes floatSlow {
  0%, 100% { transform: translate(0, 0); }
  33% { transform: translate(20px, -15px); }
  66% { transform: translate(-15px, 20px); }
}

.grid-pattern {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 40px 40px;
  opacity: 0.5;
}

.dot-matrix {
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(circle at 20% 30%, rgba(255, 255, 255, 0.15) 3px, transparent 3px),
    radial-gradient(circle at 80% 70%, rgba(255, 255, 255, 0.1) 2px, transparent 2px),
    radial-gradient(circle at 60% 20%, rgba(255, 255, 255, 0.12) 2px, transparent 2px);
  animation: twinkle 8s ease-in-out infinite;
}

@keyframes twinkle {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}

/* 品牌内容 */
.brand-content {
  position: relative;
  z-index: 10;
  max-width: 480px;
  color: #ffffff;
  animation: slideInLeft 0.8s ease-out both;
}

@keyframes slideInLeft {
  from {
    opacity: 0;
    transform: translateX(-40px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.logo-mark {
  margin-bottom: 36px;
  animation: fadeInUp 0.6s ease-out 0.2s both;
}

.logo-svg {
  width: 90px;
  height: 90px;
  filter: drop-shadow(0 8px 24px rgba(102, 126, 234, 0.4));
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.brand-title {
  font-family: 'Playfair Display', 'Georgia', serif;
  font-size: 52px;
  font-weight: 700;
  line-height: 1.2;
  margin: 0 0 24px 0;
  letter-spacing: -0.02em;
}

.title-line {
  display: block;
  text-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
}

.title-line.accent {
  font-size: 56px;
  background: linear-gradient(135deg, #ffffff 0%, rgba(255, 255, 255, 0.85) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 2px 8px rgba(255, 255, 255, 0.3));
}

.brand-description {
  font-size: 17px;
  line-height: 1.7;
  margin: 0 0 40px 0;
  opacity: 0.95;
  font-weight: 300;
  letter-spacing: 0.02em;
}

.highlight {
  display: inline-block;
  padding: 6px 16px;
  margin-top: 12px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  font-weight: 500;
  font-size: 15px;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.25);
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 40px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 15px;
  font-weight: 400;
  opacity: 0.95;
  transition: all 0.3s ease;
}

.feature-item:hover {
  transform: translateX(8px);
  opacity: 1;
}

.feature-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

.feature-item:hover .feature-icon {
  background: rgba(255, 255, 255, 0.35);
  transform: scale(1.1);
}

.version-badge {
  display: inline-flex;
  align-items: center;
  padding: 8px 20px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.08em;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.25);
}

.bottom-accent {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.5),
    rgba(255, 255, 255, 0.3),
    transparent
  );
}

/* ========== 右侧表单区 ========== */
.form-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px;
  background: #ffffff;
  position: relative;
}

.form-wrapper {
  width: 100%;
  max-width: 420px;
  animation: slideInRight 0.8s ease-out both;
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(40px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.mobile-logo {
  display: none;
}

.form-header {
  margin-bottom: 40px;
}

.form-title {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  font-size: 32px;
  font-weight: 700;
  color: #1a1a2e;
  margin: 0 0 10px 0;
  letter-spacing: -0.02em;
}

.form-subtitle {
  font-size: 15px;
  color: #6b7280;
  margin: 0;
  font-weight: 400;
}

/* 输入框标签 */
.input-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.input-label .el-icon {
  font-size: 15px;
  color: #9ca3af;
}

/* 表单样式覆盖 */
.login-form :deep(.el-input__wrapper) {
  background: #f9fafb !important;
  border: 2px solid #e5e7eb !important;
  border-radius: 12px !important;
  box-shadow: none !important;
  padding: 4px 16px !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  height: 54px !important;
}

.login-form :deep(.el-input__wrapper:hover) {
  border-color: #d1d5db !important;
  background: #ffffff !important;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  border-color: #667eea !important;
  background: #ffffff !important;
  box-shadow: 0 0 0 4px rgba(102, 126, 234, 0.1) !important;
}

.login-form :deep(.el-input__inner) {
  color: #1f2937 !important;
  font-size: 15px !important;
  font-weight: 400 !important;
  height: 44px !important;
  line-height: 44px !important;
}

.login-form :deep(.el-input__inner::placeholder) {
  color: #9ca3af !important;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 22px;
}

/* 表单选项 */
.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
}

.remember-checkbox :deep(.el-checkbox__label) {
  color: #6b7280;
  font-size: 14px;
  font-weight: 400;
}

.remember-checkbox :deep(.el-checkbox__inner) {
  background: #ffffff;
  border-color: #d1d5db;
  border-radius: 5px;
  width: 18px;
  height: 18px;
}

.remember-checkbox :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-color: transparent;
}

.forgot-link {
  color: #667eea;
  font-size: 14px;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.3s ease;
}

.forgot-link:hover {
  color: #764ba2;
  text-decoration: underline;
}

/* 登录按钮 */
.login-button {
  width: 100%;
  height: 56px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-family: inherit;
  letter-spacing: 0.04em;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.35);
}

.login-button:not(:disabled):hover {
  transform: translateY(-3px);
  box-shadow:
    0 8px 24px rgba(102, 126, 234, 0.45),
    0 0 0 1px rgba(102, 126, 234, 0.1) inset;
  background: linear-gradient(135deg, #7c94f4 0%, #8b63b5 100%);
}

.login-button:not(:disabled):active {
  transform: translateY(-1px);
}

.login-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-content,
.btn-loading {
  display: flex;
  align-items: center;
  gap: 10px;
  position: relative;
  z-index: 2;
}

.btn-loading .is-loading {
  font-size: 18px;
}

/* 演示账号区域 */
.demo-section {
  margin-top: 36px;
  padding-top: 32px;
  border-top: 1px solid #e5e7eb;
}

.demo-title {
  font-size: 13px;
  color: #9ca3af;
  font-weight: 500;
  text-align: center;
  margin: 0 0 16px 0;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.demo-buttons {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.demo-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  background: #ffffff;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  transition: all 0.3s ease;
  font-family: inherit;
}

.demo-btn:hover {
  border-color: #667eea;
  color: #667eea;
  background: rgba(102, 126, 234, 0.04);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

.demo-btn .el-icon {
  font-size: 16px;
}

/* 页脚 */
.form-footer {
  margin-top: 36px;
  text-align: center;
}

.form-footer p {
  font-size: 12px;
  color: #9ca3af;
  margin: 0;
  font-weight: 400;
}

/* ========== 响应式设计 ========== */
@media (max-width: 1024px) {
  .brand-panel {
    flex: 0 0 42%;
    padding: 48px;
  }

  .brand-title {
    font-size: 42px;
  }

  .title-line.accent {
    font-size: 46px;
  }

  .form-panel {
    padding: 48px;
  }
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
  }

  .brand-panel {
    display: none;
  }

  .mobile-logo {
    display: block;
    text-align: center;
    margin-bottom: 32px;
  }

  .mobile-logo h2 {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 32px;
    font-weight: 700;
    background: linear-gradient(135deg, #667eea, #764ba2);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    margin: 0;
  }

  .form-panel {
    padding: 32px 24px;
    justify-content: flex-start;
    padding-top: 60px;
  }

  .form-wrapper {
    max-width: 100%;
  }

  .form-title {
    font-size: 28px;
  }

  .demo-buttons {
    grid-template-columns: 1fr;
  }
}
</style>
