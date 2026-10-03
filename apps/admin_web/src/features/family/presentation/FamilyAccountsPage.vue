<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import {
  useFamilyAccountStore,
  type AdminFamily,
  type AdminFamilyAiAccount,
  type CreateAdminFamilyInput,
} from '@/features/family/application/familyAccountStore'

const store = useFamilyAccountStore()
const isCreateOpen = ref(false)
const editingFamily = ref<AdminFamily | null>(null)
const editingAiAccount = ref<AdminFamilyAiAccount | null>(null)
const modelsText = ref('')
const createForm = reactive<CreateAdminFamilyInput>({
  phone: '',
  password: '',
  guardianFamilyName: '',
  childNickname: '',
  childBirthday: '',
})

onMounted(() => {
  void store.load()
})

function openCreate(): void {
  isCreateOpen.value = true
  store.lastMessage = ''
}

function closeCreate(): void {
  isCreateOpen.value = false
}

async function createFamily(): Promise<void> {
  const succeeded = await store.create(createForm)
  if (!succeeded) {
    return
  }
  resetCreateForm()
  closeCreate()
}

function resetCreateForm(): void {
  createForm.phone = ''
  createForm.password = ''
  createForm.guardianFamilyName = ''
  createForm.childNickname = ''
  createForm.childBirthday = ''
}

function openAiEditor(family: AdminFamily): void {
  if (family.aiAccount === null) {
    store.lastMessage = ''
    return
  }
  editingFamily.value = family
  editingAiAccount.value = { ...family.aiAccount }
  modelsText.value = family.aiAccount.availableModels.join('、')
}

async function saveAiAccount(): Promise<void> {
  if (editingAiAccount.value === null) {
    return
  }
  editingAiAccount.value.availableModels = modelsText.value
    .split(/[,\s，、]+/)
    .map((model) => model.trim())
    .filter(Boolean)
  const succeeded = await store.updateAiAccount(editingAiAccount.value)
  if (succeeded) {
    editingAiAccount.value = null
    editingFamily.value = null
  }
}

function statusLabel(value: string): string {
  return value === 'active' ? '可用' : '已暂停'
}

function formatDate(value: string): string {
  if (!value) {
    return '未填写'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '未填写'
  }
  return date.toLocaleDateString('zh-CN')
}
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <p class="eyebrow">家庭与 AI 服务</p>
        <h1>家长账号</h1>
        <p class="page-description">
          每个家长账号都会自动开通一份 AI 服务。你可以在这里查看家庭绑定情况，
          并统一调整额度、模型和使用状态。
        </p>
      </div>
      <div class="header-actions">
        <button type="button" class="secondary" :disabled="store.isLoading" @click="store.load">
          重新加载
        </button>
        <button type="button" class="primary" @click="openCreate">新建家长账号</button>
      </div>
    </header>

    <Transition name="toast">
      <p v-if="store.lastMessage" class="success-message">{{ store.lastMessage }}</p>
    </Transition>
    <Transition name="toast">
      <p v-if="store.error" class="error-message">{{ store.error.message }}</p>
    </Transition>

    <Transition name="page" mode="out-in">
      <div v-if="store.isLoading" key="loading" class="state-panel">
        <span class="state-spinner" aria-hidden="true"></span>
        正在读取家长账号…
      </div>
      <div v-else-if="store.families.length === 0" key="empty" class="state-panel">
        <span class="state-icon" aria-hidden="true">☆</span>
        <span>还没有家长账号。新建后，家长可以直接登录应用使用。</span>
      </div>
      <div v-else key="table" class="table-shell">
        <table>
          <thead>
            <tr>
              <th>家长</th>
              <th>宝贝信息</th>
              <th>AI 账号</th>
              <th>剩余额度</th>
              <th>可分配模型</th>
              <th>家长账号</th>
              <th>创建时间</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="family in store.families" :key="family.parentAccountId">
              <td>
                <span class="parent-name">
                  {{ family.guardianFamilyName || family.parentDisplayName || '未填写称呼' }}
                </span>
                <span class="parent-meta">{{ family.phone || '未绑定手机号' }}</span>
              </td>
              <td>
                <span>{{ family.childNickname || '未填写宝贝姓名' }}</span>
                <span class="parent-meta">{{ formatDate(family.childBirthday) }}</span>
              </td>
              <td>
                <span v-if="family.aiAccount" :class="['status', family.aiAccount.status]">
                  {{ statusLabel(family.aiAccount.status) }}
                </span>
                <span v-else class="status missing">待开通</span>
                <span v-if="family.aiAccount" class="parent-meta">
                  {{ family.aiAccount.credentialReady ? '服务已就绪' : '服务待修复' }}
                </span>
              </td>
              <td>
                <span v-if="family.aiAccount">
                  ${{ family.aiAccount.balanceUsd.toFixed(2) }}
                </span>
                <span v-else>未开通</span>
              </td>
              <td>
                <span v-if="family.aiAccount">
                  {{ family.aiAccount.availableModels.join('、') || '使用默认模型' }}
                </span>
                <span v-else>未开通</span>
              </td>
              <td>
                <span :class="['status', family.status]">
                  {{ statusLabel(family.status) }}
                </span>
              </td>
              <td>{{ formatDate(family.createdAt) }}</td>
              <td>
                <button
                  v-if="family.aiAccount"
                  type="button"
                  :disabled="store.isSubmitting"
                  @click="openAiEditor(family)"
                >
                  管理 AI 服务
                </button>
                <button
                  v-else
                  type="button"
                  :disabled="store.isSubmitting"
                  @click="store.retryAiAccount(family.parentAccountId)"
                >
                  {{ store.isSubmitting ? '正在开通…' : '重新开通 AI 服务' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Transition>

    <Transition name="modal">
      <div v-if="isCreateOpen" class="dialog-backdrop" @click.self="closeCreate">
        <form class="dialog" @submit.prevent="createFamily">
          <h2>新建家长账号</h2>
          <p class="dialog-hint">创建后会自动开通家长 AI 账号，家长可直接登录应用。</p>
          <label>
            <span>手机号</span>
            <input
              v-model.trim="createForm.phone"
              type="tel"
              inputmode="tel"
              required
              placeholder="请输入家长手机号"
            />
          </label>
          <label>
            <span>初始密码</span>
            <input
              v-model="createForm.password"
              type="password"
              required
              minlength="8"
              placeholder="至少 8 位"
            />
          </label>
          <label>
            <span>家长姓氏</span>
            <input v-model.trim="createForm.guardianFamilyName" type="text" />
          </label>
          <label>
            <span>宝贝姓名</span>
            <input v-model.trim="createForm.childNickname" type="text" />
          </label>
          <label>
            <span>宝贝生日</span>
            <input v-model="createForm.childBirthday" type="date" />
          </label>
          <div class="dialog-actions">
            <button type="button" class="secondary" @click="closeCreate">取消</button>
            <button type="submit" class="primary" :disabled="store.isSubmitting">
              {{ store.isSubmitting ? '正在创建…' : '创建账号' }}
            </button>
          </div>
        </form>
      </div>
    </Transition>

    <Transition name="modal">
      <div
        v-if="editingAiAccount && editingFamily"
        class="dialog-backdrop"
        @click.self="editingAiAccount = null"
      >
        <form class="dialog" @submit.prevent="saveAiAccount">
          <h2>管理家长 AI 服务</h2>
          <p class="dialog-hint">
            {{ editingFamily.guardianFamilyName || editingFamily.phone || '家长账号' }}
          </p>
          <label>
            <span>服务状态</span>
            <select v-model="editingAiAccount.status">
              <option value="active">可用</option>
              <option value="suspended">暂停使用</option>
            </select>
          </label>
          <label>
            <span>剩余额度（美元）</span>
            <input
              v-model.number="editingAiAccount.balanceUsd"
              type="number"
              min="0"
              step="0.01"
            />
          </label>
          <label>
            <span>同时对话数量</span>
            <input v-model.number="editingAiAccount.concurrencyLimit" type="number" min="1" />
          </label>
          <label>
            <span>可分配模型</span>
            <input v-model="modelsText" type="text" placeholder="多个模型之间用顿号分隔" />
          </label>
          <div class="dialog-actions">
            <button type="button" class="secondary" @click="editingAiAccount = null">
              取消
            </button>
            <button type="submit" class="primary" :disabled="store.isSubmitting">
              {{ store.isSubmitting ? '正在保存…' : '保存更改' }}
            </button>
          </div>
        </form>
      </div>
    </Transition>
  </section>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 22px;
  padding: 24px;
  border: 1px solid var(--sprout-outline);
  border-radius: var(--sprout-radius-card);
  background: #ffffff;
  box-shadow: 0 8px 24px rgb(194 91 128 / 5%);
}

.eyebrow {
  margin: 0 0 6px;
  color: var(--sprout-pink-strong);
  font-size: 13px;
  font-weight: 700;
}

h1 {
  margin: 0;
  color: var(--sprout-text);
  font-size: 28px;
}

.page-description {
  max-width: 700px;
  margin: 10px 0 0;
  color: var(--sprout-text-muted);
}

.header-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

button {
  min-height: 38px;
  padding: 0 16px;
  border: 1px solid #e9a5b8;
  border-radius: var(--sprout-radius-control);
  background: #ffffff;
  color: #c94175;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

button:not(:disabled):hover {
  border-color: #d94f83;
  background: #fff7fa;
  box-shadow: 0 8px 18px rgb(217 79 131 / 10%);
}

button.primary {
  border-color: #d94f83;
  background: #d94f83;
  color: #ffffff;
}

button.primary:not(:disabled):hover {
  background: #c94175;
}

button:disabled {
  cursor: wait;
  opacity: 0.55;
}

.table-shell,
.state-panel {
  overflow: hidden;
  border: 1px solid var(--sprout-outline);
  border-radius: var(--sprout-radius-card);
  background: #ffffff;
  box-shadow: 0 10px 28px rgb(194 91 128 / 6%);
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
  white-space: nowrap;
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

tbody tr {
  transition: background-color var(--sprout-duration-fast) ease;
}

tbody tr:hover {
  background: #fff8fa;
}

.parent-name,
.parent-meta {
  display: block;
}

.parent-name {
  font-weight: 600;
}

.parent-meta {
  margin-top: 2px;
  color: #6b4f5a;
  font-size: 12px;
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

.status.suspended,
.status.disabled,
.status.missing {
  background: #fff0f2;
  color: #a12b4a;
}

.state-panel {
  display: flex;
  min-height: 132px;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 28px;
  color: #6b4f5a;
  text-align: center;
}

.state-spinner {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
  border: 2px solid #f0bdcb;
  border-top-color: #d94f83;
  border-radius: 50%;
  animation: spin 700ms linear infinite;
}

.state-icon {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 12px;
  background: #fff2f5;
  color: #d94f83;
  font-size: 20px;
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
  backdrop-filter: blur(3px);
}

.dialog {
  display: grid;
  width: min(100%, 480px);
  max-height: min(760px, calc(100vh - 40px));
  gap: 16px;
  overflow: auto;
  padding: 28px;
  border: 1px solid #f0bdcb;
  border-radius: 26px;
  background: #ffffff;
  box-shadow: 0 28px 72px rgb(74 46 59 / 22%);
}

.dialog h2 {
  margin: 0;
  color: #4a2e3b;
}

.dialog-hint {
  margin: -6px 0 0;
  color: #6b4f5a;
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
  transition:
    border-color var(--sprout-duration-fast) ease,
    box-shadow var(--sprout-duration-fast) ease,
    background-color var(--sprout-duration-fast) ease;
}

.dialog input:focus,
.dialog select:focus {
  border-color: #d94f83;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgb(217 79 131 / 12%);
  outline: none;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 6px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 900px) {
  .page-header {
    display: grid;
  }

  .table-shell {
    overflow-x: auto;
  }

  table {
    min-width: 1040px;
  }
}
</style>
