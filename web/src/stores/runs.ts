import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type RunInfo } from '@/api/client'

export const useRunsStore = defineStore('runs', () => {
  const runIds = ref<string[]>(loadFromStorage())
  const runsCache = ref<Map<string, RunInfo>>(new Map())

  function loadFromStorage(): string[] {
    try {
      const raw = localStorage.getItem('cloudperf_run_ids')
      return raw ? JSON.parse(raw) : []
    } catch {
      return []
    }
  }

  function persist() {
    localStorage.setItem('cloudperf_run_ids', JSON.stringify(runIds.value))
  }

  function addRun(id: string) {
    if (!runIds.value.includes(id)) {
      runIds.value.unshift(id)
      if (runIds.value.length > 50) runIds.value = runIds.value.slice(0, 50)
      persist()
    }
  }

  async function fetchAll() {
    const results: RunInfo[] = []
    for (const id of runIds.value) {
      try {
        const run = await api.getRun(id)
        runsCache.value.set(id, run)
        results.push(run)
      } catch {
        // run may have been deleted, skip
      }
    }
    return results
  }

  async function fetchOne(id: string) {
    const run = await api.getRun(id)
    runsCache.value.set(id, run)
    addRun(id)
    return run
  }

  return { runIds, runsCache, addRun, fetchAll, fetchOne }
})
