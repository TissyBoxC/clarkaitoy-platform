import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import {
  createAdminDownloadFilesClient,
  type AdminDownloadFilesClient,
  type DownloadArtifactKind,
  type DownloadArtifactPlatform,
  type DownloadFile,
  type DownloadFileUpload,
  type DownloadIndexStatus,
} from '@/api/adminDownloadFiles'
import { mapApiError, type ApiError } from '@/api/apiError'

export type DownloadFileStatusFilter = 'all' | 'indexed' | 'pending'
export type DownloadArtifactKindFilter = 'all' | DownloadArtifactKind

export interface DownloadFileFilters {
  directory: string
  version: string
  kind: DownloadArtifactKindFilter
  status: DownloadFileStatusFilter
}

export interface DownloadFileUploadState {
  isUploading: boolean
  progress: number
  fileName: string
  error: string
}

export const emptyDownloadIndexStatus: DownloadIndexStatus = {
  isAvailable: false,
  refreshedAt: '',
  indexedFileCount: 0,
  pendingFileCount: 0,
  errorMessage: '',
}

/// Owns the download-server inventory, filters, upload progress, and the
/// publishing index state. The page keeps this store as the single source of
/// truth so refreshes and uploads stay consistent.
export const useDownloadFilesStore = defineStore('admin-download-files', () => {
  const client: AdminDownloadFilesClient = createAdminDownloadFilesClient()
  const files = ref<DownloadFile[]>([])
  const indexStatus = ref<DownloadIndexStatus>({ ...emptyDownloadIndexStatus })
  const filters = ref<DownloadFileFilters>({
    directory: '',
    version: '',
    kind: 'all',
    status: 'all',
  })
  const isLoading = ref(false)
  const isRefreshingIndex = ref(false)
  const deletingPaths = ref<string[]>([])
  const uploadState = ref<DownloadFileUploadState>({
    isUploading: false,
    progress: 0,
    fileName: '',
    error: '',
  })
  const error = ref<ApiError | null>(null)
  const lastMessage = ref('')

  const directories = computed(() => {
    const values = new Set<string>()
    for (const file of files.value) {
      if (file.directory) {
        values.add(file.directory)
      }
    }
    return [...values].sort((left, right) => left.localeCompare(right))
  })

  const versions = computed(() => {
    const values = new Set<string>()
    for (const file of files.value) {
      if (file.version) {
        values.add(file.version)
      }
    }
    return [...values].sort((left, right) => right.localeCompare(left, undefined, { numeric: true }))
  })

  const visibleFiles = computed(() =>
    files.value.filter((file) => {
      if (filters.value.directory && file.directory !== filters.value.directory) {
        return false
      }
      if (filters.value.version && file.version !== filters.value.version) {
        return false
      }
      if (filters.value.kind !== 'all' && file.kind !== filters.value.kind) {
        return false
      }
      if (filters.value.status === 'indexed' && !file.isIndexed) {
        return false
      }
      if (filters.value.status === 'pending' && file.isIndexed) {
        return false
      }
      return true
    }),
  )

  const indexedFileCount = computed(
    () => files.value.filter((file) => file.isIndexed).length,
  )
  const pendingFileCount = computed(
    () => files.value.filter((file) => !file.isIndexed).length,
  )
  const totalSizeBytes = computed(() =>
    files.value.reduce((sum, file) => sum + file.sizeBytes, 0),
  )
  const versionArtifacts = computed(() => {
    const grouped = new Map<string, DownloadFile[]>()
    for (const file of files.value) {
      const version = file.version.trim()
      if (!version) {
        continue
      }
      const existing = grouped.get(version) ?? []
      existing.push(file)
      grouped.set(version, existing)
    }
    return [...grouped.entries()]
      .map(([version, artifacts]) => ({
        version,
        artifacts: [...artifacts].sort((left, right) =>
          left.relativePath.localeCompare(right.relativePath),
        ),
      }))
      .sort((left, right) => right.version.localeCompare(left.version, undefined, { numeric: true }))
  })

  async function load(): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      const [loadedFiles, loadedIndexStatus] = await Promise.all([
        client.loadFiles(),
        client.loadIndexStatus().catch(() => ({ ...emptyDownloadIndexStatus })),
      ])
      files.value = loadedFiles
      indexStatus.value = loadedIndexStatus
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isLoading.value = false
    }
  }

  async function upload(input: DownloadFileUpload): Promise<boolean> {
    uploadState.value = {
      isUploading: true,
      progress: 0,
      fileName: input.filename,
      error: '',
    }
    error.value = null
    lastMessage.value = ''
    try {
      await client.uploadFile(input, (progress) => {
        uploadState.value = { ...uploadState.value, progress }
      })
      uploadState.value = { ...uploadState.value, progress: 100 }
      await refreshIndex()
      await reloadFiles()
      lastMessage.value = `${input.filename} 已上传到下载服务器。`
      return true
    } catch (caught: unknown) {
      const mappedError = mapApiError(caught)
      uploadState.value = {
        ...uploadState.value,
        isUploading: false,
        error: mappedError.message,
      }
      error.value = mappedError
      return false
    } finally {
      uploadState.value = { ...uploadState.value, isUploading: false }
    }
  }

  async function remove(file: DownloadFile): Promise<boolean> {
    deletingPaths.value = [...deletingPaths.value, file.relativePath]
    error.value = null
    lastMessage.value = ''
    try {
      await client.deleteFile(file.relativePath)
      files.value = files.value.filter((item) => item.relativePath !== file.relativePath)
      lastMessage.value = `${file.name} 已从下载服务器删除。`
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      deletingPaths.value = deletingPaths.value.filter(
        (relativePath) => relativePath !== file.relativePath,
      )
    }
  }

  async function refreshIndex(): Promise<boolean> {
    isRefreshingIndex.value = true
    error.value = null
    lastMessage.value = ''
    try {
      indexStatus.value = await client.refreshIndex()
      lastMessage.value =
        indexStatus.value.pendingFileCount === 0
          ? '发布索引已刷新，所有文件都可以登记版本。'
          : `发布索引已刷新，还有 ${indexStatus.value.pendingFileCount} 个文件未登记版本。`
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isRefreshingIndex.value = false
    }
  }

  async function reloadFiles(): Promise<void> {
    try {
      files.value = await client.loadFiles()
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    }
  }

  function isDeleting(relativePath: string): boolean {
    return deletingPaths.value.includes(relativePath)
  }

  function setFilters(nextFilters: Partial<DownloadFileFilters>): void {
    filters.value = { ...filters.value, ...nextFilters }
  }

  function resetFilters(): void {
    filters.value = { directory: '', version: '', kind: 'all', status: 'all' }
  }

  function clearMessages(): void {
    error.value = null
    lastMessage.value = ''
    uploadState.value = { ...uploadState.value, error: '' }
  }

  return {
    clearMessages,
    deletingPaths,
    directories,
    error,
    files,
    filters,
    indexStatus,
    indexedFileCount,
    isLoading,
    isRefreshingIndex,
    isDeleting,
    lastMessage,
    pendingFileCount,
    reloadFiles,
    refreshIndex,
    remove,
    resetFilters,
    setFilters,
    totalSizeBytes,
    upload,
    uploadState,
    versionArtifacts,
    versions,
    visibleFiles,
    load,
  }
})

export function artifactKindLabel(value: DownloadArtifactKind): string {
  return (
    {
      resource: '资源更新',
      client: '客户端更新',
      firmware: '设备固件',
      unknown: '未标注',
    }[value] ?? '未标注'
  )
}

export function artifactPlatformLabel(value: DownloadArtifactPlatform): string {
  return (
    {
      android: 'Android',
      ios: 'iOS',
      esp32_s3: '初芽设备',
      all: '全部平台',
      unknown: '未标注',
    }[value] ?? '未标注'
  )
}
