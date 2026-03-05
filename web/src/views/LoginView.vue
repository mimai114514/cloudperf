<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('admin')
const password = ref('admin123')
const loading = ref(false)
const error = ref('')

async function submit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(username.value, password.value)
    router.push('/nodes')
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto mt-24 max-w-md">
    <Card>
      <h2 class="mb-4 text-xl font-semibold">CloudPerf 登录</h2>
      <form class="space-y-3" @submit.prevent="submit">
        <Input v-model="username" placeholder="用户名" />
        <Input v-model="password" type="password" placeholder="密码" />
        <p v-if="error" class="text-sm text-rose-400">{{ error }}</p>
        <Button type="submit" :disabled="loading">{{ loading ? '登录中...' : '登录' }}</Button>
      </form>
    </Card>
  </div>
</template>
