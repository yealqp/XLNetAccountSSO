<script setup lang="ts">
import '@cap.js/widget'

import { NAlert, NButton, NForm, NFormItem, NInput, NSpace, NText, useMessage } from 'naive-ui'
import { computed, reactive, shallowRef } from 'vue'
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

const formState = reactive({
	email: '',
	displayName: '',
	password: '',
	confirmPassword: '',
	code: '',
})

void platformStore.ensureLoaded().catch(() => {})

const capEnabled = computed(() => Boolean(platformStore.capApiEndpoint))

async function handleSendCode() {
	if (!formState.email.trim()) {
		submitError.value = '请先输入邮箱'
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
			email: formState.email,
			captcha_token: captchaToken.value,
		})
		codeSent.value = true
		message.success('验证码已发送，请查收邮箱')
	}
	catch (error) {
		submitError.value = error instanceof ApiError ? error.message : '发送验证码失败'
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

	isSubmitting.value = true
	submitError.value = ''

	try {
		await register({
			email: formState.email,
			password: formState.password,
			display_name: formState.displayName,
			code: formState.code,
		})
		await sessionStore.signIn({
			username: formState.email,
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

      <NFormItem v-if="capEnabled" label="人机验证">
        <div class="cap-shell">
          <cap-widget
            :data-cap-api-endpoint="platformStore.capApiEndpoint"
            @solve="handleCapSolve"
            @error="handleCapError"
          />
        </div>
        <NText v-if="capError" depth="3">{{ capError }}</NText>
      </NFormItem>

      <NFormItem label="验证码">
        <NSpace style="width: 100%;" :wrap="false">
          <NInput v-model:value="formState.code" placeholder="邮箱验证码" size="large" @update:value="submitError = ''" />
          <NButton secondary :loading="isSendingCode" @click="handleSendCode">
            {{ codeSent ? '重新发送' : '发送验证码' }}
          </NButton>
        </NSpace>
      </NFormItem>

      <NFormItem label="显示名称">
        <NInput v-model:value="formState.displayName" clearable placeholder="显示名称，可选" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NFormItem label="密码">
        <NInput v-model:value="formState.password" type="password" show-password-on="mousedown" placeholder="密码" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NFormItem label="确认密码">
        <NInput v-model:value="formState.confirmPassword" type="password" show-password-on="mousedown" placeholder="再次输入密码" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NAlert v-if="submitError" type="error" :show-icon="false">
        {{ submitError }}
      </NAlert>

      <NButton type="primary" size="large" block :loading="isSubmitting" attr-type="submit">
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

.cap-shell {
  width: 100%;
  overflow-x: auto;
}

.auth-link-row {
  margin-top: 16px;
}

.auth-link-row a {
  color: #9fd6ff;
}
</style>
