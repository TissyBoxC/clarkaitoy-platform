import { createRouter, createWebHistory } from 'vue-router'

import DashboardPage from '@/features/dashboard/presentation/DashboardPage.vue'
import DevicePage from '@/features/device/presentation/DevicePage.vue'
import UITextPage from '@/features/ui_text/presentation/UITextPage.vue'

/// Feature routes are registered here so removing a feature only changes its
/// page import and route entry.
export function createAdminRouter() {
  return createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
      {
        path: '/',
        name: 'dashboard',
        component: DashboardPage,
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
}
