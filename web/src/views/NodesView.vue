<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  PlusIcon, CopyIcon, Cross2Icon, CheckCircledIcon,
  DesktopIcon, CodeIcon
} from '@radix-icons/vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { api, type NodeItem } from '@/api/client'

const nodes = ref<NodeItem[]>([])
const newName = ref('')
const error = ref('')
const loading = ref(false)

// Modal state
const showModal = ref(false)
const createdNode = ref<{ id: string; name: string; token: string } | null>(null)
const copiedField = ref<string | null>(null)

async function loadNodes() {
  const res = await api.listNodes()
  nodes.value = res.items
}

async function createNode() {
  const name = newName.value.trim()
  if (!name) { error.value = 'Node name is required.'; return }
  error.value = ''
  loading.value = true
  try {
    const res = await api.createNode(name)
    createdNode.value = { id: res.node.id, name: res.node.name, token: res.token }
    showModal.value = true
    newName.value = ''
    await loadNodes()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function copyToClipboard(text: string, field: string) {
  navigator.clipboard.writeText(text)
  copiedField.value = field
  setTimeout(() => { copiedField.value = null }, 2000)
}

function agentCommand(nodeId: string, token: string) {
  return `export BACKEND_WS_URL="ws://<YOUR_BACKEND_IP>:8080/agent/v1/ws"
export NODE_ID="${nodeId}"
export NODE_TOKEN="${token}"
export PUBLIC_IP="<THIS_NODE_PUBLIC_IP>"
export AGENT_VERSION="v1"
export HEARTBEAT_INTERVAL_SEC="15"
export RECONNECT_INTERVAL_SEC="5"
export IPERF_BINARY="iperf3"
go run ./cmd/cloudperf-agent`
}

function relativeTime(dateStr?: string) {
  if (!dateStr) return 'Never'
  const diff = Date.now() - new Date(dateStr).getTime()
  if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
  return `${Math.floor(diff / 86400000)}d ago`
}

onMounted(loadNodes)
</script>

<template>
  <div class="space-y-6 animate-fade-in relative z-10">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Nodes</h2>
        <p class="text-sm text-muted-foreground mt-1">Manage your test endpoints.</p>
      </div>
    </div>

    <!-- Create Node -->
    <Card class="border-primary/20 bg-primary/5">
      <div class="flex items-center gap-2 mb-4">
        <PlusIcon class="w-5 h-5 text-primary" />
        <h3 class="font-semibold">Register New Node</h3>
      </div>
      <div class="flex gap-3">
        <Input v-model="newName" placeholder="Enter node name (e.g., us-east-1)" class="flex-1"
               @keyup.enter="createNode" />
        <Button @click="createNode" :disabled="loading">
          <PlusIcon class="w-4 h-4 mr-1.5" />
          {{ loading ? 'Creating...' : 'Create' }}
        </Button>
      </div>
      <p v-if="error" class="mt-3 text-sm text-destructive bg-destructive/10 px-3 py-2 rounded border border-destructive/20">{{ error }}</p>
    </Card>

    <!-- Node List -->
    <Card class="p-0 overflow-hidden">
      <div class="p-5 border-b border-white/5 bg-white/[0.02] flex items-center gap-2">
        <DesktopIcon class="w-5 h-5 text-primary" />
        <h3 class="font-semibold">Registered Nodes</h3>
        <span class="ml-auto text-xs text-muted-foreground tabular-nums">{{ nodes.length }} total</span>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs uppercase bg-white/[0.03] text-muted-foreground border-b border-white/5">
            <tr>
              <th class="px-6 py-3 font-medium tracking-wider">Name</th>
              <th class="px-6 py-3 font-medium tracking-wider">Status</th>
              <th class="px-6 py-3 font-medium tracking-wider">Public IP</th>
              <th class="px-6 py-3 font-medium tracking-wider">Last Heartbeat</th>
              <th class="px-6 py-3 font-medium tracking-wider">Agent Version</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-if="nodes.length === 0">
              <td colspan="5" class="px-6 py-12 text-center text-muted-foreground">
                <DesktopIcon class="w-10 h-10 mx-auto mb-3 opacity-20" />
                <p class="font-medium">No nodes registered yet</p>
                <p class="text-xs mt-1">Use the form above to register your first test endpoint.</p>
              </td>
            </tr>
            <tr v-for="node in nodes" :key="node.id" class="transition-colors hover:bg-white/[0.02]">
              <td class="px-6 py-4 font-medium">{{ node.name }}</td>
              <td class="px-6 py-4">
                <span class="inline-flex items-center gap-1.5">
                   <span class="w-2 h-2 rounded-full" :class="node.status === 'online' ? 'bg-emerald-400 shadow-sm shadow-emerald-400/50' : 'bg-zinc-500'"></span>
                   <span class="text-xs font-medium capitalize" :class="node.status === 'online' ? 'text-emerald-400' : 'text-muted-foreground'">{{ node.status || 'offline' }}</span>
                </span>
              </td>
              <td class="px-6 py-4 font-mono text-xs text-muted-foreground">{{ node.public_ip || '-' }}</td>
              <td class="px-6 py-4 text-xs text-muted-foreground tabular-nums">{{ relativeTime(node.last_heartbeat_at) }}</td>
              <td class="px-6 py-4 text-xs text-muted-foreground">{{ node.agent_version || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>

    <!-- Created Node Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showModal = false"></div>
        <div class="relative w-full max-w-lg rounded-2xl border border-white/10 bg-card/95 backdrop-blur-xl shadow-2xl p-6 space-y-5 animate-slide-up">
          <!-- Close -->
          <button @click="showModal = false" class="absolute top-4 right-4 text-muted-foreground hover:text-foreground transition-colors p-1">
            <Cross2Icon class="w-5 h-5" />
          </button>

          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center">
              <CheckCircledIcon class="w-5 h-5 text-emerald-400" />
            </div>
            <div>
              <h3 class="font-bold text-lg">Node Created Successfully</h3>
              <p class="text-sm text-muted-foreground">Save these credentials — the token is shown only once.</p>
            </div>
          </div>

          <!-- Credentials -->
          <div v-if="createdNode" class="space-y-3">
            <div>
              <label class="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1 block">Node ID</label>
              <div class="flex items-center gap-2">
                <code class="flex-1 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm font-mono break-all select-all">{{ createdNode.id }}</code>
                <Button variant="ghost" size="icon" @click="copyToClipboard(createdNode.id, 'id')">
                  <CopyIcon v-if="copiedField !== 'id'" class="w-4 h-4" />
                  <CheckCircledIcon v-else class="w-4 h-4 text-emerald-400" />
                </Button>
              </div>
            </div>
            <div>
              <label class="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1 block">Token <span class="text-destructive">(ONE-TIME DISPLAY)</span></label>
              <div class="flex items-center gap-2">
                <code class="flex-1 bg-white/5 border border-destructive/20 rounded-lg px-3 py-2 text-sm font-mono break-all select-all">{{ createdNode.token }}</code>
                <Button variant="ghost" size="icon" @click="copyToClipboard(createdNode.token, 'token')">
                  <CopyIcon v-if="copiedField !== 'token'" class="w-4 h-4" />
                  <CheckCircledIcon v-else class="w-4 h-4 text-emerald-400" />
                </Button>
              </div>
            </div>

            <!-- Agent startup command -->
            <div>
              <label class="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1 flex items-center gap-1.5">
                <CodeIcon class="w-3.5 h-3.5" /> Agent Startup Command
              </label>
              <div class="relative">
                <pre class="bg-white/5 border border-white/10 rounded-lg px-3 py-2.5 text-xs font-mono whitespace-pre-wrap break-all text-muted-foreground overflow-x-auto max-h-[150px]">{{ agentCommand(createdNode.id, createdNode.token) }}</pre>
                <Button variant="ghost" size="sm" class="absolute top-1.5 right-1.5"
                        @click="copyToClipboard(agentCommand(createdNode.id, createdNode.token), 'cmd')">
                  <CopyIcon v-if="copiedField !== 'cmd'" class="w-3.5 h-3.5 mr-1" />
                  <CheckCircledIcon v-else class="w-3.5 h-3.5 mr-1 text-emerald-400" />
                  {{ copiedField === 'cmd' ? 'Copied!' : 'Copy' }}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
