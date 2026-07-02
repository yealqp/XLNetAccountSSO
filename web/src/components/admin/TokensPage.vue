<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Button, Card, Empty, Grid, GridItem, Message, Popconfirm, Space, Table, TableColumn, Tag } from '@arco-design/web-vue'
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
    Message.error(loadError.value)
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

    Message.success(`已吊销${token.token_kind === 'access' ? ' access token' : ' refresh token'}`)
    await loadTokens()
  }
  catch (error) {
    Message.error(error instanceof ApiError ? error.message : '吊销令牌失败')
  }
  finally {
    pendingActionKey.value = ''
  }
}

async function handleRevokeClient(group: TokenGroup) {
  pendingActionKey.value = `client:${group.clientId}`

  try {
    await revokeClientTokens(group.clientId)
    Message.success('该应用的令牌已全部吊销')
    await loadTokens()
  }
  catch (error) {
    Message.error(error instanceof ApiError ? error.message : '批量吊销失败')
  }
  finally {
    pendingActionKey.value = ''
  }
}

function tokenTagType(status: TokenRecord['status']) {
  switch (status) {
    case 'active':
      return 'green'
    case 'expired':
      return 'orange'
    case 'revoked':
      return 'red'
    default:
      return 'gray'
  }
}

function tokenKindType(kind: TokenRecord['token_kind']) {
  return kind === 'access' ? 'blue' : 'purple'
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
      <Button type="text" @click="loadTokens">刷新列表</Button>
    </header>

    <Alert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <Button size="small" type="text" @click="loadTokens">重试</Button>
      </div>
    </Alert>

    <Grid :cols="{xs:1,sm:2,lg:4}" :col-gap="16" :row-gap="16">
      <GridItem>
        <StatPanel label="活跃令牌" :value="stats.active" detail="仍可用于访问或续签的有效记录。" />
      </GridItem>
      <GridItem>
        <StatPanel label="Refresh Token" :value="stats.refresh" detail="可用于续签 access token 的长期凭证。" />
      </GridItem>
      <GridItem>
        <StatPanel label="已吊销" :value="stats.revoked" detail="被手动撤销，不再允许继续使用。" />
      </GridItem>
      <GridItem>
        <StatPanel label="已过期" :value="stats.expired" detail="自然过期但仍保留审计记录。" />
      </GridItem>
    </Grid>

    <div class="token-stack">
      <Card v-if="groups.length === 0">
        <Empty description="暂无令牌记录。" />
      </Card>

      <Card v-for="group in groups" :key="group.clientId">
        <div class="token-group-header">
          <div>
            <h2 class="token-group-title">{{ group.clientName }}</h2>
            <p class="token-group-meta mono">{{ group.clientId }}</p>
          </div>

          <Space align="center" wrap>
            <Tag round color="green">{{ group.activeCount }} active</Tag>
            <Popconfirm @ok="handleRevokeClient(group)">
              <Button
                type="text"
                status="danger"
                :disabled="group.activeCount === 0"
                :loading="isActionPending(`client:${group.clientId}`)"
              >
                吊销该应用全部令牌
              </Button>
              <template #content>
                这会吊销该应用下当前账号的所有 access / refresh token。
              </template>
            </Popconfirm>
          </Space>
        </div>

          <Table :data="group.items" stripe row-key="id">
            <template #columns>
              <TableColumn v-if="manageAll" data-index="owner_username" title="所属用户" />
              <TableColumn data-index="token_kind" title="类型">
                <template #cell="{ record }">
                  <Tag size="small" :color="tokenKindType(record.token_kind)">{{ record.token_kind === 'access' ? 'Access' : 'Refresh' }}</Tag>
                </template>
              </TableColumn>
              <TableColumn data-index="status" title="状态">
                <template #cell="{ record }">
                  <Tag size="small" :color="tokenTagType(record.status)">{{ record.status === 'active' ? '活跃' : record.status === 'expired' ? '过期' : record.status === 'revoked' ? '已吊销' : record.status }}</Tag>
                </template>
              </TableColumn>
              <TableColumn data-index="scope" title="Scope">
                <template #cell="{ record }">
                  <span class="mono">{{ record.scope || '-' }}</span>
                </template>
              </TableColumn>
              <TableColumn data-index="created_at" title="签发时间" />
              <TableColumn data-index="expires_at" title="过期时间" />
              <TableColumn title="操作">
                <template #cell="{ record }">
                  <Popconfirm @ok="handleRevokeToken(record)">
                    <Button
                      size="small"
                      type="text"
                      status="danger"
                      :disabled="record.status !== 'active'"
                      :loading="isActionPending(`${record.token_kind}:${record.id}`)"
                    >
                      吊销
                    </Button>
                    <template #content>
                      吊销后该令牌将立即失效。
                    </template>
                  </Popconfirm>
                </template>
              </TableColumn>
            </template>
          </Table>
      </Card>
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
