<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Button, Card, Descriptions, DescriptionsItem, Message, Space, Tag, TypographyText as Text } from '@arco-design/web-vue'

useHead({ title: '连接信息 — XLNetAccount' })

import { getConnectionInfo } from '@/config/endpoints'

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
    Message.success('已复制地址')
  }
  catch {
    Message.error('复制失败，请手动复制')
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
      <Tag round color="blue">{{ connectionInfo.baseUrl }}</Tag>
    </header>

    <Card title="OAuth2 地址">
      <Descriptions layout="vertical" :column="1" bordered>
        <DescriptionsItem v-for="item in oauthItems" :key="item.label" :label="item.label">
          <Space vertical :size="8">
            <Space align="center" wrap>
              <Tag size="small" :color="item.method === 'GET' ? 'blue' : 'orange'">
                {{ item.method }}
              </Tag>
              <span class="mono endpoint-url">{{ item.url }}</span>
              <Button size="small" type="text" @click="copyUrl(item.url)">
                复制
              </Button>
            </Space>
            <Text type="secondary">{{ item.description }}</Text>
          </Space>
        </DescriptionsItem>
      </Descriptions>
    </Card>

    <Card title="OIDC 地址">
      <Descriptions layout="vertical" :column="1" bordered>
        <DescriptionsItem v-for="item in oidcItems" :key="item.label" :label="item.label">
          <Space vertical :size="8">
            <Space align="center" wrap>
              <Tag size="small" color="green">{{ item.method }}</Tag>
              <span class="mono endpoint-url">{{ item.url }}</span>
              <Button size="small" type="text" @click="copyUrl(item.url)">
                复制
              </Button>
            </Space>
            <Text type="secondary">{{ item.description }}</Text>
          </Space>
        </DescriptionsItem>
      </Descriptions>
    </Card>

    <Card title="userinfo 响应示例">
      <Space vertical :size="10">
        <Text type="secondary">返回字段会根据授权 scope 决定，下面是包含 `profile email roles` 时的示例。</Text>
        <pre class="code-block"><code>{{ userInfoExample }}</code></pre>
      </Space>
    </Card>
  </section>
</template>

<style scoped>
.endpoint-url {
  overflow-wrap: anywhere;
}

.code-block {
  margin: 0;
  padding: 12px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md);
  background: var(--color-surface-elevated);
  color: var(--color-ink);
  overflow: auto;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
</style>
