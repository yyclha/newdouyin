<template>
  <div class="PasswordRegister">
    <BaseHeader mode="light" backMode="dark" backImg="back">
      <template v-slot:right>
        <span class="f14" @click="$router.push('/register/help')">帮助与设置</span>
      </template>
    </BaseHeader>
    <div class="content">
      <div class="desc">
        <div class="title">手机号注册</div>
      </div>

      <LoginInput autofocus type="phone" v-model="phone" placeholder="请输入手机号" />
      <LoginInput class="mt1r" type="password" v-model="password" placeholder="请输入密码" />
      <LoginInput
        class="mt1r"
        type="password"
        v-model="confirmPassword"
        placeholder="请再次输入密码"
      />

      <div class="protocol" :class="showAnim ? 'anim-bounce' : ''">
        <Tooltip style="top: -150%; left: -10rem" v-model="showTooltip" />
        <div class="left">
          <Check v-model="isAgree" />
        </div>
        <div class="right">
          已阅读并同意
          <span
            class="link"
            @click="$router.push('/service-protocol', { type: '“抖音”用户服务协议' })"
            >用户协议</span
          >
          和
          <span class="link" @click="$router.push('/service-protocol', { type: '“抖音”隐私政策' })"
            >隐私政策</span
          >
        </div>
      </div>

      <div class="notice" v-if="notice">
        {{ notice }}
      </div>

      <dy-button
        type="primary"
        :loading="loading"
        :active="false"
        :disabled="disabled"
        @click="register"
      >
        {{ loading ? '注册中' : '注册' }}
      </dy-button>
    </div>
  </div>
</template>
<script>
import Check from '../../components/Check'
import LoginInput from './components/LoginInput'
import Tooltip from './components/Tooltip'
import Base from './Base'
import { register } from '@/api/base'
import { _notice } from '@/utils/index'

export default {
  name: 'PasswordRegister',
  extends: Base,
  components: {
    Check,
    Tooltip,
    LoginInput
  },
  data() {
    return {
      phone: '',
      password: '',
      confirmPassword: '',
      notice: '',
      isAgree: false,
      loading: false
    }
  },
  computed: {
    disabled() {
      return !(this.phone && this.password && this.confirmPassword && this.isAgree)
    }
  },
  methods: {
    async register() {
      if (this.password !== this.confirmPassword) {
        _notice('两次输入的密码不一致')
        return
      }
      if (!this.isAgree) {
        _notice('请阅读并同意用户协议和隐私政策')
        return
      }

      this.loading = true
      try {
        const res = await register(
          {},
          {
            phone: this.phone,
            password: this.password
          }
        )
        console.log('注册结果:', res.data)
        if (res.success) {
          // 注册成功后的处理逻辑，如跳转页面
          setTimeout(() => {
            _notice('注册成功!')
          }, 200)
          setTimeout(() => {
            this.$router.push('/login/password') // 跳转到密码登录页面
          }, 800) // 停留 1 秒
        } else {
          _notice(res.msg)
        }
      } catch (error) {
        this.notice = '网络错误，请稍后再试'
      } finally {
        this.loading = false
      }
    }
  }
}
</script>
<style scoped lang="less">
@import '../../assets/less/index';
@import 'Base.less';

.PasswordRegister {
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

    .desc {
      margin-bottom: 60rem;
      margin-top: 120rem;
      display: flex;
      align-items: center;
      flex-direction: column;

      .title {
        margin-bottom: 20rem;
        font-size: 20rem;
      }
    }

    .protocol {
      position: relative;
      color: gray;
      margin-top: 20rem;
      font-size: 12rem;
      display: flex;

      .left {
        padding-top: 1rem;
        margin-right: 5rem;
      }
    }
  }
}
</style>
