<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

const route = useRoute()
const isLogin = computed(() => route.path === '/login')
</script>

<template>
  <div class="min-h-screen text-slate-100">
    <header v-if="!isLogin" class="border-b border-slate-800 bg-slate-900/70 backdrop-blur">
      <div class="mx-auto flex max-w-7xl items-center gap-4 px-6 py-4">
        <h1 class="text-lg font-semibold">CloudPerf</h1>
        <nav class="flex gap-2 text-sm">
          <RouterLink class="rounded bg-slate-800 px-3 py-1 hover:bg-slate-700" to="/nodes">节点</RouterLink>
          <RouterLink class="rounded bg-slate-800 px-3 py-1 hover:bg-slate-700" to="/runs/new">即时测速</RouterLink>
          <RouterLink class="rounded bg-slate-800 px-3 py-1 hover:bg-slate-700" to="/results">历史结果</RouterLink>
        </nav>
      </div>
    </header>
    <main class="mx-auto max-w-7xl px-6 py-6">
      <RouterView />
    </main>
  </div>
</template>
