<!-- 登录页面 -->
<template>
  <div class="flex w-full h-screen">
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h3 class="title">{{ $t('login.title') }}</h3>
          <p class="sub-title">{{ $t('login.subTitle') }}</p>
          <ElForm
            ref="formRef"
            :model="formData"
            :rules="rules"
            :key="formKey"
            @keyup.enter="handleSubmit"
            style="margin-top: 25px"
          >
            <ElFormItem prop="username">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.username')"
                v-model.trim="formData.username"
              />
            </ElFormItem>
            <ElFormItem prop="password">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.password')"
                v-model.trim="formData.password"
                type="password"
                autocomplete="off"
                show-password
              />
            </ElFormItem>

            <div class="flex-cb mt-2 text-sm">
              <ElCheckbox v-model="formData.rememberPassword">{{
                $t('login.rememberPwd')
              }}</ElCheckbox>
              <RouterLink class="text-theme" :to="{ name: 'ForgetPassword' }">{{
                $t('login.forgetPwd')
              }}</RouterLink>
            </div>

            <div style="margin-top: 30px">
              <ElButton
                class="w-full custom-height"
                type="primary"
                @click="handleSubmit"
                :loading="loading"
                v-ripple
              >
                {{ $t('login.btnText') }}
              </ElButton>
            </div>

            <div class="mt-5 text-sm text-gray-600">
              <span>{{ $t('login.noAccount') }}</span>
              <RouterLink class="text-theme" :to="{ name: 'Register' }">{{
                $t('login.register')
              }}</RouterLink>
            </div>
          </ElForm>
        </div>
      </div>
    </div>

    <!-- 弹出滑动验证码 -->
    <ElDialog
      v-model="captchaVisible"
      title="安全验证"
      width="340px"
      destroy-on-close
      align-center
      :close-on-click-modal="false"
      class="captcha-dialog"
    >
      <div class="flex justify-center items-center py-2">
        <Slide
          v-if="captchaData.image"
          :config="captchaConfig"
          :data="captchaData"
          :events="captchaEvents"
          ref="slideRef"
        />
        <div v-else class="text-gray-400 py-10 flex flex-col items-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mb-2"></div>
          <span>加载验证码中...</span>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import AppConfig from '@/config'
  import { useUserStore } from '@/store/modules/user'
  import { useI18n } from 'vue-i18n'
  import { HttpError } from '@/utils/http/error'
  import { fetchLogin, fetchSlideCaptcha } from '@/api/auth'
  import { ElNotification, type FormInstance, type FormRules } from 'element-plus'
  import { useSettingStore } from '@/store/modules/setting'
  import { Slide } from 'go-captcha-vue'
  import 'go-captcha-vue/dist/style.css'

  defineOptions({ name: 'Login' })

  const settingStore = useSettingStore()
  const { isDark } = storeToRefs(settingStore)
  const { t, locale } = useI18n()
  const formKey = ref(0)

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const userStore = useUserStore()
  const router = useRouter()
  const route = useRoute()

  const systemName = AppConfig.systemInfo.name
  const formRef = ref<FormInstance>()

  const formData = reactive({
    username: '',
    password: '',
    rememberPassword: true
  })

  const rules = computed<FormRules>(() => ({
    username: [{ required: true, message: t('login.placeholder.username'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.placeholder.password'), trigger: 'blur' }]
  }))

  const loading = ref(false)

  // 滑动验证码状态
  const captchaVisible = ref(false)
  const captchaId = ref('')
  const slideRef = ref()
  const captchaData = reactive({
    image: '',
    thumb: '',
    thumbWidth: 80,
    thumbHeight: 80,
    thumbX: 0,
    thumbY: 0
  })

  const captchaConfig = reactive({
    width: 300,
    height: 220
  })

  // 加载滑动验证码挑战
  const loadCaptcha = async () => {
    try {
      captchaData.image = ''
      captchaData.thumb = ''
      const res = await fetchSlideCaptcha()
      captchaId.value = res.id
      captchaData.image = res.image
      captchaData.thumb = res.thumb
      captchaData.thumbWidth = res.w || 80
      captchaData.thumbHeight = res.h || 80
      captchaData.thumbY = res.y || 0
      captchaData.thumbX = 0
    } catch (error) {
      console.error('Failed to load captcha:', error)
      ElNotification({
        title: '错误',
        message: '加载验证码失败，请重试',
        type: 'error'
      })
    }
  }

  // 滑动验证码事件回调
  const captchaEvents = {
    confirm: async (point: { x: number; y: number }, reset: () => void) => {
      try {
        loading.value = true
        const { username, password } = formData

        const { token, refreshToken } = await fetchLogin({
          userName: username,
          password,
          captchaId: captchaId.value,
          x: Math.round(point.x),
          y: Math.round(point.y)
        })

        if (!token) {
          throw new Error('Login failed - no token received')
        }

        captchaVisible.value = false
        userStore.setToken(token, refreshToken)
        userStore.setLoginStatus(true)
        showLoginSuccessNotice()

        const redirect = route.query.redirect as string
        router.push(redirect || '/')
      } catch (error: any) {
        ElNotification({
          title: '登录失败',
          message: error.message || '滑动验证失败，请重试',
          type: 'error'
        })
        reset()
        loadCaptcha()
      } finally {
        loading.value = false
      }
    },
    refresh: () => {
      loadCaptcha()
    }
  }

  // 登录提交
  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      const valid = await formRef.value.validate()
      if (!valid) return

      // 触发滑动验证码弹窗
      captchaVisible.value = true
      await loadCaptcha()
    } catch (error) {
      console.error('[Login] validation error:', error)
    }
  }

  // 登录成功提示
  const showLoginSuccessNotice = () => {
    setTimeout(() => {
      ElNotification({
        title: t('login.success.title'),
        type: 'success',
        duration: 2500,
        zIndex: 10000,
        message: `${t('login.success.message')}, ${systemName}!`
      })
    }, 1000)
  }
</script>

<style scoped>
  @import './style.css';
</style>

<style lang="scss" scoped>
  :deep(.el-select__wrapper) {
    height: 40px !important;
  }
</style>
