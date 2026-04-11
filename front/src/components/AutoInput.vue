<template>
  <div
    ref="input"
    :placeholder="placeholder"
    class="auto-input"
    contenteditable
    @input="changeText"
    @compositionstart="handleCompositionStart"
    @compositionend="handleCompositionEnd"
  >
  </div>
</template>

<script lang="ts">
export default {
  name: 'AutoInput',
  props: {
    modelValue: String,
    placeholder: {
      type: String,
      default: '留下你的精彩评论吧'
    }
  },
  data: function () {
    return {
      isComposition: false // 标记是否正在进行中文输入
    }
  },
  watch: {
    // 监听外部传入的值变化
    modelValue(newVal) {
      // 只有当当前显示的内容和新值不一样时才更新，避免光标跳动
      if (this.$refs.input.innerText !== newVal) {
        this.$refs.input.innerText = newVal
      }
    }
  },
  mounted() {
    // 初始化时显示内容
    this.$refs.input.innerText = this.modelValue
  },
  methods: {
    handleCompositionStart() {
      this.isComposition = true;
    },
    handleCompositionEnd(e) {
      this.isComposition = false;
      // 输入完成，手动触发一次更新
      this.changeText(e);
    },
    changeText(e) {
      // 如果正在中文输入中，不更新数据
      if (this.isComposition) return;
      
      this.$emit('update:modelValue', this.$el.innerText)
    }
  }
}
</script>

<style scoped lang="less">
.auto-input {
  font-size: 14rem;
  width: 100%;
  max-height: 70rem;
  overflow-y: scroll;
  padding: 0 5rem;
  outline: none;
}

.auto-input::-webkit-scrollbar {
  width: 0 !important;
}

.auto-input:empty::before {
  content: attr(placeholder);
  color: #999999;
}

.auto-input:focus::before {
  content: none;
}
</style>