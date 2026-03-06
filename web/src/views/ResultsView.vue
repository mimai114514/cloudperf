<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import { MagnifyingGlassIcon, DownloadIcon, BarChartIcon, LoopIcon } from '@radix-icons/vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { api, type ResultItem } from '@/api/client'

const runId = ref('')
const from = ref('')
const to = ref('')
const items = ref<ResultItem[]>([])
const error = ref('')
const loading = ref(false)
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
    tooltip: { 
      trigger: 'axis',
      backgroundColor: 'rgba(15, 23, 42, 0.9)',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      textStyle: { color: '#f8fafc' },
      padding: [8, 12],
    },
    grid: { top: 40, right: 20, bottom: 40, left: 60 },
    xAxis: { 
      type: 'category', 
      data: labels,
      axisLabel: { color: 'rgba(255,255,255,0.5)' },
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.1)' } },
      splitLine: { show: false }
    },
    yAxis: { 
      type: 'value', 
      name: 'Mbps',
      nameTextStyle: { color: 'rgba(255,255,255,0.5)', padding: [0, 0, 0, 10] },
      axisLabel: { color: 'rgba(255,255,255,0.5)' },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.05)', type: 'dashed' } }
    },
    series: [{ 
      type: 'line', 
      smooth: true, 
      data,
      symbolSize: 6,
      itemStyle: { color: '#3b82f6', shadowBlur: 10, shadowColor: '#3b82f6' },
      lineStyle: { width: 3, shadowBlur: 10, shadowColor: 'rgba(59, 130, 246, 0.5)' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(59, 130, 246, 0.5)' },
          { offset: 1, color: 'rgba(59, 130, 246, 0.01)' }
        ])
      }
    }],
  }
})

async function load() {
  error.value = ''
  loading.value = true
  try {
    const q = new URLSearchParams()
    if (runId.value) q.set('run_id', runId.value)
    if (from.value) q.set('from', from.value)
    if (to.value) q.set('to', to.value)
    const query = q.toString() ? `?${q.toString()}` : ''
    items.value = (await api.listResults(query)).items
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
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
    
    // Resize observer to make chart responsive
    const ro = new ResizeObserver(() => chart?.resize())
    ro.observe(chartEl.value)
  }
})

watch(chartOption, (opt) => {
  chart?.setOption(opt)
})
</script>

<template>
  <div class="space-y-6 animate-fade-in relative z-10">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Benchmark Results</h2>
        <p class="text-sm text-muted-foreground mt-1">Analyze and export historical test records.</p>
      </div>
    </div>

    <Card class="border-primary/20 bg-primary/5">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-12 md:items-end">
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">Run ID</label>
          <Input v-model="runId" placeholder="e.g. run_xxxxx" />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">From Date (RFC3339)</label>
          <Input v-model="from" placeholder="YYYY-MM-DDTHH:mm:ssZ" />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">To Date (RFC3339)</label>
          <Input v-model="to" placeholder="YYYY-MM-DDTHH:mm:ssZ" />
        </div>
        <div class="flex items-center gap-3 md:col-span-3">
          <Button @click="load" :disabled="loading" class="flex-1 shadow-[0_0_15px_rgba(var(--primary),0.3)]">
             <LoopIcon v-if="loading" class="w-4 h-4 mr-2 animate-spin" />
             <MagnifyingGlassIcon v-else class="w-4 h-4 mr-2" />
             Query
          </Button>
          <a :href="exportHref" class="inline-flex items-center justify-center rounded-md bg-emerald-500/20 px-4 py-2 text-sm font-medium text-emerald-400 border border-emerald-500/30 transition-colors hover:bg-emerald-500/30">
            <DownloadIcon class="w-4 h-4 mr-2" />
            CSV
          </a>
        </div>
      </div>
      <p v-if="error" class="mt-3 text-sm text-destructive font-medium bg-destructive/10 px-3 py-2 rounded-md">{{ error }}</p>
    </Card>

    <div class="grid gap-6 md:grid-cols-3">
       <!-- Chart Card -->
      <Card class="md:col-span-3 p-0 overflow-hidden">
        <div class="p-5 border-b border-white/5 flex items-center gap-2 bg-white/[0.02]">
           <BarChartIcon class="w-5 h-5 text-primary" />
           <h3 class="text-lg font-semibold">Throughput Trend</h3>
        </div>
        <div class="p-4 bg-gradient-to-b from-primary/5 to-transparent">
          <!-- Show placeholder if no items -->
          <div v-show="items.length === 0" class="h-80 w-full flex items-center justify-center text-muted-foreground">
             No chart data available for the given criteria.
          </div>
          <div ref="chartEl" v-show="items.length > 0" class="h-80 w-full"></div>
        </div>
      </Card>
      
      <!-- Table Card -->
      <Card class="md:col-span-3 p-0 overflow-hidden">
        <div class="p-5 border-b border-white/5 bg-white/[0.02]">
          <h3 class="text-lg font-semibold">Raw Data</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm text-left">
            <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
              <tr>
                <th class="px-6 py-4 font-medium tracking-wider">Record ID</th>
                <th class="px-6 py-4 font-medium tracking-wider">Protocol</th>
                <th class="px-6 py-4 font-medium tracking-wider">Metrics Detail</th>
                <th class="px-6 py-4 font-medium tracking-wider text-right">Timestamp</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/5 bg-transparent">
              <tr v-if="loading && items.length === 0">
                 <td colspan="4" class="px-6 py-8 text-center text-muted-foreground">Fetching records...</td>
              </tr>
              <tr v-else-if="items.length === 0">
                 <td colspan="4" class="px-6 py-8 text-center text-muted-foreground">No records found.</td>
              </tr>
              <tr v-for="it in items" :key="it.id" class="transition-colors hover:bg-white/[0.02]">
                <td class="px-6 py-4 font-mono text-xs text-muted-foreground">{{ it.id }}</td>
                <td class="px-6 py-4">
                   <span class="px-2 py-0.5 rounded text-xs font-semibold uppercase bg-white/10 text-white">{{ it.protocol }}</span>
                </td>
                <td class="px-6 py-4 text-xs font-mono text-sky-300 max-w-[400px] truncate break-all">{{ JSON.stringify(it.metrics) }}</td>
                <td class="px-6 py-4 text-right text-muted-foreground tabular-nums">{{ it.created_at }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  </div>
</template>
