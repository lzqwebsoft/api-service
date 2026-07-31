import request from '@/utils/http'

export interface AppFileItem {
  id: number
  app_record_id: number
  app_id?: string
  app_name?: string
  version: string
  file_name: string
  download_url: string
  file_size: number
  description: string
  download_count: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface FileDownloadLogItem {
  id: number
  file_id: number
  app_id?: string
  app_name?: string
  file_name?: string
  version?: string
  ip: string
  ip_location: string
  user_agent: string
  referer: string
  channel: string
  created_at: string
}

// Fetch file list
export function fetchGetFiles(params?: {
  app_record_id?: number
  current?: number
  size?: number
}) {
  return request.get<{ list: AppFileItem[]; total: number }>({
    url: '/admin/files',
    params
  })
}

// Create new file release
export function fetchCreateFile(data: Partial<AppFileItem>) {
  return request.post<AppFileItem>({
    url: '/admin/files/create',
    data
  })
}

// Update file release
export function fetchUpdateFile(data: Partial<AppFileItem>) {
  return request.post<AppFileItem>({
    url: '/admin/files/update',
    data
  })
}

// Delete file release
export function fetchDeleteFile(data: { id: number }) {
  return request.post({
    url: '/admin/files/delete',
    data
  })
}

// Get download logs
export function fetchGetDownloadLogs(params?: {
  file_id?: number
  current?: number
  size?: number
}) {
  return request.get<{ list: FileDownloadLogItem[]; total: number }>({
    url: '/admin/files/logs',
    params
  })
}

// Upload local file to runtimes/uploads
export function fetchUploadFile(data: FormData, options?: { timeout?: number }) {
  return request.post<{ file_name: string; file_size: number; download_url: string }>({
    url: '/admin/files/upload',
    data,
    timeout: options?.timeout ?? 60000,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
