<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import { api, type NodeItem } from '@/api/client'

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
  error.value = ''
  copyMessage.value = ''
  try {
    const res = await api.createNode(name.value)
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
    copyMessage.value = `${label} 已复制`
  } catch {
    copyMessage.value = `${label} 复制失败，请手动复制`
  }
}

function closeCreatedNode() {
  createdNode.value = null
  copyMessage.value = ''
}

onMounted(load)
</script>

<template>
  <div class="space-y-5">
    <Card>
      <h2 class="mb-3 text-lg font-semibold">创建节点</h2>
      <div class="flex gap-2">
        <Input v-model="name" placeholder="节点名称" />
        <Button @click="createNode">创建</Button>
      </div>
      <p v-if="error" class="mt-2 text-sm text-rose-400">{{ error }}</p>
    </Card>

    <Card>
      <h2 class="mb-3 text-lg font-semibold">节点列表</h2>
      <div v-if="loading" class="text-sm text-slate-300">加载中...</div>
      <table v-else class="w-full text-sm">
        <thead class="text-slate-400">
          <tr>
            <th class="py-2 text-left">名称</th>
            <th class="py-2 text-left">公网IP</th>
            <th class="py-2 text-left">状态</th>
            <th class="py-2 text-left">心跳</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="n in nodes" :key="n.id" class="border-t border-slate-800">
            <td class="py-2">{{ n.name }}</td>
            <td class="py-2">{{ n.public_ip || '-' }}</td>
            <td class="py-2">{{ n.status }}</td>
            <td class="py-2">{{ n.last_heartbeat_at || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>

  <div v-if="createdNode" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/75 px-4">
    <div class="w-full max-w-xl rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h3 class="text-lg font-semibold text-white">节点创建成功</h3>
          <p class="mt-1 text-sm text-slate-300">
            节点 <span class="font-medium text-white">{{ createdNode.name }}</span> 已创建，请立即保存以下信息。
          </p>
        </div>
        <Button @click="closeCreatedNode">关闭</Button>
      </div>

      <div class="mt-5 space-y-4">
        <div class="rounded-lg border border-slate-800 bg-slate-950/70 p-4">
          <div class="mb-2 text-xs uppercase tracking-wide text-slate-400">NODE_ID</div>
          <div class="break-all font-mono text-sm text-emerald-300">{{ createdNode.id }}</div>
          <div class="mt-3">
            <Button @click="copy(createdNode.id, 'NODE_ID')">复制 NODE_ID</Button>
          </div>
        </div>

        <div class="rounded-lg border border-amber-700/50 bg-amber-950/20 p-4">
          <div class="mb-2 text-xs uppercase tracking-wide text-amber-300">NODE_TOKEN</div>
          <div class="break-all font-mono text-sm text-amber-200">{{ createdNode.token }}</div>
          <div class="mt-3">
            <Button @click="copy(createdNode.token, 'NODE_TOKEN')">复制 NODE_TOKEN</Button>
          </div>
        </div>
      </div>

      <p class="mt-4 text-sm text-amber-300">`NODE_TOKEN` 只会展示这一次，关闭后将无法再次查看。</p>
      <p v-if="copyMessage" class="mt-2 text-sm text-emerald-300">{{ copyMessage }}</p>
    </div>
  </div>
</template>
