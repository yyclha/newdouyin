import axios, { type AxiosError, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import config from '@/config'
import { _notice } from './index'
import cookie from 'js-cookie'

export const axiosInstance = axios.create({
  baseURL: config.baseUrl,
  timeout: 60000
})

axiosInstance.interceptors.request.use(
  (config) => {
    if (!config.headers['Content-Type']) {
      config.headers['Content-Type'] = 'application/json'
    }

    const token =
      window.localStorage.getItem('token') || cookie.get('token') || cookie.get('x-token') || ''

    if (!config.headers['Token']) {
      config.headers['Token'] = token
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    if (data === undefined || data === null || data === '') {
      _notice('请求失败，请稍后重试')
      return { success: false, code: 500, data: [] }
    }

    if (typeof data === 'string') {
      return { success: true, code: 200, data }
    }

    if (data.data === undefined || data.data === null) {
      data.data = data
    }

    let resCode = data.code
    if (resCode) {
      try {
        resCode = Number(resCode)
      } catch (e) {
        data.code = resCode = 500
        data.success = false
      }

      if (resCode === 0) {
        data.code = resCode = 200
        data.success = true
      }

      if (resCode !== 200) {
        _notice(response.data.msg || response.data.message || '请求失败，请稍后重试')
      } else {
        data.success = true
      }
    } else {
      data.code = 200
      data.success = true
    }

    return data
  },
  (error: AxiosError) => {
    console.log('error', error)
    const responseData: any = error.response?.data
    const backendMessage =
      responseData?.msg || responseData?.message || responseData?.data?.msg || responseData?.data?.message

    if (error.response === undefined) {
      _notice('服务器响应超时')
      return { success: false, code: 500, msg: '服务器响应超时', data: [] }
    }

    if (error.response.status >= 500) {
      const message = backendMessage || '服务器出现错误'
      _notice(message)
      return { success: false, code: 500, msg: message, data: responseData?.data ?? [] }
    }

    if (error.response.status === 404) {
      const message = backendMessage || '接口不存在'
      _notice(message)
      return { success: false, code: 404, msg: message, data: responseData?.data ?? [] }
    }

    if (error.response.status === 400) {
      const message = backendMessage || '接口报错'
      _notice(message)
      console.log('backend 400 response', responseData)
      return { success: false, code: 400, msg: message, data: responseData?.data ?? [] }
    }

    if (error.response.status === 401) {
      return {
        success: false,
        code: 401,
        msg: backendMessage || '用户未授权',
        data: responseData?.data ?? []
      }
    }

    const data: any = responseData
    if (data === null || data === undefined) {
      _notice('请求失败，请稍后重试')
      return { success: true, code: 200, data: [] }
    }

    const resCode = data.code
    console.log('IsHere:', data)
    if (data.data === undefined || data.data === null) {
      data.data = data
    }
    if (resCode && typeof resCode === 'number' && resCode !== 200) {
      _notice(backendMessage || '请求失败，请稍后重试')
    } else {
      data.code = 200
      data.success = true
    }
    return data
  }
)

export interface ApiResponse<T = any> {
  data: T
  success: boolean
  code?: number
  msg?: string
  message?: string
}

export async function request<T = any>(config: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return axiosInstance
    .request<T>(config)
    .then((response: any) => {
      if (response && typeof response === 'object' && 'success' in response && 'data' in response) {
        return response as ApiResponse<T>
      }

      if (response && typeof response === 'object' && 'data' in response) {
        return { success: true, data: response.data } as const
      }

      return { success: true, data: response } as const
    })
    .catch((err) => {
      if (err && typeof err === 'object' && 'success' in err && 'data' in err) {
        return err as ApiResponse<T>
      }

      return { success: false, data: err } as const
    })
}
