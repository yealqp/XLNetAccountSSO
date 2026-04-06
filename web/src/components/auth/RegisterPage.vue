<script setup lang="ts">
import '@cap.js/widget'

import { NAlert, NButton, NForm, NFormItem, NInput, NSpace, NText, useMessage } from 'naive-ui'
import { computed, onBeforeUnmount, reactive, shallowRef } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { register, sendRegisterCode } from '@/api/auth'
import { ApiError } from '@/api/http'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'

interface CapSolveEvent extends Event {
	detail: {
		token: string
	}
}

const route = useRoute()
const router = useRouter()
const message = useMessage()
const platformStore = usePlatformStore()
const sessionStore = useSessionStore()

const isSubmitting = shallowRef(false)
const isSendingCode = shallowRef(false)
const submitError = shallowRef('')
const codeSent = shallowRef(false)
const captchaToken = shallowRef('')
const capError = shallowRef('')
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

const capEnabled = computed(() => Boolean(platformStore.capApiEndpoint))
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
	if (isSendingCode.value) {
		return true
	}
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
		submitError.value = '请先输入邮箱'
		return
	}
	if (resendRemaining.value > 0 && resendEmail.value === normalizedEmail.value) {
		submitError.value = `请在 ${resendRemaining.value} 秒后重试`
		return
	}
	if (capEnabled.value && !captchaToken.value) {
		capError.value = '请先完成人机验证'
		return
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
		message.success('验证码已发送，请查收邮箱')
	}
	catch (error) {
		submitError.value = error instanceof ApiError ? error.message : '发送验证码失败'
		const retryAfter = extractRetryAfter(submitError.value)
		if (retryAfter > 0) {
			startResendCountdown(normalizedEmail.value, retryAfter)
		}
		message.error(submitError.value)
	}
	finally {
		isSendingCode.value = false
	}
}

async function handleSubmit() {
	if (formState.password !== formState.confirmPassword) {
		submitError.value = '两次输入的密码不一致'
		return
	}
	if (!isPasswordValid(formState.password)) {
		submitError.value = passwordHint
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
		message.success('注册成功')
		await router.replace(resolveNextTarget(route.query.next))
	}
	catch (error) {
		submitError.value = error instanceof ApiError ? error.message : '注册失败，请稍后重试'
		message.error(submitError.value)
	}
	finally {
		isSubmitting.value = false
	}
}

function handleCapSolve(event: Event) {
	const solveEvent = event as CapSolveEvent
	captchaToken.value = solveEvent.detail.token
	capError.value = ''
}

function handleCapError() {
	captchaToken.value = ''
	capError.value = '人机验证失败，请重试'
}

function resolveNextTarget(nextValue: unknown) {
	const next = Array.isArray(nextValue) ? nextValue[0] : nextValue
	if (typeof next === 'string' && next.startsWith('/') && !next.startsWith('//')) {
		return next
	}
	return '/admin'
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
	if (!matched) {
		return 0
	}
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

    <NForm label-placement="top" class="auth-form" @submit.prevent="handleSubmit">
      <NFormItem label="邮箱">
        <NInput v-model:value="formState.email" clearable placeholder="邮箱" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NFormItem label="用户名">
        <NInput v-model:value="formState.username" clearable placeholder="用户名" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NFormItem v-if="capEnabled" label="人机验证">
        <div class="cap-shell">
          <cap-widget
            :data-cap-api-endpoint="platformStore.capApiEndpoint"
            data-cap-i18n-initial-state="点击开始验证"
            data-cap-i18n-verifying-label="验证中..."
            data-cap-i18n-solved-label="验证通过"
            data-cap-i18n-error-label="验证失败，请重试"
            data-cap-i18n-verify-aria-label="点击开始人机验证"
            data-cap-i18n-verifying-aria-label="正在进行人机验证"
            data-cap-i18n-verified-aria-label="人机验证通过"
            data-cap-i18n-error-aria-label="人机验证失败，请重试"
            data-cap-i18n-wasm-disabled="当前环境未启用 WASM，验证速度可能较慢"
            data-cap-i18n-troubleshooting-label="查看帮助"
            @solve="handleCapSolve"
            @error="handleCapError"
          />
        </div>
        <NText v-if="capError" depth="3">{{ capError }}</NText>
      </NFormItem>

      <NFormItem label="验证码">
        <NSpace style="width: 100%;" :wrap="false">
          <NInput v-model:value="formState.code" placeholder="邮箱验证码" size="large" @update:value="submitError = ''" />
          <NButton secondary :loading="isSendingCode" :disabled="sendCodeDisabled" @click="handleSendCode">
            {{ sendCodeText }}
          </NButton>
        </NSpace>
      </NFormItem>

		<NFormItem label="密码">
        <NInput v-model:value="formState.password" type="password" show-password-on="mousedown" placeholder="密码" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NFormItem label="确认密码">
        <NInput v-model:value="formState.confirmPassword" type="password" show-password-on="mousedown" placeholder="再次输入密码" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NAlert type="info" :show-icon="false" class="password-hint">
        {{ passwordHint }}
      </NAlert>

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

      <NAlert v-if="submitError" type="error" :show-icon="false">
        {{ submitError }}
      </NAlert>

      <NButton type="primary" size="large" block :loading="isSubmitting" attr-type="submit" :disabled="!passwordValid || passwordMismatch">
        注册
      </NButton>
    </NForm>

    <div class="auth-link-row">
      <RouterLink to="/auth/login">已有账号？去登录</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.auth-panel-view {
  color: #eff6ff;
}

.auth-panel-header {
  margin-bottom: 24px;
}

.auth-panel-kicker {
  margin: 0 0 10px;
  color: rgba(208, 226, 248, 0.72);
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
  color: rgba(226, 236, 248, 0.68);
  font-size: 14px;
  line-height: 1.65;
}

.auth-form {
  display: grid;
  gap: 4px;
}

.password-hint {
  margin-top: -2px;
}

.password-rule-list {
	display: grid;
	gap: 6px;
	margin-top: -2px;
	padding: 10px 12px;
	border: 1px solid rgba(255, 255, 255, 0.08);
	border-radius: 8px;
	background: rgba(255, 255, 255, 0.02);
}

.password-rule {
	display: flex;
	align-items: center;
	gap: 8px;
	color: rgba(226, 236, 248, 0.64);
	font-size: 13px;
}

.password-rule.passed {
	color: #7ed6a7;
}

.cap-shell {
  width: 100%;
  overflow-x: auto;
}

.cap-shell :deep(cap-widget) {
  --cap-background: #161a22;
  --cap-border-color: #2a3342;
  --cap-border-radius: 4px;
  --cap-widget-height: 36px;
  --cap-widget-width: 100%;
  --cap-widget-padding: 14px;
  --cap-gap: 14px;
  --cap-color: #eef4ff;
  --cap-checkbox-size: 20px;
  --cap-checkbox-border: 1px solid #395174;
  --cap-checkbox-border-radius: 4px;
  --cap-checkbox-background: #0f131a;
  --cap-spinner-color: #4da3ff;
  --cap-spinner-background-color: #243042;
  display: block;
}

.cap-shell :deep(cap-widget)::part(attribution) {
  display: none;
}

.auth-link-row {
  margin-top: 16px;
}

.auth-link-row a {
  color: #9fd6ff;
}
</style>
