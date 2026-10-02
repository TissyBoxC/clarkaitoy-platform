import { createRouter, createWebHistory } from 'vue-router'

import DashboardPage from '@/features/dashboard/presentation/DashboardPage.vue'
import DevicePage from '@/features/device/presentation/DevicePage.vue'
import UITextPage from '@/features/ui_text/presentation/UITextPage.vue'
import LoginPage from '@/features/auth/presentation/LoginPage.vue'
import AiAccountsPage from '@/features/ai_gateway/presentation/AiAccountsPage.vue'
import { useAuthStore } from '@/features/auth/application/authStore'

/// Feature routes are registered here so removing a feature only changes its
/// page import and route entry.
export function createAdminRouter() {
  const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
      {
        path: '/login',
        name: 'login',
        component: LoginPage,
        meta: { public: true },
      },
      {
        path: '/',
        name: 'dashboard',
        component: DashboardPage,
      },
      {
        path: '/ai-accounts',
        name: 'ai-accounts',
        component: AiAccountsPage,
      },
      {
        path: '/devices',
        name: 'devices',
        component: DevicePage,
      },
      {
        path: '/ui-text',
        name: 'ui-text',
        component: UITextPage,
      },
    ],
  })

  router.beforeEach(async (to) => {
    const authStore = useAuthStore()
    if (authStore.isRestoring) {
      await authStore.restoreSession()
    }
    if (to.meta.public === true) {
      return authStore.isSignedIn ? '/' : true
    }
    return authStore.isSignedIn ? true : '/login'
  })

  return router
}
