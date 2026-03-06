<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { InfoCircledIcon, LightningBoltIcon } from '@radix-icons/vue'
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

const getStatusColor = (status: string) => {
  switch(status.toLowerCase()) {
    case 'running': return 'bg-sky-500/10 text-sky-400 border-sky-500/20';
    case 'completed': return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    case 'failed': case 'error': return 'bg-destructive/10 text-destructive border-destructive/20';
    case 'pending': return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    default: return 'bg-white/5 text-muted-foreground border-white/10';
  }
}
</script>

<template>
  <div class="space-y-6 animate-fade-in relative z-10">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Task Details</h2>
        <p class="text-sm text-muted-foreground mt-1">Real-time status tracking for the current test run.</p>
      </div>
    </div>

    <!-- Task Info Card -->
    <Card class="border-primary/20 bg-primary/5">
      <div class="flex items-center gap-2 mb-4">
        <InfoCircledIcon class="w-5 h-5 text-primary" />
        <h3 class="text-lg font-semibold">General Information</h3>
      </div>
      
      <p v-if="error" class="mb-4 text-sm font-medium text-destructive bg-destructive/10 px-3 py-2 rounded border border-destructive/20">{{ error }}</p>
      
      <div v-if="run" class="grid grid-cols-2 gap-4 md:grid-cols-4">
        <div class="space-y-1">
           <p class="text-xs uppercase tracking-wider font-semibold text-muted-foreground">Run ID</p>
           <p class="font-mono text-sm break-all font-medium">{{ run.id }}</p>
        </div>
        <div class="space-y-1">
           <p class="text-xs uppercase tracking-wider font-semibold text-muted-foreground">Status</p>
           <div :class="['inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold border capitalize', getStatusColor(run.status)]">
              {{ run.status }}
           </div>
        </div>
        <div class="space-y-1">
           <p class="text-xs uppercase tracking-wider font-semibold text-muted-foreground">Protocol</p>
           <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-white/10 text-white uppercase shadow-sm">
             {{ run.protocol }}
           </span>
        </div>
        <div class="space-y-1">
           <p class="text-xs uppercase tracking-wider font-semibold text-muted-foreground">Mode</p>
           <p class="text-sm font-medium capitalize">{{ run.mode.replace(/_/g, ' ') }}</p>
        </div>
      </div>
      <div v-else-if="!error" class="text-sm text-muted-foreground flex items-center gap-2">
         <span class="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></span>
         Loading task details...
      </div>
    </Card>

    <!-- Pairs Table Card -->
    <Card class="p-0 overflow-hidden flex flex-col">
      <div class="p-5 border-b border-white/5 bg-white/[0.02]">
        <h3 class="text-lg font-semibold flex items-center gap-2">
           <LightningBoltIcon class="w-5 h-5 text-primary" /> Edge Pair Progress
        </h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left whitespace-nowrap">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-6 py-4 font-medium tracking-wider">Source Node ID</th>
              <th class="px-6 py-4 font-medium tracking-wider">Target Node ID</th>
              <th class="px-6 py-4 font-medium tracking-wider">Status</th>
              <th class="px-6 py-4 font-medium tracking-wider w-1/3">Error Info</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5 bg-transparent">
            <tr v-if="pairs.length === 0 && !error">
               <td colspan="4" class="px-6 py-8 text-center text-muted-foreground flex items-center justify-center gap-2">
                 <span class="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></span>
                 Awaiting pair details...
               </td>
            </tr>
            <tr v-for="p in pairs" :key="p.id" class="transition-colors hover:bg-white/[0.02]">
              <td class="px-6 py-4 font-mono text-xs text-muted-foreground">{{ p.source_node_id }}</td>
              <td class="px-6 py-4 font-mono text-xs text-muted-foreground">{{ p.target_node_id }}</td>
              <td class="px-6 py-4">
                 <div :class="['inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold border capitalize', getStatusColor(p.status)]">
                    {{ p.status }}
                 </div>
              </td>
              <td class="px-6 py-4 font-mono text-xs max-w-xs truncate" :class="p.error_message ? 'text-destructive' : 'text-muted-foreground/50'">
                 {{ p.error_message || '-' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>
  </div>
</template>
