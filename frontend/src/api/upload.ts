import { request } from '@/utils/request'

export function uploadAvatar(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  return request({
    url: '/upload/avatar',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function uploadCover(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  return request({
    url: '/upload/cover',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function initVideoUpload(payload: {
  upload_id?: string
  file_name: string
  file_size: number
  chunk_size: number
  total_chunks: number
  content_type: string
  description: string
  tags: string
  private_status: number
}) {
  return request({
    url: '/upload/video/init',
    method: 'post',
    data: payload
  })
}

export function uploadVideoChunk(payload: {
  upload_id: string
  chunk_index: number
  total_chunks: number
  chunk_hash: string
  file: Blob
  onUploadProgress?: (event: any) => void
}) {
  const formData = new FormData()
  formData.append('upload_id', payload.upload_id)
  formData.append('chunk_index', String(payload.chunk_index))
  formData.append('total_chunks', String(payload.total_chunks))
  formData.append('chunk_hash', payload.chunk_hash)
  formData.append('file', payload.file, `chunk-${payload.chunk_index}.part`)

  return request({
    url: '/upload/video/chunk',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    onUploadProgress: payload.onUploadProgress
  })
}

export function completeVideoUpload(payload: { upload_id: string }) {
  return request({
    url: '/upload/video/complete',
    method: 'post',
    data: payload
  })
}

export function getVideoUploadStatus(taskId: string) {
  return request({
    url: '/upload/video/status',
    method: 'get',
    params: {
      task_id: taskId
    }
  })
}
