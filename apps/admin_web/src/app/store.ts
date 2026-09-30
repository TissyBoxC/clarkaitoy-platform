import { createPinia } from 'pinia'

/// Creates the global Pinia instance for the admin application.
export function createAdminStore() {
  return createPinia()
}
