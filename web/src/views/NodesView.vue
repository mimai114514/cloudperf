<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { CheckIcon, CopyIcon, PlusIcon, CodeIcon, Cross2Icon } from '@radix-icons/vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import { api, type NodeItem } from '@/api/client'
import { cn } from '@/lib/utils'

const name = ref('')
const nodes = ref<NodeItem[]>([])
const loading = ref(false)
const error = ref('')
const copyMessage = ref('')
const createdNode = ref<{ id: string; name: string; token: string } | null>(null)

async function load() {
  loading.value = true
  try {
    const data = await api.listNodes()
    nodes.value = data.items
  } finally {
    loading.value = false
  }
}

async function createNode() {
  if (!name.value.trim()) return
  error.value = ''
  copyMessage.value = ''
  try {
    const res = await api.createNode(name.value.trim())
    createdNode.value = {
      id: res.node.id,
      name: res.node.name,
      token: res.token,
    }
    name.value = ''
    await load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

async function copy(text: string, label: string) {
  try {
    await navigator.clipboard.writeText(text)
    copyMessage.value = `${label} copied to clipboard!`
    setTimeout(() => { copyMessage.value = '' }, 3000)
  } catch {
    copyMessage.value = `Failed to copy ${label}`
  }
}

function closeCreatedNode() {
  createdNode.value = null
  copyMessage.value = ''
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 animate-fade-in relative">
    
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Nodes</h2>
        <p class="text-sm text-muted-foreground mt-1">Manage your testing nodes across different regions.</p>
      </div>
    </div>

    <div class="grid gap-6 md:grid-cols-3">
      <!-- Create Node Card -->
      <Card class="md:col-span-1 border-primary/20 bg-primary/5">
        <h3 class="mb-4 text-lg font-semibold flex items-center gap-2">
          <PlusIcon class="w-5 h-5 text-primary" /> New Node
        </h3>
        <div class="flex flex-col gap-3">
          <Input v-model="name" placeholder="E.g., us-west-1-aws" @keydown.enter="createNode"/>
          <Button @click="createNode" class="w-full flex justify-center gap-2">
            Create Node
          </Button>
        </div>
        <p v-if="error" class="mt-3 text-sm text-destructive font-medium bg-destructive/10 px-3 py-2 rounded-md">{{ error }}</p>
      </Card>

      <!-- Nodes List Card -->
      <Card class="md:col-span-2 overflow-hidden flex flex-col p-0">
        <div class="p-6 border-b border-white/5 bg-white/[0.02]">
          <h3 class="text-lg font-semibold flex items-center gap-2">
            <CodeIcon class="w-5 h-5 text-muted-foreground" /> Registered Nodes
          </h3>
        </div>
        
        <div class="flex-1 overflow-x-auto">
          <table class="w-full text-sm text-left whitespace-nowrap">
            <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
              <tr>
                <th class="px-6 py-4 font-medium tracking-wider">Node Name</th>
                <th class="px-6 py-4 font-medium tracking-wider">Public IP</th>
                <th class="px-6 py-4 font-medium tracking-wider">Status</th>
                <th class="px-6 py-4 font-medium tracking-wider text-right">Heartbeat</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/5 bg-transparent">
              <tr v-if="loading && nodes.length === 0">
                <td colspan="4" class="px-6 py-8 text-center text-muted-foreground">Loading nodes data...</td>
              </tr>
              <tr v-else-if="nodes.length === 0">
                <td colspan="4" class="px-6 py-8 text-center text-muted-foreground">No nodes registered yet.</td>
              </tr>
              <tr v-for="n in nodes" :key="n.id" class="transition-colors hover:bg-white/[0.02] group">
                <td class="px-6 py-4 font-medium text-foreground">{{ n.name }}</td>
                <td class="px-6 py-4 font-mono text-xs text-muted-foreground group-hover:text-foreground transition-colors">{{ n.public_ip || 'N/A' }}</td>
                <td class="px-6 py-4">
                  <div class="inline-flex items-center gap-2 px-2.5 py-1 rounded-full text-xs font-semibold"
                       :class="n.status === 'online' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-destructive/10 text-destructive border border-destructive/20'">
                    <span class="w-1.5 h-1.5 rounded-full" :class="n.status === 'online' ? 'bg-emerald-400 animate-pulse' : 'bg-destructive'"></span>
                    {{ n.status === 'online' ? 'Online' : 'Offline' }}
                  </div>
                </td>
                <td class="px-6 py-4 text-right text-muted-foreground tabular-nums">{{ n.last_heartbeat_at || 'Never' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </Card>
    </div>

    <!-- Creation Success Modal -->
    <transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div v-if="createdNode" class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6">
        <div class="absolute inset-0 bg-background/80 backdrop-blur-sm transition-opacity" @click="closeCreatedNode"></div>
        
        <div class="relative w-full max-w-lg rounded-2xl border border-white/10 bg-card p-6 shadow-2xl overflow-hidden glass-panel">
          <!-- decorative blur -->
          <div class="absolute -top-10 -right-10 w-40 h-40 bg-emerald-500/20 rounded-full blur-3xl pointer-events-none"></div>

          <div class="flex items-start justify-between mb-6 relative z-10">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-full bg-emerald-500/20 text-emerald-400 ring-1 ring-emerald-500/30">
                <CheckIcon class="h-6 w-6" />
              </div>
              <div>
                <h3 class="text-xl font-semibold text-foreground">Node Created</h3>
                <p class="text-sm text-muted-foreground">Save the following credentials immediately.</p>
              </div>
            </div>
            <button @click="closeCreatedNode" class="text-muted-foreground hover:text-foreground transition-colors rounded-full p-1 hover:bg-white/10">
              <Cross2Icon class="w-5 h-5"/>
            </button>
          </div>

          <div class="space-y-4 relative z-10">
            <!-- NODE ID -->
            <div class="rounded-xl border border-white/10 bg-black/40 p-4">
              <div class="flex items-center justify-between mb-2">
                <span class="text-xs uppercase tracking-wider font-semibold text-muted-foreground">NODE_ID</span>
                <Button variant="ghost" size="sm" class="h-7 px-2 text-xs" @click="copy(createdNode.id, 'NODE_ID')">
                  <CopyIcon class="w-3 h-3 mr-1" /> Copy
                </Button>
              </div>
              <div class="font-mono text-sm break-all text-sky-400 bg-sky-400/10 rounded-md p-2 border border-sky-400/20">{{ createdNode.id }}</div>
            </div>

            <!-- NODE TOKEN -->
            <div class="rounded-xl border border-amber-500/30 bg-amber-500/5 p-4 relative overflow-hidden">
               <div class="absolute inset-0 bg-gradient-to-r from-amber-500/10 to-transparent pointer-events-none"></div>
              <div class="flex items-center justify-between mb-2 relative z-10">
                <span class="text-xs uppercase tracking-wider font-semibold text-amber-500/70">NODE_TOKEN</span>
                <Button variant="outline" size="sm" class="h-7 px-2 text-xs border-amber-500/30 text-amber-400 hover:bg-amber-500/20 hover:text-amber-300" @click="copy(createdNode.token, 'NODE_TOKEN')">
                  <CopyIcon class="w-3 h-3 mr-1" /> Copy
                </Button>
              </div>
              <div class="font-mono text-sm break-all text-amber-400 bg-black/40 rounded-md p-2 relative z-10 border border-black/50">{{ createdNode.token }}</div>
            </div>
          </div>

          <div class="mt-6 flex items-center justify-between relative z-10">
            <p class="text-xs text-amber-500/80 font-medium">
              ⚠️ The token is only shown once. Keep it secret.
            </p>
            <p v-if="copyMessage" class="text-xs text-emerald-400 animate-slide-up">{{ copyMessage }}</p>
          </div>
          
          <div class="mt-6 pt-6 border-t border-white/5 flex justify-end relative z-10">
             <Button @click="closeCreatedNode">Done</Button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>
