import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isFinance = computed(() => user.value?.role === 'finance')
  const canManualAudit = computed(() => user.value?.role === 'admin' || user.value?.role === 'finance')

  const login = async (username, password) => {
    try {
      const response = await loginApi(username, password)
      token.value = response.data.token
      user.value = {
        id: response.data.user.id,
        username: response.data.user.username,
        real_name: response.data.user.real_name,
        role: response.data.user.role
      }
      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))
      return response
    } catch (error) {
      throw error
    }
  }

  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    token,
    user,
    isAuthenticated,
    isAdmin,
    isFinance,
    canManualAudit,
    login,
    logout
  }
})
