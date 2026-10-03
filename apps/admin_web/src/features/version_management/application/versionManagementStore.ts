import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { mapApiError, type ApiError } from '@/api/apiError'
import {
  createAdminVersionManagementClient,
  type AdminServiceVersion,
  type AdminServiceVersionOperation,
  type ServiceVersionSnapshot,
} from '@/api/adminVersionManagement'
import { createHttpClient } from '@/api/httpClient'

const activeOperationStatuses = new Set(['queued', 'running', 'recovering'])

/// Owns the brand-wide service inventory and the state of upgrade operations.
///
/// The page drives polling while an operation is active and stops it on
/// unmount. This keeps requests scoped to the visible operations page.
export const useVersionManagementStore = defineStore('admin-version-management', () => {
  const client = createAdminVersionManagementClient(createHttpClient())
  const snapshot = ref<ServiceVersionSnapshot | null>(null)
  const operations = ref<AdminServiceVersionOperation[]>([])
  const selectedOperation = ref<AdminServiceVersionOperation | null>(null)
  const isLoading = ref(false)
  const isChecking = ref(false)
  const isUpgradingAll = ref(false)
  const isRefreshingOperations = ref(false)
  const upgradingServiceIds = ref<string[]>([])
  const error = ref<ApiError | null>(null)
  const lastMessage = ref('')

  const services = computed(() => snapshot.value?.services ?? [])
  const outdatedServices = computed(() =>
    services.value.filter((service) => service.status === 'outdated' && service.canUpgrade),
  )
  const hasActiveOperations = computed(() =>
    operations.value.some((operation) => activeOperationStatuses.has(operation.status)),
  )
  const isUpgrading = computed(() => isUpgradingAll.value || upgradingServiceIds.value.length > 0)

  async function load(): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      snapshot.value = await client.loadServiceVersions()
      operations.value = await client.loadServiceVersionOperations()
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isLoading.value = false
    }
  }

  async function checkForUpdates(): Promise<void> {
    isChecking.value = true
    error.value = null
    lastMessage.value = ''
    try {
      snapshot.value = await client.checkServiceVersions()
      operations.value = await client.loadServiceVersionOperations()
      const count = outdatedServices.value.length
      lastMessage.value =
        count === 0 ? '所有服务都已是当前版本。' : `检查完成，有 ${count} 个服务可以升级。`
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isChecking.value = false
    }
  }

  async function upgradeService(service: AdminServiceVersion): Promise<boolean> {
    if (!service.canUpgrade || isServiceUpgrading(service.id)) {
      return false
    }
    upgradingServiceIds.value = [...upgradingServiceIds.value, service.id]
    error.value = null
    lastMessage.value = ''
    try {
      const operation = await client.upgradeService(service.id)
      if (operation === null) {
        throw new Error('missing service upgrade operation')
      }
      upsertOperation(operation)
      selectedOperation.value = operation
      lastMessage.value = `${service.displayName || service.id} 的升级已开始。`
      await refreshOperations()
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      upgradingServiceIds.value = upgradingServiceIds.value.filter(
        (serviceId) => serviceId !== service.id,
      )
    }
  }

  async function upgradeAll(): Promise<boolean> {
    if (outdatedServices.value.length === 0 || isUpgradingAll.value) {
      return false
    }
    isUpgradingAll.value = true
    error.value = null
    lastMessage.value = ''
    try {
      const operation = await client.upgradeAllServices()
      if (operation === null) {
        throw new Error('missing all-service upgrade operation')
      }
      upsertOperation(operation)
      selectedOperation.value = operation
      lastMessage.value = '全部服务升级已开始，完成后会自动更新状态。'
      await refreshOperations()
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isUpgradingAll.value = false
    }
  }

  /// Refreshes operation rows and reloads service state after completion.
  async function refreshOperations(): Promise<void> {
    if (isRefreshingOperations.value) {
      return
    }
    isRefreshingOperations.value = true
    try {
      const previousStatuses = new Map(
        operations.value.map((operation) => [operation.id, operation.status]),
      )
      const nextOperations = await client.loadServiceVersionOperations()
      operations.value = sortOperations(nextOperations)
      const completedOperation = operations.value.find((operation) => {
        const previousStatus = previousStatuses.get(operation.id)
        return (
          previousStatus !== undefined &&
          activeOperationStatuses.has(previousStatus) &&
          !activeOperationStatuses.has(operation.status)
        )
      })
      if (completedOperation !== undefined) {
        selectedOperation.value = completedOperation
        lastMessage.value =
          completedOperation.status === 'succeeded'
            ? '服务升级已完成。'
            : completedOperation.message || '服务升级没有完成，请查看操作记录后重试。'
      }
      if (!hasActiveOperations.value) {
        snapshot.value = await client.loadServiceVersions()
      }
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isRefreshingOperations.value = false
    }
  }

  async function selectOperation(operationId: string): Promise<void> {
    error.value = null
    try {
      const operation = await client.loadServiceVersionOperation(operationId)
      if (operation !== null) {
        upsertOperation(operation)
        selectedOperation.value = operation
      }
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    }
  }

  function closeOperation(): void {
    selectedOperation.value = null
  }

  function isServiceUpgrading(serviceId: string): boolean {
    return upgradingServiceIds.value.includes(serviceId)
  }

  function operationForService(serviceId: string): AdminServiceVersionOperation | null {
    return operations.value.find((operation) => operation.targetService === serviceId) ?? null
  }

  function upsertOperation(operation: AdminServiceVersionOperation): void {
    const existingIndex = operations.value.findIndex((item) => item.id === operation.id)
    const nextOperations = [...operations.value]
    if (existingIndex === -1) {
      nextOperations.unshift(operation)
    } else {
      nextOperations[existingIndex] = operation
    }
    operations.value = sortOperations(nextOperations)
  }

  return {
    closeOperation,
    error,
    hasActiveOperations,
    isChecking,
    isLoading,
    isServiceUpgrading,
    isUpgrading,
    isUpgradingAll,
    lastMessage,
    operationForService,
    operations,
    outdatedServices,
    refreshOperations,
    selectedOperation,
    selectOperation,
    services,
    snapshot,
    checkForUpdates,
    load,
    upgradeAll,
    upgradeService,
  }
})

function sortOperations(
  operations: AdminServiceVersionOperation[],
): AdminServiceVersionOperation[] {
  return [...operations].sort((left, right) => right.startedAt.localeCompare(left.startedAt))
}
