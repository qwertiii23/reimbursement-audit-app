<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup>
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Playfair+Display:wght@400;500;600;700&family=Inter:wght@300;400;500;600;700&display=swap');

:root {
  /* 主色调 - 紫蓝渐变 */
  --primary-color: #667eea;
  --primary-light: rgba(102, 126, 234, 0.1);
  --primary-dark: #5a67d8;
  --primary-glow: rgba(102, 126, 234, 0.35);

  /* 辅助色 - 紫罗兰 */
  --secondary-color: #764ba2;
  --secondary-light: rgba(118, 75, 162, 0.1);
  --secondary-glow: rgba(118, 75, 162, 0.3);

  /* 点缀色 - 粉色 */
  --accent-color: #f093fb;
  --accent-light: rgba(240, 147, 251, 0.1);

  /* 背景色 - 清新白色 */
  --background: #f8fafc;
  --background-deep: #ffffff;
  --surface: #ffffff;
  --surface-hover: #f1f5f9;
  --glass: rgba(255, 255, 255, 0.8);
  --glass-border: rgba(0, 0, 0, 0.08);

  /* 文字色 - 深灰层次 */
  --text-primary: #1a202c;
  --text-secondary: #4a5568;
  --text-tertiary: #a0aec0;
  --text-gold: #667eea;

  /* 边框和分割线 */
  --border: #e2e8f0;
  --divider: #edf2f7;

  /* 阴影 - 柔和轻盈 */
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.07), 0 2px 4px rgba(0, 0, 0, 0.05);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.08), 0 4px 6px rgba(0, 0, 0, 0.05);
  --shadow-xl: 0 20px 25px rgba(0, 0, 0, 0.1), 0 10px 10px rgba(0, 0, 0, 0.04);
  --shadow-glow: 0 4px 20px var(--primary-glow);
  --shadow-cyan: 0 4px 20px var(--secondary-glow);

  /* 圆角 - 现代圆润 */
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-xl: 24px;
  --radius-full: 9999px;

  /* 间距 - 舒适宽松 */
  --space-xs: 6px;
  --space-sm: 12px;
  --space-md: 20px;
  --space-lg: 32px;
  --space-xl: 48px;
  --space-2xl: 64px;

  /* 字体 - 现代优雅 */
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-display: 'Playfair Display', Georgia, serif;

  /* 过渡 - 流畅自然 */
  --transition-fast: 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-normal: 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-slow: 0.5s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-bounce: 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: var(--font-sans);
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);
  background: var(--background);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  overflow-x: hidden;
}

#app {
  min-height: 100vh;
  background: var(--background);
}

/* 滚动条样式 - 浅色透明 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
  border-radius: var(--radius-sm);
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, var(--primary-color), var(--secondary-color));
  border-radius: var(--radius-sm);
  border: 2px solid var(--background);
  transition: all var(--transition-fast);
}

::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg, var(--primary-dark), var(--secondary-color));
  box-shadow: var(--shadow-glow);
}

/* 通用动画 */
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

@keyframes fadeInScale {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes shimmer {
  0% { background-position: -200% center; }
  100% { background-position: 200% center; }
}

.fade-in {
  animation: fadeInUp var(--transition-normal) ease-out;
}

/* 通用卡片样式 - 白色简洁 */
.card {
  background: var(--surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
  padding: var(--space-lg);
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg,
    var(--primary-color),
    var(--secondary-color),
    var(--accent-color)
  );
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-4px);
  border-color: var(--primary-color);
}

.card:hover::before {
  opacity: 1;
}

/* 通用按钮样式 - 渐变质感 */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-md);
  font-weight: 500;
  font-family: var(--font-sans);
  transition: all var(--transition-normal);
  border: none;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.15), transparent);
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.btn:hover::before {
  opacity: 1;
}

.btn-primary {
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: white;
  box-shadow: var(--shadow-md), 0 2px 8px var(--primary-glow);
  font-weight: 600;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg), var(--shadow-glow);
  background: linear-gradient(135deg, #7c94f4, #8b63b5);
}

.btn-secondary {
  background: var(--surface);
  color: var(--text-primary);
  border: 2px solid var(--border);
}

.btn-secondary:hover {
  background: var(--surface-hover);
  border-color: var(--primary-color);
  color: var(--primary-color);
  box-shadow: var(--shadow-md), 0 0 12px var(--primary-light);
  transform: translateY(-2px);
}

/* 通用标题样式 - 优雅衬线 */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-display);
  font-weight: 600;
  line-height: 1.2;
  margin-bottom: var(--space-md);
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

h1 {
  font-size: 36px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--text-primary) 0%, var(--primary-color) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

h2 {
  font-size: 28px;
  font-weight: 600;
}

h3 {
  font-size: 22px;
  font-weight: 600;
}

/* 通用文本样式 */
.text-primary { color: var(--text-primary); }
.text-secondary { color: var(--text-secondary); }
.text-tertiary { color: var(--text-tertiary); }
.text-gold { color: var(--text-gold); }

/* 通用间距类 */
.mt-1 { margin-top: var(--space-xs); }
.mt-2 { margin-top: var(--space-sm); }
.mt-3 { margin-top: var(--space-md); }
.mt-4 { margin-top: var(--space-lg); }
.mt-5 { margin-top: var(--space-xl); }

.mb-1 { margin-bottom: var(--space-xs); }
.mb-2 { margin-bottom: var(--space-sm); }
.mb-3 { margin-bottom: var(--space-md); }
.mb-4 { margin-bottom: var(--space-lg); }
.mb-5 { margin-bottom: var(--space-xl); }

.ml-1 { margin-left: var(--space-xs); }
.ml-2 { margin-left: var(--space-sm); }
.ml-3 { margin-left: var(--space-md); }
.ml-4 { margin-left: var(--space-lg); }

.mr-1 { margin-right: var(--space-xs); }
.mr-2 { margin-right: var(--space-sm); }
.mr-3 { margin-right: var(--space-md); }
.mr-4 { margin-right: var(--space-lg); }

.p-1 { padding: var(--space-xs); }
.p-2 { padding: var(--space-sm); }
.p-3 { padding: var(--space-md); }
.p-4 { padding: var(--space-lg); }
.p-5 { padding: var(--space-xl); }

/* Element Plus 全局覆盖 - 白色主题 */
.el-button--primary {
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color)) !important;
  border: none !important;
  box-shadow: 0 2px 8px var(--primary-glow) !important;
}

.el-button--primary:hover {
  background: linear-gradient(135deg, #7c94f4, #8b63b5) !important;
  box-shadow: var(--shadow-glow) !important;
  transform: translateY(-2px);
}

.el-input__wrapper {
  background: var(--surface) !important;
  border: 2px solid var(--border) !important;
  box-shadow: none !important;
  transition: all var(--transition-normal) !important;
  border-radius: 10px !important;
}

.el-input__wrapper:hover {
  border-color: #cbd5e0 !important;
}

.el-input__wrapper.is-focus {
  border-color: var(--primary-color) !important;
  box-shadow: 0 0 0 4px var(--primary-light) !important;
}

.el-card {
  background: var(--surface) !important;
  border: 1px solid var(--border) !important;
  border-radius: var(--radius-lg) !important;
  box-shadow: var(--shadow-sm) !important;
}

.el-table {
  background: transparent !important;
  color: var(--text-secondary) !important;
}

.el-table th.el-table__cell {
  background: var(--background) !important;
  color: var(--text-primary) !important;
  border-bottom: 2px solid var(--border) !important;
  font-weight: 600 !important;
}

.el-table td.el-table__cell {
  border-bottom: 1px solid var(--divider) !important;
}

.el-table tr:hover > td.el-table__cell {
  background: var(--primary-light) !important;
}

.el-tag {
  border: none !important;
  border-radius: 6px !important;
}
</style>
