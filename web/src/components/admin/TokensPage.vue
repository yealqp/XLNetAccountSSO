<script setup lang="ts">
import { useHead } from '@unhead/vue'
import {
  NAlert,
  NButton,
  NCard,
  NEmpty,
  NGrid,
  NGridItem,
  NPopconfirm,
  NSpace,
  NTable,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRoute } from 'vue-router'

useHead({ title: '令牌管理 — XLNetAccount' })

import {
  fetchTokens,
  fetchManagedTokens,
  revokeAccessToken,
  revokeClientTokens,
  revokeRefreshToken,
} from '@/api/admin'
import { ApiError } from '@/api/http'
import StatPanel from '@/components/admin/StatPanel.vue'
import type { TokenRecord } from '@/types/api'

interface TokenGroup {
  clientId: string
  clientName: string
  activeCount: number
  items: TokenRecord[]
}

const message = useMessage()
const route = useRoute()
const tokens = shallowRef<TokenRecord[]>([])
const pendingActionKey = shallowRef('')
const loadError = shallowRef('')
const manageAll = computed(() => route.meta.manageScope === 'all')

const stats = computed(() => ({
  active: tokens.value.filter(token => token.status === 'active').length,
  refresh: tokens.value.filter(token => token.token_kind === 'refresh').length,
  revoked: tokens.value.filter(token => token.status === 'revoked').length,
  expired: tokens.value.filter(token => token.status === 'expired').length,
}))

const groups = computed<TokenGroup[]>(() => {
  const byClient = new Map<string, TokenGroup>()

  for (const token of tokens.value) {
    const existing = byClient.get(token.client_id)
    if (existing) {
      existing.items.push(token)
      if (token.status === 'active') {
        existing.activeCount += 1
      }
      continue
    }

    byClient.set(token.client_id, {
      clientId: token.client_id,
      clientName: token.client_name,
      activeCount: token.status === 'active' ? 1 : 0,
      items: [token],
    })
  }

  return [...byClient.values()]
})

onMounted(() => {
  void loadTokens()
})

watch(() => route.fullPath, () => {
  void loadTokens()
})

async function loadTokens() {
  loadError.value = ''

  try {
    tokens.value = manageAll.value ? await fetchManagedTokens() : await fetchTokens()
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '加载令牌失败'
    message.error(loadError.value)
  }
}

async function handleRevokeToken(token: TokenRecord) {
  pendingActionKey.value = `${token.token_kind}:${token.id}`

  try {
    if (token.token_kind === 'access') {
      await revokeAccessToken(token.id)
    }
    else {
      await revokeRefreshToken(token.id)
    }

    message.success(`已吊销${token.token_kind === 'access' ? ' access token' : ' refresh token'}`)
    await loadTokens()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '吊销令牌失败')
  }
  finally {
    pendingActionKey.value = ''
  }
}

async function handleRevokeClient(group: TokenGroup) {
  pendingActionKey.value = `client:${group.clientId}`

  try {
    await revokeClientTokens(group.clientId)
    message.success('该应用的令牌已全部吊销')
    await loadTokens()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '批量吊销失败')
  }
  finally {
    pendingActionKey.value = ''
  }
}

function tokenTagType(status: TokenRecord['status']) {
  switch (status) {
    case 'active':
      return 'success'
    case 'expired':
      return 'warning'
    default:
      return 'error'
  }
}

function tokenKindType(kind: TokenRecord['token_kind']) {
  return kind === 'access' ? 'info' : 'default'
}

function isActionPending(actionKey: string) {
  return pendingActionKey.value === actionKey
}
</script>

<template>
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">{{ manageAll ? '令牌管理' : '令牌' }}</h1>
        <p class="page-subtitle">{{ manageAll ? '查看全部令牌并执行吊销操作。' : '查看并吊销您当前账号的令牌。' }}</p>
      </div>
      <NButton tertiary @click="loadTokens">刷新列表</NButton>
    </header>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <NButton size="small" tertiary @click="loadTokens">重试</NButton>
      </div>
    </NAlert>

    <NGrid cols="1 s:2 l:4" responsive="screen" :x-gap="16" :y-gap="16">
      <NGridItem>
        <StatPanel label="活跃令牌" :value="stats.active" detail="仍可用于访问或续签的有效记录。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="Refresh Token" :value="stats.refresh" detail="可用于续签 access token 的长期凭证。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="已吊销" :value="stats.revoked" detail="被手动撤销，不再允许继续使用。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="已过期" :value="stats.expired" detail="自然过期但仍保留审计记录。" />
      </NGridItem>
    </NGrid>

    <div class="token-stack">
      <NCard v-if="groups.length === 0">
        <NEmpty description="暂无令牌记录。" />
      </NCard>

      <NCard v-for="group in groups" :key="group.clientId">
        <div class="token-group-header">
          <div>
            <h2 class="token-group-title">{{ group.clientName }}</h2>
            <p class="token-group-meta mono">{{ group.clientId }}</p>
          </div>

          <NSpace align="center" :wrap="true">
            <NTag round type="success">{{ group.activeCount }} active</NTag>
            <NPopconfirm @positive-click="handleRevokeClient(group)">
              <template #trigger>
                <NButton
                  tertiary
                  type="error"
                  :disabled="group.activeCount === 0"
                  :loading="isActionPending(`client:${group.clientId}`)"
                >
                  吊销该应用全部令牌
                </NButton>
              </template>
              这会吊销该应用下当前账号的所有 access / refresh token。
            </NPopconfirm>
          </NSpace>
        </div>

        <div class="table-scroll">
          <NTable striped>
            <thead>
              <tr>
                <th v-if="manageAll">所属用户</th>
                <th>类型</th>
                <th>状态</th>
                <th>Scope</th>
                <th>签发时间</th>
                <th>过期时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="token in group.items" :key="token.id">
                <td v-if="manageAll">{{ token.owner_username || '-' }}</td>
                <td>
                  <NTag size="small" :type="tokenKindType(token.token_kind)">
                    {{ token.token_kind }}
                  </NTag>
                </td>
                <td>
                  <NTag size="small" :type="tokenTagType(token.status)">
                    {{ token.status }}
                  </NTag>
                </td>
                <td>
                  <span class="mono">{{ token.scope || '-' }}</span>
                </td>
                <td>{{ token.created_at }}</td>
                <td>{{ token.expires_at }}</td>
                <td>
                  <NPopconfirm @positive-click="handleRevokeToken(token)">
                    <template #trigger>
                      <NButton
                        size="small"
                        tertiary
                        type="error"
                        :disabled="token.status !== 'active'"
                        :loading="isActionPending(`${token.token_kind}:${token.id}`)"
                      >
                        吊销
                      </NButton>
                    </template>
                    吊销后该令牌将立即失效。
                  </NPopconfirm>
                </td>
              </tr>
            </tbody>
          </NTable>
        </div>
      </NCard>
    </div>
  </section>
</template>

<style scoped>
.token-stack {
  display: grid;
  gap: 16px;
}

.token-group-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.token-group-title {
  margin: 0;
  font-size: 20px;
}

.token-group-meta {
  margin: 8px 0 0;
  color: var(--color-charcoal);
}

.table-scroll {
  overflow-x: auto;
}

.page-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

@media (max-width: 820px) {
  .token-group-header {
    flex-direction: column;
  }

  .page-alert {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
