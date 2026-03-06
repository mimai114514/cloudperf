<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import { MagnifyingGlassIcon, DownloadIcon, BarChartIcon, LoopIcon, Cross2Icon } from '@radix-icons/vue'
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

// Toggle: tcp | udp
const activeTab = ref<'tcp' | 'udp'>('tcp')

onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  const qRunId = params.get('run_id')
  if (qRunId) { runId.value = qRunId }
})

const chartOption = computed(() => {
  const data = items.value
    .map((it) => {
      if (it.protocol === 'tcp') {
        return Number((it.metrics as any).receiver_mbps ?? (it.metrics as any).sender_mbps ?? 0)
      }
      return Number((it.metrics as any).mbps ?? 0)
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
    grid: { top: 30, right: 20, bottom: 10, left: 60 },
    xAxis: {
      type: 'category',
      data: data.map((_, i) => String(i + 1)),
      show: false,
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

function formatNum(v: unknown): string {
  if (v == null) return '-'
  const n = Number(v)
  if (isNaN(n)) return String(v)
  return n.toFixed(2)
}

const tcpResults = computed(() => items.value.filter(it => it.protocol === 'tcp'))
const udpResults = computed(() => items.value.filter(it => it.protocol === 'udp'))
const activeResults = computed(() => activeTab.value === 'tcp' ? tcpResults.value : udpResults.value)

function exportHref() {
  const q = new URLSearchParams()
  if (runId.value) q.set('run_id', runId.value)
  if (from.value) q.set('from', from.value)
  if (to.value) q.set('to', to.value)
  const query = q.toString() ? `?${q.toString()}` : ''
  return api.exportResults(query)
}

function clearFilters() {
  runId.value = ''
  from.value = ''
  to.value = ''
}

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

async function refresh() {
  clearFilters()
  await load()
}

onMounted(async () => {
  await load()
  await nextTick()
  if (chartEl.value) {
    chart = echarts.init(chartEl.value)
    chart.setOption(chartOption.value)
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
    <div>
      <h2 class="text-3xl font-bold tracking-tight">Results</h2>
      <p class="text-sm text-muted-foreground mt-1">Analyze and export benchmark data.</p>
    </div>

    <!-- Throughput Trend -->
    <Card class="p-0 overflow-hidden">
      <div class="p-5 border-b border-white/5 flex items-center gap-2 bg-white/[0.02]">
        <BarChartIcon class="w-5 h-5 text-primary" />
        <h3 class="font-semibold">Throughput Trend</h3>
      </div>
      <div class="p-4">
        <div v-show="items.length === 0" class="h-64 w-full flex items-center justify-center text-muted-foreground">
          No chart data available.
        </div>
        <div ref="chartEl" v-show="items.length > 0" class="h-64 w-full"></div>
      </div>
    </Card>

    <!-- Detailed Results with TCP/UDP Toggle -->
    <Card class="p-0 overflow-hidden">
      <div class="p-4 border-b border-white/5 bg-white/[0.02] flex items-center gap-3">
        <h3 class="font-semibold text-sm mr-2">Detailed Results</h3>
        <!-- Toggle Buttons -->
        <div class="inline-flex rounded-lg border border-white/10 p-0.5 bg-white/[0.03]">
          <button @click="activeTab = 'tcp'"
            class="px-3.5 py-1.5 rounded-md text-xs font-semibold transition-all"
            :class="activeTab === 'tcp'
              ? 'bg-sky-500/20 text-sky-400 border border-sky-500/30 shadow-sm'
              : 'text-muted-foreground hover:text-foreground border border-transparent'">
            TCP <span class="ml-1 opacity-70 tabular-nums">{{ tcpResults.length }}</span>
          </button>
          <button @click="activeTab = 'udp'"
            class="px-3.5 py-1.5 rounded-md text-xs font-semibold transition-all"
            :class="activeTab === 'udp'
              ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30 shadow-sm'
              : 'text-muted-foreground hover:text-foreground border border-transparent'">
            UDP <span class="ml-1 opacity-70 tabular-nums">{{ udpResults.length }}</span>
          </button>
        </div>
        <span class="ml-auto text-xs text-muted-foreground tabular-nums">{{ activeResults.length }} result(s)</span>
      </div>
      <div class="overflow-x-auto">
        <!-- TCP Table -->
        <table v-if="activeTab === 'tcp'" class="w-full text-sm text-left">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-5 py-2.5 font-medium">Send Mbps</th>
              <th class="px-5 py-2.5 font-medium">Recv Mbps</th>
              <th class="px-5 py-2.5 font-medium">Retransmits</th>
              <th class="px-5 py-2.5 font-medium text-right">Time</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-if="tcpResults.length === 0">
              <td colspan="4" class="px-5 py-8 text-center text-muted-foreground text-xs">No TCP results.</td>
            </tr>
            <tr v-for="it in tcpResults" :key="it.id" class="hover:bg-white/[0.02]">
              <td class="px-5 py-2.5 tabular-nums font-medium text-sky-300">{{ formatNum((it.metrics as any).sender_mbps) }}</td>
              <td class="px-5 py-2.5 tabular-nums font-medium text-emerald-300">{{ formatNum((it.metrics as any).receiver_mbps) }}</td>
              <td class="px-5 py-2.5 tabular-nums text-muted-foreground">{{ (it.metrics as any).retransmits ?? '-' }}</td>
              <td class="px-5 py-2.5 text-right text-muted-foreground text-xs tabular-nums">{{ it.created_at }}</td>
            </tr>
          </tbody>
        </table>

        <!-- UDP Table -->
        <table v-else class="w-full text-sm text-left">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-5 py-2.5 font-medium">Mbps</th>
              <th class="px-5 py-2.5 font-medium">Jitter (ms)</th>
              <th class="px-5 py-2.5 font-medium">Packet Loss</th>
              <th class="px-5 py-2.5 font-medium text-right">Time</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-if="udpResults.length === 0">
              <td colspan="4" class="px-5 py-8 text-center text-muted-foreground text-xs">No UDP results.</td>
            </tr>
            <tr v-for="it in udpResults" :key="it.id" class="hover:bg-white/[0.02]">
              <td class="px-5 py-2.5 tabular-nums font-medium text-sky-300">{{ formatNum((it.metrics as any).mbps) }}</td>
              <td class="px-5 py-2.5 tabular-nums text-amber-300">{{ formatNum((it.metrics as any).jitter_ms) }}</td>
              <td class="px-5 py-2.5 tabular-nums text-red-300">
                {{ (it.metrics as any).lost_packets ?? '-' }}/{{ (it.metrics as any).total_packets ?? '?' }}
                <span v-if="(it.metrics as any).lost_percent != null" class="text-muted-foreground"> ({{ formatNum((it.metrics as any).lost_percent) }}%)</span>
              </td>
              <td class="px-5 py-2.5 text-right text-muted-foreground text-xs tabular-nums">{{ it.created_at }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>

    <!-- Query & Export (bottom) -->
    <Card class="border-primary/20 bg-primary/5">
      <div class="flex items-center gap-2 mb-4">
        <MagnifyingGlassIcon class="w-5 h-5 text-primary" />
        <h3 class="font-semibold">Query & Export</h3>
      </div>
      <div class="grid grid-cols-1 gap-4 md:grid-cols-12 md:items-end">
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">Run ID</label>
          <Input v-model="runId" placeholder="e.g. run_xxxxx" />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">From</label>
          <Input v-model="from" placeholder="YYYY-MM-DDTHH:mm:ssZ" />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">To</label>
          <Input v-model="to" placeholder="YYYY-MM-DDTHH:mm:ssZ" />
        </div>
        <div class="flex items-center gap-2 md:col-span-3">
          <Button @click="load" :disabled="loading" class="flex-1">
            <LoopIcon v-if="loading" class="w-4 h-4 mr-1.5 animate-spin" />
            <MagnifyingGlassIcon v-else class="w-4 h-4 mr-1.5" />
            Query
          </Button>
          <Button variant="ghost" size="icon" @click="clearFilters" title="Clear filters"
            class="shrink-0 text-muted-foreground hover:text-foreground">
            <Cross2Icon class="w-4 h-4" />
          </Button>
          <a :href="exportHref()" class="inline-flex items-center justify-center rounded-md bg-emerald-500/20 px-3 py-2 text-sm font-medium text-emerald-400 border border-emerald-500/30 transition-colors hover:bg-emerald-500/30 shrink-0">
            <DownloadIcon class="w-4 h-4 mr-1.5" /> CSV
          </a>
        </div>
      </div>
      <div class="flex items-center gap-2 mt-3">
        <Button variant="outline" size="sm" @click="refresh">
          <LoopIcon class="w-3.5 h-3.5 mr-1.5" /> Refresh All
        </Button>
        <span class="text-xs text-muted-foreground">Clears filters and reloads all results.</span>
      </div>
      <p v-if="error" class="mt-3 text-sm text-destructive bg-destructive/10 px-3 py-2 rounded border border-destructive/20">{{ error }}</p>
    </Card>
  </div>
</template>
