<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { GearIcon, PlayIcon, Cross1Icon } from '@radix-icons/vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import { api, type NodeItem } from '@/api/client'

const router = useRouter()
const nodes = ref<NodeItem[]>([])
const protocol = ref<'tcp' | 'udp'>('tcp')
const duration = ref('10')
const parallel = ref('1')
const udpBandwidth = ref('100M')
const selected = ref<Record<string, boolean>>({})
const error = ref('')
const loading = ref(false)

const pairs = computed(() => {
  const out: Array<{ source_node_id: string; target_node_id: string; key: string }> = []
  for (const s of nodes.value) {
    for (const t of nodes.value) {
      if (s.id === t.id) continue
      const key = `${s.id}->${t.id}`
      if (selected.value[key]) {
        out.push({ source_node_id: s.id, target_node_id: t.id, key })
      }
    }
  }
  return out
})

async function loadNodes() {
  const res = await api.listNodes()
  nodes.value = res.items
}

async function createRun() {
  if (pairs.value.length === 0) {
    error.value = 'Please select at least one direction pair.'
    return
  }
  error.value = ''
  loading.value = true
  try {
    const payload = {
      mode: pairs.value.length === 1 ? 'one_to_one' : 'many_to_many',
      protocol: protocol.value,
      params: {
        duration: Number(duration.value),
        parallel_streams: Number(parallel.value),
        udp_bandwidth: udpBandwidth.value,
      },
      pairs: pairs.value.map((p) => ({ source_node_id: p.source_node_id, target_node_id: p.target_node_id })),
    }
    const res = await api.createRun(payload)
    router.push(`/runs/${res.run_id}`)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(loadNodes)
</script>

<template>
  <div class="space-y-6 animate-fade-in relative z-10">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">New Test Run</h2>
        <p class="text-sm text-muted-foreground mt-1">Configure and establish network performance testing between your nodes.</p>
      </div>
    </div>

    <!-- Configuration Panel -->
    <Card class="border-primary/20 bg-primary/5">
      <h3 class="mb-5 text-lg font-semibold flex items-center gap-2 text-primary">
        <GearIcon class="w-5 h-5"/> Test Parameters
      </h3>
      <div class="grid grid-cols-1 gap-6 md:grid-cols-4">
        <div class="space-y-2">
           <label class="text-sm font-medium text-foreground">Protocol</label>
           <select v-model="protocol" class="flex h-9 w-full rounded-md border border-input bg-background/50 backdrop-blur-sm px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring hover:bg-background/80 hover:border-input/80">
             <option value="tcp">TCP</option>
             <option value="udp">UDP</option>
           </select>
        </div>
        <div class="space-y-2">
           <label class="text-sm font-medium text-foreground">Duration (s)</label>
           <Input v-model="duration" type="number" />
        </div>
        <div class="space-y-2">
           <label class="text-sm font-medium text-foreground">Parallel Streams</label>
           <Input v-model="parallel" type="number" min="1" />
        </div>
        <div class="space-y-2">
           <label class="text-sm font-medium text-foreground transition-opacity" :class="protocol !== 'udp' ? 'opacity-50' : ''">UDP Bandwidth</label>
           <Input v-model="udpBandwidth" placeholder="100M" :disabled="protocol !== 'udp'" />
        </div>
      </div>
    </Card>

    <!-- Node Matrix Panel -->
    <Card class="overflow-hidden p-0 flex flex-col">
      <div class="p-6 border-b border-white/5 bg-white/[0.02] flex items-center justify-between">
         <h3 class="text-lg font-semibold">Node Direction Matrix (Source ➡️ Target)</h3>
         <div class="text-sm font-medium bg-primary/10 text-primary px-3 py-1 rounded-full border border-primary/20">
            Selected: <span class="font-bold">{{ pairs.length }}</span> pairs
         </div>
      </div>
      
      <div class="overflow-x-auto p-4 bg-black/20">
        <table class="w-full min-w-[700px] border-collapse text-sm">
          <thead>
            <tr>
              <th class="border border-white/10 bg-white/5 px-4 py-3 text-left font-medium text-muted-foreground whitespace-nowrap">
                 Source \ Target
              </th>
              <th v-for="t in nodes" :key="t.id" class="border border-white/10 bg-white/5 px-4 py-3 font-medium text-center truncate max-w-[120px]">
                 {{ t.name }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="nodes.length === 0">
               <td :colspan="nodes.length + 1" class="border border-white/10 px-4 py-8 text-center text-muted-foreground">
                  No nodes available. Add nodes in the Nodes page first.
               </td>
            </tr>
            <tr v-for="s in nodes" :key="s.id" class="group/row">
              <td class="border border-white/10 bg-white/5 px-4 py-3 font-medium truncate max-w-[120px] transition-colors group-hover/row:bg-white/10">
                 {{ s.name }}
              </td>
              <td v-for="t in nodes" :key="`${s.id}-${t.id}`" 
                  class="border border-white/10 px-4 py-2 text-center transition-colors hover:bg-white/[0.02]"
                  :class="{'bg-primary/5': selected[`${s.id}->${t.id}`]}">
                <template v-if="s.id !== t.id">
                  <label class="flex items-center justify-center w-full h-full cursor-pointer p-2">
                     <input v-model="selected[`${s.id}->${t.id}`]" type="checkbox" 
                            class="w-4 h-4 rounded border-white/20 bg-black/40 text-primary focus:ring-primary focus:ring-offset-0 focus:ring-offset-transparent cursor-pointer transition-all" />
                  </label>
                </template>
                <span v-else class="flex items-center justify-center text-white/10 h-full p-2">
                   <Cross1Icon class="w-4 h-4"/>
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <div class="p-6 bg-white/[0.01] border-t border-white/5 flex items-center justify-between">
         <p v-if="error" class="text-sm font-medium text-destructive bg-destructive/10 px-3 py-1.5 rounded border border-destructive/20">{{ error }}</p>
         <div v-else></div> <!-- Spacer -->
         
         <Button :disabled="loading || pairs.length === 0 || nodes.length < 2" @click="createRun" size="lg" class="shadow-[0_0_20px_rgba(var(--primary),0.3)]">
            <span v-if="loading" class="flex items-center gap-2">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent"></span>
              Starting...
            </span>
            <span v-else class="flex items-center justify-center gap-2 font-bold">
              <PlayIcon class="w-5 h-5"/> Start Benchmark
            </span>
         </Button>
      </div>
    </Card>
  </div>
</template>
