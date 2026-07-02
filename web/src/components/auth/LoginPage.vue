<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Button, Form, FormItem, Input, Message, VerificationCode } from '@arco-design/web-vue'
import { computed, reactive, shallowRef } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

useHead({ title: '登录 — XLNetAccount' })

import { ApiError } from '@/api/http'
import { startPasskeyLogin } from '@/api/passkeys'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'
import { buildNextQuery, resolveNextTarget } from '@/utils/authNext'
import { describePasskeyError, getPasskeyCredential, getPasskeySupportMessage } from '@/utils/webauthn'

const route = useRoute()
const router = useRouter()
const platformStore = usePlatformStore()
const sessionStore = useSessionStore()
void platformStore.ensureLoaded().catch(() => {})

const showRegister = computed(() => platformStore.allowRegistration)
const registerLink = computed(() => ({
  name: 'register',
  query: buildNextQuery(route.query.next),
}))

const isSubmitting = shallowRef(false)
const isPasskeySubmitting = shallowRef(false)
const submitError = shallowRef('')
const passkeySupportMessage = getPasskeySupportMessage()
const showTotpStep = shallowRef(false)
const totpSessionId = shallowRef('')
const totpCode = shallowRef('')
const isTotpSubmitting = shallowRef(false)
const formState = reactive({
  username: '',
  password: '',
})

async function handleSubmit() {
  isSubmitting.value = true
  submitError.value = ''

  try {
    const result = await sessionStore.signIn({
      username: formState.username,
      password: formState.password,
    })

    if (result.type === 'totp') {
      showTotpStep.value = true
      totpSessionId.value = result.sessionId
      Message.info('请输入两步验证码')
      return
    }

    Message.success('登录成功')
    await router.replace(resolveNextTarget(route.query.next))
  }
  catch (error) {
    const fallbackMessage = error instanceof ApiError ? error.message : '登录失败，请稍后重试'
    submitError.value = fallbackMessage
    Message.error(fallbackMessage)
  }
  finally {
    isSubmitting.value = false
  }
}

async function handleTotpVerify() {
  isTotpSubmitting.value = true
  submitError.value = ''

  try {
    await sessionStore.signInWithTOTP({
      totp_session_id: totpSessionId.value,
      code: totpCode.value,
    })

    Message.success('登录成功')
    await router.replace(resolveNextTarget(route.query.next))
  }
  catch (error) {
    const fallbackMessage = error instanceof ApiError ? error.message : '验证失败'
    submitError.value = fallbackMessage
    Message.error(fallbackMessage)
  }
  finally {
    isTotpSubmitting.value = false
  }
}

function handleBackToLogin() {
  showTotpStep.value = false
  totpSessionId.value = ''
  totpCode.value = ''
  submitError.value = ''
}

async function handlePasskeySignIn() {
  isPasskeySubmitting.value = true
  submitError.value = ''

  try {
    const start = await startPasskeyLogin()
    const credential = await getPasskeyCredential(start.options)

    await sessionStore.signInWithPasskey({
      session_id: start.session_id,
      credential,
    })

    Message.success('登录成功')
    await router.replace(resolveNextTarget(route.query.next))
  }
  catch (error) {
    const fallbackMessage = error instanceof ApiError ? error.message : describePasskeyError(error)
    submitError.value = fallbackMessage
    Message.error(fallbackMessage)
  }
  finally {
    isPasskeySubmitting.value = false
  }
}
</script>

<template>
  <div class="auth-panel-view">
    <div class="auth-panel-header">
      <p class="auth-panel-kicker">{{ showTotpStep ? '两步验证' : '欢迎回来' }}</p>
      <h2 class="auth-panel-title">{{ showTotpStep ? '输入验证码' : `登录到 ${platformStore.displayName}` }}</h2>
      <p class="auth-panel-subtitle">{{ showTotpStep ? '请在下方输入验证器 App 中显示的 6 位动态码。' : '继续管理账号、客户端与令牌。' }}</p>
    </div>

    <template v-if="!showTotpStep">
      <Form :model="formState" layout="vertical" class="auth-form" @submit="handleSubmit">
        <FormItem label="邮箱或用户名">
          <Input v-model="formState.username" clearable placeholder="邮箱或用户名" size="large" @input="submitError = ''" />
        </FormItem>

        <FormItem label="密码">
          <Input
            v-model="formState.password"
            type="password"
                        placeholder="密码"
            size="large"
            @input="submitError = ''"
          />
        </FormItem>

        <Alert v-if="passkeySupportMessage" type="warning" :show-icon="false">
          {{ passkeySupportMessage }}
        </Alert>

        <div class="auth-action-stack">
          <Button type="primary" size="large" block :loading="isSubmitting" :disabled="isPasskeySubmitting" html-type="submit">
            登录
          </Button>
          <Button
            type="secondary"
            size="large"
            block
            html-type="button"
            data-testid="passkey-login-button"
            :loading="isPasskeySubmitting"
            :disabled="Boolean(passkeySupportMessage) || isSubmitting"
            @click="handlePasskeySignIn"
          >
            使用通行密钥登录
          </Button>
        </div>
      </Form>

      <div v-if="showRegister" class="auth-link-row">
        <RouterLink :to="registerLink">没有账号？立即注册</RouterLink>
      </div>
    </template>

    <template v-else>
      <Form :model="{code: totpCode}" layout="vertical" class="auth-form" @submit="handleTotpVerify">
        <FormItem label="验证码">
          <div class="totp-code-wrapper">
            <VerificationCode
              v-model="totpCode"
              :length="6"
              size="large"
              @change="submitError = ''"
              @finish="handleTotpVerify"
            />
          </div>
        </FormItem>

        <div class="auth-action-stack">
          <Button type="primary" size="large" block :loading="isTotpSubmitting" html-type="submit">
            验证
          </Button>
          <Button type="secondary" size="large" block html-type="button" @click="handleBackToLogin">
            返回
          </Button>
        </div>
      </Form>
    </template>
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

.auth-link-row {
  margin-top: 16px;
}

.auth-action-stack {
  display: grid;
  gap: 12px;
}

.auth-link-row a {
  color: var(--color-accent-blue);
}

.totp-code-wrapper {
  display: flex;
  justify-content: center;
  padding: 4px 0;
}
</style>
