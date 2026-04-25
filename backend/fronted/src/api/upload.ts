import { request } from '@/utils/request'


export function uploadAvatar(file: File) {
  // 创建 FormData 对象
  const formData = new FormData()
  formData.append('file', file) // 将文件附加到 FormData

  return request({
    url: '/upload/avatar',
    method: 'post',
    data: formData, // 使用 FormData 作为请求数据
    headers: {
      'Content-Type': 'multipart/form-data' // 设置请求头，表明这是文件上传
    }
  })
}

export function uploadCover(file: File) {
  // 创建 FormData 对象
  const formData = new FormData()
  formData.append('file', file) // 将文件附加到 FormData、
  return  request({
    url:'/upload/cover',
    method:'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data' // 设置请求头，表明这是文件上传
    }
  })
}

export function uploadVideo(file: FormData, params?: Record<string, any>) {
  // 创建 FormData 对象
  const formData = new FormData()
  formData.append('file', file) // 添加视频文件
  if (params) {
    Object.keys(params).forEach((key) => {
      formData.append(key, params[key]) // 添加其他参数
    })
  }

  return request({
    url: '/upload/video',
    method: 'post',
    data: formData, // 使用 FormData 作为请求体
    headers: {
      'Content-Type': 'multipart/form-data' // 设置 multipart/form-data
    }
  })
}