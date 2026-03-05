<script setup lang="ts">
defineProps<{ modelValue?: string | number; type?: string; placeholder?: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <input
    :type="type ?? 'text'"
    :value="modelValue"
    :placeholder="placeholder"
    class="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm outline-none ring-sky-500 focus:ring"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
