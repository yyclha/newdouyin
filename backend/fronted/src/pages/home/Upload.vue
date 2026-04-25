<template>
  <div class="LocalVideoUpload">
    <BaseHeader mode="light" backMode="dark" backImg="back">
    </BaseHeader>
    <div class="content">
      <div class="video-preview" v-if="videoSrc">
        <video controls :src="videoSrc" style="width: 100%; max-height: 30vh"></video>
      </div>
      <div v-else class="video-placeholder" @click="triggerVideoUpload">
        <span >请先选择本地视频</span>
      </div>
      <input
        type="file"
        ref="fileInput"
        style="display: none"
        accept="video/*"
        @change="handleVideoUpload"
      />
      <div class="textarea-ctn">
          <textarea
            name=""
            id=""
            cols="30"
            rows="4"
            v-model="description"
            placeholder="添加作品描述"
          ></textarea>
      </div>
      <div class="textarea-ctn">
        <textarea
          name=""
          id=""
          cols="30"
          rows="4"
          v-model="tags"
          placeholder="添加标签（用#分隔）"
        ></textarea>
      </div>
      <div class="parameters">
        <div class="permission-setting mt1r">
          <span>公开权限</span>
          <select v-model="private_status">
            <option value=0>公开</option>
            <option value=1>好友可见</option>
            <option value=2>仅自己可见</option>
          </select>
        </div>
      </div>
      <div class="notice" v-if="notice">
        {{ notice }}
      </div>
      <dy-button
        type="primary"
        :loading="loading"
        :disabled="!videoFile || !description"
        @click="uploadVideo"
      >
        {{ loading ? '上传中' : '上传' }}
      </dy-button>
    </div>
  </div>
</template>

<script>

import { uploadVideo } from '@/api/upload'
import { _notice } from '@/utils'

export default {
  name: 'LocalVideoUpload',
  data() {
    return {
      videoFile: null,
      videoSrc: '',
      description: '',
      tags: '',
      private_status: 0,
      notice: '',
      loading: false,
    }
  },
  methods: {
    triggerVideoUpload() {
      this.$refs.fileInput.click();
    },
    handleVideoUpload(event) {
      const file = event.target.files[0];
      if (!file) return;

      // Validate file type
      const allowedTypes = ['video/mp4', 'video/avi', 'video/mov'];
      if (!allowedTypes.includes(file.type)) {
        _notice('仅支持 MP4/AVI/MOV 格式的视频');
        return;
      }

      // Validate file size (e.g., limit to 50MB)
      const maxSize = 50 * 1024 * 1024; // 50MB
      if (file.size > maxSize) {
        _notice('视频文件不能超过 50MB');
        return;
      }

      this.videoFile = file;
      this.videoSrc = URL.createObjectURL(file);
    },
    async uploadVideo() {
      if (!this.videoFile || !this.description) {
        this.notice = '请完整填写上传信息'
        return
      }
      this.loading = true
      try {
        let formData = new FormData()
        formData =  this.videoFile
        const res = await uploadVideo(formData, {
          description:this.description,
          tags: this.tags,
          private_status: this.private_status
        })
        if (res.success) {
          _notice('视频上传成功')
          this.$router.push('/home')
        } else {
          _notice(res.msg || '上传失败')
        }
      } catch (error) {
        this.notice = '网络错误，请稍后再试'
      } finally {
        this.loading = false
      }
    },
  },
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
    .textarea-ctn {
      width: 100%;
      background: white;
      padding: 15rem;
      box-sizing: border-box;
      margin-top: 10rem;
      border-radius: 2px;
      border: 1px solid black; /* 设置边框为1像素、实线、黑色 */

      textarea {
        font-family: 'Microsoft YaHei UI',fangsong;
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
