<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="admin-layout">
    <aside class="admin-sidebar">
      <RouterLink class="brand" to="/">Clarkaitoy</RouterLink>
      <nav class="admin-nav" aria-label="Primary navigation">
        <RouterLink to="/">Overview</RouterLink>
        <RouterLink to="/devices">Devices</RouterLink>
      </nav>
    </aside>
    <main class="admin-content">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.admin-layout {
  display: grid;
  min-height: 100vh;
  grid-template-columns: 220px minmax(0, 1fr);
  background: #f3f6f4;
}

.admin-sidebar {
  padding: 24px 16px;
  color: #f7fbf8;
  background: #17483e;
}

.brand {
  display: block;
  margin-bottom: 32px;
  color: inherit;
  font-size: 18px;
  font-weight: 700;
  text-decoration: none;
}

.admin-nav {
  display: grid;
  gap: 8px;
}

.admin-nav a {
  padding: 10px 12px;
  color: inherit;
  text-decoration: none;
  border-radius: 6px;
}

.admin-nav a.router-link-active {
  background: #2f6b5f;
}

.admin-content {
  min-width: 0;
  padding: 28px 32px;
}

@media (max-width: 720px) {
  .admin-layout {
    grid-template-columns: 1fr;
  }

  .admin-sidebar {
    padding: 16px;
  }

  .brand {
    margin-bottom: 16px;
  }

  .admin-nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .admin-content {
    padding: 20px 16px;
  }
}
</style>
