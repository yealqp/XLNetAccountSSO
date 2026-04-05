<script setup lang="ts">
import { NAlert, NButton, NForm, NFormItem, NInput, useMessage } from 'naive-ui'
import { reactive, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ApiError } from '@/api/http'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const platformStore = usePlatformStore()
const sessionStore = useSessionStore()
void platformStore.ensureLoaded().catch(() => {})

const isSubmitting = shallowRef(false)
const submitError = shallowRef('')
const formState = reactive({
  username: 'admin',
  password: 'Admin123!',
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
      <p class="auth-panel-kicker">欢迎回来</p>
      <h2 class="auth-panel-title">登录到 {{ platformStore.displayName }}</h2>
      <p class="auth-panel-subtitle">继续管理账号、客户端与令牌。</p>
    </div>

    <NForm label-placement="top" class="auth-form" @submit.prevent="handleSubmit">
      <NFormItem label="用户名">
        <NInput v-model:value="formState.username" clearable placeholder="用户名" size="large" @update:value="submitError = ''" />
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

      <NButton type="primary" size="large" block :loading="isSubmitting" attr-type="submit">
        登录
      </NButton>
    </NForm>

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

</style>
