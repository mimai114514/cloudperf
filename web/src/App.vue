<script setup lang="ts">
import { RouterView, RouterLink, useRoute } from 'vue-router'
import {
  DashboardIcon, DesktopIcon, RocketIcon, BarChartIcon, ExitIcon
} from '@radix-icons/vue'
import { useAuthStore } from '@/stores/auth'
import { computed } from 'vue'

const auth = useAuthStore()
const route = useRoute()

const navItems = [
  { to: '/', label: 'Dashboard', icon: DashboardIcon },
  { to: '/nodes', label: 'Nodes', icon: DesktopIcon },
  { to: '/runs', label: 'Runs', icon: RocketIcon },
  { to: '/results', label: 'Results', icon: BarChartIcon },
]

const isLoggedIn = computed(() => !!auth.userId)
const isLoginPage = computed(() => route.path === '/login')

async function logout() {
  await auth.logout()
  window.location.href = '/login'
}
</script>

<template>
  <div class="min-h-screen relative">
    <!-- Background glow effects -->
    <div class="fixed inset-0 -z-10 overflow-hidden pointer-events-none">
      <div class="absolute -top-1/4 -left-1/4 w-1/2 h-1/2 rounded-full bg-primary/[0.03] blur-[100px]"></div>
      <div class="absolute -bottom-1/4 -right-1/4 w-1/2 h-1/2 rounded-full bg-sky-500/[0.02] blur-[100px]"></div>
    </div>

    <!-- Header -->
    <header v-if="isLoggedIn && !isLoginPage"
            class="sticky top-0 z-50 w-full border-b border-white/5 bg-background/80 backdrop-blur-xl">
      <div class="container flex h-14 items-center">
        <!-- Logo -->
        <RouterLink to="/" class="flex items-center gap-2.5 mr-8 group">
          <div class="h-7 w-7 rounded-lg bg-gradient-to-br from-primary to-sky-400 flex items-center justify-center shadow-lg shadow-primary/20 transition-transform group-hover:scale-110">
            <span class="text-xs font-black text-white">CP</span>
          </div>
          <span class="font-bold text-sm tracking-tight hidden sm:inline">CloudPerf</span>
        </RouterLink>

        <!-- Navigation -->
        <nav class="flex items-center gap-1 flex-1">
          <RouterLink v-for="item in navItems" :key="item.to" :to="item.to"
            class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium transition-all"
            :class="[
              route.path === item.to || (item.to !== '/' && route.path.startsWith(item.to))
                ? 'text-primary bg-primary/10'
                : 'text-muted-foreground hover:text-foreground hover:bg-white/5'
            ]">
            <component :is="item.icon" class="w-4 h-4" />
            <span class="hidden sm:inline">{{ item.label }}</span>
          </RouterLink>
        </nav>

        <!-- User actions -->
        <button @click="logout"
          class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors">
          <ExitIcon class="w-4 h-4" />
          <span class="hidden sm:inline">Logout</span>
        </button>
      </div>
    </header>

    <!-- Main content -->
    <main :class="isLoginPage ? '' : 'container py-8'">
      <RouterView v-slot="{ Component, route: viewRoute }">
        <transition name="page" mode="out-in">
          <component :is="Component" :key="viewRoute.path" />
        </transition>
      </RouterView>
    </main>
  </div>
</template>

<style>
.page-enter-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.page-leave-active {
  transition: opacity 0.1s ease;
}
.page-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
.page-leave-to {
  opacity: 0;
}
</style>
