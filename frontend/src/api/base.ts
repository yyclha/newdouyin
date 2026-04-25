import { request } from '@/utils/request'

export function register(params?: any, data?: any) {
  return request({ url: '/base/register', method: 'post', params, data })
}

export function login(params?: any, data?: any) {
  return request({ url: '/base/login', method: 'post', params, data })
}
