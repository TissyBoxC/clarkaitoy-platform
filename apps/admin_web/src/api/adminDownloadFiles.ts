import type { AxiosInstance, AxiosProgressEvent } from 'axios'

import { createHttpClient } from '@/api/httpClient'

export type DownloadArtifactKind = 'resource' | 'client' | 'firmware' | 'unknown'

export type DownloadArtifactPlatform =
  | 'android'
  | 'ios'
  | 'esp32_s3'
  | 'all'
  | 'unknown'

/// One real file on the download server, enriched with the release metadata
/// the publishing flow uses when it registers an update.
export interface DownloadFile {
  id: string
  name: string
  relativePath: string
  directory: string
  sizeBytes: number
  modifiedAt: string
  sha256: string
  version: string
  kind: DownloadArtifactKind
  platform: DownloadArtifactPlatform
  isIndexed: boolean
}

export interface DownloadFileQuery {
  directory?: string
  version?: string
}

export interface DownloadFileUpload {
  file: File
  directory: string
  filename: string
  version: string
  kind: DownloadArtifactKind
  platform: DownloadArtifactPlatform
}

export interface DownloadIndexStatus {
  isAvailable: boolean
  refreshedAt: string
  indexedFileCount: number
  pendingFileCount: number
  errorMessage: string
}

/// Upload progress is reported as a percentage so the page can render a stable
/// progress bar without knowing the transport implementation.
export type UploadProgressHandler = (percent: number) => void

export interface AdminDownloadFilesClient {
  loadFiles(query?: DownloadFileQuery): Promise<DownloadFile[]>
  uploadFile(input: DownloadFileUpload, onProgress?: UploadProgressHandler): Promise<DownloadFile>
  deleteFile(relativePath: string): Promise<void>
  refreshIndex(): Promise<DownloadIndexStatus>
  loadIndexStatus(): Promise<DownloadIndexStatus>
}

/// Creates the client for the download server inventory and its publishing
/// index. Relative paths are validated here so no caller can escape the
/// configured download root or the release directory contract.
export function createAdminDownloadFilesClient(
  httpClient: AxiosInstance = createHttpClient(),
): AdminDownloadFilesClient {
  return {
    async loadFiles(query: DownloadFileQuery = {}): Promise<DownloadFile[]> {
      const response = await httpClient.get('/api/v1/admin/storage/files', {
        params: compactParams({
          directory: query.directory?.trim(),
          version: query.version?.trim(),
        }),
      })
      return toDownloadFiles(response.data?.data)
    },

    async uploadFile(
      input: DownloadFileUpload,
      onProgress?: UploadProgressHandler,
    ): Promise<DownloadFile> {
      const relativePath = joinRelativePath(input.directory, input.filename)
      const formData = new FormData()
      formData.append('file', input.file)
      formData.append('relative_path', relativePath)
      formData.append('directory', normalizeRelativePath(input.directory))
      formData.append('filename', input.filename.trim())
      formData.append('version', input.version.trim())
      formData.append('kind', input.kind)
      formData.append('platform', input.platform)

      const response = await httpClient.post('/api/v1/admin/storage/files/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: (event: AxiosProgressEvent) => {
          if (onProgress === undefined) {
            return
          }
          onProgress(uploadPercent(event))
        },
      })
      return toDownloadFile(response.data?.data?.file)
    },

    async deleteFile(relativePath: string): Promise<void> {
      await httpClient.delete(
        `/api/v1/admin/storage/files/${encodeRelativePath(relativePath)}`,
      )
    },

    async refreshIndex(): Promise<DownloadIndexStatus> {
      const response = await httpClient.post('/api/v1/admin/storage/index/refresh')
      return toIndexStatus(response.data?.data)
    },

    async loadIndexStatus(): Promise<DownloadIndexStatus> {
      const response = await httpClient.get('/api/v1/admin/storage/index/status')
      return toIndexStatus(response.data?.data)
    },
  }
}

export function normalizeRelativePath(value: string): string {
  return value
    .replaceAll('\\', '/')
    .split('/')
    .map((segment) => segment.trim())
    .filter((segment) => segment !== '' && segment !== '.')
    .reduce<string[]>((segments, segment) => {
      if (segment === '..') {
        segments.pop()
        return segments
      }
      segments.push(segment)
      return segments
    }, [])
    .join('/')
}

export function joinRelativePath(directory: string, filename: string): string {
  const normalizedDirectory = normalizeRelativePath(directory)
  const normalizedFilename = normalizeRelativePath(filename)
  if (normalizedDirectory === '') {
    return normalizedFilename
  }
  if (normalizedFilename === '') {
    return normalizedDirectory
  }
  return `${normalizedDirectory}/${normalizedFilename}`
}

function encodeRelativePath(value: string): string {
  return normalizeRelativePath(value)
    .split('/')
    .map((segment) => encodeURIComponent(segment))
    .join('/')
}

function compactParams(input: Record<string, string | undefined>): Record<string, string> {
  const params: Record<string, string> = {}
  for (const [key, value] of Object.entries(input)) {
    if (value !== undefined && value !== '') {
      params[key] = value
    }
  }
  return params
}

function uploadPercent(event: AxiosProgressEvent): number {
  const total = event.total ?? 0
  if (total <= 0) {
    return event.loaded > 0 ? 1 : 0
  }
  return Math.min(100, Math.max(0, Math.round((event.loaded / total) * 100)))
}

function toDownloadFiles(value: unknown): DownloadFile[] {
  const record = recordValue(value)
  const source = Array.isArray(record.files) ? record.files : Array.isArray(value) ? value : []
  return source.flatMap((item: unknown) => {
    const file = toNullableDownloadFile(item)
    return file === null ? [] : [file]
  })
}

function toDownloadFile(value: unknown): DownloadFile {
  const file = toNullableDownloadFile(value)
  if (file === null) {
    throw new Error('download file response is missing a relative path or file name')
  }
  return file
}

function toNullableDownloadFile(value: unknown): DownloadFile | null {
  const record = recordValue(value)
  const relativePath = stringValue(record.relative_path ?? record.relativePath).trim()
  const name = stringValue(record.name ?? record.filename).trim() || basename(relativePath)
  if (!name || !relativePath) {
    return null
  }

  return {
    id: stringValue(record.id).trim() || relativePath,
    name,
    relativePath,
    directory: stringValue(record.directory).trim() || dirname(relativePath),
    sizeBytes: numberValue(record.size_bytes ?? record.sizeBytes ?? record.size),
    modifiedAt: stringValue(
      record.modified_at ?? record.modifiedAt ?? record.updated_at ?? record.updatedAt,
    ),
    sha256: stringValue(record.sha256),
    version: stringValue(record.version),
    kind: artifactKind(record.kind),
    platform: artifactPlatform(record.platform),
    isIndexed: record.is_indexed === true || record.indexed === true,
  }
}

function toIndexStatus(value: unknown): DownloadIndexStatus {
  const record = recordValue(value)
  const refreshedAt = stringValue(
    record.refreshed_at ?? record.refreshedAt ?? record.generated_at ?? record.generatedAt,
  )
  return {
    isAvailable: record.available !== false && record.is_available !== false,
    refreshedAt,
    indexedFileCount: numberValue(
      record.indexed_file_count ?? record.indexedFileCount ?? record.file_count,
    ),
    pendingFileCount: numberValue(
      record.pending_file_count ?? record.pendingFileCount ?? record.pending_count,
    ),
    errorMessage: stringValue(record.error_message ?? record.errorMessage),
  }
}

function artifactKind(value: unknown): DownloadArtifactKind {
  switch (value) {
    case 'resource':
    case 'client':
    case 'firmware':
      return value
    default:
      return 'unknown'
  }
}

function artifactPlatform(value: unknown): DownloadArtifactPlatform {
  switch (value) {
    case 'android':
    case 'ios':
    case 'esp32_s3':
    case 'all':
      return value
    default:
      return 'unknown'
  }
}

function basename(value: string): string {
  const segments = normalizeRelativePath(value).split('/')
  return segments.at(-1) ?? ''
}

function dirname(value: string): string {
  const segments = normalizeRelativePath(value).split('/')
  segments.pop()
  return segments.join('/')
}

function recordValue(value: unknown): Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}
}

function stringValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function numberValue(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0
}
