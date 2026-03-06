<script setup lang="ts">
import { cn } from '@/lib/utils'
defineProps<{ class?: any }>()
</script>

<template>
  <section :class="cn('rounded-xl border border-white/5 bg-card/40 backdrop-blur-xl p-6 shadow-lg transition-all duration-300 hover:shadow-xl hover:border-white/10 hover:-translate-y-0.5 relative overflow-hidden group', $props.class)">
    <!-- 卡片内部微微发光的装饰元素 -->
    <div class="absolute inset-0 bg-gradient-to-br from-white/5 to-transparent opacity-0 transition-opacity duration-500 group-hover:opacity-100 pointer-events-none"></div>
    <div class="relative z-10">
      <slot />
    </div>
  </section>
</template>
