<script setup lang="ts">
import { onMounted, ref } from 'vue'

import {
  useAiAccountStore,
  type AdminAiAccount,
} from '@/features/ai_gateway/application/aiAccountStore'

const store = useAiAccountStore()
const editing = ref<AdminAiAccount | null>(null)
const modelsText = ref('')
const saveMessage = ref('')

onMounted(() => {
  void store.load()
})

function openEditor(account: AdminAiAccount): void {
  editing.value = { ...account }
  modelsText.value = account.availableModels.join(', ')
  saveMessage.value = ''
}

async function save(): Promise<void> {
  if (editing.value === null) {
    return
  }
  editing.value.availableModels = modelsText.value
    .split(',')
    .map((model) => model.trim())
    .filter(Boolean)
  const succeeded = await store.update(editing.value)
  if (succeeded) {
    saveMessage.value = '已保存'
    editing.value = null
  }
}
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <p class="eyebrow">AI 服务</p>
        <h1>家长 AI 账号</h1>
        <p>这里只管理余额、模型和账号状态，不会显示家长或设备使用的密钥。</p>
      </div>
      <button type="button" :disabled="store.isLoading" @click="store.load">
        重新加载
      </button>
    </header>

    <p v-if="saveMessage" class="success-message">{{ saveMessage }}</p>
    <p v-if="store.error" class="error-message">{{ store.error.message }}</p>

    <div v-if="store.isLoading" class="state-panel">正在读取家长账号…</div>
    <div v-else-if="store.accounts.length === 0" class="state-panel">
      还没有家长账号。家长完成注册后，账号会显示在这里。
    </div>
    <div v-else class="table-shell">
      <table>
        <thead>
          <tr>
            <th>家长账号</th>
            <th>状态</th>
            <th>余额</th>
            <th>同时对话</th>
            <th>可分配模型</th>
            <th>家长已选</th>
            <th>凭证</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="account in store.accounts" :key="account.providerAccountId">
            <td class="account-id">{{ account.providerAccountId }}</td>
            <td>
              <span :class="['status', account.status]">
                {{ account.status === 'active' ? '可用' : '已暂停' }}
              </span>
            </td>
            <td>${{ account.balanceUsd.toFixed(2) }}</td>
            <td>{{ account.concurrencyLimit }}</td>
            <td>{{ account.availableModels.join('、') || '使用默认模型' }}</td>
            <td>
              {{
                account.selectedModels.length === 0
                  ? '全部可用'
                  : account.selectedModels.join('、')
              }}
            </td>
            <td>{{ account.credentialReady ? '已就绪' : '待修复' }}</td>
            <td>
              <button type="button" @click="openEditor(account)">管理</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="editing" class="dialog-backdrop" @click.self="editing = null">
      <form class="dialog" @submit.prevent="save">
        <h2>调整家长 AI 额度</h2>
        <p class="account-id">{{ editing.providerAccountId }}</p>
        <label>
          <span>账号状态</span>
          <select v-model="editing.status">
            <option value="active">可用</option>
            <option value="suspended">暂停使用</option>
          </select>
        </label>
        <label>
          <span>可用余额（美元）</span>
          <input v-model.number="editing.balanceUsd" type="number" min="0" step="0.01" />
        </label>
        <label>
          <span>同时对话数量</span>
          <input v-model.number="editing.concurrencyLimit" type="number" min="1" />
        </label>
        <label>
          <span>可分配模型</span>
          <input v-model="modelsText" placeholder="用逗号分隔，留空使用默认模型" />
        </label>
        <div class="dialog-actions">
          <button type="button" class="secondary" @click="editing = null">
            取消
          </button>
          <button type="submit">保存更改</button>
        </div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 22px;
}

.eyebrow {
  margin: 0 0 6px;
  color: #d94f83;
  font-size: 13px;
  font-weight: 700;
}

h1 {
  margin: 0;
  color: #4a2e3b;
  font-size: 28px;
}

.page-header p:not(.eyebrow) {
  max-width: 680px;
  margin: 8px 0 0;
  color: #6b4f5a;
}

button {
  min-height: 38px;
  padding: 0 16px;
  border: 1px solid #e9a5b8;
  border-radius: 19px;
  background: #ffffff;
  color: #c94175;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.55;
}

.table-shell,
.state-panel {
  overflow: hidden;
  border: 1px solid #f0bdcb;
  border-radius: 20px;
  background: #ffffff;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 14px 16px;
  border-bottom: 1px solid #f7d9e2;
  text-align: left;
}

th {
  color: #6b4f5a;
  font-size: 13px;
}

td {
  color: #4a2e3b;
}

tbody tr:last-child td {
  border-bottom: 0;
}

.account-id {
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 13px;
  overflow-wrap: anywhere;
}

.status {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 999px;
  background: #fff3cd;
  color: #755600;
  font-size: 13px;
}

.status.active {
  background: #e7f8ee;
  color: #1d6b3f;
}

.state-panel {
  padding: 28px;
  color: #6b4f5a;
}

.error-message {
  color: #b3261e;
}

.success-message {
  color: #1d6b3f;
}

.dialog-backdrop {
  position: fixed;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 20px;
  background: rgb(74 46 59 / 35%);
}

.dialog {
  display: grid;
  width: min(100%, 460px);
  gap: 16px;
  padding: 28px;
  border: 1px solid #f0bdcb;
  border-radius: 26px;
  background: #ffffff;
}

.dialog h2 {
  margin: 0;
  color: #4a2e3b;
}

.dialog label {
  display: grid;
  gap: 7px;
  color: #4a2e3b;
  font-weight: 600;
}

.dialog input,
.dialog select {
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid #f0bdcb;
  border-radius: 14px;
  background: #fff8fa;
  color: #4a2e3b;
  font: inherit;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 6px;
}

.dialog button[type='submit'] {
  border-color: #d94f83;
  background: #d94f83;
  color: #ffffff;
}
</style>
