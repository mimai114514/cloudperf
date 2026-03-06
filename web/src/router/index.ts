import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import NodesView from '@/views/NodesView.vue'
import RunsListView from '@/views/RunsListView.vue'
import NewRunView from '@/views/NewRunView.vue'
import RunDetailView from '@/views/RunDetailView.vue'
import ResultsView from '@/views/ResultsView.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginView },
    { path: '/', component: DashboardView },
    { path: '/nodes', component: NodesView },
    { path: '/runs', component: RunsListView },
    { path: '/runs/new', component: NewRunView },
    { path: '/runs/:id', component: RunDetailView },
    { path: '/results', component: ResultsView },
    { path: '/:pathMatch(.*)*', redirect: '/' },
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
    return '/'
  }
  return true
})

export default router
