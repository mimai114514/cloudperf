<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
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
    error.value = '请至少勾选一个方向组合'
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
  <div class="space-y-5">
    <Card>
      <h2 class="mb-3 text-lg font-semibold">创建即时测速任务</h2>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <label class="text-sm">协议
          <select v-model="protocol" class="mt-1 w-full rounded-md border border-slate-700 bg-slate-900 px-2 py-2">
            <option value="tcp">TCP</option>
            <option value="udp">UDP</option>
          </select>
        </label>
        <label class="text-sm">时长(s)
          <Input v-model="duration" type="number" />
        </label>
        <label class="text-sm">并发流
          <Input v-model="parallel" type="number" />
        </label>
        <label class="text-sm">UDP带宽
          <Input v-model="udpBandwidth" placeholder="100M" :disabled="protocol !== 'udp'" />
        </label>
      </div>
    </Card>

    <Card>
      <h2 class="mb-3 text-lg font-semibold">节点组合 Grid（单向）</h2>
      <div class="overflow-auto">
        <table class="w-full min-w-[700px] border-collapse text-xs">
          <thead>
            <tr>
              <th class="border border-slate-800 px-2 py-2 text-left">Source \ Target</th>
              <th v-for="t in nodes" :key="t.id" class="border border-slate-800 px-2 py-2">{{ t.name }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in nodes" :key="s.id">
              <td class="border border-slate-800 px-2 py-2 font-medium">{{ s.name }}</td>
              <td v-for="t in nodes" :key="`${s.id}-${t.id}`" class="border border-slate-800 px-2 py-2 text-center">
                <template v-if="s.id !== t.id">
                  <input v-model="selected[`${s.id}->${t.id}`]" type="checkbox" />
                </template>
                <span v-else class="text-slate-600">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="mt-2 text-xs text-slate-400">已选择 {{ pairs.length }} 条单向组合</p>
      <p v-if="error" class="mt-2 text-sm text-rose-400">{{ error }}</p>
      <div class="mt-3">
        <Button :disabled="loading" @click="createRun">{{ loading ? '创建中...' : '开始测速' }}</Button>
      </div>
    </Card>
  </div>
</template>
