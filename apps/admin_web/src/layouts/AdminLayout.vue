<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="admin-layout">
    <aside class="admin-sidebar">
      <RouterLink class="brand" to="/">
        <img
          class="brand-avatar"
          src="/brand/sprout/brand_avatar.png"
          alt=""
          width="48"
          height="48"
        />
        <span>
          <strong>如此萌屋</strong>
          <small>芽系列 · 初芽</small>
        </span>
      </RouterLink>
      <nav class="admin-nav" aria-label="Primary navigation">
        <RouterLink to="/">运营概览</RouterLink>
        <RouterLink to="/devices">设备管理</RouterLink>
        <RouterLink to="/ui-text">界面文案</RouterLink>
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
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 32px;
  color: inherit;
  text-decoration: none;
}

.brand-avatar {
  flex: 0 0 auto;
  border-radius: 50%;
}

.brand span {
  display: grid;
  gap: 2px;
}

.brand strong {
  font-size: 18px;
}

.brand small {
  color: #cbe3d8;
  font-size: 12px;
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
