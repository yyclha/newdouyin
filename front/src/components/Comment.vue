<template>
  <from-bottom-dialog
    :page-id="pageId"
    :modelValue="modelValue"
    @update:modelValue="(e) => $emit('update:modelValue', e)"
    @cancel="cancel"
    @scroll="handleDialogScroll"
    :show-heng-gang="false"
    maskMode="light"
    :height="height"
    tag="comment"
    mode="white"
  >
    <template v-slot:header>
      <div class="title">
        <dy-back mode="dark" img="close" direction="right" style="opacity: 0" />
        <div class="num">{{ _formatNumber(totalComments) }}条评论</div>
        <div class="right">
          <Icon icon="prime:arrow-up-right-and-arrow-down-left-from-center" @click.stop="_no" />
          <Icon icon="ic:round-close" v-click="cancel" />
        </div>
      </div>
    </template>
    <div class="comment">
      <div v-if="comments.length" class="wrapper">
        <div class="items">
          <div
            v-for="(item, i) in comments"
            :key="item.comment_id || i"
            class="item"
            v-longpress="() => showOptions(item)"
          >
            <div class="main">
              <div class="content">
                <img :src="_checkImgUrl(item.avatar)" alt="" class="head-image" />
                <div class="comment-container">
                  <div class="name">{{ item.nickname }}</div>
                  <div class="detail" :class="item.user_buried && 'gray'">
                    {{ item.user_buried ? '该评论已折叠' : item.content }}
                  </div>
                  <div class="time-wrapper">
                    <div class="left">
                      <div class="time">
                        {{ _time(item.create_time) }}{{ item.ip_location ? ` · ${item.ip_location}` : '' }}
                      </div>
                      <div class="reply-text">回复</div>
                    </div>
                    <div class="right d-flex" style="gap: 10rem">
                      <div class="love" :class="item.user_digged && 'loved'" @click="loved(item)">
                        <Icon
                          icon="icon-park-solid:like"
                          v-show="item.user_digged"
                          class="love-image"
                        />
                        <Icon
                          icon="icon-park-outline:like"
                          v-show="!item.user_digged"
                          class="love-image"
                        />
                        <span v-if="item.digg_count">{{ _formatNumber(item.digg_count) }}</span>
                      </div>
                      <div class="love" @click="item.user_buried = !item.user_buried">
                        <Icon
                          v-if="item.user_buried"
                          icon="icon-park-solid:dislike-two"
                          class="love-image"
                        />
                        <Icon v-else icon="icon-park-outline:dislike" class="love-image" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="Number(item.sub_comment_count)" class="replies">
              <template v-if="item.showChildren">
                <div
                  v-for="(child, childIndex) in item.children"
                  :key="child.comment_id || childIndex"
                  class="reply"
                >
                  <div class="content">
                    <img :src="_checkImgUrl(child.avatar)" alt="" class="head-image" />
                    <div class="comment-container">
                      <div class="name">
                        {{ child.nickname }}
                        <div v-if="child.replay" class="reply-user"></div>
                        {{ child.replay }}
                      </div>
                      <div class="detail">{{ child.content }}</div>
                      <div class="time-wrapper">
                        <div class="left">
                          <div class="time">
                            {{ _time(child.create_time) }}{{ child.ip_location ? ` · ${child.ip_location}` : '' }}
                          </div>
                          <div class="reply-text">回复</div>
                        </div>
                        <div class="love" :class="child.user_digged && 'loved'" @click="loved(child)">
                          <Icon
                            icon="icon-park-solid:like"
                            v-show="child.user_digged"
                            class="love-image"
                          />
                          <Icon
                            icon="icon-park-outline:like"
                            v-show="!child.user_digged"
                            class="love-image"
                          />
                          <span v-if="child.digg_count">{{ _formatNumber(child.digg_count) }}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
              <Loading
                v-if="loadChildren && loadChildrenItemCId === item.comment_id"
                :type="'small'"
                :is-full-screen="false"
              />
              <div v-else class="more" @click="handShowChildren(item)">
                <div class="gang"></div>
                <span>{{ item.showChildren ? '收起回复' : `展开${item.sub_comment_count}条回复` }}</span>
                <Icon icon="ep:arrow-down-bold" />
              </div>
            </div>
          </div>
        </div>

        <div v-if="hasMore" class="load-more" @click="loadMore">
          <span v-if="!loadingMore">加载更多评论</span>
          <span v-else>加载中...</span>
        </div>
        <no-more v-else />
      </div>

      <div v-else-if="commentsLoaded" class="empty-state">暂无评论</div>
      <Loading v-else style="position: absolute" />

      <transition name="fade">
        <BaseMask v-if="isCall" mode="lightgray" @click="isCall = false" />
      </transition>

      <div class="input-toolbar">
        <transition name="fade">
          <div v-if="isCall" class="call-friend">
            <div
              v-for="(friend, i) in friends?.all || []"
              :key="i"
              class="friend"
              @click="toggleCall(friend)"
            >
              <img
                :style="friend.select ? 'opacity: .5;' : ''"
                class="avatar"
                :src="_checkImgUrl(friend.avatar_small?.url_list?.[0])"
                alt=""
              />
              <span>{{ friend.nickname }}</span>
              <img
                v-if="friend.select"
                class="checked"
                src="../assets/img/icon/components/check/check-red-share.png"
              />
            </div>
          </div>
        </transition>

        <div class="toolbar">
          <div class="input-wrapper">
            <AutoInput v-model="comment" placeholder="留下你的精彩评论吧" />
            <div class="right">
              <img src="../assets/img/icon/message/call.png" @click="isCall = !isCall" />
              <img src="../assets/img/icon/message/emoji-black.png" @click="_no" />
            </div>
          </div>
          <img v-if="comment" src="../assets/img/icon/message/up.png" @click="send(item)" />
        </div>
      </div>

      <ConfirmDialog title="私信聊天" ok-text="发送" v-model:visible="showPrivateChat">
        <Search mode="light" v-model="test" :isShowSearchIcon="false" />
      </ConfirmDialog>
    </div>
  </from-bottom-dialog>
</template>

<script lang="ts">
import AutoInput from './AutoInput.vue'
import ConfirmDialog from './dialog/ConfirmDialog.vue'
import { mapState } from 'pinia'
import FromBottomDialog from './dialog/FromBottomDialog.vue'
import Loading from './Loading.vue'
import Search from './Search.vue'
import {
  _checkImgUrl,
  _formatNumber,
  _no,
  _showSelectDialog,
  _showSimpleConfirmDialog,
  _sleep,
  _time,
  sampleSize
} from '@/utils'
import { useBaseStore } from '@/store/pinia'
import { commentDigg, deleteComment, videoComment, videoComments } from '@/api/videos'

export default {
  name: 'Comment',
  components: {
    AutoInput,
    ConfirmDialog,
    FromBottomDialog,
    Loading,
    Search
  },
  props: {
    modelValue: {
      type: Boolean,
      default() {
        return false
      }
    },
    item: {
      type: Object,
      default: () => {
        return {}
      }
    },
    videoId: {
      type: String,
      default: null
    },
    pageId: {
      type: String,
      default: 'home-index'
    },
    height: {
      type: String,
      default: 'calc(var(--vh, 1vh) * 70)'
    }
  },
  computed: {
    ...mapState(useBaseStore, ['friends'])
  },
  watch: {
    modelValue(newVal) {
      if (newVal) {
        this.getData()
      } else {
        this.resetCommentState()
      }
    }
  },
  data() {
    return {
      comment: '',
      test: '',
      comments: [],
      totalComments: 0,
      pageNo: 0,
      pageSize: 20,
      hasMore: false,
      loadingMore: false,
      commentsLoaded: false,
      options: [
        { id: 1, name: '私信回复' },
        { id: 2, name: '举报' },
        { id: 3, name: '复制' },
        { id: 4, name: '删除评论' }
      ],
      selectRow: {},
      showPrivateChat: false,
      isInput: false,
      isCall: false,
      loadChildren: false,
      loadChildrenItemCId: -1
    }
  },
  mounted() {},
  methods: {
    _no,
    _time,
    _formatNumber,
    _checkImgUrl,
    resetCommentState() {
      this.comments = []
      this.totalComments = 0
      this.pageNo = 0
      this.hasMore = false
      this.loadingMore = false
      this.commentsLoaded = false
    },
    normalizeComment(comment) {
      return {
        ...comment,
        comment_id: comment?.comment_id != null ? String(comment.comment_id) : '',
        showChildren: Boolean(comment.showChildren),
        children: Array.isArray(comment.children) ? comment.children : [],
        digg_count: Number(comment.digg_count || 0),
        user_digged: Number(comment.user_digged || 0),
        sub_comment_count: Number(comment.sub_comment_count || 0),
        user_buried: Boolean(comment.user_buried)
      }
    },
    ensureCommentsArray() {
      if (!Array.isArray(this.comments)) {
        this.comments = []
      }
      return this.comments
    },
    resetSelectStatus() {
      this.friends?.all?.forEach((item) => {
        item.select = false
      })
    },
    async handShowChildren(item) {
      this.loadChildrenItemCId = item.comment_id
      this.loadChildren = true
      await _sleep(500)
      this.loadChildren = false
      if (item.showChildren) {
        item.children = item.children.concat(sampleSize(this.ensureCommentsArray(), 10))
      } else {
        item.children = sampleSize(this.ensureCommentsArray(), 3)
        item.showChildren = true
      }
    },
    async send(item) {
      const content = this.comment.trim()
      if (!content) {
        return
      }

      const baseStore = useBaseStore()
      const avatar = baseStore.userinfo.avatar_small?.url_list?.[0] || ''
      const commentData = {
        ip_location: baseStore.userinfo.ip_location,
        aweme_id: this.videoId,
        content
      }

      try {
        const response = await videoComment({}, commentData)
        if (!response?.success) {
          console.error('Comment request failed:', response)
          return
        }

        const currentComments = this.ensureCommentsArray()
        const nextComment = this.normalizeComment({
          comment_id: response.data?.comment_id != null ? String(response.data.comment_id) : `temp-${Date.now()}`,
          avatar,
          nickname: baseStore.userinfo.nickname,
          content,
          ip_location: baseStore.userinfo.ip_location,
          digg_count: 0,
          user_digged: 0,
          user_buried: false,
          sub_comment_count: 0,
          create_time: Math.floor(Date.now() / 1000),
          children: []
        })

        this.comments = [nextComment, ...currentComments]
        this.totalComments = Number(this.totalComments || 0) + 1
        this.hasMore = this.comments.length < this.totalComments
        if (item?.statistics) {
          item.statistics.comment_count = Number(item.statistics.comment_count || 0) + 1
        }
        this.comment = ''
        this.isCall = false
        this.resetSelectStatus()
      } catch (error) {
        console.error('Error sending comment:', error)
      }
    },
    async getData(reset = true) {
      if (reset) {
        this.commentsLoaded = false
        this.pageNo = 0
      }

      const res: any = await videoComments({
        aweme_id: this.videoId,
        pageNo: this.pageNo,
        pageSize: this.pageSize
      })

      if (res.success) {
        const list = Array.isArray(res.data?.list)
          ? res.data.list.map((comment) => this.normalizeComment(comment))
          : []
        this.totalComments = Number(res.data?.total || 0)
        this.hasMore = Boolean(res.data?.hasMore)
        const currentComments = reset ? [] : this.ensureCommentsArray()
        this.comments = reset ? list : currentComments.concat(list)
        if (this.item?.statistics) {
          this.item.statistics.comment_count = this.totalComments
        }
      } else if (reset) {
        this.comments = []
        this.totalComments = 0
        this.hasMore = false
        if (this.item?.statistics) {
          this.item.statistics.comment_count = 0
        }
      }

      this.commentsLoaded = true
    },
    async loadMore() {
      if (this.loadingMore || !this.hasMore) {
        return
      }

      this.loadingMore = true
      this.pageNo += 1
      try {
        await this.getData(false)
      } finally {
        this.loadingMore = false
      }
    },
    handleDialogScroll(e) {
      const target = e?.target
      if (!target || this.loadingMore || !this.hasMore) {
        return
      }
      if (target.scrollHeight - target.clientHeight - target.scrollTop <= 80) {
        this.loadMore()
      }
    },
    cancel() {
      this.$emit('update:modelValue', false)
      this.$emit('close')
    },
    toggleCall(item) {
      item.select = !item.select
      const nickname = item.nickname
      if (this.comment.includes('@' + nickname)) {
        this.comment = this.comment.replace(`@${nickname} `, '')
      } else {
        this.comment += `@${nickname} `
      }
    },
    async loved(row) {
      const nextAction = !Boolean(row.user_digged)
      const prevDigged = Boolean(row.user_digged)
      const prevDiggCount = Number(row.digg_count || 0)

      row.user_digged = nextAction ? 1 : 0
      row.digg_count = Math.max(prevDiggCount + (nextAction ? 1 : -1), 0)

      const res: any = await commentDigg(
        {},
        {
          comment_id: String(row.comment_id),
          action: nextAction
        }
      )

      if (!res?.success) {
        row.user_digged = prevDigged ? 1 : 0
        row.digg_count = prevDiggCount
      }
    },
    async deleteCommentRow(row) {
      const res: any = await deleteComment(
        {},
        {
          comment_id: String(row.comment_id)
        }
      )

      if (!res?.success) {
        return
      }

      const currentComments = this.ensureCommentsArray()
      this.comments = currentComments.filter((item) => item.comment_id !== row.comment_id)
      this.totalComments = Math.max(Number(this.totalComments || 0) - 1, 0)
      this.hasMore = this.comments.length < this.totalComments
      if (this.item?.statistics) {
        this.item.statistics.comment_count = Math.max(
          Number(this.item.statistics.comment_count || 0) - 1,
          0
        )
      }
    },
    showOptions(row) {
      _showSelectDialog(this.options, (e) => {
        if (e.id === 1) {
          this.selectRow = row
          this.showPrivateChat = true
        }
        if (e.id === 4) {
          _showSimpleConfirmDialog('确认删除这条评论吗？', () => {
            this.deleteCommentRow(row)
          })
        }
      })
    }
  }
}
</script>

<style lang="less" scoped>
@import '../assets/less/index';

.title {
  box-sizing: border-box;
  width: 100%;
  height: 40rem;
  padding: 0 15rem;
  background: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 10rem 10rem 0 0;

  .num {
    width: 100%;
    position: absolute;
    font-size: 12rem;
    font-weight: bold;
    text-align: center;
  }

  .right {
    display: flex;
    gap: 12rem;
    position: relative;
    z-index: 9;

    svg {
      color: #000;
      background: rgb(242, 242, 242);
      padding: 4rem;
      font-size: 16rem;
      border-radius: 50%;
    }
  }
}

.gray {
  color: var(--second-text-color);
}

.comment {
  color: #000;
  width: 100%;
  background: #fff;
  z-index: 5;

  .empty-state {
    position: absolute;
    left: 50%;
    top: 40%;
    transform: translate(-50%, -50%);
    color: var(--second-text-color);
    font-size: 14rem;
  }

  .wrapper {
    width: 100%;
    position: relative;
    padding-bottom: 60rem;
  }

  .load-more {
    padding: 12rem 0 20rem;
    color: var(--second-text-color);
    font-size: 13rem;
    text-align: center;
  }

  .items {
    width: 100%;

    .item {
      width: 100%;
      margin-bottom: 15rem;

      .main {
        width: 100%;
        padding: 5rem 0;
        display: flex;

        &:active {
          background: #53535321;
        }

        .head-image {
          margin-left: 15rem;
          margin-right: 10rem;
          width: 37rem;
          height: 37rem;
          border-radius: 50%;
        }
      }

      .replies {
        padding-left: 55rem;

        .reply {
          padding: 5rem 0 5rem 5rem;
          display: flex;

          &:active {
            background: #53535321;
          }

          .head-image {
            margin-right: 10rem;
            width: 20rem;
            height: 20rem;
            border-radius: 50%;
          }
        }

        .more {
          font-size: 13rem;
          margin: 5rem;
          display: flex;
          align-items: center;
          color: gray;

          .gang {
            background: #d5d5d5;
            width: 20rem;
            margin-right: 10rem;
            height: 1px;
          }

          span {
            margin-right: 5rem;
          }

          svg {
            font-size: 10rem;
          }
        }
      }

      .content {
        width: 100%;
        display: flex;
        font-size: 14rem;

        .comment-container {
          flex: 1;
          margin-right: 20rem;

          .name {
            color: var(--second-text-color);
            margin-bottom: 5rem;
            display: flex;
            align-items: center;

            .reply-user {
              margin-left: 5rem;
              width: 0;
              height: 0;
              border: 5rem solid transparent;
              border-left: 6rem solid gray;
            }
          }

          .detail {
            margin-bottom: 5rem;
          }

          .time-wrapper {
            display: flex;
            align-items: center;
            justify-content: space-between;
            font-size: 13rem;

            .left {
              display: flex;

              .time {
                color: #c4c3c3;
                margin-right: 10rem;
              }

              .reply-text {
                color: var(--second-text-color);
              }
            }

            .love {
              color: gray;
              display: flex;
              align-items: center;

              &.loved {
                color: rgb(231, 58, 87);
              }

              .love-image {
                font-size: 17rem;
                margin-right: 4rem;
              }

              span {
                word-break: keep-all;
              }
            }
          }
        }
      }
    }
  }

  @normal-bg-color: rgb(35, 38, 47);
  @chat-bg-color: rgb(105, 143, 244);

  .input-toolbar {
    border-radius: 10rem 10rem 0 0;
    background: white;
    position: fixed;
    width: 100%;
    bottom: 0;
    z-index: 3;

    @space-width: 18rem;
    @icon-width: 48rem;

    .call-friend {
      padding-top: 30rem;
      overflow-x: scroll;
      display: flex;
      padding-right: @space-width;

      .friend {
        width: @icon-width;
        position: relative;
        margin-left: @space-width;
        margin-bottom: @space-width;
        font-size: 10rem;
        display: flex;
        flex-direction: column;
        align-items: center;

        .avatar {
          width: @icon-width;
          height: @icon-width;
          border-radius: 50%;
        }

        span {
          margin-top: 5rem;
          text-align: center;
          width: @icon-width;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .checked {
          position: absolute;
          top: @icon-width - 1.5;
          right: -2px;
          width: 20rem;
          height: 20rem;
          border-radius: 50%;
        }
      }
    }

    .toolbar {
      @icon-width: 25rem;
      display: flex;
      align-items: center;
      padding: 10rem 15rem;
      border-top: 1px solid #e2e1e1;

      .input-wrapper {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: space-between;
        box-sizing: border-box;
        padding: 5rem 10rem;
        background: #eee;
        border-radius: 20rem;

        .right {
          display: flex;
          align-items: center;
        }

        .auto-input {
          width: calc(100% - 160rem);
        }
      }

      img {
        width: @icon-width;
        height: @icon-width;
        border-radius: 50%;
        margin-left: 15rem;
      }
    }
  }
}

.comment-enter-active,
.comment-leave-active {
  transition: all 0.15s ease;
}

.comment-enter-from,
.comment-leave-to {
  transform: translateY(60vh);
}
</style>
