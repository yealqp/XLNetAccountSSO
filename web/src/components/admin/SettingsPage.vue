<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { NAlert, NButton, NCard, NForm, NFormItem, NGrid, NGridItem, NInput, NSpace, NSpin, NText, useMessage } from 'naive-ui'
import { computed, onBeforeUnmount, onMounted, reactive, shallowRef } from 'vue'

useHead({ title: '普通设置 — XLNetAccount' })

import { sendProfilePasswordCode, updateProfile } from '@/api/auth'
import { ApiError } from '@/api/http'
import { deletePasskey, fetchPasskeys, finishPasskeyRegistration, startPasskeyRegistration } from '@/api/passkeys'
import { disableTOTP, fetchTOTPStatus, startTOTPSetup, verifyTOTPSetup } from '@/api/totp'
import { useSessionStore } from '@/stores/session'
import type { PasskeyRecord } from '@/types/api'
import { createPasskeyCredential, describePasskeyError, getPasskeySupportMessage } from '@/utils/webauthn'

const message = useMessage()
const sessionStore = useSessionStore()

const isSavingProfile = shallowRef(false)
const isSendingProfileCode = shallowRef(false)
const loadError = shallowRef('')
const isLoadingPasskeys = shallowRef(false)
const isCreatingPasskey = shallowRef(false)
const deletingPasskeyId = shallowRef('')
const passkeyError = shallowRef('')
const passkeys = shallowRef<PasskeyRecord[]>([])

const profileState = reactive({
  username: '',
  email: '',
  password: '',
  code: '',
})
const passkeyState = reactive({
  name: '',
})

const passwordHint = '密码需为 8-20 位，且包含大写字母、小写字母和数字'
const resendRemaining = shallowRef(0)
let resendTimer: ReturnType<typeof setInterval> | null = null
const passkeySupportMessage = getPasskeySupportMessage()

const totpEnabled = shallowRef(false)
const isLoadingTOTPStatus = shallowRef(false)
const isEnablingTOTP = shallowRef(false)
const isDisablingTOTP = shallowRef(false)
const showTotpSetup = shallowRef(false)
const totpSetupData = shallowRef<{ secret: string; uri: string; qr_data_uri: string; setup_session_id: string } | null>(null)
const totpSetupCode = shallowRef('')
const totpSetupError = shallowRef('')
const showTotpDisableConfirm = shallowRef(false)
const totpError = shallowRef('')

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
  void loadPasskeys()
  void loadTOTPStatus()
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

async function loadPasskeys() {
  isLoadingPasskeys.value = true
  passkeyError.value = ''
  try {
    const response = await fetchPasskeys()
    passkeys.value = response.items
  }
  catch (error) {
    passkeyError.value = error instanceof ApiError ? error.message : '加载通行密钥失败'
  }
  finally {
    isLoadingPasskeys.value = false
  }
}

async function handleCreatePasskey() {
  isCreatingPasskey.value = true
  passkeyError.value = ''
  try {
    const start = await startPasskeyRegistration()
    const credential = await createPasskeyCredential(start.options)
    const response = await finishPasskeyRegistration({
      session_id: start.session_id,
      name: passkeyState.name.trim() || undefined,
      credential,
    })
    passkeys.value = [response.credential, ...passkeys.value.filter(item => item.id !== response.credential.id)]
    passkeyState.name = ''
    message.success('通行密钥已添加')
  }
  catch (error) {
    passkeyError.value = error instanceof ApiError ? error.message : describePasskeyError(error)
    message.error(passkeyError.value)
  }
  finally {
    isCreatingPasskey.value = false
  }
}

async function handleDeletePasskey(passkeyId: string) {
  deletingPasskeyId.value = passkeyId
  passkeyError.value = ''
  try {
    await deletePasskey(passkeyId)
    passkeys.value = passkeys.value.filter(item => item.id !== passkeyId)
    message.success('通行密钥已删除')
  }
  catch (error) {
    passkeyError.value = error instanceof ApiError ? error.message : '删除通行密钥失败'
    message.error(passkeyError.value)
  }
  finally {
    deletingPasskeyId.value = ''
  }
}

function formatPasskeyTime(value: string | null) {
  if (!value) {
    return '未使用'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN', { hour12: false })
}

async function loadTOTPStatus() {
  isLoadingTOTPStatus.value = true
  totpError.value = ''
  try {
    const status = await fetchTOTPStatus()
    totpEnabled.value = status.enabled
  }
  catch (error) {
    totpError.value = error instanceof ApiError ? error.message : '加载两步验证状态失败'
  }
  finally {
    isLoadingTOTPStatus.value = false
  }
}

async function handleEnableTOTP() {
  isEnablingTOTP.value = true
  totpSetupError.value = ''
  try {
    const data = await startTOTPSetup()
    totpSetupData.value = data
    showTotpSetup.value = true
  }
  catch (error) {
    totpSetupError.value = error instanceof ApiError ? error.message : '开启两步验证失败'
    message.error(totpSetupError.value)
  }
  finally {
    isEnablingTOTP.value = false
  }
}

async function handleVerifyTOTPSetup() {
  if (!totpSetupData.value) {
    return
  }
  isEnablingTOTP.value = true
  totpSetupError.value = ''
  try {
    await verifyTOTPSetup({
      setup_session_id: totpSetupData.value.setup_session_id,
      code: totpSetupCode.value,
    })
    totpEnabled.value = true
    showTotpSetup.value = false
    totpSetupData.value = null
    totpSetupCode.value = ''
    message.success('两步验证已启用')
  }
  catch (error) {
    totpSetupError.value = error instanceof ApiError ? error.message : '验证失败'
    message.error(totpSetupError.value)
  }
  finally {
    isEnablingTOTP.value = false
  }
}

function handleCancelTOTPSetup() {
  showTotpSetup.value = false
  totpSetupData.value = null
  totpSetupCode.value = ''
  totpSetupError.value = ''
}

async function handleDisableTOTP() {
  isDisablingTOTP.value = true
  totpError.value = ''
  try {
    await disableTOTP()
    totpEnabled.value = false
    showTotpDisableConfirm.value = false
    message.success('两步验证已关闭')
  }
  catch (error) {
    totpError.value = error instanceof ApiError ? error.message : '关闭两步验证失败'
    message.error(totpError.value)
  }
  finally {
    isDisablingTOTP.value = false
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

    <NGrid cols="1 l:2" responsive="screen" :x-gap="12" :y-gap="12">
      <NGridItem>
        <NCard title="通行密钥" size="small">
          <NSpace vertical :size="12">
            <NText depth="3">使用通行密钥完成无密码登录，保留密码作为回退方式。</NText>

            <NAlert v-if="passkeySupportMessage" type="warning" :show-icon="false">
              {{ passkeySupportMessage }}
            </NAlert>

            <NAlert v-if="passkeyError" type="error" :show-icon="false">
              {{ passkeyError }}
            </NAlert>

            <NForm label-placement="top" @submit.prevent="handleCreatePasskey">
              <NFormItem label="通行密钥名称">
                <NSpace style="width: 100%;" :wrap="false">
                  <NInput v-model:value="passkeyState.name" placeholder="当前设备（可选）" />
                  <NButton
                    type="primary"
                    attr-type="submit"
                    data-testid="passkey-enroll-button"
                    :loading="isCreatingPasskey"
                    :disabled="Boolean(passkeySupportMessage)"
                  >
                    添加通行密钥
                  </NButton>
                </NSpace>
              </NFormItem>
            </NForm>

            <NSpin :show="isLoadingPasskeys">
              <div v-if="passkeys.length > 0" class="passkey-list">
                <div v-for="passkey in passkeys" :key="passkey.id" class="passkey-item">
                  <div class="passkey-item-main">
                    <div class="passkey-name">{{ passkey.name }}</div>
                    <NText depth="3">创建于 {{ formatPasskeyTime(passkey.created_at) }}</NText>
                    <NText depth="3">最近使用 {{ formatPasskeyTime(passkey.last_used_at) }}</NText>
                  </div>
                  <NButton
                    secondary
                    type="error"
                    data-testid="passkey-delete-button"
                    :loading="deletingPasskeyId === passkey.id"
                    @click="handleDeletePasskey(passkey.id)"
                  >
                    删除
                  </NButton>
                </div>
              </div>
              <NText v-else depth="3">当前还没有绑定通行密钥。</NText>
            </NSpin>
          </NSpace>
        </NCard>
      </NGridItem>
      <NGridItem>
        <NCard title="两步验证（2FA）" size="small">
          <NSpace vertical :size="12">
            <NText depth="3">启用后登录需额外输入验证器 App 生成的 6 位动态码，增强账号安全性。</NText>

            <NAlert v-if="totpError" type="error" :show-icon="false">
              {{ totpError }}
            </NAlert>

            <NSpin :show="isLoadingTOTPStatus">
              <template v-if="!totpEnabled && !showTotpSetup">
                <NButton type="primary" data-testid="totp-enable-button" :loading="isEnablingTOTP" @click="handleEnableTOTP">
                  启用两步验证
                </NButton>
              </template>

              <template v-if="showTotpSetup && totpSetupData">
                <NAlert type="info" :show-icon="false">
                  <template #header>
                    使用验证器 App 扫描以下二维码
                  </template>
                  <div style="text-align: center; margin: 12px 0;">
                    <img :src="totpSetupData.qr_data_uri" alt="TOTP QR Code" style="width: 180px; height: 180px; border-radius: 6px;" />
                  </div>
                  <div style="text-align: center; margin-bottom: 8px;">
                    <NText depth="3">无法扫描？手动输入密钥：</NText>
                    <NText code>{{ totpSetupData.secret }}</NText>
                  </div>
                </NAlert>

                <div class="totp-verify-row">
                  <NInput
                    v-model:value="totpSetupCode"
                    placeholder="输入 6 位动态码"
                    maxlength="6"
                  />
                  <NButton type="primary" :loading="isEnablingTOTP" @click="handleVerifyTOTPSetup">
                    验证
                  </NButton>
                  <NButton secondary @click="handleCancelTOTPSetup">
                    取消
                  </NButton>
                </div>

                <NAlert v-if="totpSetupError" type="error" :show-icon="false">
                  {{ totpSetupError }}
                </NAlert>
              </template>

              <template v-if="totpEnabled && !showTotpSetup">
                <div class="totp-status-row">
                  <NText type="success">两步验证已启用</NText>
                  <template v-if="!showTotpDisableConfirm">
                    <NButton secondary type="error" data-testid="totp-disable-button" @click="showTotpDisableConfirm = true">
                      关闭两步验证
                    </NButton>
                  </template>
                  <template v-else>
                    <NSpace>
                      <NText depth="3">确认关闭？</NText>
                      <NButton type="error" size="small" :loading="isDisablingTOTP" @click="handleDisableTOTP">
                        确认关闭
                      </NButton>
                      <NButton secondary size="small" @click="showTotpDisableConfirm = false">
                        取消
                      </NButton>
                    </NSpace>
                  </template>
                </div>
              </template>
            </NSpin>
          </NSpace>
        </NCard>
      </NGridItem>
    </NGrid>
  </section>
</template>

<style scoped>
.passkey-list {
  display: grid;
  gap: 12px;
}

.passkey-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.02);
}

.passkey-item-main {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.passkey-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-ink);
  overflow-wrap: anywhere;
}

.password-rule-list {
  display: grid;
  gap: 6px;
  margin: 0 0 16px;
  padding: 10px 12px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.02);
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

.totp-verify-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.totp-status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

@media (max-width: 720px) {
  .passkey-item {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
