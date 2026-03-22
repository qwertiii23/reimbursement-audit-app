import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/Layout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'reimbursement',
        name: 'Reimbursement',
        component: () => import('@/views/Reimbursement.vue'),
        meta: { title: '报销单管理' }
      },
      {
        path: 'reimbursement/:id',
        name: 'ReimbursementDetail',
        component: () => import('@/views/ReimbursementDetail.vue'),
        meta: { title: '报销单详情' }
      },
      {
        path: 'reimbursement/:id/edit',
        name: 'ReimbursementEdit',
        component: () => import('@/views/ReimbursementEdit.vue'),
        meta: { title: '编辑报销单' }
      },
      {
        path: 'audit',
        name: 'Audit',
        component: () => import('@/views/Audit.vue'),
        meta: { title: '审核管理', requiresAdmin: true }
      },
      {
        path: 'audit/:id',
        name: 'AuditDetail',
        component: () => import('@/views/AuditDetail.vue'),
        meta: { title: '审核详情', requiresAdmin: true }
      },
      {
        path: 'audit/:id/flow',
        name: 'AuditFlow',
        component: () => import('@/views/FlowLogs.vue'),
        meta: { title: '流程日志' }
      },
      {
        path: 'reimbursement/:id/flow',
        name: 'ReimbursementFlow',
        component: () => import('@/views/FlowLogs.vue'),
        meta: { title: '流程日志' }
      },
      {
        path: 'rule-engine',
        name: 'RuleEngine',
        component: () => import('@/views/RuleEngineManagement.vue'),
        meta: { title: '规则引擎管理', requiresAdmin: true }
      },
      {
        path: 'knowledge',
        name: 'Knowledge',
        component: () => import('@/views/KnowledgeManagement.vue'),
        meta: { title: '知识库' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const isAuthenticated = userStore.isAuthenticated
  const userRole = userStore.user?.role

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresAdmin && userRole !== 'admin') {
    next('/')
  } else if (to.path === '/login' && isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router
