<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  DashboardIcon, DesktopIcon, RocketIcon, TargetIcon,
  CheckCircledIcon, CrossCircledIcon, LightningBoltIcon, PlusIcon,
  ArrowRightIcon
} from '@radix-icons/vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import { api, type NodeItem, type RunInfo } from '@/api/client'
import { useRunsStore } from '@/stores/runs'

const router = useRouter()
const runsStore = useRunsStore()

const nodes = ref<NodeItem[]>([])
const recentRuns = ref<RunInfo[]>([])
const loading = ref(true)

const onlineNodes = computed(() => nodes.value.filter(n => n.status === 'online').length)
const totalNodes = computed(() => nodes.value.length)

const completedRuns = computed(() => recentRuns.value.filter(r => r.status === 'completed' || r.status === 'success').length)
const failedRuns = computed(() => recentRuns.value.filter(r => r.status === 'failed' || r.status === 'error').length)
const totalRuns = computed(() => recentRuns.value.length)

const getStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'running': return 'text-sky-400 bg-sky-500/10 border-sky-500/20'
    case 'completed': case 'success': return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
    case 'failed': case 'error': return 'text-red-400 bg-red-500/10 border-red-500/20'
    case 'pending': return 'text-amber-400 bg-amber-500/10 border-amber-500/20'
    default: return 'text-muted-foreground bg-white/5 border-white/10'
  }
}

onMounted(async () => {
  try {
    const [nodesRes] = await Promise.all([api.listNodes()])
    nodes.value = nodesRes.items
    recentRuns.value = await runsStore.fetchAll()
  } catch {
    // silent fail on dashboard
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-8 animate-fade-in relative z-10">
    <!-- Header -->
    <div>
      <h1 class="text-3xl font-bold tracking-tight flex items-center gap-3">
        <DashboardIcon class="w-8 h-8 text-primary" />
        Dashboard
      </h1>
      <p class="text-muted-foreground mt-1">System overview and quick actions.</p>
    </div>

    <!-- Stats Cards -->
    <div class="grid gap-4 md:grid-cols-3">
      <!-- Online Nodes -->
      <Card class="relative overflow-hidden group">
        <div class="absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r from-transparent via-emerald-500/80 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Online Nodes</p>
            <p class="text-3xl font-bold mt-1 tabular-nums">
              <span class="text-emerald-400">{{ onlineNodes }}</span>
              <span class="text-lg text-muted-foreground font-normal"> / {{ totalNodes }}</span>
            </p>
          </div>
          <div class="h-12 w-12 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center">
            <DesktopIcon class="w-6 h-6 text-emerald-400" />
          </div>
        </div>
      </Card>

      <!-- Completed Runs -->
      <Card class="relative overflow-hidden group">
        <div class="absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r from-transparent via-sky-500/80 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Total Test Runs</p>
            <p class="text-3xl font-bold mt-1 tabular-nums">
              {{ totalRuns }}
            </p>
          </div>
          <div class="h-12 w-12 rounded-xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center">
            <RocketIcon class="w-6 h-6 text-sky-400" />
          </div>
        </div>
      </Card>

      <!-- Success Rate -->
      <Card class="relative overflow-hidden group">
        <div class="absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r from-transparent via-primary/80 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Success / Failed</p>
            <p class="text-3xl font-bold mt-1 tabular-nums">
              <span class="text-emerald-400">{{ completedRuns }}</span>
              <span class="text-muted-foreground mx-1">/</span>
              <span class="text-red-400">{{ failedRuns }}</span>
            </p>
          </div>
          <div class="h-12 w-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center">
            <TargetIcon class="w-6 h-6 text-primary" />
          </div>
        </div>
      </Card>
    </div>

    <!-- Quick Actions -->
    <div class="grid gap-4 md:grid-cols-2">
      <button @click="router.push('/nodes')"
        class="group relative flex items-center gap-5 rounded-xl border border-white/10 bg-card/40 backdrop-blur-xl p-6 text-left transition-all hover:border-emerald-500/30 hover:bg-emerald-500/5 hover:shadow-xl hover:-translate-y-0.5 focus:outline-none">
        <div class="h-14 w-14 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center shrink-0 transition-transform group-hover:scale-110">
          <PlusIcon class="w-7 h-7 text-emerald-400" />
        </div>
        <div class="flex-1 min-w-0">
          <h3 class="font-semibold text-lg">Manage Nodes</h3>
          <p class="text-sm text-muted-foreground mt-0.5">Register, view and manage your test endpoints.</p>
        </div>
        <ArrowRightIcon class="w-5 h-5 text-muted-foreground group-hover:text-emerald-400 transition-colors shrink-0" />
      </button>

      <button @click="router.push('/runs/new')"
        class="group relative flex items-center gap-5 rounded-xl border border-white/10 bg-card/40 backdrop-blur-xl p-6 text-left transition-all hover:border-primary/30 hover:bg-primary/5 hover:shadow-xl hover:-translate-y-0.5 focus:outline-none">
        <div class="h-14 w-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0 transition-transform group-hover:scale-110">
          <LightningBoltIcon class="w-7 h-7 text-primary" />
        </div>
        <div class="flex-1 min-w-0">
          <h3 class="font-semibold text-lg">New Benchmark</h3>
          <p class="text-sm text-muted-foreground mt-0.5">Configure and run a network performance test.</p>
        </div>
        <ArrowRightIcon class="w-5 h-5 text-muted-foreground group-hover:text-primary transition-colors shrink-0" />
      </button>
    </div>

    <!-- Recent Runs -->
    <Card class="p-0 overflow-hidden">
      <div class="p-5 border-b border-white/5 bg-white/[0.02] flex items-center justify-between">
        <h3 class="text-lg font-semibold flex items-center gap-2">
          <RocketIcon class="w-5 h-5 text-primary" /> Recent Test Runs
        </h3>
        <Button variant="ghost" size="sm" @click="router.push('/runs')" class="text-xs">
          View All <ArrowRightIcon class="w-3.5 h-3.5 ml-1" />
        </Button>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-6 py-3 font-medium tracking-wider">Run ID</th>
              <th class="px-6 py-3 font-medium tracking-wider">Protocol</th>
              <th class="px-6 py-3 font-medium tracking-wider">Status</th>
              <th class="px-6 py-3 font-medium tracking-wider text-right">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-if="loading">
              <td colspan="4" class="px-6 py-8 text-center text-muted-foreground">
                <span class="inline-flex items-center gap-2"><span class="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></span> Loading...</span>
              </td>
            </tr>
            <tr v-else-if="recentRuns.length === 0">
              <td colspan="4" class="px-6 py-10 text-center text-muted-foreground">
                <RocketIcon class="w-8 h-8 mx-auto mb-2 opacity-30" />
                <p>No test runs yet.</p>
                <p class="text-xs mt-1">Create your first benchmark from the quick action above.</p>
              </td>
            </tr>
            <tr v-for="run in recentRuns.slice(0, 5)" :key="run.id"
                class="transition-colors hover:bg-white/[0.02] cursor-pointer"
                @click="router.push(`/runs/${run.id}`)">
              <td class="px-6 py-4 font-mono text-xs text-muted-foreground">{{ run.id.slice(0, 24) }}…</td>
              <td class="px-6 py-4">
                <span class="px-2 py-0.5 rounded text-xs font-semibold uppercase bg-white/10">{{ run.protocol }}</span>
              </td>
              <td class="px-6 py-4">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded text-xs font-semibold border capitalize', getStatusColor(run.status)]">
                  {{ run.status }}
                </span>
              </td>
              <td class="px-6 py-4 text-right">
                <ArrowRightIcon class="w-4 h-4 text-muted-foreground inline-block" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>
  </div>
</template>
