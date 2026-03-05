import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import NodesView from '@/views/NodesView.vue'
import NewRunView from '@/views/NewRunView.vue'
import RunDetailView from '@/views/RunDetailView.vue'
import ResultsView from '@/views/ResultsView.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginView },
    { path: '/nodes', component: NodesView },
    { path: '/runs/new', component: NewRunView },
    { path: '/runs/:id', component: RunDetailView },
    { path: '/results', component: ResultsView },
    { path: '/:pathMatch(.*)*', redirect: '/nodes' },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (auth.userId === null) {
    await auth.refresh()
  }
  if (to.path !== '/login' && !auth.userId) {
    return '/login'
  }
  if (to.path === '/login' && auth.userId) {
    return '/nodes'
  }
  return true
})

export default router
