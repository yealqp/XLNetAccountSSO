<script setup lang="ts">
import { NAlert, NButton, NForm, NFormItem, NInput, useMessage } from 'naive-ui'
import { computed, reactive, shallowRef } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { ApiError } from '@/api/http'
import { startPasskeyLogin } from '@/api/passkeys'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'
import { buildNextQuery, resolveNextTarget } from '@/utils/authNext'
import { describePasskeyError, getPasskeyCredential, getPasskeySupportMessage } from '@/utils/webauthn'

const route = useRoute()
const router = useRouter()
const message = useMessage()
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
const formState = reactive({
  username: '',
  password: '',
})

async function handleSubmit() {
  isSubmitting.value = true
  submitError.value = ''

  try {
    await sessionStore.signIn({
      username: formState.username,
      password: formState.password,
    })

    message.success('登录成功')
    await router.replace(resolveNextTarget(route.query.next))
  }
  catch (error) {
    const fallbackMessage = error instanceof ApiError ? error.message : '登录失败，请稍后重试'
    submitError.value = fallbackMessage
    message.error(fallbackMessage)
  }
  finally {
    isSubmitting.value = false
  }
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

    message.success('登录成功')
    await router.replace(resolveNextTarget(route.query.next))
  }
  catch (error) {
    const fallbackMessage = error instanceof ApiError ? error.message : describePasskeyError(error)
    submitError.value = fallbackMessage
    message.error(fallbackMessage)
  }
  finally {
    isPasskeySubmitting.value = false
  }
}
</script>

<template>
  <div class="auth-panel-view">
    <div class="auth-panel-header">
      <p class="auth-panel-kicker">欢迎回来</p>
      <h2 class="auth-panel-title">登录到 {{ platformStore.displayName }}</h2>
      <p class="auth-panel-subtitle">继续管理账号、客户端与令牌。</p>
    </div>

    <NForm label-placement="top" class="auth-form" @submit.prevent="handleSubmit">
      <NFormItem label="邮箱或用户名">
        <NInput v-model:value="formState.username" clearable placeholder="邮箱或用户名" size="large" @update:value="submitError = ''" />
      </NFormItem>

      <NFormItem label="密码">
        <NInput
          v-model:value="formState.password"
          type="password"
          show-password-on="mousedown"
          placeholder="密码"
          size="large"
          @update:value="submitError = ''"
        />
      </NFormItem>

      <NAlert v-if="submitError" type="error" :show-icon="false">
        {{ submitError }}
      </NAlert>

      <NAlert v-if="passkeySupportMessage" type="warning" :show-icon="false">
        {{ passkeySupportMessage }}
      </NAlert>

      <div class="auth-action-stack">
        <NButton type="primary" size="large" block :loading="isSubmitting" :disabled="isPasskeySubmitting" attr-type="submit">
          登录
        </NButton>
        <NButton
          secondary
          size="large"
          block
          attr-type="button"
          data-testid="passkey-login-button"
          :loading="isPasskeySubmitting"
          :disabled="Boolean(passkeySupportMessage) || isSubmitting"
          @click="handlePasskeySignIn"
        >
          使用通行密钥登录
        </NButton>
      </div>
    </NForm>

    <div v-if="showRegister" class="auth-link-row">
      <RouterLink :to="registerLink">没有账号？立即注册</RouterLink>
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

.auth-link-row {
  margin-top: 16px;
}

.auth-action-stack {
  display: grid;
  gap: 12px;
}

.auth-link-row a {
  color: #9fd6ff;
}

</style>
