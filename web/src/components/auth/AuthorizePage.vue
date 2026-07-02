<script setup lang="ts">
import { useHead } from '@unhead/vue'
import {
  Button,
  Card,
  Divider,
  Message,
  Result,
  Space,
  Spin,
  Tag,
  TypographyText as Text,
} from '@arco-design/web-vue'
import { computed, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

useHead({ title: '授权确认 — XLNetAccount' })

import { ApiError } from '@/api/http'
import { decideAuthorization, previewAuthorization } from '@/api/oauth'
import { useViewport } from '@/composables/useViewport'
import { resolveServerUrl } from '@/config/endpoints'
import { useSessionStore } from '@/stores/session'
import type { AuthorizationPreview } from '@/types/api'

const route = useRoute()
const router = useRouter()
const { isMobile } = useViewport()
const sessionStore = useSessionStore()

const preview = shallowRef<AuthorizationPreview | null>(null)
const isLoading = shallowRef(true)
const isSubmitting = shallowRef(false)

const scopeItems = computed(() => preview.value?.client.requested_scope_details ?? [])
const actionsVertical = computed(() => isMobile.value)
const clientInitial = computed(() => (preview.value?.client.name?.trim().charAt(0) || 'C').toUpperCase())
const clientTitle = computed(() => preview.value ? `${preview.value.client.name} 请求使用您的信息` : '授权确认')
const clientSubtitle = computed(() => preview.value?.client.description?.trim() || '该应用希望获取以下信息。')
const hasClientIcon = computed(() => Boolean(preview.value?.client.icon_url?.trim()))
const resolvedClientIconUrl = computed(() => resolveServerUrl(preview.value?.client.icon_url))
const consentHint = computed(() => {
  if (!preview.value) {
    return ''
  }

  return `点击同意后 ${preview.value.client.name} 将获得以上权限，您将被重定向到 ${preview.value.client.redirect_uri}`
})

watch(preview, (nextPreview) => {
  if (nextPreview?.client.trusted && !isSubmitting.value) {
    void handleDecision(true)
  }
})

void loadPreview()

async function loadPreview() {
  isLoading.value = true

  try {
    if (!sessionStore.authenticated && !(await sessionStore.ensureSession())) {
      await redirectToLogin()
      return
    }

    preview.value = await previewAuthorization(buildAuthParams())
  }
  catch (error) {
    if (isUnauthorizedError(error)) {
      await redirectToLogin()
      return
    }

    Message.error(error instanceof ApiError ? error.message : '加载授权请求失败')
  }
  finally {
    isLoading.value = false
  }
}

async function handleDecision(approved: boolean) {
  isSubmitting.value = true

  try {
    const params = buildAuthParams()
    const response = await decideAuthorization({
      response_type: params.get('response_type') ?? '',
      client_id: params.get('client_id') ?? '',
      redirect_uri: params.get('redirect_uri') ?? '',
      scope: params.get('scope') ?? '',
      state: params.get('state') ?? '',
      nonce: params.get('nonce') ?? '',
      code_challenge: params.get('code_challenge') ?? '',
      code_challenge_method: params.get('code_challenge_method') ?? '',
      approved,
    })

    window.location.href = response.redirect_to
  }
  catch (error) {
    if (isUnauthorizedError(error)) {
      isSubmitting.value = false
      await redirectToLogin()
      return
    }

    Message.error(error instanceof ApiError ? error.message : '提交授权结果失败')
    isSubmitting.value = false
  }
}

function isUnauthorizedError(error: unknown) {
  return error instanceof ApiError && error.status === 401
}

async function redirectToLogin() {
  await router.replace({
    name: 'login',
    query: { next: route.fullPath },
  })
}

function buildAuthParams() {
  const params = new URLSearchParams()
  params.set('response_type', getQueryValue('response_type'))
  params.set('client_id', getQueryValue('client_id'))
  params.set('redirect_uri', getQueryValue('redirect_uri'))
  params.set('scope', getQueryValue('scope'))
  params.set('state', getQueryValue('state'))
  params.set('nonce', getQueryValue('nonce'))
  params.set('code_challenge', getQueryValue('code_challenge'))
  params.set('code_challenge_method', getQueryValue('code_challenge_method'))
  return params
}

function getQueryValue(key: string) {
  const value = route.query[key]
  return Array.isArray(value) ? value[0] ?? '' : value ?? ''
}
</script>

<template>
  <div class="auth-panel-view">
    <Spin :loading="isLoading">
      <div v-if="preview" class="auth-scope-body">
        <div class="client-hero">
          <p class="auth-panel-kicker">授权确认</p>
          <img
            v-if="hasClientIcon"
            :src="resolvedClientIconUrl"
            :alt="preview.client.name"
            class="client-avatar client-avatar-image"
          >
          <div v-else class="client-avatar client-avatar-fallback">
            {{ clientInitial }}
          </div>

          <div class="client-title-block">
            <h2 class="client-title">{{ clientTitle }}</h2>
            <p class="client-subtitle">{{ clientSubtitle }}</p>
          </div>

          <Tag v-if="preview.client.trusted" type="success" round>Trusted</Tag>
        </div>

        <Divider style="margin: 0;" />

        <div class="scope-section">
          <Space align="center" justify="space-between" wrap fill>
            <Text strong>请求的 Scope</Text>
            <Tag round type="info">{{ scopeItems.length }}</Tag>
          </Space>

          <div class="scope-grid">
            <Card v-for="scope in scopeItems" :key="scope.key" size="small" embedded>
              <div class="scope-card-body">
                <Tag size="small" type="info" round>{{ scope.label }}</Tag>
                <Text class="scope-desc-text">{{ scope.description }}</Text>
              </div>
            </Card>
          </div>

          <Text class="scope-hint-text">{{ consentHint }}</Text>
        </div>

        <Result
          v-if="scopeItems.length === 0"
          status="warning"
          title="未请求任何可用 Scope"
          description="请返回客户端检查授权参数。"
        />

        <div class="action-row">
          <Space justify="end" :vertical="actionsVertical" :size="12" fill>
            <Button :disabled="isSubmitting" :block="isMobile" @click="handleDecision(false)">
              拒绝
            </Button>
            <Button type="primary" :block="isMobile" :loading="isSubmitting" @click="handleDecision(true)">
              同意
            </Button>
          </Space>
        </div>
      </div>
    </Spin>
  </div>
</template>

<style scoped>
.auth-panel-view {
  color: var(--color-ink);
}

.auth-panel-kicker {
  margin: 0;
  color: var(--color-charcoal);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

/* ── body ── */

.auth-scope-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  width: 100%;
}

/* ── hero ── */

.client-hero {
  display: flex;
  align-items: center;
  flex-direction: column;
  gap: 12px;
  text-align: center;
}

.client-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 10px;
  background: rgba(52, 159, 244, 0.18);
  color: var(--color-accent-blue);
  flex-shrink: 0;
}

.client-avatar-image {
  object-fit: cover;
  border: 1px solid var(--color-hairline);
}

.client-avatar-fallback {
  font-size: 22px;
  font-weight: 700;
}

.client-title-block {
  display: grid;
  gap: 8px;
}

.client-title {
  margin: 0;
  font-size: clamp(24px, 2.8vw, 30px);
  font-weight: 650;
  letter-spacing: -0.03em;
}

.client-subtitle {
  margin: 0;
  color: var(--color-charcoal);
  font-size: 14px;
  line-height: 1.65;
}

/* ── scope section ── */

.scope-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.scope-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.scope-card-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.scope-desc-text {
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
  overflow-wrap: break-word;
}

.scope-hint-text {
  display: block;
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-charcoal);
  word-break: break-word;
  overflow-wrap: break-word;
}

/* ── actions ── */

.action-row {
  width: 100%;
  min-width: 0;
}

@media (max-width: 720px) {
  .scope-grid {
    grid-template-columns: 1fr;
  }
}
</style>
