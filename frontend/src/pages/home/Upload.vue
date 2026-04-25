<template>
  <div class="LocalVideoUpload">
    <BaseHeader mode="light" backMode="dark" backImg="back"> </BaseHeader>
    <div class="content">
      <div v-if="videoSrc" class="video-preview">
        <video controls :src="videoSrc" style="width: 100%; max-height: 30vh"></video>
      </div>
      <div v-else class="video-placeholder" @click="triggerVideoUpload">
        <span>选择本地视频</span>
      </div>
      <input
        ref="fileInput"
        type="file"
        style="display: none"
        accept="video/mp4,video/quicktime,video/x-msvideo"
        @change="handleVideoUpload"
      />
      <div class="textarea-ctn">
        <textarea
          v-model.trim="description"
          cols="30"
          rows="4"
          placeholder="添加作品描述"
        ></textarea>
      </div>
      <div class="textarea-ctn">
        <textarea
          v-model.trim="tags"
          cols="30"
          rows="4"
          placeholder="添加话题，例如 #旅行 #日常"
        ></textarea>
      </div>
      <div class="parameters">
        <div class="permission-setting mt1r">
          <span>可见范围</span>
          <select v-model.number="private_status">
            <option :value="0">公开</option>
            <option :value="1">好友可见</option>
            <option :value="2">仅自己可见</option>
          </select>
        </div>
      </div>
      <div v-if="loading || uploadProgress > 0" class="upload-progress">
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: `${uploadProgress}%` }"></div>
        </div>
        <div class="progress-meta">
          <span>{{ progressText }}</span>
          <span>{{ Math.round(uploadProgress) }}%</span>
        </div>
      </div>
      <div v-if="notice" class="notice">
        {{ notice }}
      </div>
      <dy-button
        type="primary"
        :loading="loading"
        :disabled="loading || !videoFile || !description"
        @click="submitVideo"
      >
        {{ loading ? '上传中...' : '发布' }}
      </dy-button>
    </div>
  </div>
</template>

<script>
import { completeVideoUpload, initVideoUpload, uploadVideoChunk } from '@/api/upload'
import { _notice } from '@/utils'
import { useBaseStore } from '@/store/pinia'

const MAX_VIDEO_SIZE = 50 * 1024 * 1024
const CHUNK_SIZE = 2 * 1024 * 1024
const CHUNK_RETRY_TIMES = 3
const ALLOWED_VIDEO_TYPES = ['video/mp4', 'video/quicktime', 'video/x-msvideo']
const VIDEO_UPLOAD_SESSION_STORAGE_KEY = 'video-upload-sessions'

export default {
  name: 'LocalVideoUpload',
  data() {
    return {
      baseStore: useBaseStore(),
      videoFile: null,
      videoSrc: '',
      description: '',
      tags: '',
      private_status: 0,
      notice: '',
      loading: false,
      uploadProgress: 0,
      progressText: '等待上传'
    }
  },
  beforeUnmount() {
    this.revokeVideoPreview()
  },
  methods: {
    triggerVideoUpload() {
      this.$refs.fileInput?.click()
    },
    revokeVideoPreview() {
      if (this.videoSrc) {
        URL.revokeObjectURL(this.videoSrc)
        this.videoSrc = ''
      }
    },
    resetFileInput() {
      if (this.$refs.fileInput) {
        this.$refs.fileInput.value = ''
      }
    },
    handleVideoUpload(event) {
      const input = event.target
      const file = input?.files?.[0]
      if (!file) {
        return
      }

      if (!ALLOWED_VIDEO_TYPES.includes(file.type)) {
        this.notice = '仅支持 MP4、MOV 和 AVI 格式的视频'
        _notice(this.notice)
        this.resetFileInput()
        return
      }

      if (file.size > MAX_VIDEO_SIZE) {
        this.notice = '视频大小不能超过 50MB'
        _notice(this.notice)
        this.resetFileInput()
        return
      }

      this.notice = ''
      this.uploadProgress = 0
      this.progressText = '等待上传'
      this.videoFile = file
      this.revokeVideoPreview()
      this.videoSrc = URL.createObjectURL(file)
    },
    isRequestSuccess(res) {
      return !!res && res.success !== false && Number(res.code || 200) === 200
    },
    getResponseMessage(res, fallbackMessage) {
      return res?.msg || res?.message || fallbackMessage
    },
    shouldRetryChunkUpload(res) {
      const code = Number(res?.code || 0)
      return ![400, 401, 403, 404].includes(code)
    },
    updateUploadProgress(loadedBytes, totalBytes, text) {
      const percent = totalBytes > 0 ? Math.min((loadedBytes / totalBytes) * 100, 100) : 0
      this.uploadProgress = Number(percent.toFixed(2))
      this.progressText = text
    },
    getChunkTotal(file) {
      return Math.ceil(file.size / CHUNK_SIZE)
    },
    getChunkBlob(file, chunkIndex) {
      const start = chunkIndex * CHUNK_SIZE
      const end = Math.min(start + CHUNK_SIZE, file.size)
      return file.slice(start, end)
    },
    calculateUploadedBytes(file, uploadedChunks, totalChunks) {
      return [...uploadedChunks].reduce((sum, chunkIndex) => {
        if (chunkIndex < 0 || chunkIndex >= totalChunks) {
          return sum
        }
        return sum + this.getChunkBlob(file, chunkIndex).size
      }, 0)
    },
    async hashBlob(blob) {
      if (!window.crypto?.subtle) {
        throw new Error('当前浏览器不支持分片完整性校验')
      }

      const hashBuffer = await window.crypto.subtle.digest('SHA-256', await blob.arrayBuffer())
      return Array.from(new Uint8Array(hashBuffer))
        .map((item) => item.toString(16).padStart(2, '0'))
        .join('')
    },
    getFileFingerprint(file) {
      return [file.name, file.size, file.lastModified, file.type].join(':')
    },
    readUploadSessionMap() {
      try {
        const raw = window.localStorage.getItem(VIDEO_UPLOAD_SESSION_STORAGE_KEY)
        if (!raw) {
          return {}
        }

        const parsed = JSON.parse(raw)
        return parsed && typeof parsed === 'object' ? parsed : {}
      } catch (error) {
        return {}
      }
    },
    writeUploadSessionMap(sessionMap) {
      window.localStorage.setItem(VIDEO_UPLOAD_SESSION_STORAGE_KEY, JSON.stringify(sessionMap))
    },
    getSavedUploadId(file) {
      const sessionMap = this.readUploadSessionMap()
      return sessionMap[this.getFileFingerprint(file)] || ''
    },
    saveUploadId(file, uploadId) {
      if (!uploadId) {
        return
      }

      const sessionMap = this.readUploadSessionMap()
      sessionMap[this.getFileFingerprint(file)] = uploadId
      this.writeUploadSessionMap(sessionMap)
    },
    clearUploadId(file) {
      const fingerprint = this.getFileFingerprint(file)
      const sessionMap = this.readUploadSessionMap()
      if (!sessionMap[fingerprint]) {
        return
      }

      delete sessionMap[fingerprint]
      this.writeUploadSessionMap(sessionMap)
    },
    async uploadChunkWithRetry({
      uploadId,
      chunkIndex,
      totalChunks,
      chunkBlob,
      chunkHash,
      uploadedBytes,
      fileSize
    }) {
      for (let attempt = 1; attempt <= CHUNK_RETRY_TIMES; attempt += 1) {
        const res = await uploadVideoChunk({
          upload_id: uploadId,
          chunk_index: chunkIndex,
          total_chunks: totalChunks,
          chunk_hash: chunkHash,
          file: chunkBlob,
          onUploadProgress: (event) => {
            const currentLoaded = event?.loaded || 0
            this.updateUploadProgress(
              Math.min(uploadedBytes + currentLoaded, fileSize),
              fileSize,
              `正在上传分片 ${chunkIndex + 1}/${totalChunks}`
            )
          }
        })

        if (this.isRequestSuccess(res)) {
          return res.data ?? {}
        }

        if (!this.shouldRetryChunkUpload(res)) {
          throw new Error(this.getResponseMessage(res, `第 ${chunkIndex + 1} 个分片上传失败`))
        }

        if (attempt === CHUNK_RETRY_TIMES) {
          throw new Error(this.getResponseMessage(res, `第 ${chunkIndex + 1} 个分片上传失败`))
        }
      }

      throw new Error(`第 ${chunkIndex + 1} 个分片上传失败`)
    },
    async handleUploadFinished(payload) {
      const uploadStatus = payload?.status || 'done'
      const isQueued = uploadStatus === 'queued'
      const successNotice = isQueued
        ? '视频已进入后台处理队列'
        : '视频上传成功'

      this.updateUploadProgress(1, 1, isQueued ? '已进入后台处理队列' : '上传完成')
      _notice(successNotice)
      this.notice = successNotice
      this.baseStore.markMeRefresh(true)
      if (!isQueued) {
        await this.baseStore.refreshPanel()
      }
      if (this.videoFile) {
        this.clearUploadId(this.videoFile)
      }
      this.revokeVideoPreview()
      this.videoFile = null
      this.description = ''
      this.tags = ''
      this.private_status = 0
      this.resetFileInput()
      this.$router.push('/home')
    },
    async submitVideo() {
      if (!this.videoFile) {
        this.notice = '请先选择视频'
        return
      }
      if (!this.description) {
        this.notice = '请输入作品描述'
        return
      }

      const file = this.videoFile
      const totalChunks = this.getChunkTotal(file)
      this.notice = ''
      this.loading = true
      this.uploadProgress = 0
      this.progressText = '正在初始化上传会话'

      try {
        const savedUploadId = this.getSavedUploadId(file)
        const initRes = await initVideoUpload({
          upload_id: savedUploadId || undefined,
          file_name: file.name,
          file_size: file.size,
          chunk_size: CHUNK_SIZE,
          total_chunks: totalChunks,
          content_type: file.type,
          description: this.description,
          tags: this.tags,
          private_status: this.private_status
        })

        if (!this.isRequestSuccess(initRes)) {
          this.notice = this.getResponseMessage(initRes, '初始化上传失败')
          _notice(this.notice)
          return
        }

        const initData = initRes.data ?? {}
        const uploadId = initData.uploadId || initData.upload_id || savedUploadId
        if (!uploadId) {
          throw new Error('服务器未返回 uploadId')
        }
        this.saveUploadId(file, uploadId)

        if (initData.status === 'queued') {
          await this.handleUploadFinished(initData)
          return
        }

        const uploadedChunks = new Set((initData.uploadedChunks || []).map((item) => Number(item)))
        let uploadedBytes = this.calculateUploadedBytes(file, uploadedChunks, totalChunks)
        this.updateUploadProgress(uploadedBytes, file.size, '正在继续上传分片')

        for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex += 1) {
          if (uploadedChunks.has(chunkIndex)) {
            continue
          }

          const chunkBlob = this.getChunkBlob(file, chunkIndex)
          this.progressText = `正在校验分片 ${chunkIndex + 1}/${totalChunks}`
          const chunkHash = await this.hashBlob(chunkBlob)
          await this.uploadChunkWithRetry({
            uploadId,
            chunkIndex,
            totalChunks,
            chunkBlob,
            chunkHash,
            uploadedBytes,
            fileSize: file.size
          })
          uploadedBytes += chunkBlob.size
          this.updateUploadProgress(
            uploadedBytes,
            file.size,
            `已上传 ${chunkIndex + 1}/${totalChunks} 个分片`
          )
        }

        this.progressText = '服务器正在合并视频'
        const completeRes = await completeVideoUpload({
          upload_id: uploadId
        })
        if (!this.isRequestSuccess(completeRes)) {
          const missingChunks = completeRes?.data?.missingChunks
          const missingSuffix =
            Array.isArray(missingChunks) && missingChunks.length
              ? ` 缺失分片：${missingChunks.join(', ')}`
              : ''
          this.notice = `${this.getResponseMessage(completeRes, '合并视频失败')}${missingSuffix}`
          _notice(this.notice)
          return
        }

        await this.handleUploadFinished(completeRes.data ?? {})
      } catch (error) {
        this.notice = error?.message || '上传失败，请稍后重试'
        _notice(this.notice)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped lang="less">
@import '../../assets/less/index';

.LocalVideoUpload {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  top: 0;
  overflow: auto;
  color: black;
  font-size: 14rem;
  background: white;

  .content {
    padding: 60rem 30rem;

    .video-preview {
      margin-bottom: 30rem;
      border: 1px solid #ddd;
      border-radius: 10rem;
      overflow: hidden;
    }

    .video-placeholder {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 30vh;
      background: #f7f7f7;
      border: 1px dashed #ccc;
      border-radius: 10rem;
      margin-bottom: 30rem;
      color: #999;
      font-size: 16rem;
    }

    .parameters {
      .permission-setting {
        margin-bottom: 20rem;
        display: flex;
        justify-content: space-between;
        align-items: center;

        select {
          padding: 10rem;
          font-size: 14rem;
          border: 1px solid #ddd;
          border-radius: 5rem;
        }
      }
    }

    .upload-progress {
      margin: 20rem 0;

      .progress-track {
        width: 100%;
        height: 8rem;
        border-radius: 999rem;
        background: #ececec;
        overflow: hidden;
      }

      .progress-fill {
        height: 100%;
        border-radius: 999rem;
        background: #fe2c55;
        transition: width 0.2s ease;
      }

      .progress-meta {
        margin-top: 10rem;
        display: flex;
        justify-content: space-between;
        color: #666;
        font-size: 12rem;
      }
    }

    .notice {
      margin: 20rem 0;
      color: #d93025;
      line-height: 1.5;
    }

    .textarea-ctn {
      width: 100%;
      background: white;
      padding: 15rem;
      box-sizing: border-box;
      margin-top: 10rem;
      border-radius: 2px;
      border: 1px solid black;

      textarea {
        font-family: 'Microsoft YaHei UI', fangsong;
        outline: none;
        width: 100%;
        border: none;
        background: transparent;
        color: black;

        &::placeholder {
          color: var(--second-text-color);
        }
      }
    }
  }
}
</style>
