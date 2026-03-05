<script setup lang="ts">
defineProps<{ type?: 'button' | 'submit'; disabled?: boolean }>()
</script>

<template>
  <button
    :type="type ?? 'button'"
    :disabled="disabled"
    class="inline-flex items-center justify-center rounded-md bg-sky-500 px-3 py-2 text-sm font-medium text-slate-950 transition hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-50"
  >
    <slot />
  </button>
</template>
