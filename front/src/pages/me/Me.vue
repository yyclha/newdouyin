<template>
  <div class="Me">
    <SlideRowList name="baseSlide" style="width: 100%" v-model:active-index="baseActiveIndex">
      <SlideItem>
        <div ref="float" class="float" :class="floatFixed ? 'fixed' : ''">
          <div class="left" @click="$nav('/me/edit-userinfo')">
            <Icon icon="ri:edit-fill" />
            <span>编辑资料</span>
          </div>
          <transition name="fade">
            <div class="center" v-if="floatShowName">
              <p class="name f14 mt1r mb1r">{{ userinfo.nickname }}</p>
            </div>
          </transition>
          <div class="right">
            <div class="item" @click="$nav('/me/request-update')">
              <Icon class="finger" icon="fluent-emoji-high-contrast:middle-finger" />
            </div>
            <div class="item" @click="$nav('/message/visitors')">
              <Icon icon="eva:people-outline" />
            </div>
            <div class="item" @click="_no">
              <Icon icon="ic:round-search" />
            </div>
            <div class="item" @click.stop="baseActiveIndex = 1">
              <Icon icon="ic:round-menu" />
            </div>
          </div>
        </div>

        <div class="scroll" ref="scroll" @touchstart="touchStart" @touchmove="touchMove" @touchend="touchEnd">
          <div ref="desc" class="desc">
            <header
              ref="header"
              :style="{ backgroundImage: `url(${_checkImgUrl(userinfo.cover_url)})` }"
              @click="previewCover = _checkImgUrl(userinfo.cover_url)"
            >
              <div class="info">
                <img
                  :src="_checkImgUrl(userinfo.avatar_small)"
                  class="avatar"
                  @click.stop="previewAvatar = _checkImgUrl(userinfo.avatar_large)"
                />
                <div class="right">
                  <p class="name">{{ userinfo.nickname }}</p>
                  <div class="number mb1r">
                    <span class="mr1r" v-if="userinfo.is_private">私密账号</span>
                    <span>抖音号 {{ _getUserDouyinId({ author: userinfo }) }}</span>
                    <img
                      src="../../assets/img/icon/me/qrcode-gray.png"
                      alt=""
                      @click.stop="$nav('/me/my-card')"
                    />
                  </div>
                </div>
              </div>
            </header>

            <div class="detail">
              <div class="head">
                <div class="heat">
                  <div class="text" @click="isShowStarCount = true">
                    <span class="num">{{ _formatNumber(userinfo.total_favorited) }}</span>
                    <span>获赞</span>
                  </div>
                  <div class="text" @click="$nav('/people/follow-and-fans', { type: 0 })">
                    <span class="num">{{ _formatNumber(userinfo.following_count) }}</span>
                    <span>关注</span>
                  </div>
                  <div class="text" @click="$nav('/people/find-acquaintance', { type: 1 })">
                    <span class="num">{{ _formatNumber(baseStore.friends.all.length) }}</span>
                    <span>朋友</span>
                  </div>
                  <div class="text" @click="$nav('/people/follow-and-fans', { type: 1 })">
                    <span class="num">{{ _formatNumber(userinfo.follower_count) }}</span>
                    <span>粉丝</span>
                  </div>
                </div>
                <div class="button" @click="$nav('/people/find-acquaintance')">添加朋友</div>
              </div>

              <div class="signature" @click="$nav('/me/edit-userinfo-item', { type: 3 })">
                <template v-if="!userinfo.signature">
                  <span>添加简介，让更多人认识你</span>
                  <img src="../../assets/img/icon/me/write-gray.png" alt="" />
                </template>
                <div v-else class="text" v-html="userinfo.signature"></div>
              </div>

              <div class="more" @click="$nav('/me/edit-userinfo')">
                <div class="age item" v-if="userinfo.user_age !== -1">
                  <img v-if="userinfo.gender == 2" src="../../assets/img/icon/me/woman.png" alt="" />
                  <img v-if="userinfo.gender == 1" src="../../assets/img/icon/me/man.png" alt="" />
                  <span>{{ userinfo.user_age }}岁</span>
                </div>
                <div class="item" v-if="userinfo.province || userinfo.city">
                  {{ userinfo.province }}
                  <template v-if="userinfo.province && userinfo.city"> - </template>
                  {{ userinfo.city }}
                </div>
                <div class="item" v-if="userinfo.school?.name">{{ userinfo.school.name }}</div>
              </div>

              <div class="other">
                <div class="item" @click="_no">
                  <Icon icon="iconamoon:shopping-card-light" />
                  <span>电商橱窗</span>
                </div>
                <div class="item" @click="$nav('/me/my-music')">
                  <Icon icon="iconamoon:music-2-light" />
                  <span>我的音乐</span>
                </div>
                <div class="item" @click="_no">
                  <Icon icon="streamline:chat-two-bubbles-oval" />
                  <span>我的群聊</span>
                </div>
                <div class="item" @click="_no">
                  <Icon icon="iconamoon:shopping-card-light" />
                  <span>商品橱窗</span>
                </div>
              </div>
            </div>
          </div>

          <Indicator
            name="videoList"
            tabStyleWidth="25%"
            :tabTexts="['作品', '私密', '喜欢', '收藏']"
            v-model:active-index="contentIndex"
          />

          <div class="video-panel">
            <div v-show="contentIndex === 0" class="SlideItem" @scroll="scroll">
              <Posters
                v-if="videos.my.total !== -1"
                :list="videos.my.list"
                :canDelete="true"
                @delete="confirmDeleteVideo"
              />
              <Loading v-if="loadings.loading0" :is-full-screen="false" />
              <no-more v-else />
            </div>

            <div v-show="contentIndex === 1" class="SlideItem" @scroll="scroll">
              <div class="notice">
                <img src="../../assets/img/icon/me/lock-gray.png" alt="" />
                <span>仅自己可见的内容会展示在这里</span>
              </div>
              <Posters
                v-if="videos.private.total !== -1"
                mode="date"
                :list="videos.private.list"
                :canDelete="true"
                @delete="confirmDeleteVideo"
              />
              <Loading v-if="loadings.loading1" :is-full-screen="false" />
              <no-more v-else />
            </div>

            <div v-show="contentIndex === 2" class="SlideItem" @scroll="scroll">
              <div class="notice">
                <img src="../../assets/img/icon/me/lock-gray.png" alt="" />
                <span>这里显示你点赞过的作品</span>
              </div>
              <Posters v-if="videos.like.total !== -1" :list="videos.like.list" />
              <Loading v-if="loadings.loading2" :is-full-screen="false" />
              <no-more v-else />
            </div>

            <div v-show="contentIndex === 3" class="SlideItem" @scroll="scroll">
              <div class="notice">
                <img src="../../assets/img/icon/me/lock-gray.png" alt="" />
                <span>这里显示你收藏的视频和音乐</span>
              </div>
              <div class="collect" ref="collect">
                <div class="video" v-if="videos.collect.video.total !== -1">
                  <div class="top" @click="$nav('/me/collect/video-collect')">
                    <div class="left">
                      <img src="../../assets/img/icon/me/video-whitegray.png" alt="" />
                      <span>视频</span>
                    </div>
                    <div class="right">
                      <span>查看全部</span>
                      <dy-back direction="right"></dy-back>
                    </div>
                  </div>
                  <Posters :list="videos.collect.video.list" />
                </div>

                <div class="music" v-if="videos.collect.music.total !== -1">
                  <div class="top" @click="$nav('/me/collect/music-collect')">
                    <div class="left">
                      <img src="../../assets/img/icon/me/music-whitegray.png" alt="" />
                      <span>音乐</span>
                    </div>
                    <div class="right">
                      <span>查看全部</span>
                      <dy-back direction="right"></dy-back>
                    </div>
                  </div>
                  <div class="list">
                    <div
                      v-for="(i, j) in videos.collect.music.list.slice(0, 3)"
                      :key="j"
                      class="item"
                      @click.stop="$nav('/home/music', i)"
                    >
                      <img class="poster" :src="_checkImgUrl(i.cover)" alt="" />
                      <div class="title">{{ i.name }}</div>
                    </div>
                  </div>
                </div>
              </div>
              <Loading v-if="loadings.loading3" :is-full-screen="false" />
              <no-more v-else />
            </div>
          </div>
        </div>

        <BaseFooter v-bind:init-tab="5" />
        <transition name="fade">
          <div class="mask" v-if="baseActiveIndex === 1" @click="baseActiveIndex = 0"></div>
        </transition>
      </SlideItem>

      <SlideItem style="width: 70vw; overflow: auto">
        <transition name="fade1">
          <div class="ul" v-if="!isMoreFunction">
            <div class="li" @click="_no">
              <img src="../../assets/img/icon/newicon/left_menu/shopping.png" alt="" />
              <span>我的商城</span>
            </div>
            <div class="li" @click="_no">
              <img src="../../assets/img/icon/newicon/left_menu/wallet.png" alt="" />
              <span>钱包</span>
            </div>
            <div class="line"></div>
            <div class="li" @click="$nav('/me/my-card')">
              <img src="../../assets/img/icon/newicon/left_menu/qrcode.png" alt="" />
              <span>我的二维码</span>
            </div>
            <div class="li" @click="$nav('/me/right-menu/look-history')">
              <img src="../../assets/img/icon/newicon/left_menu/time.png" alt="" />
              <span>观看历史</span>
            </div>
            <div class="li" @click="_no">
              <img src="../../assets/img/icon/newicon/left_menu/clock.png" alt="" />
              <span>稍后再看</span>
            </div>
            <div class="li" @click="_no">
              <img src="../../assets/img/icon/newicon/left_menu/workbench.png" alt="" />
              <span>创作者服务中心</span>
            </div>
            <div class="line"></div>
            <div class="li" @click="_no">
              <img src="../../assets/img/icon/newicon/left_menu/bytedance-mini-app.png" alt="" />
              <span>小程序</span>
            </div>
            <div class="li" @click="_no">
              <img src="../../assets/img/icon/newicon/left_menu/gongyi.png" alt="" />
              <span>公益</span>
            </div>
            <div class="li" @click="$nav('/me/right-menu/minor-protection/index')">
              <img src="../../assets/img/icon/newicon/left_menu/umbrella.png" alt="" />
              <span>青少年模式</span>
            </div>
            <div class="li" @click="_no">
              <img src="../../assets/img/icon/newicon/left_menu/headset.png" alt="" />
              <span>我的客服</span>
            </div>
            <div class="li" @click="$nav('/me/right-menu/setting')">
              <img src="../../assets/img/icon/newicon/left_menu/setting-one.png" alt="" />
              <span>设置</span>
            </div>
          </div>
          <div v-else class="more-function">
            <div class="title">更多功能</div>
            <div class="functions">
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/quan.png" alt="" />
                <span>圈子</span>
              </div>
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/sd-card.png" alt="" />
                <span>离线模式</span>
              </div>
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/alarmmmmmmmmmmmm.png" alt="" />
                <span>视频管理</span>
              </div>
            </div>
            <div class="title">更多服务</div>
            <div class="functions">
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/sun-one.png" alt="" />
                <span>稍后再看</span>
              </div>
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/download.png" alt="" />
                <span>离线缓存</span>
              </div>
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/hot.png" alt="" />
                <span>热门内容</span>
              </div>
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/shop.png" alt="" />
                <span>商城</span>
              </div>
              <div class="function" @click="_no">
                <img src="../../assets/img/icon/newicon/left_menu/yuandi.png" alt="" />
                <span>同城</span>
              </div>
            </div>
          </div>
        </transition>
        <div class="button-ctn">
          <div class="button" v-if="!isMoreFunction" @click="isMoreFunction = true">
            <img src="../../assets/img/icon/newicon/left_menu/more.png" alt="" />
            <span>更多功能</span>
          </div>
          <div class="button" v-if="isMoreFunction" @click="isMoreFunction = false">
            <span>收起</span>
          </div>
        </div>
      </SlideItem>
    </SlideRowList>

    <transition name="fade">
      <div class="preview-img" v-if="previewCover" @click="previewCover = ''">
        <img class="resource" :src="previewCover" alt="" />
        <img
          class="upload"
          src="@/assets/img/icon/components/video/upload.png"
          alt="Upload"
          @click.stop="upload_cover"
        />
        <img
          class="download"
          src="@/assets/img/icon/components/video/download.png"
          alt="Download"
          @click.stop="_no"
        />
      </div>
    </transition>

    <transition name="fade">
      <div class="preview-img" v-if="previewAvatar" @click="previewAvatar = ''">
        <img class="resource" :src="previewAvatar" alt="" />
        <img
          class="upload"
          src="@/assets/img/icon/components/video/upload.png"
          alt="Upload"
          @click.stop="upload_avatar"
        />
        <img
          class="download"
          src="@/assets/img/icon/components/video/download.png"
          alt="Download"
          @click.stop="_no"
        />
      </div>
    </transition>

    <ConfirmDialog
      v-model:visible="isShowStarCount"
      :subtitle="`“${userinfo.nickname}”累计获得 ${_formatNumber(userinfo.total_favorited)} 个赞`"
      okText="知道了"
      cancelText="关闭"
      @ok="isShowStarCount = false"
      @cancel="isShowStarCount = false"
    >
      <template v-slot:header>
        <img style="width: 100%" src="../../assets/img/icon/star-bg.png" alt="" />
      </template>
    </ConfirmDialog>
  </div>
</template>

<script>
import Posters from '../../components/Posters'
import Indicator from '../../components/slide/Indicator'
import { nextTick } from 'vue'
import { mapState } from 'pinia'

import bus from '../../utils/bus'
import ConfirmDialog from '../../components/dialog/ConfirmDialog'
import {
  _checkImgUrl,
  _formatNumber,
  _getUserDouyinId,
  _no,
  _notice,
  _showConfirmDialog,
  _stopPropagation
} from '@/utils'
import { collectVideo, deleteMyVideo, likeVideo, myVideo, privateVideo } from '@/api/videos'
import { useBaseStore } from '@/store/pinia'
import { uploadAvatar, uploadCover } from '@/api/upload'
import SlideRowList from '@/components/slide/SlideRowList.vue'

export default {
  name: 'Me',
  components: { SlideRowList, Posters, Indicator, ConfirmDialog },
  data() {
    return {
      baseStore: useBaseStore(),
      previewCover: '',
      previewAvatar: '',
      contentIndex: 0,
      baseActiveIndex: 0,
      isShowStarCount: false,
      floatFixed: false,
      floatShowName: false,
      isMoreFunction: false,
      videos: {
        my: { list: [], total: -1, pageNo: 0 },
        private: { list: [], total: -1, pageNo: 0 },
        like: { list: [], total: -1, pageNo: 0 },
        collect: {
          video: { list: [], total: -1, pageNo: 0 },
          music: { list: [], total: -1 }
        }
      },
      pageSize: 15,
      loadings: {
        loading0: false,
        loading1: false,
        loading2: false,
        loading3: false
      },
      canScroll: true
    }
  },
  computed: {
    ...mapState(useBaseStore, ['userinfo'])
  },
  watch: {
    contentIndex(newVal) {
      this.changeIndex(newVal)
    }
  },
  mounted() {
    nextTick(async () => {
      await this.baseStore.init()
      await this.changeIndex(0)
    })
    bus.on('baseSlide-moved', () => (this.canScroll = false))
    bus.on('baseSlide-end', () => (this.canScroll = true))
  },
  activated() {
    if (this.userinfo.uid && this.baseStore.meNeedRefresh) {
      this.reloadMeData()
    }
  },
  methods: {
    async reloadMeData() {
      this.videos.my = { list: [], total: -1, pageNo: 0 }
      this.videos.private = { list: [], total: -1, pageNo: 0 }
      this.videos.like = { list: [], total: -1, pageNo: 0 }
      this.videos.collect = {
        video: { list: [], total: -1, pageNo: 0 },
        music: { list: [], total: -1 }
      }
      this.setLoadingFalse()
      await this.baseStore.init()
      await this.changeIndex(this.contentIndex)
      this.baseStore.markMeRefresh(false)
    },
    confirmDeleteVideo(item) {
      _showConfirmDialog('确认删除该作品吗？', '删除后将无法恢复，请谨慎操作。', 'gray', () =>
        this.handleDeleteVideo(item)
      )
    },
    async handleDeleteVideo(item) {
      const awemeId = item?.aweme_id
      if (!awemeId) return

      const res = await deleteMyVideo({ aweme_id: awemeId })
      if (!res.success) return

      _notice('删除成功')

      if (this.contentIndex === 0) {
        this.videos.my.list = this.videos.my.list.filter((video) => video.aweme_id !== awemeId)
        this.videos.my.total = Math.max((this.videos.my.total || 0) - 1, 0)
      } else if (this.contentIndex === 1) {
        this.videos.private.list = this.videos.private.list.filter((video) => video.aweme_id !== awemeId)
        this.videos.private.total = Math.max((this.videos.private.total || 0) - 1, 0)
      }

      this.baseStore.markMeRefresh(true)
      await this.baseStore.refreshPanel()
    },
    _no,
    upload_cover() {
      const fileInput = document.createElement('input')
      fileInput.type = 'file'
      fileInput.accept = 'image/*'
      fileInput.onchange = async () => {
        const file = fileInput.files[0]
        await uploadCover(file)
      }
      fileInput.click()
      this.$nav('/me')
    },
    upload_avatar() {
      const fileInput = document.createElement('input')
      fileInput.type = 'file'
      fileInput.accept = 'image/*'
      fileInput.onchange = async () => {
        const file = fileInput.files[0]
        await uploadAvatar(file)
      }
      fileInput.click()
      this.$nav('/me')
    },
    _getUserDouyinId,
    _checkImgUrl,
    _formatNumber,
    $nav(path) {
      this.$router.push(path)
    },
    setLoadingFalse() {
      this.loadings.loading0 = false
      this.loadings.loading1 = false
      this.loadings.loading2 = false
      this.loadings.loading3 = false
    },
    async requestVideoList(tabIndex, retry = true) {
      let res
      switch (tabIndex) {
        case 0:
          res = await myVideo({ pageNo: this.videos.my.pageNo, pageSize: this.pageSize })
          if (!res.success && retry) {
            await this.baseStore.init()
            return this.requestVideoList(tabIndex, false)
          }
          if (res.success) this.videos.my = res.data
          break
        case 1:
          res = await privateVideo({ pageNo: this.videos.private.pageNo, pageSize: this.pageSize })
          if (!res.success && retry) {
            await this.baseStore.init()
            return this.requestVideoList(tabIndex, false)
          }
          if (res.success) this.videos.private = res.data
          break
        case 2:
          res = await likeVideo({ pageNo: this.videos.like.pageNo, pageSize: this.pageSize })
          if (!res.success && retry) {
            await this.baseStore.init()
            return this.requestVideoList(tabIndex, false)
          }
          if (res.success) this.videos.like = res.data
          break
      }
      return res
    },
    async changeIndex(newVal) {
      if (this.loadings['loading' + newVal]) return
      const keys = ['my', 'private', 'like', 'collect']
      const videoOb = this.videos[keys[newVal]]
      if (newVal === 3) {
        if (videoOb.video.total === -1) {
          this.loadings.loading3 = true
          const res = await collectVideo({
            pageNo: this.videos.collect.video.pageNo,
            pageSize: this.pageSize
          })
          if (res.success) this.videos.collect = res.data
        }
      } else if (videoOb.total === -1) {
        this.loadings['loading' + newVal] = true
        await this.requestVideoList(newVal)
      }
      this.setLoadingFalse()
    },
    async loadMoreData() {
      if (this.loadings['loading' + this.contentIndex]) return
      if (this.contentIndex === 3) return

      const keys = ['my', 'private', 'like']
      const key = keys[this.contentIndex]
      const videoOb = this.videos[key]
      if (videoOb.total <= videoOb.list.length) return

      videoOb.pageNo++
      this.loadings['loading' + this.contentIndex] = true
      const res = await this.requestVideoList(this.contentIndex, false)
      this.loadings['loading' + this.contentIndex] = false
      if (res?.success && res.data?.list) {
        videoOb.list = videoOb.list.concat(res.data.list)
        videoOb.total = res.data.total
      }
    },
    async scroll() {
      if (!this.canScroll) return
      await this.loadMoreData()
    },
    touchStart() {},
    touchMove() {},
    touchEnd() {},
    click(e) {
      if (this.baseActiveIndex === 1) {
        this.baseActiveIndex = 0
        _stopPropagation(e)
      }
    }
  }
}
</script>

<style scoped lang="less">
@import 'Me';
</style>
