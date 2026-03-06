<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  GearIcon, PlayIcon, Cross1Icon, CheckCircledIcon,
  ChevronLeftIcon, ChevronRightIcon, DesktopIcon
} from '@radix-icons/vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import { api, type NodeItem } from '@/api/client'
import { useRunsStore } from '@/stores/runs'

const router = useRouter()
const runsStore = useRunsStore()
const nodes = ref<NodeItem[]>([])
const error = ref('')
const loading = ref(false)

// Wizard step
const step = ref(1)

// Step 1: Protocol & params
const protocol = ref<'tcp' | 'udp'>('tcp')
const duration = ref('10')
const parallel = ref('1')
const udpBandwidth = ref('100M')

// Step 2: Node pairs
const selected = ref<Record<string, boolean>>({})

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

const onlineNodes = computed(() => nodes.value.filter(n => n.status === 'online'))

function togglePair(sourceId: string, targetId: string) {
  const key = `${sourceId}->${targetId}`
  selected.value[key] = !selected.value[key]
}

function selectAll() {
  for (const s of nodes.value) {
    for (const t of nodes.value) {
      if (s.id !== t.id) selected.value[`${s.id}->${t.id}`] = true
    }
  }
}
function clearAll() {
  selected.value = {}
}

function removePair(key: string) {
  delete selected.value[key]
}

function getNodeName(id: string) {
  return nodes.value.find(n => n.id === id)?.name ?? id.slice(0, 16)
}

// Step navigation
function nextStep() {
  if (step.value === 1) {
    step.value = 2
  } else if (step.value === 2) {
    if (pairs.value.length === 0) {
      error.value = 'Please select at least one node pair.'
      return
    }
    error.value = ''
    step.value = 3
  }
}
function prevStep() {
  error.value = ''
  step.value = Math.max(1, step.value - 1)
}

async function createRun() {
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
      pairs: pairs.value.map(p => ({ source_node_id: p.source_node_id, target_node_id: p.target_node_id })),
    }
    const res = await api.createRun(payload)
    runsStore.addRun(res.run_id)
    router.push(`/runs/${res.run_id}`)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const res = await api.listNodes()
  nodes.value = res.items
})
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-6 animate-fade-in relative z-10">
    <!-- Header -->
    <div>
      <h2 class="text-3xl font-bold tracking-tight">New Benchmark</h2>
      <p class="text-sm text-muted-foreground mt-1">Configure and launch a network performance test.</p>
    </div>

    <!-- Stepper -->
    <div class="flex items-center gap-2">
      <template v-for="(s, idx) in ['Protocol & Params', 'Select Pairs', 'Confirm & Launch']" :key="idx">
        <div v-if="idx > 0" class="flex-1 h-px" :class="step > idx ? 'bg-primary/60' : 'bg-white/10'"></div>
        <button @click="idx + 1 < step ? step = idx + 1 : null"
          class="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium transition-all shrink-0"
          :class="[
            step === idx + 1 ? 'bg-primary/10 text-primary border border-primary/30' :
            step > idx + 1 ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 cursor-pointer' :
            'bg-white/5 text-muted-foreground border border-white/10'
          ]">
          <span v-if="step > idx + 1" class="w-4 h-4"><CheckCircledIcon class="w-4 h-4" /></span>
          <span v-else class="w-4 h-4 flex items-center justify-center text-[10px] font-bold">{{ idx + 1 }}</span>
          <span class="hidden sm:inline">{{ s }}</span>
        </button>
      </template>
    </div>

    <!-- Step 1. Protocol & Parameters -->
    <Card v-if="step === 1" class="border-primary/20 bg-primary/5">
      <div class="flex items-center gap-2 mb-6">
        <GearIcon class="w-5 h-5 text-primary" />
        <h3 class="text-lg font-semibold">Protocol & Parameters</h3>
      </div>

      <!-- Protocol selector -->
      <div class="mb-6">
        <label class="text-sm font-medium text-muted-foreground mb-3 block">Select Protocol</label>
        <div class="grid grid-cols-2 gap-3">
          <button @click="protocol = 'tcp'"
            class="flex flex-col items-center gap-2 rounded-xl border-2 p-5 transition-all"
            :class="protocol === 'tcp'
              ? 'border-primary bg-primary/10 text-primary shadow-lg shadow-primary/10'
              : 'border-white/10 hover:border-white/20 text-muted-foreground'">
            <span class="text-2xl font-bold">TCP</span>
            <span class="text-xs">Reliable throughput test</span>
          </button>
          <button @click="protocol = 'udp'"
            class="flex flex-col items-center gap-2 rounded-xl border-2 p-5 transition-all"
            :class="protocol === 'udp'
              ? 'border-primary bg-primary/10 text-primary shadow-lg shadow-primary/10'
              : 'border-white/10 hover:border-white/20 text-muted-foreground'">
            <span class="text-2xl font-bold">UDP</span>
            <span class="text-xs">Jitter & packet loss test</span>
          </button>
        </div>
      </div>

      <!-- Parameters -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div class="space-y-2">
          <label class="text-sm font-medium text-foreground">Duration (seconds)</label>
          <Input v-model="duration" type="number" min="1" max="300" />
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-foreground">Parallel Streams</label>
          <Input v-model="parallel" type="number" min="1" max="20" />
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-foreground transition-opacity" :class="protocol !== 'udp' ? 'opacity-40' : ''">UDP Bandwidth</label>
          <Input v-model="udpBandwidth" placeholder="100M" :disabled="protocol !== 'udp'" />
        </div>
      </div>

      <div class="flex justify-end mt-6">
        <Button @click="nextStep">
          Next <ChevronRightIcon class="w-4 h-4 ml-1" />
        </Button>
      </div>
    </Card>

    <!-- Step 2. Select Node Pairs -->
    <Card v-if="step === 2" class="border-primary/20 bg-primary/5">
      <div class="flex items-center justify-between mb-5">
        <div class="flex items-center gap-2">
          <DesktopIcon class="w-5 h-5 text-primary" />
          <h3 class="text-lg font-semibold">Select Node Pairs</h3>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="ghost" size="sm" @click="selectAll">Select All</Button>
          <Button variant="ghost" size="sm" @click="clearAll">Clear</Button>
        </div>
      </div>

      <!-- Matrix view (compact for ≤ 8 nodes) -->
      <div v-if="nodes.length <= 8" class="overflow-x-auto mb-4">
        <table class="w-full min-w-[400px] border-collapse text-sm">
          <thead>
            <tr>
              <th class="border border-white/10 bg-white/5 px-3 py-2.5 text-left font-medium text-muted-foreground text-xs uppercase tracking-wider">
                Src \ Tgt
              </th>
              <th v-for="t in nodes" :key="t.id" class="border border-white/10 bg-white/5 px-3 py-2.5 font-medium text-center text-xs truncate max-w-[100px]">
                {{ t.name }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in nodes" :key="s.id">
              <td class="border border-white/10 bg-white/5 px-3 py-2.5 font-medium text-xs truncate max-w-[100px]">{{ s.name }}</td>
              <td v-for="t in nodes" :key="`${s.id}-${t.id}`"
                  class="border border-white/10 px-3 py-2 text-center transition-colors"
                  :class="{'bg-primary/10': selected[`${s.id}->${t.id}`]}">
                <template v-if="s.id !== t.id">
                  <label class="flex items-center justify-center cursor-pointer p-1">
                    <input :checked="!!selected[`${s.id}->${t.id}`]"
                           @change="togglePair(s.id, t.id)" type="checkbox"
                           class="w-4 h-4 rounded border-white/20 bg-black/40 text-primary focus:ring-primary cursor-pointer" />
                  </label>
                </template>
                <span v-else class="text-white/10"><Cross1Icon class="w-3.5 h-3.5 inline-block" /></span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- List view for many nodes -->
      <div v-else class="grid gap-2 sm:grid-cols-2 mb-4 max-h-[400px] overflow-y-auto pr-1">
        <template v-for="s in nodes" :key="s.id">
          <template v-for="t in nodes" :key="`${s.id}-${t.id}`">
            <label v-if="s.id !== t.id"
              class="flex items-center gap-3 rounded-lg border px-3 py-2.5 cursor-pointer transition-colors text-sm"
              :class="selected[`${s.id}->${t.id}`]
                ? 'border-primary/30 bg-primary/10'
                : 'border-white/10 hover:border-white/20'">
              <input :checked="!!selected[`${s.id}->${t.id}`]"
                     @change="togglePair(s.id, t.id)" type="checkbox"
                     class="w-4 h-4 rounded border-white/20 bg-black/40 text-primary focus:ring-primary cursor-pointer shrink-0" />
              <span class="truncate">{{ s.name }} → {{ t.name }}</span>
            </label>
          </template>
        </template>
      </div>

      <!-- Selected pairs summary -->
      <div class="flex items-center gap-2 text-sm mb-4">
        <span :class="pairs.length > 0 ? 'text-primary' : 'text-muted-foreground'" class="font-medium">{{ pairs.length }} pair(s) selected</span>
      </div>

      <p v-if="error" class="text-sm text-destructive bg-destructive/10 px-3 py-2 rounded border border-destructive/20 mb-4">{{ error }}</p>

      <div class="flex justify-between">
        <Button variant="outline" @click="prevStep">
          <ChevronLeftIcon class="w-4 h-4 mr-1" /> Back
        </Button>
        <Button @click="nextStep">
          Next <ChevronRightIcon class="w-4 h-4 ml-1" />
        </Button>
      </div>
    </Card>

    <!-- Step 3. Confirm & Launch -->
    <Card v-if="step === 3" class="border-primary/20 bg-primary/5">
      <div class="flex items-center gap-2 mb-6">
        <PlayIcon class="w-5 h-5 text-primary" />
        <h3 class="text-lg font-semibold">Confirm & Launch</h3>
      </div>

      <!-- Summary -->
      <div class="space-y-4 mb-6">
        <div class="grid grid-cols-3 gap-4">
          <div class="rounded-lg border border-white/10 bg-white/5 p-4 text-center">
            <p class="text-xs uppercase tracking-wider text-muted-foreground mb-1">Protocol</p>
            <p class="text-xl font-bold uppercase">{{ protocol }}</p>
          </div>
          <div class="rounded-lg border border-white/10 bg-white/5 p-4 text-center">
            <p class="text-xs uppercase tracking-wider text-muted-foreground mb-1">Duration</p>
            <p class="text-xl font-bold">{{ duration }}s</p>
          </div>
          <div class="rounded-lg border border-white/10 bg-white/5 p-4 text-center">
            <p class="text-xs uppercase tracking-wider text-muted-foreground mb-1">Streams</p>
            <p class="text-xl font-bold">{{ parallel }}</p>
          </div>
        </div>

        <div v-if="protocol === 'udp'" class="rounded-lg border border-white/10 bg-white/5 p-4">
          <span class="text-xs uppercase tracking-wider text-muted-foreground">UDP Bandwidth:</span>
          <span class="ml-2 font-semibold">{{ udpBandwidth }}</span>
        </div>

        <div>
          <p class="text-xs uppercase tracking-wider text-muted-foreground mb-2">Test Pairs ({{ pairs.length }})</p>
          <div class="space-y-1.5 max-h-[200px] overflow-y-auto pr-1">
            <div v-for="p in pairs" :key="p.key"
              class="flex items-center justify-between rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 text-sm">
              <span>
                <span class="font-medium">{{ getNodeName(p.source_node_id) }}</span>
                <span class="text-muted-foreground mx-2">→</span>
                <span class="font-medium">{{ getNodeName(p.target_node_id) }}</span>
              </span>
              <button @click="removePair(p.key)" class="text-muted-foreground hover:text-destructive transition-colors p-1">
                <Cross1Icon class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <p v-if="error" class="text-sm text-destructive bg-destructive/10 px-3 py-2 rounded border border-destructive/20 mb-4">{{ error }}</p>

      <div class="flex justify-between">
        <Button variant="outline" @click="prevStep">
          <ChevronLeftIcon class="w-4 h-4 mr-1" /> Back
        </Button>
        <Button :disabled="loading || pairs.length === 0" @click="createRun" size="lg"
                class="shadow-[0_0_20px_rgba(var(--primary),0.3)]">
          <span v-if="loading" class="flex items-center gap-2">
            <span class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent"></span>
            Starting...
          </span>
          <span v-else class="flex items-center gap-2 font-bold">
            <PlayIcon class="w-5 h-5" /> Launch Benchmark
          </span>
        </Button>
      </div>
    </Card>
  </div>
</template>
