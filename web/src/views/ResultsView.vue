<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { api, type ResultItem } from '@/api/client'

const runId = ref('')
const from = ref('')
const to = ref('')
const items = ref<ResultItem[]>([])
const error = ref('')
const chartEl = ref<HTMLDivElement | null>(null)
let chart: echarts.ECharts | null = null

const chartOption = computed(() => {
  const labels = items.value.map((_, i) => String(i + 1)).reverse()
  const data = items.value
    .map((it) => {
      if (it.protocol === 'tcp') {
        return Number(it.metrics.receiver_mbps ?? it.metrics.sender_mbps ?? 0)
      }
      return Number(it.metrics.mbps ?? 0)
    })
    .reverse()

  return {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: labels },
    yAxis: { type: 'value', name: 'Mbps' },
    series: [{ type: 'line', smooth: true, data, areaStyle: {} }],
  }
})

async function load() {
  error.value = ''
  try {
    const q = new URLSearchParams()
    if (runId.value) q.set('run_id', runId.value)
    if (from.value) q.set('from', from.value)
    if (to.value) q.set('to', to.value)
    const query = q.toString() ? `?${q.toString()}` : ''
    items.value = (await api.listResults(query)).items
  } catch (e) {
    error.value = (e as Error).message
  }
}

const exportHref = computed(() => {
  const q = new URLSearchParams()
  if (runId.value) q.set('run_id', runId.value)
  if (from.value) q.set('from', from.value)
  if (to.value) q.set('to', to.value)
  const query = q.toString() ? `?${q.toString()}` : ''
  return api.exportResults(query)
})

onMounted(async () => {
  await load()
  await nextTick()
  if (chartEl.value) {
    chart = echarts.init(chartEl.value)
    chart.setOption(chartOption.value)
  }
})

watch(chartOption, (opt) => {
  chart?.setOption(opt)
})
</script>

<template>
  <div class="space-y-5">
    <Card>
      <h2 class="mb-3 text-lg font-semibold">历史结果筛选</h2>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <Input v-model="runId" placeholder="Run ID" />
        <Input v-model="from" placeholder="From RFC3339" />
        <Input v-model="to" placeholder="To RFC3339" />
        <div class="flex items-center gap-2">
          <Button @click="load">查询</Button>
          <a :href="exportHref" class="rounded-md bg-emerald-500 px-3 py-2 text-sm font-medium text-slate-950">导出CSV</a>
        </div>
      </div>
      <p v-if="error" class="mt-2 text-sm text-rose-400">{{ error }}</p>
    </Card>

    <Card>
      <h3 class="mb-2 text-base font-semibold">吞吐趋势</h3>
      <div ref="chartEl" class="h-80 w-full"></div>
    </Card>

    <Card>
      <h3 class="mb-2 text-base font-semibold">结果列表</h3>
      <table class="w-full text-sm">
        <thead class="text-slate-400">
          <tr>
            <th class="py-2 text-left">ID</th>
            <th class="py-2 text-left">协议</th>
            <th class="py-2 text-left">指标</th>
            <th class="py-2 text-left">时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="it in items" :key="it.id" class="border-t border-slate-800">
            <td class="py-2">{{ it.id }}</td>
            <td class="py-2">{{ it.protocol }}</td>
            <td class="py-2">{{ JSON.stringify(it.metrics) }}</td>
            <td class="py-2">{{ it.created_at }}</td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
