<script setup lang="ts">
import { NAlert, NButton, NCard, NForm, NFormItem, NInput, NSpace, NText, useMessage } from 'naive-ui'
import { computed, onBeforeUnmount, onMounted, reactive, shallowRef } from 'vue'

import { sendProfilePasswordCode, updateProfile } from '@/api/auth'
import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'

const message = useMessage()
const sessionStore = useSessionStore()

const isSavingProfile = shallowRef(false)
const isSendingProfileCode = shallowRef(false)
const loadError = shallowRef('')

const profileState = reactive({
  username: '',
  email: '',
  password: '',
  code: '',
})

const passwordHint = '密码需为 8-20 位，且包含大写字母、小写字母和数字'
const resendRemaining = shallowRef(0)
let resendTimer: ReturnType<typeof setInterval> | null = null

const profilePasswordChecks = computed(() => {
  const password = profileState.password
  if (!password) {
    return []
  }
  return [
    { label: '8-20 位长度', passed: password.length >= 8 && password.length <= 20 },
    { label: '包含大写字母', passed: /[A-Z]/.test(password) },
    { label: '包含小写字母', passed: /[a-z]/.test(password) },
    { label: '包含数字', passed: /\d/.test(password) },
  ]
})
const profilePasswordValid = computed(() => profileState.password === '' || profilePasswordChecks.value.every(item => item.passed))
const sendCodeDisabled = computed(() => isSendingProfileCode.value || resendRemaining.value > 0)

onMounted(() => {
  profileState.username = sessionStore.user?.username ?? ''
  profileState.email = sessionStore.user?.email ?? ''
})

onBeforeUnmount(() => {
  clearResendTimer()
})

async function handleSaveProfile() {
  if (!profilePasswordValid.value) {
    loadError.value = passwordHint
    message.error(loadError.value)
    return
  }
  if (profileState.password && !profileState.code.trim()) {
    loadError.value = '修改密码需要邮箱验证码'
    message.error(loadError.value)
    return
  }
  isSavingProfile.value = true
  loadError.value = ''
  try {
    const response = await updateProfile({
      username: profileState.username,
      password: profileState.password,
      code: profileState.code,
    })
    sessionStore.setUser(response.user ?? null)
    profileState.email = response.user?.email ?? profileState.email
    profileState.password = ''
    profileState.code = ''
    message.success('个人信息已保存')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '保存个人信息失败'
    message.error(loadError.value)
  }
  finally {
    isSavingProfile.value = false
  }
}

async function handleSendProfileCode() {
  isSendingProfileCode.value = true
  loadError.value = ''
  try {
    await sendProfilePasswordCode()
    startResendCountdown(60)
    message.success('验证码已发送至当前邮箱')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '发送验证码失败'
    const retryAfter = extractRetryAfter(loadError.value)
    if (retryAfter > 0) {
      startResendCountdown(retryAfter)
    }
    message.error(loadError.value)
  }
  finally {
    isSendingProfileCode.value = false
  }
}

function startResendCountdown(seconds: number) {
  clearResendTimer()
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
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">普通设置</h1>
        <p class="page-subtitle">修改当前账号信息。</p>
      </div>
    </header>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      {{ loadError }}
    </NAlert>

    <NCard title="个人信息">
      <NForm label-placement="top" @submit.prevent="handleSaveProfile">
        <NFormItem label="用户名">
          <NInput v-model:value="profileState.username" placeholder="用户名" />
        </NFormItem>

        <NFormItem label="邮箱">
          <NInput v-model:value="profileState.email" placeholder="邮箱" disabled />
        </NFormItem>

        <NFormItem label="新密码">
          <NInput v-model:value="profileState.password" type="password" show-password-on="click" placeholder="留空则不修改" />
        </NFormItem>

        <NText v-if="profileState.password" depth="3">{{ passwordHint }}</NText>

        <div v-if="profilePasswordChecks.length > 0" class="password-rule-list">
          <div v-for="rule in profilePasswordChecks" :key="rule.label" class="password-rule" :class="{ passed: rule.passed }">
            <span>{{ rule.passed ? '✓' : '○' }}</span>
            <span>{{ rule.label }}</span>
          </div>
        </div>

        <NFormItem label="邮箱验证码">
          <NSpace style="width: 100%;" :wrap="false">
            <NInput v-model:value="profileState.code" placeholder="修改密码时必填" />
            <NButton secondary :loading="isSendingProfileCode" :disabled="sendCodeDisabled" @click="handleSendProfileCode">
              {{ resendRemaining > 0 ? `${resendRemaining}s` : '发送验证码' }}
            </NButton>
          </NSpace>
        </NFormItem>

        <NButton type="primary" attr-type="submit" :loading="isSavingProfile">
          保存个人信息
        </NButton>
      </NForm>
    </NCard>
  </section>
</template>

<style scoped>
.password-rule-list {
  display: grid;
  gap: 6px;
  margin: 0 0 16px;
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
</style>
