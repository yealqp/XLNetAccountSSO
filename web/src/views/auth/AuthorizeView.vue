<script setup lang="ts">
import {
  NAlert,
  NButton,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NSpace,
  NSpin,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'
import { computed, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ApiError } from '@/api/http'
import { decideAuthorization, previewAuthorization } from '@/api/oauth'
import { useViewport } from '@/composables/useViewport'
import type { AuthorizationPreview } from '@/types/api'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const { isMobile } = useViewport()

const preview = shallowRef<AuthorizationPreview | null>(null)
const isLoading = shallowRef(true)
const isSubmitting = shallowRef(false)
const errorMessage = shallowRef('')

const scopeItems = computed(() => preview.value?.client.requested_scope_details ?? [])
const descriptionColumns = computed(() => (isMobile.value ? 1 : 2))

watch(preview, (nextPreview) => {
  if (nextPreview?.client.trusted && !isSubmitting.value) {
    void handleDecision(true)
  }
})

void loadPreview()

async function loadPreview() {
  isLoading.value = true
  errorMessage.value = ''

  try {
    preview.value = await previewAuthorization(buildAuthParams())
  }
  catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      await router.replace({
        name: 'login',
        query: { next: route.fullPath },
      })
      return
    }

    errorMessage.value = error instanceof ApiError ? error.message : '加载授权请求失败'
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
    message.error(error instanceof ApiError ? error.message : '提交授权结果失败')
    isSubmitting.value = false
  }
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
  <NCard title="授权确认" class="authorize-card">
    <NSpin :show="isLoading">
      <NAlert v-if="errorMessage" type="error" :show-icon="false">
        {{ errorMessage }}
      </NAlert>

      <template v-else-if="preview">
        <NSpace vertical :size="16">
          <div>
            <NSpace align="center" :wrap="true">
              <NText strong>{{ preview.client.name }}</NText>
              <NTag v-if="preview.client.trusted" type="success" round>Trusted</NTag>
            </NSpace>
            <NText v-if="preview.client.description" depth="3">{{ preview.client.description }}</NText>
          </div>

          <NDescriptions bordered label-placement="top" :column="descriptionColumns">
            <NDescriptionsItem label="客户端 ID">
              <span class="mono">{{ preview.client.client_id }}</span>
            </NDescriptionsItem>
            <NDescriptionsItem label="客户端类型">
              {{ preview.client.client_type }}
            </NDescriptionsItem>
            <NDescriptionsItem label="回调地址">
              <span class="mono">{{ preview.client.redirect_uri }}</span>
            </NDescriptionsItem>
            <NDescriptionsItem label="当前用户">
              {{ preview.user.display_name }}
            </NDescriptionsItem>
          </NDescriptions>

          <NDivider style="margin: 0;" />

          <div>
            <NSpace align="center" justify="space-between" :wrap="true">
              <NText strong>Scope</NText>
              <NTag round type="info">{{ scopeItems.length }}</NTag>
            </NSpace>

            <div class="scope-grid">
              <NCard v-for="scope in scopeItems" :key="scope.key" size="small" embedded>
                <NSpace vertical :size="6">
                  <NSpace align="center" :wrap="true">
                    <NTag size="small" type="info" round>{{ scope.label }}</NTag>
                    <span class="mono">{{ scope.key }}</span>
                  </NSpace>
                  <NText depth="3">{{ scope.description }}</NText>
                </NSpace>
              </NCard>
            </div>
          </div>

          <NAlert v-if="preview.client.trusted" type="success" :show-icon="false">
            Trusted client，将自动通过。
          </NAlert>

          <NSpace justify="end">
            <NButton :disabled="isSubmitting" @click="handleDecision(false)">
              拒绝
            </NButton>
            <NButton type="primary" :loading="isSubmitting" @click="handleDecision(true)">
              同意
            </NButton>
          </NSpace>
        </NSpace>
      </template>
    </NSpin>
  </NCard>
</template>

<style scoped>
.authorize-card {
  max-width: 720px;
  margin: 0 auto;
}

.scope-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 12px;
}

@media (max-width: 720px) {
  .scope-grid {
    grid-template-columns: 1fr;
  }
}
</style>
