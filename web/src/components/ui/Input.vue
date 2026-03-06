<script setup lang="ts">
import { cn } from '@/lib/utils'

defineProps<{ modelValue?: string | number; type?: string; placeholder?: string; disabled?: boolean; class?: any }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <input
    :type="type ?? 'text'"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :class="cn(
      'flex h-9 w-full rounded-md border border-input bg-background/50 backdrop-blur-sm px-3 py-1 text-sm shadow-sm transition-all file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:border-ring/50 disabled:cursor-not-allowed disabled:opacity-50 hover:bg-background/80 hover:border-input/80',
      $props.class
    )"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
