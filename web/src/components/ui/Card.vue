<script setup lang="ts">
import { cn } from '@/lib/utils'
defineProps<{ class?: any }>()
</script>

<template>
  <section :class="cn('rounded-xl border border-white/5 bg-card/40 backdrop-blur-xl p-6 shadow-lg transition-all duration-300 relative overflow-hidden', $props.class)">
    <div class="relative z-10">
      <slot />
    </div>
  </section>
</template>
