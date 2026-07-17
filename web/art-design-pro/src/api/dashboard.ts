import request from '@/utils/http'

export interface TrendItem {
  date: string
  count: number
}

export interface DashboardStats {
  totalApps: number
  activeApps: number
  totalTokens: number
  trend: TrendItem[]
}

/**
 * 获取控制面板统计数据
 */
export function fetchGetDashboardStats() {
  return request.get<DashboardStats>({
    url: '/admin/dashboard/stats'
  })
}
