import { defineStore } from 'pinia'
import { ref } from 'vue'

import { mapApiError, type ApiError } from '@/api/apiError'
import { createHttpClient } from '@/api/httpClient'

export type DeviceCommandType =
  | 'refresh_configuration'
  | 'reconnect_network'
  | 'resync_time'

/// Runtime projection used by the operations console.
export interface DeviceRuntime {
  isOnline: boolean
  connectionState: string
  transport: string
  networkQuality: string
  rssiDbm: number
  latencyMs: number
  packetLossPercent: number
  timeSyncState: string
  lastSyncedAt: string
  offlineState: string
  offlineReason: string
  fallbackActive: boolean
  pendingTelemetry: number
  reportedAt: string
  receivedAt: string
}

/// Bound device and its latest runtime snapshot.
export interface AdminDevice {
  deviceId: string
  deviceName: string
  hardwareModel: string
  firmwareVersion: string
  capabilities: string[]
  boundAt: string
  updatedAt: string
  runtime: DeviceRuntime | null
}

export interface DeviceCommand {
  commandId: string
  deviceId: string
  commandType: DeviceCommandType
  status: string
  requestId: string
  createdAt: string
  deliveredAt: string
  acknowledgedAt: string
  resultCode: string
}

/// Loads device state and owns administrator maintenance commands.
export const useDeviceStore = defineStore('admin-devices', () => {
  const httpClient = createHttpClient()
  const devices = ref<AdminDevice[]>([])
  const commands = ref<Record<string, DeviceCommand[]>>({})
  const isLoading = ref(false)
  const isSubmitting = ref(false)
  const error = ref<ApiError | null>(null)
  const lastMessage = ref('')

  async function load(): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      const response = await httpClient.get('/api/v1/admin/devices')
      devices.value = (response.data.data.devices ?? []).map(toDevice)
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isLoading.value = false
    }
  }

  async function loadCommands(deviceId: string): Promise<void> {
    error.value = null
    try {
      const response = await httpClient.get(
        `/api/v1/admin/devices/${deviceId}/commands`,
      )
      commands.value = {
        ...commands.value,
        [deviceId]: (response.data.data.commands ?? []).map(toCommand),
      }
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    }
  }

  async function sendCommand(
    deviceId: string,
    commandType: DeviceCommandType,
  ): Promise<boolean> {
    isSubmitting.value = true
    error.value = null
    lastMessage.value = ''
    try {
      await httpClient.post(
        `/api/v1/admin/devices/${deviceId}/commands`,
        { command_type: commandType },
      )
      lastMessage.value = '操作已发送，设备将在下一次连接时执行。'
      await loadCommands(deviceId)
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    commands,
    devices,
    error,
    isLoading,
    isSubmitting,
    lastMessage,
    load,
    loadCommands,
    sendCommand,
  }
})

function toDevice(value: Record<string, unknown>): AdminDevice {
  return {
    deviceId: String(value.device_id ?? ''),
    deviceName: String(value.device_name ?? ''),
    hardwareModel: String(value.hardware_model ?? ''),
    firmwareVersion: String(value.firmware_version ?? ''),
    capabilities: stringList(value.capabilities),
    boundAt: String(value.bound_at ?? ''),
    updatedAt: String(value.updated_at ?? ''),
    runtime: isRecord(value.runtime) ? toRuntime(value.runtime) : null,
  }
}

function toRuntime(value: Record<string, unknown>): DeviceRuntime {
  const connection = record(value.connection)
  const quality = record(value.network_quality)
  const timeSync = record(value.time_sync)
  const offline = record(value.offline)
  return {
    isOnline: value.is_online === true,
    connectionState: String(connection.state ?? ''),
    transport: String(connection.transport ?? ''),
    networkQuality: String(quality.level ?? ''),
    rssiDbm: number(quality.rssi_dbm),
    latencyMs: number(quality.latency_ms),
    packetLossPercent: number(quality.packet_loss_percent),
    timeSyncState: String(timeSync.state ?? ''),
    lastSyncedAt: String(timeSync.last_synced_at ?? ''),
    offlineState: String(offline.state ?? ''),
    offlineReason: String(offline.reason ?? ''),
    fallbackActive: offline.fallback_active === true,
    pendingTelemetry: number(offline.pending_telemetry),
    reportedAt: String(value.reported_at ?? ''),
    receivedAt: String(value.received_at ?? ''),
  }
}

function toCommand(value: Record<string, unknown>): DeviceCommand {
  return {
    commandId: String(value.command_id ?? ''),
    deviceId: String(value.device_id ?? ''),
    commandType: String(value.command_type ?? '') as DeviceCommandType,
    status: String(value.status ?? ''),
    requestId: String(value.request_id ?? ''),
    createdAt: String(value.created_at ?? ''),
    deliveredAt: String(value.delivered_at ?? ''),
    acknowledgedAt: String(value.acknowledged_at ?? ''),
    resultCode: String(value.result_code ?? ''),
  }
}

function record(value: unknown): Record<string, unknown> {
  return isRecord(value) ? value : {}
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function stringList(value: unknown): string[] {
  return Array.isArray(value) ? value.map(String) : []
}

function number(value: unknown): number {
  return typeof value === 'number' ? value : Number(value ?? 0)
}
