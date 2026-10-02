<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'

import { useAuthStore } from '@/features/auth/application/authStore'

const router = useRouter()
const authStore = useAuthStore()

async function logout(): Promise<void> {
  await authStore.logout()
  await router.replace('/login')
}
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
      <nav class="admin-nav" aria-label="主要导航">
        <RouterLink to="/">运营概览</RouterLink>
        <RouterLink to="/ai-accounts">家长 AI 账号</RouterLink>
        <RouterLink to="/devices">设备管理</RouterLink>
        <RouterLink to="/ui-text">界面文案</RouterLink>
      </nav>
      <div class="account-panel">
        <p>{{ authStore.account?.displayName }}</p>
        <small>{{ authStore.account?.email }}</small>
        <button type="button" @click="logout">退出登录</button>
      </div>
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
  grid-template-columns: 248px minmax(0, 1fr);
  background: #fffbfc;
}

.admin-sidebar {
  display: flex;
  flex-direction: column;
  padding: 24px 18px;
  border-right: 1px solid #f0bdcb;
  background: #ffffff;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 32px;
  color: #4a2e3b;
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
  color: #6b4f5a;
  font-size: 12px;
}

.admin-nav {
  display: grid;
  gap: 8px;
}

.admin-nav a {
  padding: 11px 14px;
  color: #6b4f5a;
  text-decoration: none;
  border-radius: 14px;
}

.admin-nav a.router-link-active {
  background: #fff0f4;
  color: #c94175;
  font-weight: 700;
}

.account-panel {
  display: grid;
  gap: 4px;
  margin-top: auto;
  padding-top: 20px;
  border-top: 1px solid #f7d9e2;
}

.account-panel p,
.account-panel small {
  margin: 0;
  overflow-wrap: anywhere;
}

.account-panel p {
  color: #4a2e3b;
  font-weight: 700;
}

.account-panel small {
  color: #6b4f5a;
}

.account-panel button {
  min-height: 38px;
  margin-top: 10px;
  border: 1px solid #f0bdcb;
  border-radius: 19px;
  background: #ffffff;
  color: #c94175;
  font: inherit;
  cursor: pointer;
}

.admin-content {
  min-width: 0;
  padding: 32px;
}

@media (max-width: 760px) {
  .admin-layout {
    grid-template-columns: 1fr;
  }

  .admin-sidebar {
    position: static;
    border-right: 0;
    border-bottom: 1px solid #f0bdcb;
  }

  .brand {
    margin-bottom: 18px;
  }

  .admin-nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .account-panel {
    margin-top: 18px;
  }

  .admin-content {
    padding: 22px 16px;
  }
}
</style>
