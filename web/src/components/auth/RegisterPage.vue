<script setup lang="ts">
import { useHead } from '@unhead/vue'

import { Button, Form, FormItem, Input, Message } from '@arco-design/web-vue'
import { computed, onBeforeUnmount, reactive, shallowRef } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

useHead({ title: '注册 — XLNetAccount' })

import { register, sendRegisterCode } from '@/api/auth'
import { solveCaptcha } from '@/utils/captcha'
import CaptchaCard from '@/components/ui/CaptchaCard.vue'
import { ApiError } from '@/api/http'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'
import { buildNextQuery, resolveNextTarget } from '@/utils/authNext'

const route = useRoute()
const router = useRouter()
const platformStore = usePlatformStore()
const sessionStore = useSessionStore()

const isSubmitting = shallowRef(false)
const isSendingCode = shallowRef(false)
const submitError = shallowRef('')
const codeSent = shallowRef(false)
const captchaToken = shallowRef('')
const resendEmail = shallowRef('')
const resendRemaining = shallowRef(0)

let resendTimer: ReturnType<typeof setInterval> | null = null

const formState = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  code: '',
})

void platformStore.ensureLoaded().catch(() => {})

const capEnabled = computed(() => Boolean(platformStore.capWidgetEndpoint))
const loginLink = computed(() => ({
  name: 'login',
  query: buildNextQuery(route.query.next),
}))
const passwordHint = '密码需为 8-20 位，且包含大写字母、小写字母和数字'
const passwordChecks = computed(() => {
  const password = formState.password
  return [
    { label: '8-20 位长度', passed: password.length >= 8 && password.length <= 20 },
    { label: '包含大写字母', passed: /[A-Z]/.test(password) },
    { label: '包含小写字母', passed: /[a-z]/.test(password) },
    { label: '包含数字', passed: /\d/.test(password) },
  ]
})
const passwordValid = computed(() => passwordChecks.value.every(item => item.passed))
const passwordMismatch = computed(() => formState.confirmPassword !== '' && formState.password !== formState.confirmPassword)
const normalizedEmail = computed(() => formState.email.trim().toLowerCase())
const sendCodeDisabled = computed(() => {
  if (isSendingCode.value) return true
  return resendRemaining.value > 0 && resendEmail.value === normalizedEmail.value
})
const sendCodeText = computed(() => {
  if (sendCodeDisabled.value && resendRemaining.value > 0 && resendEmail.value === normalizedEmail.value) {
    return `${resendRemaining.value}s`
  }
  return codeSent.value ? '重新发送' : '发送验证码'
})

onBeforeUnmount(() => {
  clearResendTimer()
})

async function handleSendCode() {
  if (!normalizedEmail.value) {
    Message.error('请先输入邮箱')
    return
  }
  if (resendRemaining.value > 0 && resendEmail.value === normalizedEmail.value) {
    Message.error(`请在 ${resendRemaining.value} 秒后重试`)
    return
  }

  // 自动完成人机验证（无感，后台 PoW 求解）
  if (capEnabled.value && !captchaToken.value) {
    isSendingCode.value = true
    Message.loading('正在进行人机验证，请稍候...')
    try {
      const token = await solveCaptcha(platformStore.capWidgetEndpoint)
      captchaToken.value = token
    } catch (error) {
      Message.clear()
      isSendingCode.value = false
      Message.error(error instanceof Error ? error.message : '人机验证失败，请重试')
      return
    }
    Message.clear()
  }

  isSendingCode.value = true
  submitError.value = ''

  try {
    await sendRegisterCode({
      email: normalizedEmail.value,
      captcha_token: captchaToken.value,
    })
    codeSent.value = true
    startResendCountdown(normalizedEmail.value, 60)
    Message.success('验证码已发送，请查收邮箱')
  }
  catch (error) {
    const msg = error instanceof ApiError ? error.message : '发送验证码失败'
    submitError.value = msg
    const retryAfter = extractRetryAfter(msg)
    if (retryAfter > 0) {
      startResendCountdown(normalizedEmail.value, retryAfter)
    }
    Message.error(msg)
  }
  finally {
    isSendingCode.value = false
  }
}

async function handleSubmit() {
  if (formState.password !== formState.confirmPassword) {
    Message.error('两次输入的密码不一致')
    return
  }
  if (!isPasswordValid(formState.password)) {
    Message.error(passwordHint)
    return
  }

  isSubmitting.value = true
  submitError.value = ''

  try {
    await register({
      username: formState.username,
      email: formState.email,
      password: formState.password,
      code: formState.code,
    })
    await sessionStore.signIn({
      username: formState.username,
      password: formState.password,
    })
    Message.success('注册成功')
    await router.replace(resolveNextTarget(route.query.next))
  }
  catch (error) {
    const msg = error instanceof ApiError ? error.message : '注册失败，请稍后重试'
    submitError.value = msg
    Message.error(msg)
  }
  finally {
    isSubmitting.value = false
  }
}

function isPasswordValid(password: string) {
  return /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,20}$/.test(password)
}

function startResendCountdown(email: string, seconds: number) {
  clearResendTimer()
  resendEmail.value = email
  resendRemaining.value = seconds
  resendTimer = setInterval(() => {
    if (resendRemaining.value <= 1) {
      resendRemaining.value = 0
      clearResendTimer()
      return
    }
    resendRemaining.value -= 1
  }, 1000)
}

function clearResendTimer() {
  if (resendTimer) {
    clearInterval(resendTimer)
    resendTimer = null
  }
}

function extractRetryAfter(message: string) {
  const matched = message.match(/(\d+)\s*秒/)
  if (!matched) return 0
  return Number(matched[1] || 0)
}
</script>

<template>
  <div class="auth-panel-view">
    <div class="auth-panel-header">
      <p class="auth-panel-kicker">创建账号</p>
      <h2 class="auth-panel-title">注册 {{ platformStore.displayName }}</h2>
      <p class="auth-panel-subtitle">使用邮箱验证码创建账号。</p>
    </div>

    <Form :model="formState" layout="vertical" class="auth-form" @submit="handleSubmit">
      <FormItem label="用户名">
        <Input v-model="formState.username" clearable placeholder="用户名" size="large" @input="submitError = ''" />
      </FormItem>

      <FormItem label="邮箱">
        <div class="email-row">
          <Input v-model="formState.email" clearable placeholder="邮箱" size="large" @input="submitError = ''" />
          <Button
            type="secondary"
            :loading="isSendingCode"
            :disabled="!formState.email.trim() || sendCodeDisabled"
            @click="handleSendCode"
          >
            {{ sendCodeText }}
          </Button>
        </div>
      </FormItem>

      <FormItem v-if="formState.email.trim()" label="验证码">
        <Input v-model="formState.code" placeholder="邮箱验证码" size="large" @input="submitError = ''" />
      </FormItem>

      <FormItem v-if="capEnabled" label="人机验证">
        <CaptchaCard
          :api-endpoint="platformStore.capWidgetEndpoint"
          @solve="(token) => captchaToken = token"
          @error="() => captchaToken = ''"
        />
      </FormItem>

      <FormItem label="密码">
        <Input v-model="formState.password" type="password" placeholder="密码" size="large" @input="submitError = ''" />
      </FormItem>

      <FormItem label="确认密码">
        <Input v-model="formState.confirmPassword" type="password" placeholder="再次输入密码" size="large" @input="submitError = ''" />
      </FormItem>

      <div class="password-rule-list">
        <div v-for="rule in passwordChecks" :key="rule.label" class="password-rule" :class="{ passed: rule.passed }">
          <span>{{ rule.passed ? '✓' : '○' }}</span>
          <span>{{ rule.label }}</span>
        </div>
        <div class="password-rule" :class="{ passed: !passwordMismatch && formState.confirmPassword !== '' }">
          <span>{{ !passwordMismatch && formState.confirmPassword !== '' ? '✓' : '○' }}</span>
          <span>两次输入密码一致</span>
        </div>
      </div>

      <Button type="primary" size="large" long :loading="isSubmitting" html-type="submit" :disabled="!passwordValid || passwordMismatch">
        注册
      </Button>
    </Form>

    <div class="auth-link-row">
      <RouterLink :to="loginLink">已有账号？去登录</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.auth-panel-view {
  color: var(--color-ink);
}

.auth-panel-header {
  margin-bottom: 24px;
}

.auth-panel-kicker {
  margin: 0 0 10px;
  color: var(--color-charcoal);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.auth-panel-title {
  margin: 0;
  font-size: clamp(26px, 2.8vw, 32px);
  font-weight: 650;
  letter-spacing: -0.03em;
}

.auth-panel-subtitle {
  margin: 10px 0 0;
  color: var(--color-charcoal);
  font-size: 14px;
  line-height: 1.65;
}

.auth-form {
  display: grid;
  gap: 4px;
}

.password-rule-list {
  display: grid;
  gap: 6px;
  margin-top: -2px;
  padding: 10px 12px;
  border: 1px solid var(--color-hairline);
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.02);
}

.password-rule {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-charcoal);
  font-size: 13px;
}

.password-rule.passed {
  color: var(--color-accent-green);
}

.email-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.email-row .arco-input {
  flex: 1;
  min-width: 0;
}

.auth-link-row {
  margin-top: 16px;
}

.auth-link-row a {
  color: var(--color-accent-blue);
}
</style>
