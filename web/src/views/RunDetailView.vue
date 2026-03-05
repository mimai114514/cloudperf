<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import Card from '@/components/ui/Card.vue'
import { api, type RunInfo, type RunPair } from '@/api/client'

const route = useRoute()
const run = ref<RunInfo | null>(null)
const pairs = ref<RunPair[]>([])
const error = ref('')
let timer: number | null = null

async function load() {
  try {
    const id = String(route.params.id)
    run.value = await api.getRun(id)
    pairs.value = (await api.getRunPairs(id)).items
  } catch (e) {
    error.value = (e as Error).message
  }
}

onMounted(async () => {
  await load()
  timer = window.setInterval(load, 3000)
})

onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-5">
    <Card>
      <h2 class="text-lg font-semibold">任务详情</h2>
      <p v-if="error" class="mt-2 text-sm text-rose-400">{{ error }}</p>
      <div v-if="run" class="mt-2 grid grid-cols-2 gap-3 text-sm md:grid-cols-4">
        <div>Run ID：{{ run.id }}</div>
        <div>状态：{{ run.status }}</div>
        <div>协议：{{ run.protocol }}</div>
        <div>模式：{{ run.mode }}</div>
      </div>
    </Card>

    <Card>
      <h3 class="mb-2 text-base font-semibold">Pair 进度</h3>
      <table class="w-full text-sm">
        <thead class="text-slate-400">
          <tr>
            <th class="py-2 text-left">source</th>
            <th class="py-2 text-left">target</th>
            <th class="py-2 text-left">status</th>
            <th class="py-2 text-left">error</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in pairs" :key="p.id" class="border-t border-slate-800">
            <td class="py-2">{{ p.source_node_id }}</td>
            <td class="py-2">{{ p.target_node_id }}</td>
            <td class="py-2">{{ p.status }}</td>
            <td class="py-2">{{ p.error_message || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
