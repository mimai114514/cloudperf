<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import { api, type NodeItem } from '@/api/client'

const name = ref('')
const nodes = ref<NodeItem[]>([])
const loading = ref(false)
const newToken = ref('')
const error = ref('')

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
  newToken.value = ''
  try {
    const res = await api.createNode(name.value)
    newToken.value = res.token
    name.value = ''
    await load()
  } catch (e) {
    error.value = (e as Error).message
  }
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
      <p v-if="newToken" class="mt-2 text-sm text-amber-300">一次性 Token：{{ newToken }}</p>
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
</template>
