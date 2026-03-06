<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { DesktopIcon, ActivityLogIcon, TargetIcon } from '@radix-icons/vue'

const route = useRoute()
const isLogin = computed(() => route.path === '/login')
</script>

<template>
  <div class="min-h-screen text-foreground bg-background selection:bg-primary/30 antialiased font-sans flex flex-col relative z-0">
    <!-- Background Glow Elements (behind everything) -->
    <div class="pointer-events-none fixed inset-0 flex justify-center overflow-hidden z-[-1]">
       <div class="w-[50rem] h-[50rem] rounded-full bg-primary/5 blur-3xl opacity-50 -translate-y-[40%]"></div>
    </div>

    <!-- Header -->
    <header v-if="!isLogin" class="sticky top-0 z-50 w-full border-b border-white/5 bg-background/60 backdrop-blur-xl supports-[backdrop-filter]:bg-background/40">
      <div class="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <div class="flex items-center gap-8">
          <RouterLink to="/" class="flex items-center gap-2 transition-opacity hover:opacity-80">
            <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/20 text-primary ring-1 ring-primary/30 shadow-[0_0_15px_rgba(var(--primary),0.3)]">
              <DesktopIcon class="h-5 w-5" />
            </div>
            <span class="text-xl font-bold tracking-tight bg-gradient-to-br from-white to-white/40 bg-clip-text text-transparent">CloudPerf</span>
          </RouterLink>
          
          <nav class="hidden md:flex items-center gap-2 text-sm font-medium">
            <RouterLink 
              class="flex items-center gap-2 rounded-md px-3 py-2 text-muted-foreground transition-all duration-300 hover:bg-white/5 hover:text-foreground" 
              active-class="!bg-white/10 !text-foreground shadow-sm ring-1 ring-white/10"
              to="/nodes">
              <DesktopIcon class="h-4 w-4" /> 节点
            </RouterLink>
            <RouterLink 
              class="flex items-center gap-2 rounded-md px-3 py-2 text-muted-foreground transition-all duration-300 hover:bg-white/5 hover:text-foreground" 
              active-class="!bg-white/10 !text-foreground shadow-sm ring-1 ring-white/10"
              to="/runs/new">
              <TargetIcon class="h-4 w-4" /> 即时测速
            </RouterLink>
            <RouterLink 
              class="flex items-center gap-2 rounded-md px-3 py-2 text-muted-foreground transition-all duration-300 hover:bg-white/5 hover:text-foreground" 
              active-class="!bg-white/10 !text-foreground shadow-sm ring-1 ring-white/10"
              to="/results">
              <ActivityLogIcon class="h-4 w-4" /> 历史结果
            </RouterLink>
          </nav>
        </div>
      </div>
    </header>

    <!-- Main Content with Transition -->
    <main class="flex-1 mx-auto w-full max-w-7xl px-4 sm:px-6 py-8 relative">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<style>
.page-enter-active,
.page-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
  transform: translateY(6px) scale(0.99);
}
</style>
