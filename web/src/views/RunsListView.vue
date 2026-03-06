<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { RocketIcon, PlusIcon, ArrowRightIcon, ReloadIcon } from '@radix-icons/vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import { type RunInfo } from '@/api/client'
import { useRunsStore } from '@/stores/runs'

const router = useRouter()
const runsStore = useRunsStore()
const runs = ref<RunInfo[]>([])
const loading = ref(true)

const getStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'running': return 'text-sky-400 bg-sky-500/10 border-sky-500/20'
    case 'completed': case 'success': return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
    case 'failed': case 'error': return 'text-red-400 bg-red-500/10 border-red-500/20'
    case 'pending': return 'text-amber-400 bg-amber-500/10 border-amber-500/20'
    default: return 'text-muted-foreground bg-white/5 border-white/10'
  }
}

async function load() {
  loading.value = true
  try {
    runs.value = await runsStore.fetchAll()
  } catch { /* silent */ }
  finally { loading.value = false }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 animate-fade-in relative z-10">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Test Runs</h2>
        <p class="text-sm text-muted-foreground mt-1">View and manage all your benchmark tasks.</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="outline" size="sm" @click="load" :disabled="loading">
          <ReloadIcon class="w-4 h-4 mr-1.5" :class="loading ? 'animate-spin' : ''" /> Refresh
        </Button>
        <Button @click="router.push('/runs/new')" class="shadow-[0_0_15px_rgba(var(--primary),0.3)]">
          <PlusIcon class="w-4 h-4 mr-1.5" /> New Run
        </Button>
      </div>
    </div>

    <Card class="p-0 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-6 py-4 font-medium tracking-wider">Run ID</th>
              <th class="px-6 py-4 font-medium tracking-wider">Protocol</th>
              <th class="px-6 py-4 font-medium tracking-wider">Mode</th>
              <th class="px-6 py-4 font-medium tracking-wider">Status</th>
              <th class="px-6 py-4 font-medium tracking-wider">Pairs</th>
              <th class="px-6 py-4 font-medium tracking-wider text-right">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-if="loading && runs.length === 0">
              <td colspan="6" class="px-6 py-10 text-center text-muted-foreground">
                <span class="inline-flex items-center gap-2"><span class="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></span> Fetching runs...</span>
              </td>
            </tr>
            <tr v-else-if="runs.length === 0">
              <td colspan="6" class="px-6 py-12 text-center text-muted-foreground">
                <RocketIcon class="w-10 h-10 mx-auto mb-3 opacity-20" />
                <p class="font-medium">No test runs found</p>
                <p class="text-xs mt-1 mb-4">Start your first benchmark to see results here.</p>
                <Button size="sm" @click="router.push('/runs/new')">
                  <PlusIcon class="w-4 h-4 mr-1.5" /> Create First Run
                </Button>
              </td>
            </tr>
            <tr v-for="run in runs" :key="run.id"
                class="transition-colors hover:bg-white/[0.02] cursor-pointer"
                @click="router.push(`/runs/${run.id}`)">
              <td class="px-6 py-4 font-mono text-xs text-muted-foreground">{{ run.id }}</td>
              <td class="px-6 py-4">
                <span class="px-2 py-0.5 rounded text-xs font-semibold uppercase bg-white/10">{{ run.protocol }}</span>
              </td>
              <td class="px-6 py-4 text-sm capitalize text-muted-foreground">{{ run.mode.replace(/_/g, ' ') }}</td>
              <td class="px-6 py-4">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded text-xs font-semibold border capitalize', getStatusColor(run.status)]">
                  {{ run.status }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm tabular-nums">
                <template v-if="run.summary">
                  <span class="text-emerald-400">{{ run.summary.success ?? 0 }}</span>
                  <span class="text-muted-foreground/50 mx-0.5">/</span>
                  <span class="text-red-400">{{ run.summary.failed ?? 0 }}</span>
                  <span class="text-muted-foreground/50 mx-0.5">/</span>
                  <span class="text-muted-foreground">{{ run.summary.total ?? 0 }}</span>
                </template>
                <span v-else class="text-muted-foreground">-</span>
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
