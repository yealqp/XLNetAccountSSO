<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Button, Card, Form, FormItem, Grid, GridItem, Input, Message, Popconfirm, Space, Spin, TypographyText as Text } from '@arco-design/web-vue'
import { computed, onBeforeUnmount, onMounted, reactive, shallowRef } from 'vue'
import { useRoute } from 'vue-router'

useHead({ title: '用户中心 — XLNetAccount' })

import { sendProfilePasswordCode, updateProfile } from '@/api/auth'
import { ApiError } from '@/api/http'
import { deletePasskey, fetchPasskeys, finishPasskeyRegistration, startPasskeyRegistration } from '@/api/passkeys'
import { disableTOTP, fetchTOTPStatus, startTOTPSetup, verifyTOTPSetup } from '@/api/totp'
import { useSessionStore } from '@/stores/session'
import type { PasskeyRecord } from '@/types/api'
import { createPasskeyCredential, describePasskeyError, getPasskeySupportMessage } from '@/utils/webauthn'
import type { OAuthBindingRecord } from '@/types/api'
import { fetchOAuthBindings, unlinkOAuthBinding, getOAuthBindUrl } from '@/api/oauth-binding'
import { oauthProviders } from '@/constants/oauth'

const sessionStore = useSessionStore()
const route = useRoute()

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
const totpError = shallowRef('')

const oauthBindings = shallowRef<OAuthBindingRecord[]>([])
const isLoadingBindings = shallowRef(false)
const unlinkingId = shallowRef('')

// oauthProviders imported from constants/oauth

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
  void loadOAuthBindings()

  // Handle OAuth bind callback result
  if (route.query.oauth_bind_success) {
    Message.success('已绑定 ' + route.query.oauth_bind_success + ' 账号')
    void loadOAuthBindings()
  } else if (route.query.oauth_bind_error) {
    const msg = {
      already_bound: '该第三方账号已被其他用户绑定',
      already_bound_to_user: '已绑定过该类型的账号',
      link_failed: '绑定失败，请重试',
    }[route.query.oauth_bind_error as string] || '绑定失败'
    Message.error(msg)
  }
})

onBeforeUnmount(() => {
  clearResendTimer()
})

async function handleSaveProfile() {
  if (!profilePasswordValid.value) {
    loadError.value = passwordHint
    Message.error(loadError.value)
    return
  }
  if (profileState.password && !profileState.code.trim()) {
    loadError.value = '修改密码需要邮箱验证码'
    Message.error(loadError.value)
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
    Message.success('个人信息已保存')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '保存个人信息失败'
    Message.error(loadError.value)
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
    Message.success('验证码已发送至当前邮箱')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '发送验证码失败'
    const retryAfter = extractRetryAfter(loadError.value)
    if (retryAfter > 0) {
      startResendCountdown(retryAfter)
    }
    Message.error(loadError.value)
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
    Message.success('通行密钥已添加')
  }
  catch (error) {
    passkeyError.value = error instanceof ApiError ? error.message : describePasskeyError(error)
    Message.error(passkeyError.value)
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
    Message.success('通行密钥已删除')
  }
  catch (error) {
    passkeyError.value = error instanceof ApiError ? error.message : '删除通行密钥失败'
    Message.error(passkeyError.value)
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

async function loadOAuthBindings() {
  isLoadingBindings.value = true
  try {
    const res = await fetchOAuthBindings()
    oauthBindings.value = res.items
  } catch (error) {
    // silently fail
  } finally {
    isLoadingBindings.value = false
  }
}

async function handleUnlinkOAuth(id: string) {
  unlinkingId.value = id
  try {
    await unlinkOAuthBinding(id)
    oauthBindings.value = oauthBindings.value.filter(b => b.id !== id)
    Message.success('已解绑')
  } catch (error) {
    Message.error(error instanceof ApiError ? error.message : '解绑失败')
  } finally {
    unlinkingId.value = ''
  }
}

function isBound(provider: string) {
  return oauthBindings.value.some(b => b.provider === provider)
}

function getBinding(provider: string) {
  return oauthBindings.value.find(b => b.provider === provider)
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
    Message.error(totpSetupError.value)
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
    Message.success('两步验证已启用')
  }
  catch (error) {
    totpSetupError.value = error instanceof ApiError ? error.message : '验证失败'
    Message.error(totpSetupError.value)
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
    Message.success('两步验证已关闭')
  }
  catch (error) {
    totpError.value = error instanceof ApiError ? error.message : '关闭两步验证失败'
    Message.error(totpError.value)
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

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    Message.success('已复制')
  }
  catch {
    Message.error('复制失败')
  }
}
</script>

<template>
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">用户中心</h1>
        <p class="page-subtitle">管理个人信息、通行密钥与安全设置。</p>
      </div>
    </header>

    <Card title="个人信息">
      <Form :model="profileState" layout="vertical" @submit="handleSaveProfile">
        <FormItem label="用户名">
          <Input v-model="profileState.username" placeholder="用户名" />
        </FormItem>

        <FormItem label="邮箱">
          <Input v-model="profileState.email" placeholder="邮箱" disabled />
        </FormItem>

        <FormItem label="新密码">
          <Input v-model="profileState.password" type="password" placeholder="留空则不修改" />
        </FormItem>

        <Text v-if="profileState.password" type="secondary">{{ passwordHint }}</Text>

        <div v-if="profilePasswordChecks.length > 0" class="password-rule-list">
          <div v-for="rule in profilePasswordChecks" :key="rule.label" class="password-rule" :class="{ passed: rule.passed }">
            <span>{{ rule.passed ? '✓' : '○' }}</span>
            <span>{{ rule.label }}</span>
          </div>
        </div>

        <FormItem label="邮箱验证码">
          <Space style="width: 100%;">
            <Input v-model="profileState.code" placeholder="修改密码时必填" />
            <Button type="secondary" :loading="isSendingProfileCode" :disabled="sendCodeDisabled" @click="handleSendProfileCode">
              {{ resendRemaining > 0 ? `${resendRemaining}s` : '发送验证码' }}
            </Button>
          </Space>
        </FormItem>

        <Button type="primary" html-type="submit" :loading="isSavingProfile">
          保存个人信息
        </Button>
      </Form>
    </Card>

    <Grid :cols="{xs:1,lg:2}" :col-gap="16" :row-gap="16">
      <GridItem>
        <Card title="通行密钥" size="small">
          <div class="card-body">
            <Text type="secondary">使用通行密钥完成无密码登录，保留密码作为回退方式。</Text>

            <Alert v-if="passkeySupportMessage" type="warning" :show-icon="false">
              {{ passkeySupportMessage }}
            </Alert>

            <div class="passkey-add-row">
              <Input v-model="passkeyState.name" placeholder="当前设备（可选）" @keyup.enter="handleCreatePasskey" />
              <Button
                type="primary"
                data-testid="passkey-enroll-button"
                :loading="isCreatingPasskey"
                :disabled="Boolean(passkeySupportMessage)"
                @click="handleCreatePasskey"
              >
                添加
              </Button>
            </div>

            <Spin :loading="isLoadingPasskeys">
              <div v-if="passkeys.length > 0" class="passkey-list">
                <div v-for="passkey in passkeys" :key="passkey.id" class="passkey-item">
                  <div class="passkey-item-main">
                    <div class="passkey-name">{{ passkey.name }}</div>
                    <div class="passkey-meta">
                      <span>创建于 {{ formatPasskeyTime(passkey.created_at) }}</span>
                      <span>·</span>
                      <span>最近使用 {{ formatPasskeyTime(passkey.last_used_at) }}</span>
                    </div>
                  </div>
                  <Popconfirm @ok="handleDeletePasskey(passkey.id)">
                    <Button
                      type="secondary"
                      status="danger"
                      size="small"
                      data-testid="passkey-delete-button"
                      :loading="deletingPasskeyId === passkey.id"
                    >
                      删除
                    </Button>
                    <template #content>
                      删除后需要重新绑定该通行密钥。
                    </template>
                  </Popconfirm>
                </div>
              </div>
              <Text v-else type="secondary">当前还没有绑定通行密钥。</Text>
            </Spin>
          </div>
        </Card>
      </GridItem>

      <GridItem>
      <div class="card-stack">
        <Card title="两步验证（2FA）" size="small">
          <div class="card-body">
            <Text type="secondary">启用后登录需额外输入验证器 App 生成的 6 位动态码，增强账号安全性。</Text>

            <Spin :loading="isLoadingTOTPStatus">
              <template v-if="!totpEnabled && !showTotpSetup">
                <Button type="primary" data-testid="totp-enable-button" :loading="isEnablingTOTP" @click="handleEnableTOTP">
                  启用两步验证
                </Button>
              </template>

              <template v-if="showTotpSetup && totpSetupData">
                <div class="totp-qr-section">
                  <img :src="totpSetupData.qr_data_uri" alt="TOTP QR Code" />
                  <div class="totp-secret-row">
                    <Text type="secondary">无法扫描？</Text>
                    <Text code class="totp-secret">{{ totpSetupData.secret }}</Text>
                    <Button size="mini" type="text" @click="copyText(totpSetupData.secret)">
                      复制
                    </Button>
                  </div>
                </div>

                <div class="totp-verify-row">
                  <Input
                    v-model="totpSetupCode"
                    placeholder="输入 6 位动态码"
                    max-length="6"
                  />
                  <Button type="primary" :loading="isEnablingTOTP" @click="handleVerifyTOTPSetup">
                    验证
                  </Button>
                  <Button @click="handleCancelTOTPSetup">
                    取消
                  </Button>
                </div>
              </template>

              <template v-if="totpEnabled && !showTotpSetup">
                <div class="totp-status-row">
                  <div class="totp-status-badge">
                    <span class="totp-status-dot" />
                    <Text>两步验证已启用</Text>
                  </div>
                  <Popconfirm @ok="handleDisableTOTP">
                    <Button type="secondary" status="danger" data-testid="totp-disable-button" :loading="isDisablingTOTP">
                      关闭两步验证
                    </Button>
                    <template #content>
                      关闭后需重新启用才能使用两步验证。
                    </template>
                  </Popconfirm>
                </div>
              </template>
            </Spin>
          </div>
        </Card>
      
      
        <Card title="第三方账号绑定" size="small">
          <div class="card-body">
            <Text type="secondary">绑定后可用第三方账号快速登录。</Text>
            <Spin :loading="isLoadingBindings">
              <div class="oauth-bind-list">
                <div v-for="p in oauthProviders" :key="p.id" class="oauth-bind-item">
                  <span class="oauth-bind-icon" v-html="p.icon"></span>
                  <span class="oauth-bind-label">{{ p.label }}</span>
                  <template v-if="isBound(p.id)">
                    <Text type="success" class="oauth-bind-status">已绑定 {{ getBinding(p.id)?.name || '' }}</Text>
                    <Popconfirm @ok="handleUnlinkOAuth(getBinding(p.id)!.id)">
                      <Button size="mini" type="text" status="danger" :loading="unlinkingId === getBinding(p.id)?.id">
                        解绑
                      </Button>
                      <template #content>
                        解绑后该第三方账号将无法使用此方式登录。
                      </template>
                    </Popconfirm>
                  </template>
                  <template v-else>
                    <Text type="secondary" class="oauth-bind-status">未绑定</Text>
                    <a :href="getOAuthBindUrl(p.id)" class="oauth-bind-btn">绑定</a>
                  </template>
                </div>
              </div>
            </Spin>
          </div>
        </Card>
      </div>
      </GridItem>
    </Grid>
  </section>
</template>

<style scoped>
.card-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.card-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ── 通行密钥 ── */

.passkey-list {
  display: grid;
  gap: 8px;
}

.passkey-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md);
  background: rgba(0, 0, 0, 0.02);
}

.passkey-item:last-child {
  margin-bottom: 0;
}

.passkey-item-main {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.passkey-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-ink);
  overflow-wrap: anywhere;
}

.passkey-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--color-charcoal);
}

.passkey-add-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ── 密码规则 ── */

.password-rule-list {
  display: grid;
  gap: 6px;
  margin: 0 0 16px;
  padding: 10px 12px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md);
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

/* ── 两步验证 ── */

.totp-qr-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.totp-qr-section img {
  width: 160px;
  height: 160px;
  border-radius: 8px;
  border: 1px solid var(--color-hairline-strong);
}

.totp-secret-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.totp-secret {
  font-size: 13px;
  letter-spacing: 1px;
  user-select: all;
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

.totp-status-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.totp-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-accent-green);
  flex-shrink: 0;
}

.oauth-bind-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.oauth-bind-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md);
  background: rgba(0, 0, 0, 0.02);
}

.oauth-bind-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.oauth-bind-icon :deep(svg) {
  display: block;
}

.oauth-bind-label {
  font-weight: 500;
  font-size: 14px;
  color: var(--color-ink);
  min-width: 60px;
}

.oauth-bind-status {
  flex: 1;
  font-size: 12px;
}

.oauth-bind-btn {
  font-size: 12px;
  color: var(--color-accent-blue);
  text-decoration: none;
  cursor: pointer;
}

.oauth-bind-btn:hover {
  text-decoration: underline;
}

@media (max-width: 720px) {
  .passkey-item {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
