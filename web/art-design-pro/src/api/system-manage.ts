import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

// 获取用户列表
export function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/admin/users',
    params
  })
}

// 新增管理员用户
export function fetchCreateUser(data: any) {
  return request.post({
    url: '/admin/users/create',
    data
  })
}

// 更新管理员用户
export function fetchUpdateUser(data: any) {
  return request.post({
    url: '/admin/users/update',
    data
  })
}

// 删除管理员用户
export function fetchDeleteUser(data: { id: number }) {
  return request.post({
    url: '/admin/users/delete',
    data
  })
}

// 获取当前登录用户个人资料
export function fetchGetUserProfile() {
  return request.get<any>({
    url: '/admin/user/profile'
  })
}

// 更新当前登录用户个人资料
export function fetchUpdateUserProfile(data: any) {
  return request.post({
    url: '/admin/user/profile',
    data
  })
}

// 修改当前登录用户密码
export function fetchUpdateUserPassword(data: any) {
  return request.post({
    url: '/admin/user/password',
    data
  })
}

// 获取角色列表
export function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: '/admin/role/list',
    params
  })
}

// 新增角色
export function fetchCreateRole(data: Partial<Api.SystemManage.RoleListItem>) {
  return request.post({
    url: '/admin/role/add',
    data
  })
}

// 更新角色
export function fetchUpdateRole(data: Partial<Api.SystemManage.RoleListItem>) {
  return request.post({
    url: '/admin/role/update',
    data
  })
}

// 删除角色
export function fetchDeleteRole(data: { roleId: number }) {
  return request.post({
    url: '/admin/role/delete',
    data
  })
}

// 获取角色关联的菜单ID列表
export function fetchGetRoleMenuIds(roleId: number) {
  return request.get<number[]>({
    url: '/admin/role/menu_ids',
    params: { roleId }
  })
}

// 设置角色关联的菜单ID列表
export function fetchSetRoleMenus(roleId: number, menuIds: number[]) {
  return request.post({
    url: '/admin/role/set_menus',
    data: { roleId, menuIds }
  })
}

// 获取菜单列表
export function fetchGetMenuList() {
  return request.get<AppRouteRecord[]>({
    url: '/admin/menus'
  })
}
