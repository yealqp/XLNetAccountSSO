<script setup lang="ts">
import { useHead } from '@unhead/vue'
import {
  NButton,
  NCard,
  NCode,
  NDescriptions,
  NDescriptionsItem,
  NSpace,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'

useHead({ title: '连接信息 — XLNetAccount' })

import { getConnectionInfo } from '@/config/endpoints'

const message = useMessage()
const connectionInfo = getConnectionInfo()

const oauthItems = [
  {
    label: '授权地址',
    method: 'GET',
    url: connectionInfo.authorizeUrl,
    description: '浏览器跳转到该地址发起授权码流程。',
  },
  {
    label: '令牌地址',
    method: 'POST',
    url: connectionInfo.tokenUrl,
    description: '使用授权码或刷新令牌换取令牌。',
  },
  {
    label: '用户信息地址',
    method: 'GET',
    url: connectionInfo.userInfoUrl,
    description: '携带 Bearer Token 获取用户信息。',
  },
  {
    label: '令牌吊销地址',
    method: 'POST',
    url: connectionInfo.revokeUrl,
    description: '撤销 access token 或 refresh token。',
  },
  {
    label: '令牌校验地址',
    method: 'POST',
    url: connectionInfo.introspectUrl,
    description: '资源服务可用于校验 opaque token。',
  },
]

const oidcItems = [
  {
    label: 'Issuer',
    method: 'GET',
    url: connectionInfo.issuer,
    description: 'OIDC Issuer 标识。',
  },
  {
    label: 'OpenID Configuration',
    method: 'GET',
    url: connectionInfo.openidConfigurationUrl,
    description: 'OIDC 元数据发现地址。',
  },
  {
    label: 'JWKS',
    method: 'GET',
    url: connectionInfo.jwksUrl,
    description: '公钥集合地址，用于校验 ID Token。',
  },
]

const userInfoExample = `{
  "sub": "user_123456",
  "preferred_username": "account_user",
  "name": "Account User",
  "email": "user@example.com",
  "roles": ["admin"]
}`

async function copyUrl(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    message.success('已复制地址')
  }
  catch {
    message.error('复制失败，请手动复制')
  }
}
</script>

<template>
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">连接信息</h1>
        <p class="page-subtitle">OAuth2 / OIDC 固定地址与用户信息响应示例。</p>
      </div>
      <NTag round type="info">{{ connectionInfo.baseUrl }}</NTag>
    </header>

    <NCard title="OAuth2 地址">
      <NDescriptions label-placement="top" :column="1" bordered>
        <NDescriptionsItem v-for="item in oauthItems" :key="item.label" :label="item.label">
          <NSpace vertical :size="8">
            <NSpace align="center" :wrap="true">
              <NTag size="small" :type="item.method === 'GET' ? 'info' : 'warning'">
                {{ item.method }}
              </NTag>
              <span class="mono endpoint-url">{{ item.url }}</span>
              <NButton size="small" tertiary @click="copyUrl(item.url)">
                复制
              </NButton>
            </NSpace>
            <NText depth="3">{{ item.description }}</NText>
          </NSpace>
        </NDescriptionsItem>
      </NDescriptions>
    </NCard>

    <NCard title="OIDC 地址">
      <NDescriptions label-placement="top" :column="1" bordered>
        <NDescriptionsItem v-for="item in oidcItems" :key="item.label" :label="item.label">
          <NSpace vertical :size="8">
            <NSpace align="center" :wrap="true">
              <NTag size="small" type="success">{{ item.method }}</NTag>
              <span class="mono endpoint-url">{{ item.url }}</span>
              <NButton size="small" tertiary @click="copyUrl(item.url)">
                复制
              </NButton>
            </NSpace>
            <NText depth="3">{{ item.description }}</NText>
          </NSpace>
        </NDescriptionsItem>
      </NDescriptions>
    </NCard>

    <NCard title="userinfo 响应示例">
      <NSpace vertical :size="10">
        <NText depth="3">返回字段会根据授权 scope 决定，下面是包含 `profile email roles` 时的示例。</NText>
        <NCode :code="userInfoExample" language="json" word-wrap />
      </NSpace>
    </NCard>
  </section>
</template>

<style scoped>
.endpoint-url {
  overflow-wrap: anywhere;
}
</style>
