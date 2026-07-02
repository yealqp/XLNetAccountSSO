<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Button, Form, FormItem, Input, Message, VerificationCode } from '@arco-design/web-vue'
import { computed, onMounted, reactive, shallowRef } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

useHead({ title: '登录 — XLNetAccount' })

import { ApiError } from '@/api/http'
import { startPasskeyLogin } from '@/api/passkeys'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'
import { buildNextQuery, resolveNextTarget } from '@/utils/authNext'
import { describePasskeyError, getPasskeyCredential, getPasskeySupportMessage } from '@/utils/webauthn'
import { solveCaptcha } from '@/utils/captcha'
import CaptchaCard from '@/components/ui/CaptchaCard.vue'
import { getServerBaseUrl } from '@/config/endpoints'

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

const oauthProviders = [
	{
		id: 'microsoft', label: 'Microsoft',
		icon: '<svg viewBox="0 0 21 21" width="18" height="18"><path fill="#f25022" d="M10 0H0v10h10z"/><path fill="#7fba00" d="M21 0H11v10h10z"/><path fill="#00a4ef" d="M10 11H0v10h10z"/><path fill="#ffb900" d="M21 11H11v10h10z"/></svg>',
	},
	{
		id: 'google', label: 'Google',
		icon: '<svg viewBox="0 0 24 24" width="18" height="18"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/></svg>',
	},
	{
		id: 'github', label: 'GitHub',
		icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd"/></svg>',
	},
]

onMounted(() => {
	const token = route.query.oauth_token as string | undefined
	if (token) {
		import('@/utils/authToken').then(({ setAuthToken }) => {
			setAuthToken(token)
			sessionStore.syncSession().then(() => {
				Message.success('登录成功')
				router.replace(resolveNextTarget(route.query.next))
			})
		})
	}
})

const isSubmitting = shallowRef(false)
const isPasskeySubmitting = shallowRef(false)
const submitError = shallowRef('')
const passkeySupportMessage = getPasskeySupportMessage()
const captchaToken = shallowRef('')
const capEnabled = computed(() => Boolean(platformStore.capWidgetEndpoint))
const showTotpStep = shallowRef(false)
const totpSessionId = shallowRef('')
const totpCode = shallowRef('')
const isTotpSubmitting = shallowRef(false)
const formState = reactive({
  username: '',
  password: '',
})

async function handleSubmit() {
  submitError.value = ''

  // 自动完成人机验证
  if (capEnabled.value && !captchaToken.value) {
    isSubmitting.value = true
    Message.loading('正在进行人机验证，请稍候...')
    try {
      const token = await solveCaptcha(platformStore.capWidgetEndpoint)
      captchaToken.value = token
    } catch (error) {
      Message.clear()
      isSubmitting.value = false
      Message.error(error instanceof Error ? error.message : '人机验证失败，请重试')
      return
    }
    Message.clear()
  }

  isSubmitting.value = true

  try {
    const result = await sessionStore.signIn({
      username: formState.username,
      password: formState.password,
      captcha_token: captchaToken.value || undefined,
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

        <FormItem v-if="capEnabled" label="人机验证">
          <CaptchaCard
            :api-endpoint="platformStore.capWidgetEndpoint"
            @solve="(token) => captchaToken = token"
            @error="() => captchaToken = ''"
          />
        </FormItem>

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

        <div class="oauth-divider">
          <span class="oauth-divider-text">第三方登录</span>
        </div>

        <div class="oauth-providers">
          <a
            v-for="p in oauthProviders"
            :key="p.id"
            :href="getServerBaseUrl() + `/api/auth/oauth/${p.id}/login`"
            class="oauth-btn"
          >
            <span class="oauth-btn-icon" v-html="p.icon"></span>
            <span>{{ p.label }}</span>
          </a>
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

.oauth-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 4px 0;
}

.oauth-divider::before,
.oauth-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--color-hairline);
}

.oauth-divider-text {
  font-size: 11px;
  color: var(--color-charcoal);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  white-space: nowrap;
}

.oauth-providers {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.oauth-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid var(--color-hairline-strong);
  border-radius: var(--radius-md, 8px);
  background: var(--color-surface-card);
  color: var(--color-ink);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  transition: border-color 0.15s, background 0.15s;
}

.oauth-btn:hover {
  border-color: var(--color-accent-blue);
  background: rgba(59, 158, 255, 0.06);
  color: var(--color-ink);
}

.oauth-btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.oauth-btn-icon :deep(svg) {
  display: block;
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
