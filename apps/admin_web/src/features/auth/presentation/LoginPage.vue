<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/features/auth/application/authStore'

const router = useRouter()
const authStore = useAuthStore()
const email = ref('')
const password = ref('')
const verificationCode = ref('')
const isSubmitting = ref(false)

async function submit(): Promise<void> {
  if (authStore.isMFARequired) {
    if (!/^\d{6}$/.test(verificationCode.value.trim())) {
      return
    }
    isSubmitting.value = true
    const succeeded = await authStore.completeMFA(verificationCode.value.trim())
    isSubmitting.value = false
    if (succeeded) {
      await router.replace('/')
    }
    return
  }
  if (email.value.trim() === '' || password.value === '') {
    return
  }
  isSubmitting.value = true
  const succeeded = await authStore.login(email.value.trim(), password.value)
  isSubmitting.value = false
  if (succeeded) {
    await router.replace('/')
  }
}

function returnToPassword(): void {
  verificationCode.value = ''
  authStore.cancelMFA()
}

function handleCodeInput(event: Event): void {
  const input = event.target as HTMLInputElement
  verificationCode.value = input.value.replace(/\D/g, '').slice(0, 6)
}
</script>

<template>
  <main class="login-page">
    <section class="login-panel">
      <img
        class="brand-avatar"
        src="/brand/sprout/brand_avatar.png"
        alt=""
        width="72"
        height="72"
      />
      <p class="brand-name">如此萌屋</p>
      <h1>管理后台登录</h1>
      <p class="description">
        {{ authStore.isMFARequired ? '请输入验证器中的 6 位验证码。' : '登录后管理家长账号、AI 额度和设备绑定。' }}
      </p>

      <form class="login-form" @submit.prevent="submit">
        <template v-if="authStore.isMFARequired">
          <label>
            <span>6 位验证码</span>
            <input
              :value="verificationCode"
              type="text"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="6"
              required
              @input="handleCodeInput"
            />
          </label>
        </template>
        <template v-else>
          <label>
            <span>管理员邮箱</span>
            <input
              v-model="email"
              type="email"
              autocomplete="username"
              required
            />
          </label>
          <label>
            <span>密码</span>
            <input
              v-model="password"
              type="password"
              autocomplete="current-password"
              required
            />
          </label>
        </template>
        <p v-if="authStore.error" class="error-message">
          {{ authStore.error.message }}
        </p>
        <button type="submit" :disabled="isSubmitting">
          {{ isSubmitting ? '正在登录…' : authStore.isMFARequired ? '验证并登录' : '登录管理后台' }}
        </button>
        <button
          v-if="authStore.isMFARequired"
          type="button"
          class="secondary-button"
          :disabled="isSubmitting"
          @click="returnToPassword"
        >
          返回密码登录
        </button>
      </form>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(circle at 10% 10%, #fff2f5 0, transparent 34%),
    #fffbfc;
}

.login-panel {
  width: min(100%, 420px);
  padding: 36px;
  border: 1px solid #f0bdcb;
  border-radius: 28px;
  background: #ffffff;
  box-shadow: 0 18px 48px rgb(217 79 131 / 10%);
  text-align: center;
}

.brand-avatar {
  border-radius: 50%;
}

.brand-name {
  margin: 12px 0 4px;
  color: #d94f83;
  font-weight: 700;
}

h1 {
  margin: 0;
  color: #4a2e3b;
  font-size: 28px;
}

.description {
  margin: 10px 0 28px;
  color: #6b4f5a;
}

.login-form {
  display: grid;
  gap: 18px;
  text-align: left;
}

label {
  display: grid;
  gap: 8px;
  color: #4a2e3b;
  font-weight: 600;
}

input {
  min-height: 48px;
  padding: 0 16px;
  border: 1px solid #f0bdcb;
  border-radius: 16px;
  background: #fff2f5;
  color: #4a2e3b;
  font: inherit;
}

input:focus {
  border-color: #d94f83;
  outline: 2px solid rgb(217 79 131 / 18%);
}

button {
  min-height: 50px;
  border: 0;
  border-radius: 25px;
  background: #d94f83;
  color: #ffffff;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.65;
}

.secondary-button {
  border: 1px solid #f0bdcb;
  background: #ffffff;
  color: #b23a68;
}

.error-message {
  margin: 0;
  color: #b3261e;
}
</style>
