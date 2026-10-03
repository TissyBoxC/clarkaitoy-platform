import { defineStore } from 'pinia'
import { ref } from 'vue'

import { mapApiError, type ApiError } from '@/api/apiError'
import {
  createAdminOperationsClient,
  type ReleaseArtifact,
} from '@/api/adminOperations'
import { createHttpClient } from '@/api/httpClient'
import { getRuntimeConfig } from '@/config/runtimeConfig'

export type ReleaseKind = 'resource' | 'client' | 'firmware'
export type ReleaseChannel = 'stable' | 'beta' | 'canary'
export type ReleasePlatform = 'android' | 'ios' | 'esp32_s3' | 'all'

export interface AdminRelease {
  version: string
  channel: string
  kind: string
  platform: string
  downloadUrl: string
  sha256: string
  releaseNotes: string
  isMandatory: boolean
  minimumSupportedVersion: string
  status: string
  publishedAt: string
  createdAt: string
}

export interface CreateReleaseInput {
  version: string
  channel: ReleaseChannel
  kind: ReleaseKind
  platform: ReleasePlatform
  downloadUrl: string
  sha256: string
  releaseNotes: string
  isMandatory: boolean
  minimumSupportedVersion: string
}

export interface ReleaseArtifactLookup {
  version: string
  kind: ReleaseKind
  platform: ReleasePlatform
}

// 管理端只操作版本清单；文件本身由发布流程上传并校验后再登记。
export const useReleaseStore = defineStore('admin-releases', () => {
  const httpClient = createHttpClient()
  const operationsClient = createAdminOperationsClient(httpClient)
  const runtimeConfig = getRuntimeConfig()
  const releases = ref<AdminRelease[]>([])
  const isLoading = ref(false)
  const isResolvingArtifact = ref(false)
  const isSubmitting = ref(false)
  const error = ref<ApiError | null>(null)
  const lastMessage = ref('')

  async function load(): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      const response = await httpClient.get('/api/v1/admin/releases')
      const records = response.data.data?.releases ?? []
      releases.value = Array.isArray(records) ? records.map(toRelease) : []
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isLoading.value = false
    }
  }

  async function create(input: CreateReleaseInput): Promise<boolean> {
    isSubmitting.value = true
    error.value = null
    lastMessage.value = ''
    try {
      await httpClient.post('/api/v1/admin/releases', {
        version: input.version.trim(),
        channel: input.channel,
        kind: input.kind,
        platform: input.platform,
        download_url: input.downloadUrl.trim(),
        sha256: input.sha256.trim(),
        release_notes: input.releaseNotes.trim(),
        is_mandatory: input.isMandatory,
        min_supported_version: input.minimumSupportedVersion.trim(),
      })
      await load()
      lastMessage.value = '更新版本已登记。'
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  async function resolveArtifact(
    input: ReleaseArtifactLookup,
  ): Promise<CreateReleaseInput | null> {
    isResolvingArtifact.value = true
    error.value = null
    lastMessage.value = ''
    try {
      const artifact = await operationsClient.loadReleaseArtifact(
        input.version,
        input.kind,
        input.platform,
      )
      if (artifact === null) {
        error.value = {
          kind: 'not_found',
          message: '这个版本还没有可用的更新文件，请确认文件已经上传。',
          retryable: false,
        }
        return null
      }
      const releaseInput = artifactToCreateInput(artifact, input)
      if (releaseInput === null) {
        error.value = {
          kind: 'unexpected',
          message: '更新文件的信息不完整，请确认下载地址和校验值已经生成。',
          retryable: false,
        }
        return null
      }
      lastMessage.value = `已找到 ${artifact.version} 的更新文件。`
      return releaseInput
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return null
    } finally {
      isResolvingArtifact.value = false
    }
  }

  async function publish(version: string): Promise<boolean> {
    isSubmitting.value = true
    error.value = null
    lastMessage.value = ''
    try {
      await httpClient.put(
        `/api/v1/admin/releases/${encodeURIComponent(version)}/publish`,
      )
      await load()
      lastMessage.value = `版本 ${version} 已发布。`
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  async function remove(version: string): Promise<boolean> {
    isSubmitting.value = true
    error.value = null
    lastMessage.value = ''
    try {
      await httpClient.delete(
        `/api/v1/admin/releases/${encodeURIComponent(version)}`,
      )
      await load()
      lastMessage.value = `版本 ${version} 已删除。`
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    adminBuildVersion: runtimeConfig.adminBuildVersion,
    error,
    isLoading,
    isSubmitting,
    isResolvingArtifact,
    lastMessage,
    releases,
    create,
    load,
    publish,
    resolveArtifact,
    remove,
  }
})

function artifactToCreateInput(
  artifact: ReleaseArtifact,
  lookup: ReleaseArtifactLookup,
): CreateReleaseInput | null {
  if (!artifact.downloadUrl || !artifact.sha256) {
    return null
  }
  return {
    version: artifact.version || lookup.version,
    channel: 'stable',
    kind: normalizeReleaseKind(artifact.kind, lookup.kind),
    platform: normalizeReleasePlatform(artifact.platform, lookup.platform),
    downloadUrl: artifact.downloadUrl,
    sha256: artifact.sha256,
    releaseNotes: '',
    isMandatory: false,
    minimumSupportedVersion: '',
  }
}

function normalizeReleaseKind(
  value: string,
  fallback: ReleaseKind,
): ReleaseKind {
  return value === 'resource' || value === 'client' || value === 'firmware'
    ? value
    : fallback
}

function normalizeReleasePlatform(
  value: string,
  fallback: ReleasePlatform,
): ReleasePlatform {
  return value === 'android' ||
    value === 'ios' ||
    value === 'esp32_s3' ||
    value === 'all'
    ? value
    : fallback
}

function toRelease(value: Record<string, unknown>): AdminRelease {
  return {
    version: stringValue(value.version),
    channel: stringValue(value.channel, 'stable'),
    kind: stringValue(value.kind, 'resource'),
    platform: stringValue(value.platform, 'all'),
    downloadUrl: stringValue(value.download_url),
    sha256: stringValue(value.sha256),
    releaseNotes: stringValue(value.release_notes),
    isMandatory: value.is_mandatory === true,
    minimumSupportedVersion: stringValue(value.min_supported_version),
    status: stringValue(value.status, 'draft'),
    publishedAt: stringValue(value.published_at),
    createdAt: stringValue(value.created_at),
  }
}

function stringValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}
