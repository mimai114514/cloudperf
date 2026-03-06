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

// Read run_id from URL query
onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  const qRunId = params.get('run_id')
  if (qRunId) {
    runId.value = qRunId
  }
})

const chartOption = computed(() => {
  const labels = items.value.map((_, i) => String(i + 1)).reverse()
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

// Structured metric extraction
function formatMetrics(item: ResultItem) {
  const m = item.metrics as Record<string, unknown>
  if (item.protocol === 'tcp') {
    return {
      sender: formatNum(m.sender_mbps),
      receiver: formatNum(m.receiver_mbps),
      retransmits: m.retransmits != null ? String(m.retransmits) : '-',
    }
  } else {
    return {
      mbps: formatNum(m.mbps),
      jitter: formatNum(m.jitter_ms),
      lost: m.lost_packets != null ? `${m.lost_packets}/${m.total_packets ?? '?'}` : '-',
      lossPercent: formatNum(m.lost_percent),
    }
  }
}

function formatNum(v: unknown): string {
  if (v == null) return '-'
  const n = Number(v)
  if (isNaN(n)) return String(v)
  return n.toFixed(2)
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
    const ro = new ResizeObserver(() => chart?.resize())
    ro.observe(chartEl.value)
  }
})

watch(chartOption, (opt) => {
  chart?.setOption(opt)
})

// Determine which columns to show
const hasTcp = computed(() => items.value.some(it => it.protocol === 'tcp'))
const hasUdp = computed(() => items.value.some(it => it.protocol === 'udp'))
</script>

<template>
  <div class="space-y-6 animate-fade-in relative z-10">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Results</h2>
        <p class="text-sm text-muted-foreground mt-1">Query, analyze and export benchmark data.</p>
      </div>
    </div>

    <!-- Filters -->
    <Card class="border-primary/20 bg-primary/5">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-12 md:items-end">
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">Run ID</label>
          <Input v-model="runId" placeholder="e.g. run_xxxxx" />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">From (RFC3339)</label>
          <Input v-model="from" placeholder="YYYY-MM-DDTHH:mm:ssZ" />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-xs font-medium text-muted-foreground ml-1">To (RFC3339)</label>
          <Input v-model="to" placeholder="YYYY-MM-DDTHH:mm:ssZ" />
        </div>
        <div class="flex items-center gap-3 md:col-span-3">
          <Button @click="load" :disabled="loading" class="flex-1">
            <LoopIcon v-if="loading" class="w-4 h-4 mr-2 animate-spin" />
            <MagnifyingGlassIcon v-else class="w-4 h-4 mr-2" />
            Query
          </Button>
          <a :href="exportHref" class="inline-flex items-center justify-center rounded-md bg-emerald-500/20 px-4 py-2 text-sm font-medium text-emerald-400 border border-emerald-500/30 transition-colors hover:bg-emerald-500/30">
            <DownloadIcon class="w-4 h-4 mr-2" /> CSV
          </a>
        </div>
      </div>
      <p v-if="error" class="mt-3 text-sm text-destructive bg-destructive/10 px-3 py-2 rounded border border-destructive/20">{{ error }}</p>
    </Card>

    <!-- Chart -->
    <Card class="p-0 overflow-hidden">
      <div class="p-5 border-b border-white/5 flex items-center gap-2 bg-white/[0.02]">
        <BarChartIcon class="w-5 h-5 text-primary" />
        <h3 class="font-semibold">Throughput Trend</h3>
      </div>
      <div class="p-4 bg-gradient-to-b from-primary/5 to-transparent">
        <div v-show="items.length === 0" class="h-80 w-full flex items-center justify-center text-muted-foreground">
          No chart data available.
        </div>
        <div ref="chartEl" v-show="items.length > 0" class="h-80 w-full"></div>
      </div>
    </Card>

    <!-- Structured Data Table -->
    <Card class="p-0 overflow-hidden">
      <div class="p-5 border-b border-white/5 bg-white/[0.02]">
        <h3 class="font-semibold">Detailed Results</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-5 py-3 font-medium tracking-wider">Protocol</th>
              <th class="px-5 py-3 font-medium tracking-wider" v-if="hasTcp">Send Mbps</th>
              <th class="px-5 py-3 font-medium tracking-wider" v-if="hasTcp">Recv Mbps</th>
              <th class="px-5 py-3 font-medium tracking-wider" v-if="hasTcp">Retransmits</th>
              <th class="px-5 py-3 font-medium tracking-wider" v-if="hasUdp">Mbps</th>
              <th class="px-5 py-3 font-medium tracking-wider" v-if="hasUdp">Jitter (ms)</th>
              <th class="px-5 py-3 font-medium tracking-wider" v-if="hasUdp">Packet Loss</th>
              <th class="px-5 py-3 font-medium tracking-wider text-right">Time</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-if="loading && items.length === 0">
              <td :colspan="8" class="px-5 py-8 text-center text-muted-foreground">Fetching data...</td>
            </tr>
            <tr v-else-if="items.length === 0">
              <td :colspan="8" class="px-5 py-8 text-center text-muted-foreground">No results found.</td>
            </tr>
            <tr v-for="it in items" :key="it.id" class="transition-colors hover:bg-white/[0.02]">
              <td class="px-5 py-3">
                <span class="px-2 py-0.5 rounded text-xs font-semibold uppercase bg-white/10">{{ it.protocol }}</span>
              </td>
              <!-- TCP columns -->
              <template v-if="hasTcp">
                <td class="px-5 py-3 tabular-nums font-medium" :class="it.protocol === 'tcp' ? 'text-sky-300' : 'text-muted-foreground/30'">
                  {{ it.protocol === 'tcp' ? formatMetrics(it).sender : '-' }}
                </td>
                <td class="px-5 py-3 tabular-nums font-medium" :class="it.protocol === 'tcp' ? 'text-emerald-300' : 'text-muted-foreground/30'">
                  {{ it.protocol === 'tcp' ? formatMetrics(it).receiver : '-' }}
                </td>
                <td class="px-5 py-3 tabular-nums" :class="it.protocol === 'tcp' ? 'text-muted-foreground' : 'text-muted-foreground/30'">
                  {{ it.protocol === 'tcp' ? formatMetrics(it).retransmits : '-' }}
                </td>
              </template>
              <!-- UDP columns -->
              <template v-if="hasUdp">
                <td class="px-5 py-3 tabular-nums font-medium" :class="it.protocol === 'udp' ? 'text-sky-300' : 'text-muted-foreground/30'">
                  {{ it.protocol === 'udp' ? formatMetrics(it).mbps : '-' }}
                </td>
                <td class="px-5 py-3 tabular-nums" :class="it.protocol === 'udp' ? 'text-amber-300' : 'text-muted-foreground/30'">
                  {{ it.protocol === 'udp' ? formatMetrics(it).jitter : '-' }}
                </td>
                <td class="px-5 py-3 tabular-nums" :class="it.protocol === 'udp' ? 'text-red-300' : 'text-muted-foreground/30'">
                  {{ it.protocol === 'udp' ? `${formatMetrics(it).lost} (${formatMetrics(it).lossPercent}%)` : '-' }}
                </td>
              </template>
              <td class="px-5 py-3 text-right text-muted-foreground tabular-nums text-xs">{{ it.created_at }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>
  </div>
</template>
