import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const userId = ref<string | null>(null)

  async function refresh() {
    try {
      const me = await api.me()
      userId.value = me.user_id
    } catch {
      userId.value = null
    }
  }

  async function login(username: string, password: string) {
    await api.login(username, password)
    await refresh()
  }

  async function logout() {
    await api.logout()
    userId.value = null
  }

  return { userId, refresh, login, logout }
})
