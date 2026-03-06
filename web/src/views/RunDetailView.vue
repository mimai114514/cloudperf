<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  InfoCircledIcon, LightningBoltIcon, BarChartIcon,
  ChevronLeftIcon, ArrowRightIcon, CheckCircledIcon, CrossCircledIcon,
  ClockIcon, LoopIcon
} from '@radix-icons/vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import { api, type RunInfo, type RunPair, type ResultItem } from '@/api/client'
import { useRunsStore } from '@/stores/runs'

const route = useRoute()
const router = useRouter()
const runsStore = useRunsStore()
const run = ref<RunInfo | null>(null)
const pairs = ref<RunPair[]>([])
const results = ref<ResultItem[]>([])
const error = ref('')
let timer: number | null = null

const runId = computed(() => String(route.params.id))

// Progress computation
const progress = computed(() => {
  if (!run.value?.summary) return { total: 0, success: 0, failed: 0, running: 0, pending: 0, pct: 0 }
  const s = run.value.summary as Record<string, number>
  const total = s.total || 0
  const success = s.success || 0
  const failed = s.failed || 0
  const running = s.running || 0
  const pending = s.pending || 0
  const done = success + failed
  return { total, success, failed, running, pending, pct: total > 0 ? Math.round((done / total) * 100) : 0 }
})

const isFinished = computed(() => {
  const s = run.value?.status?.toLowerCase()
  return s === 'completed' || s === 'failed' || s === 'success'
})

const getStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'running': return 'text-sky-400 bg-sky-500/10 border-sky-500/20'
    case 'completed': case 'success': return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
    case 'failed': case 'error': return 'text-red-400 bg-red-500/10 border-red-500/20'
    case 'pending': return 'text-amber-400 bg-amber-500/10 border-amber-500/20'
    default: return 'text-muted-foreground bg-white/5 border-white/10'
  }
}

// Extract human-readable metrics from a result item
function formatMetrics(item: ResultItem) {
  const m = item.metrics as Record<string, unknown>
  if (item.protocol === 'tcp') {
    return {
      'Sender Mbps': formatNum(m.sender_mbps),
      'Receiver Mbps': formatNum(m.receiver_mbps),
      'Retransmits': m.retransmits != null ? String(m.retransmits) : '-',
    }
  } else {
    return {
      'Mbps': formatNum(m.mbps),
      'Jitter (ms)': formatNum(m.jitter_ms),
      'Lost / Total': m.lost_packets != null ? `${m.lost_packets} / ${m.total_packets ?? '?'}` : '-',
      'Loss %': formatNum(m.lost_percent),
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
  try {
    run.value = await api.getRun(runId.value)
    pairs.value = (await api.getRunPairs(runId.value)).items
    runsStore.addRun(runId.value)

    // Fetch results for this run
    try {
      const res = await api.listResults(`?run_id=${runId.value}`)
      results.value = res.items
    } catch { /* may not have results yet */ }
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
  <div class="space-y-6 animate-fade-in relative z-10">
    <!-- Header with back -->
    <div class="flex items-center gap-4">
      <Button variant="ghost" size="icon" @click="router.push('/runs')">
        <ChevronLeftIcon class="w-5 h-5" />
      </Button>
      <div class="flex-1 min-w-0">
        <h2 class="text-3xl font-bold tracking-tight">Run Details</h2>
        <p class="text-sm text-muted-foreground mt-0.5 font-mono truncate">{{ runId }}</p>
      </div>
      <div v-if="run">
        <span :class="['inline-flex items-center px-3 py-1 rounded-full text-sm font-semibold border capitalize', getStatusColor(run.status)]">
          <LoopIcon v-if="run.status === 'running'" class="w-3.5 h-3.5 mr-1.5 animate-spin" />
          <CheckCircledIcon v-else-if="run.status === 'completed' || run.status === 'success'" class="w-3.5 h-3.5 mr-1.5" />
          <CrossCircledIcon v-else-if="run.status === 'failed'" class="w-3.5 h-3.5 mr-1.5" />
          <ClockIcon v-else class="w-3.5 h-3.5 mr-1.5" />
          {{ run.status }}
        </span>
      </div>
    </div>

    <p v-if="error" class="text-sm text-destructive bg-destructive/10 px-3 py-2 rounded border border-destructive/20">{{ error }}</p>

    <!-- Progress Bar -->
    <div v-if="run && progress.total > 0" class="space-y-2">
      <div class="flex items-center justify-between text-xs text-muted-foreground">
        <span>Progress</span>
        <span class="tabular-nums font-medium">{{ progress.pct }}% ({{ progress.success + progress.failed }} / {{ progress.total }})</span>
      </div>
      <div class="h-2.5 rounded-full bg-white/5 overflow-hidden flex">
        <div class="h-full bg-emerald-500 transition-all duration-500 ease-out rounded-l-full" :style="{ width: `${(progress.success / progress.total) * 100}%` }"></div>
        <div class="h-full bg-red-500 transition-all duration-500 ease-out" :style="{ width: `${(progress.failed / progress.total) * 100}%` }"></div>
        <div class="h-full bg-sky-500 animate-pulse transition-all duration-500 ease-out" :style="{ width: `${(progress.running / progress.total) * 100}%` }"></div>
      </div>
      <div class="flex gap-4 text-xs">
        <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-emerald-500"></span> Success: {{ progress.success }}</span>
        <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-red-500"></span> Failed: {{ progress.failed }}</span>
        <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-sky-500"></span> Running: {{ progress.running }}</span>
        <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-amber-500"></span> Pending: {{ progress.pending }}</span>
      </div>
    </div>

    <!-- Run Configuration -->
    <Card v-if="run" class="border-primary/20">
      <div class="flex items-center gap-2 mb-4">
        <InfoCircledIcon class="w-5 h-5 text-primary" />
        <h3 class="font-semibold">Configuration</h3>
      </div>
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div class="rounded-lg border border-white/10 bg-white/5 p-3 text-center">
          <p class="text-[10px] uppercase tracking-wider text-muted-foreground mb-0.5">Protocol</p>
          <p class="font-bold uppercase">{{ run.protocol }}</p>
        </div>
        <div class="rounded-lg border border-white/10 bg-white/5 p-3 text-center">
          <p class="text-[10px] uppercase tracking-wider text-muted-foreground mb-0.5">Mode</p>
          <p class="font-bold capitalize">{{ run.mode.replace(/_/g, ' ') }}</p>
        </div>
        <div class="rounded-lg border border-white/10 bg-white/5 p-3 text-center">
          <p class="text-[10px] uppercase tracking-wider text-muted-foreground mb-0.5">Duration</p>
          <p class="font-bold">{{ run.params?.duration ?? '-' }}s</p>
        </div>
        <div class="rounded-lg border border-white/10 bg-white/5 p-3 text-center">
          <p class="text-[10px] uppercase tracking-wider text-muted-foreground mb-0.5">Streams</p>
          <p class="font-bold">{{ run.params?.parallel_streams ?? '-' }}</p>
        </div>
      </div>
    </Card>

    <!-- Pair Progress Table -->
    <Card class="p-0 overflow-hidden">
      <div class="p-5 border-b border-white/5 bg-white/[0.02] flex items-center gap-2">
        <LightningBoltIcon class="w-5 h-5 text-primary" />
        <h3 class="font-semibold">Pair Progress</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-6 py-3 font-medium tracking-wider">Source</th>
              <th class="px-6 py-3 font-medium tracking-wider">Target</th>
              <th class="px-6 py-3 font-medium tracking-wider">Status</th>
              <th class="px-6 py-3 font-medium tracking-wider">Error</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-if="pairs.length === 0 && !error">
              <td colspan="4" class="px-6 py-6 text-center text-muted-foreground">
                <span class="inline-flex items-center gap-2"><span class="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></span> Loading pairs...</span>
              </td>
            </tr>
            <tr v-for="p in pairs" :key="p.id" class="transition-colors hover:bg-white/[0.02]">
              <td class="px-6 py-3 font-mono text-xs text-muted-foreground">{{ p.source_node_id.slice(0, 24) }}</td>
              <td class="px-6 py-3 font-mono text-xs text-muted-foreground">{{ p.target_node_id.slice(0, 24) }}</td>
              <td class="px-6 py-3">
                <span :class="['inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold border capitalize', getStatusColor(p.status)]">
                  <LoopIcon v-if="p.status === 'running'" class="w-3 h-3 mr-1 animate-spin" />
                  {{ p.status }}
                </span>
              </td>
              <td class="px-6 py-3 font-mono text-xs max-w-xs truncate" :class="p.error_message ? 'text-destructive' : 'text-muted-foreground/40'">
                {{ p.error_message || '-' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>

    <!-- Test Results (structured) -->
    <Card v-if="results.length > 0" class="p-0 overflow-hidden">
      <div class="p-5 border-b border-white/5 bg-white/[0.02] flex items-center justify-between">
        <div class="flex items-center gap-2">
          <BarChartIcon class="w-5 h-5 text-primary" />
          <h3 class="font-semibold">Benchmark Results</h3>
        </div>
        <Button variant="outline" size="sm" @click="router.push(`/results?run_id=${runId}`)">
          Full Analysis <ArrowRightIcon class="w-3.5 h-3.5 ml-1" />
        </Button>
      </div>

      <div class="divide-y divide-white/5">
        <div v-for="item in results" :key="item.id" class="p-5">
          <div class="flex items-center gap-3 mb-3">
            <span class="px-2 py-0.5 rounded text-xs font-semibold uppercase bg-white/10">{{ item.protocol }}</span>
            <span class="text-xs text-muted-foreground">Pair: {{ item.pair_id.slice(0, 20) }}…</span>
            <span class="text-xs text-muted-foreground ml-auto tabular-nums">{{ item.created_at }}</span>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div v-for="(val, label) in formatMetrics(item)" :key="label"
                 class="rounded-lg border border-white/10 bg-white/[0.03] p-3 text-center">
              <p class="text-[10px] uppercase tracking-wider text-muted-foreground mb-0.5">{{ label }}</p>
              <p class="text-lg font-bold tabular-nums" :class="val !== '-' ? 'text-foreground' : 'text-muted-foreground'">{{ val }}</p>
            </div>
          </div>
        </div>
      </div>
    </Card>

    <!-- Empty results state -->
    <Card v-else-if="isFinished && results.length === 0" class="text-center py-8">
      <BarChartIcon class="w-8 h-8 mx-auto mb-2 text-muted-foreground/30" />
      <p class="text-muted-foreground">No result data available for this run.</p>
    </Card>
  </div>
</template>
