<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { CheckCircledIcon, LightningBoltIcon } from '@radix-icons/vue'
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
  <div class="flex min-h-[80vh] items-center justify-center p-4">
    <div class="w-full max-w-md animate-fade-in relative z-10">
      <!-- Glow effect behind card -->
      <div class="absolute -inset-1 rounded-2xl bg-gradient-to-br from-primary/30 to-primary/0 opcaity-20 blur-xl"></div>
      
      <Card class="border-white/10 p-8 shadow-2xl backdrop-blur-3xl bg-background/50">
        <div class="mb-8 flex flex-col items-center">
          <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/20 text-primary shadow-[0_0_20px_rgba(var(--primary),0.3)] ring-1 ring-white/10">
            <LightningBoltIcon class="h-7 w-7" />
          </div>
          <h2 class="text-2xl font-bold tracking-tight text-foreground">Welcome Back</h2>
          <p class="mt-2 text-sm text-muted-foreground">Sign in to your CloudPerf account</p>
        </div>

        <form class="space-y-5" @submit.prevent="submit">
          <div class="space-y-1">
            <label class="text-xs font-medium text-muted-foreground ml-1">Username</label>
            <Input v-model="username" placeholder="Enter your username" class="h-11" />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-muted-foreground ml-1">Password</label>
            <Input v-model="password" type="password" placeholder="••••••••" class="h-11" />
          </div>
          
          <div v-if="error" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive border border-destructive/20 flex items-center gap-2">
            {{ error }}
          </div>
          
          <Button type="submit" class="w-full h-11 text-base shadow-[0_0_15px_rgba(var(--primary),0.4)]" :disabled="loading">
            <span v-if="loading" class="flex items-center gap-2">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent"></span>
              Signing in...
            </span>
            <span v-else class="flex items-center justify-center gap-2">
              Sign In <CheckCircledIcon v-if="!loading" class="w-4 h-4 ml-1 opacity-70"/>
            </span>
          </Button>
        </form>
      </Card>
    </div>
  </div>
</template>
