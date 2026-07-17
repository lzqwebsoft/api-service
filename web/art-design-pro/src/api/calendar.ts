import request from '@/utils/http'

// List calendar exceptions
export function fetchGetCalendarList(params: {
  region?: string
  is_workday?: boolean
  year?: number
  current?: number
  size?: number
}) {
  return request.get<{
    list: any[]
    total: number
    totalCount: number
    holidayCount: number
    workdayCount: number
    years: string[]
  }>({
    url: '/admin/calendar',
    params
  })
}

// Add calendar exception
export function fetchAddCalendar(data: {
  date: string
  region: string
  is_workday: boolean
  description: string
}) {
  return request.post({
    url: '/admin/calendar/add',
    data
  })
}

// Update calendar exception
export function fetchUpdateCalendar(data: {
  date: string
  region: string
  is_workday: boolean
  description: string
}) {
  return request.post({
    url: '/admin/calendar/update',
    data
  })
}

// Delete calendar exception
export function fetchDeleteCalendar(data: { date: string; region: string }) {
  return request.post({
    url: '/admin/calendar/delete',
    data
  })
}

// List standard holidays
export function fetchGetHolidayList(params?: {
  current?: number
  size?: number
  name?: string
  type?: string
  regions?: string
}) {
  return request.get<{
    list: any[]
    total: number
    totalCount: number
    solarCount: number
    weekdayCount: number
    industryCount: number
  }>({
    url: '/admin/holiday',
    params
  })
}

// Add holiday definition
export function fetchAddHoliday(data: {
  name: string
  type: string
  month?: number
  day?: number
  week_number?: number
  day_of_week?: number
  regions?: string
  description?: string
}) {
  return request.post({
    url: '/admin/holiday/add',
    data
  })
}

// Update holiday definition
export function fetchUpdateHoliday(data: {
  id: number
  name: string
  type: string
  month?: number
  day?: number
  week_number?: number
  day_of_week?: number
  regions?: string
  description?: string
}) {
  return request.post({
    url: '/admin/holiday/update',
    data
  })
}

// Delete holiday definition
export function fetchDeleteHoliday(data: { id: number }) {
  return request.post({
    url: '/admin/holiday/delete',
    data
  })
}
