import request from '@/utils/http'

// List apps
export function fetchGetApps() {
  return request.get<any[]>({
    url: '/admin/apps'
  })
}

// Register new app version
export function fetchRegisterApp(data: { app_id: string; name: string; version: string }) {
  return request.post({
    url: '/admin/apps/register',
    data
  })
}

// Toggle app status
export function fetchToggleApp(data: { app_id: string; version: string; is_active: boolean }) {
  return request.post({
    url: '/admin/apps/toggle',
    data
  })
}

// Delete app version
export function fetchDeleteApp(data: { app_id: string; version: string }) {
  return request.post({
    url: '/admin/apps/delete',
    data
  })
}

// Get tokens for specific app or all tokens
export function fetchGetTokens(params?: { app_id?: string; version?: string }) {
  return request.get<any>({
    url: '/admin/tokens',
    params
  })
}

// Generate token
export function fetchGenerateToken(data: { app_id: string; version: string; platform: string }) {
  return request.post<any>({
    url: '/admin/tokens/generate',
    data
  })
}

// Revoke token
export function fetchRevokeToken(data: { token: string }) {
  return request.post({
    url: '/admin/tokens/revoke',
    data
  })
}

// List blacklisted tokens
export function fetchGetBlacklist() {
  return request.get<any[]>({
    url: '/admin/blacklist'
  })
}

// Add token to blacklist
export function fetchAddBlacklist(data: { token_id?: number; token?: string; user_uuid: string }) {
  return request.post({
    url: '/admin/blacklist/add',
    data
  })
}

// Remove token from blacklist
export function fetchDeleteBlacklist(data: { id: number }) {
  return request.post({
    url: '/admin/blacklist/delete',
    data
  })
}

// List access logs
export function fetchGetLogs(params?: { current?: number; size?: number }) {
  return request.get<{ list: any[]; total: number; blacklistedKeys: Record<string, boolean> }>({
    url: '/admin/logs',
    params
  })
}

// Add token from logs to blacklist (One-click)
export function fetchAddLogBlacklist(data: { token_id: number; token: string; user_uuid: string }) {
  return request.post({
    url: '/admin/logs/blacklist',
    data
  })
}

// List user feedback
export function fetchGetFeedback(params?: { current?: number; size?: number }) {
  return request.get<{ list: any[]; total: number }>({
    url: '/admin/feedback',
    params
  })
}

// Update user feedback processing status
export function fetchUpdateFeedbackStatus(data: { id: number; status: number }) {
  return request.post({
    url: '/admin/feedback/status',
    data
  })
}

// Delete user feedback record
export function fetchDeleteFeedback(data: { id: number }) {
  return request.post({
    url: '/admin/feedback/delete',
    data
  })
}
