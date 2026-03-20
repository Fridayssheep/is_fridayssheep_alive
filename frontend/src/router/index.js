// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import GithubStatus from '../views/GithubStatus.vue'
import WorkstationStatus from '../views/WorkstationStatus.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/github',
      name: 'github',
      component: GithubStatus,
    },
    {
      path: '/workstation',
      name: 'workstation',
      component: WorkstationStatus,
    },
  ],
})

export default router
