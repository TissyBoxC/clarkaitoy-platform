import { createApp } from 'vue'

import App from '@/App.vue'
import { createAdminRouter } from '@/app/router'
import { createAdminStore } from '@/app/store'

/// Creates the admin application with its router and global store.
export function createAdminApp() {
  const app = createApp(App)

  app.use(createAdminStore())
  app.use(createAdminRouter())

  return app
}
