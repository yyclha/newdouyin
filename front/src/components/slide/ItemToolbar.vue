<script setup lang="ts">
import BaseMusic from '../BaseMusic.vue'
import { _checkImgUrl, _formatNumber, cloneDeep } from '@/utils'
import bus, { EVENT_KEY } from '@/utils/bus'
import { Icon } from '@iconify/vue'
import { useClick } from '@/utils/hooks/useClick'
import { computed, inject } from 'vue'
import { videoCollect, videoDigg } from '@/api/videos'
import { userAttention } from '@/api/user'
import { useBaseStore } from '@/store/pinia'

const props = defineProps({
  isMy: {
    type: Boolean,
    default: () => {
      return false
    }
  },
  item: {
    type: Object,
    default: () => {
      return {}
    }
  }
})

const position = inject<any>('position')
const baseStore = useBaseStore()
const isAttention = computed(() => {
  const uid = String(props.item?.author?.uid || '')
  if (!uid) return !!props.item?.isAttention || !!props.item?.author?.follow_status

  return (
    !!props.item?.isAttention ||
    !!props.item?.author?.follow_status ||
    baseStore.AwemeStatus.Attentions.some((id) => String(id) === uid)
  )
})

const emit = defineEmits(['update:item', 'goUserInfo', 'showComments', 'showShare', 'goMusic'])

function _updateItem(item) {
  emit('update:item', item)
  bus.emit(EVENT_KEY.UPDATE_ITEM, { position: position.value, item })
}

function syncAttentionState(nextItem, followed) {
  nextItem.isAttention = followed

  if (nextItem.author) {
    nextItem.author.follow_status = followed ? 1 : 0
  }

  const uid = String(nextItem.author?.uid || '')
  if (!uid) return

  const attentions = Array.isArray(baseStore.AwemeStatus.Attentions)
    ? [...baseStore.AwemeStatus.Attentions]
    : []

  if (followed) {
    if (!attentions.includes(uid)) {
      attentions.push(uid)
    }
    baseStore.AwemeStatus.Attentions = attentions
    return
  }

  baseStore.AwemeStatus.Attentions = attentions.filter((id) => String(id) !== uid)
}

function syncLovedState(awemeId, loved) {
  const likes = Array.isArray(baseStore.AwemeStatus.Likes) ? [...baseStore.AwemeStatus.Likes] : []
  const targetId = String(awemeId)

  if (loved) {
    if (!likes.includes(targetId)) {
      likes.push(targetId)
    }
    baseStore.AwemeStatus.Likes = likes
    return
  }

  baseStore.AwemeStatus.Likes = likes.filter((id) => String(id) !== targetId)
}

async function loved() {
  try {
    console.log('aweme_id:', props)
    const res = await videoDigg(
      {},
      {
        aweme_id: props.item.aweme_id,
        action: !props.item.isLoved
      }
    )

    if (res.success) {
      const nextItem = cloneDeep(props.item)
      const nextLoved = !nextItem.isLoved
      const diggDelta = nextLoved ? 1 : -1

      nextItem.isLoved = nextLoved
      nextItem.statistics.digg_count = Math.max(
        (nextItem.statistics?.digg_count || 0) + diggDelta,
        0
      )

      if (nextItem.author) {
        nextItem.author.total_favorited = Math.max(
          (nextItem.author.total_favorited || 0) + diggDelta,
          0
        )

        if (String(nextItem.author.uid) === String(baseStore.userinfo.uid)) {
          baseStore.userinfo = {
            ...baseStore.userinfo,
            total_favorited: nextItem.author.total_favorited
          }
        }
      }

      syncLovedState(nextItem.aweme_id, nextLoved)
      _updateItem(nextItem)
    } else {
      console.error('点赞失败:', res.data.msg)
    }
  } catch (error) {
    console.error('点赞请求异常:', error)
  }
}

async function collecte() {
  try {
    console.log('aweme_id:', props)
    const res = await videoCollect(
      {},
      {
        aweme_id: props.item.aweme_id,
        action: !props.item.isCollect
      }
    )

    if (res.success) {
      const nextItem = cloneDeep(props.item)
      const nextCollect = !nextItem.isCollect

      nextItem.isCollect = nextCollect
      nextItem.statistics.collect_count = Math.max(
        (nextItem.statistics?.collect_count || 0) + (nextCollect ? 1 : -1),
        0
      )

      _updateItem(nextItem)
    } else {
      console.error('收藏失败:', res.data.msg)
    }
  } catch (error) {
    console.error('收藏请求异常:', error)
  }
}

async function attention(e) {
  console.log('关注', props.item)
  e.currentTarget.classList.add('attention')

  try {
    const nextAttention = !isAttention.value
    const res = await userAttention(
      {},
      {
        following_id: String(props.item.author.uid),
        action: nextAttention
      }
    )

    if (res.success) {
      const nextItem = cloneDeep(props.item)
      syncAttentionState(nextItem, nextAttention)
      _updateItem(nextItem)
    } else {
      console.error('关注失败:', res.data.msg)
    }
  } catch (error) {
    console.error('关注请求异常:', error)
  }
}

function showComments() {
  bus.emit(EVENT_KEY.OPEN_COMMENTS, props.item.aweme_id)
}

const vClick = useClick()
</script>

<template>
  <div class="toolbar mb1r">
    <div class="avatar-ctn mb2r">
      <img
        class="avatar"
        :src="_checkImgUrl(item.author.avatar_small)"
        alt=""
        v-click="() => bus.emit(EVENT_KEY.GO_USERINFO)"
      />
      <transition name="fade">
        <div
          v-if="!isAttention"
          v-click="attention"
          class="options"
        >
          <img class="no" src="../../assets/img/icon/add-light.png" alt="" />
          <img class="yes" src="../../assets/img/icon/ok-red.png" alt="" />
        </div>
      </transition>
    </div>
    <div class="love mb2r" v-click="loved">
      <div>
        <img src="../../assets/img/icon/love.svg" class="love-image" v-if="!item.isLoved" />
        <img src="../../assets/img/icon/loved.svg" class="love-image" v-if="item.isLoved" />
      </div>
      <span>{{ _formatNumber(item.statistics.digg_count) }}</span>
    </div>
    <div class="message mb2r" v-click="showComments">
      <Icon icon="mage:message-dots-round-fill" class="icon" style="color: white" />
      <span>{{ _formatNumber(item.statistics.comment_count) }}</span>
    </div>
    <!--TODO     -->
    <div class="message mb2r" v-click="collecte">
      <Icon
        v-if="item.isCollect"
        icon="ic:round-star"
        class="icon"
        style="color: rgb(252, 179, 3)"
      />
      <Icon v-else icon="ic:round-star" class="icon" style="color: white" />
      <span>{{ _formatNumber(item.statistics.collect_count) }}</span>
    </div>
    <div v-if="!props.isMy" class="share mb2r" v-click="() => bus.emit(EVENT_KEY.SHOW_SHARE)">
      <img src="../../assets/img/icon/share-white-full.png" alt="" class="share-image" />
      <span>{{ _formatNumber(item.statistics.share_count) }}</span>
    </div>
    <div v-else class="share mb2r" v-click="() => bus.emit(EVENT_KEY.SHOW_SHARE)">
      <img src="../../assets/img/icon/menu-white.png" alt="" class="share-image" />
    </div>
    <!--    <BaseMusic-->
    <!--        :cover="item.music.cover"-->
    <!--        v-click="$router.push('/home/music')"-->
    <!--    /> -->
    <BaseMusic />
  </div>
</template>

<style scoped lang="less">
.toolbar {
  //width: 40px;
  position: absolute;
  bottom: 0;
  right: 10rem;
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;

  .avatar-ctn {
    position: relative;

    @w: 45rem;

    .avatar {
      width: @w;
      height: @w;
      border: 3rem solid white;
      border-radius: 50%;
    }

    .options {
      position: absolute;
      border-radius: 50%;
      margin: auto;
      left: 0;
      right: 0;
      bottom: -5px;
      background: red;
      //background: black;
      width: 18rem;
      height: 18rem;
      display: flex;
      justify-content: center;
      align-items: center;
      transition: all 1s;

      img {
        position: absolute;
        width: 14rem;
        height: 14rem;
        transition: all 1s;
      }

      .yes {
        opacity: 0;
        transform: rotate(-180deg);
      }

      &.attention {
        background: white;

        .no {
          opacity: 0;
          transform: rotate(180deg);
        }

        .yes {
          opacity: 1;
          transform: rotate(0deg);
        }
      }
    }
  }

  .love,
  .message,
  .share {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;

    @width: 35rem;

    img {
      width: @width;
      height: @width;
    }

    span {
      font-size: 12rem;
    }
  }

  .icon {
    font-size: 40rem;
  }

  .loved {
    background: red;
  }
}
</style>
