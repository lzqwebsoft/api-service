/**
 * 快速入口配置
 * 包含：应用列表、快速链接等配置
 */
import type { FastEnterConfig } from '@/types/config'

const fastEnterConfig: FastEnterConfig = {
  // 显示条件（屏幕宽度）
  minWidth: 1200,
  // 应用列表
  applications: [
    {
      name: '工作台',
      description: '系统概览与数据统计',
      icon: 'ri:pie-chart-line',
      iconColor: '#377dff',
      enabled: true,
      order: 1,
      routeName: 'Console'
    },
    {
      name: '应用管理',
      description: '管理客户端 App 及 Token 签发',
      icon: 'ri:apps-line',
      iconColor: '#ffb100',
      enabled: true,
      order: 2,
      routeName: 'Apps'
    },
    {
      name: 'Token 黑名单',
      description: 'Token 撤销与黑名单管制',
      icon: 'ri:forbid-line',
      iconColor: '#ff6b6b',
      enabled: true,
      order: 3,
      routeName: 'Blacklist'
    },
    {
      name: '鉴权日志',
      description: 'API 访问日志与拦截记录',
      icon: 'ri:file-list-line',
      iconColor: '#10b981',
      enabled: true,
      order: 4,
      routeName: 'Logs'
    },
    {
      name: '节假日日历',
      description: '排班与法定节假日调休管理',
      icon: 'ri:calendar-todo-line',
      iconColor: '#8b5cf6',
      enabled: true,
      order: 5,
      routeName: 'Holiday'
    },
    {
      name: '用户管理',
      description: '系统管理员账号与角色配置',
      icon: 'ri:user-3-line',
      iconColor: '#ec4899',
      enabled: true,
      order: 6,
      routeName: 'User'
    }
  ],
  // 快速链接
  quickLinks: [
    {
      name: '工作台',
      enabled: true,
      order: 1,
      routeName: 'Console'
    },
    {
      name: '应用管理',
      enabled: true,
      order: 2,
      routeName: 'Apps'
    },
    {
      name: 'Token 黑名单',
      enabled: true,
      order: 3,
      routeName: 'Blacklist'
    },
    {
      name: '鉴权日志',
      enabled: true,
      order: 4,
      routeName: 'Logs'
    },
    {
      name: '用户管理',
      enabled: true,
      order: 5,
      routeName: 'User'
    },
    {
      name: '个人中心',
      enabled: true,
      order: 6,
      routeName: 'UserCenter'
    }
  ]
}

export default Object.freeze(fastEnterConfig)
