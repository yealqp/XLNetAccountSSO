<script setup lang="ts">
import { useHead } from '@unhead/vue'
import {
  Alert,
  Avatar,
  Button,
  Card,
  Empty,
  Message,
  Modal,
  Popconfirm,
  Space,
  Table,
  TableColumn,
  Tag,
} from '@arco-design/web-vue'
import { computed, h, onMounted, shallowRef, watch } from 'vue'
import { useRoute } from 'vue-router'

useHead({ title: '应用管理 — XLNetAccount' })

import { ApiError } from '@/api/http'
import { deleteClient, deleteManagedClient, fetchClients, fetchManagedClients } from '@/api/admin'
import ClientFormDrawer from '@/components/admin/ClientFormDrawer.vue'
import { resolveServerUrl } from '@/config/endpoints'
import type { OAuthClientRecord } from '@/types/api'

const route = useRoute()

const clients = shallowRef<OAuthClientRecord[]>([])
const drawerVisible = shallowRef(false)
const drawerSaving = shallowRef(false)
const activeClient = shallowRef<OAuthClientRecord | null>(null)
const loadError = shallowRef('')
const manageAll = computed(() => route.meta.manageScope === 'all')

onMounted(loadClients)

watch(() => route.fullPath, () => {
  void loadClients()
})

async function loadClients() {
  loadError.value = ''

  try {
    clients.value = manageAll.value ? await fetchManagedClients() : await fetchClients()
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '加载客户端失败'
    Message.error(loadError.value)
  }
}

function openCreateDrawer() {
  activeClient.value = null
  drawerVisible.value = true
}

function openEditDrawer(client: OAuthClientRecord) {
  activeClient.value = client
  drawerVisible.value = true
}

async function handleSaved(client: OAuthClientRecord) {
  if (client.client_secret) {
    openSecretDialog(client.client_secret)
  }
  await loadClients()
}

async function handleDelete(client: OAuthClientRecord) {
  try {
    await (manageAll.value ? deleteManagedClient(client.id) : deleteClient(client.id))
    Message.success('客户端已删除')
    await loadClients()
  }
  catch (error) {
    Message.error(error instanceof ApiError ? error.message : '删除客户端失败')
  }
}

function openSecretDialog(secret: string) {
  Modal.warning({
    title: 'Client Secret',
    content: () => h('div', { style: { display: 'grid', gap: '12px' } }, [
      h('span', '仅展示一次，请立即保存。'),
      h('pre', { style: { margin: 0, padding: '12px', border: '1px solid var(--color-hairline)', borderRadius: 'var(--radius-md)', background: 'var(--color-surface-elevated)', color: 'var(--color-ink)', overflow: 'auto' } }, [h('code', secret)]),
    ]),
    okText: '我已保存',
    cancelText: '复制',
    onCancel: () => {
      navigator.clipboard.writeText(secret).then(
        () => Message.success('已复制 Client Secret'),
        () => Message.error('复制失败，请手动复制'),
      )
    },
  })
}
</script>

<template>
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">{{ manageAll ? '应用管理' : '应用' }}</h1>
        <p class="page-subtitle">{{ manageAll ? '查看全部应用并执行管理操作。' : '查看并维护您创建的应用。' }}</p>
      </div>
      <Button type="primary" @click="openCreateDrawer">新建客户端</Button>
    </header>

    <Alert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <Button size="small" type="text" @click="loadClients">重试</Button>
      </div>
    </Alert>

    <Card>
      <Empty v-if="clients.length === 0" description="暂无客户端。" />

      <Table v-else :data="clients" stripe row-key="id">
        <template #columns>
          <TableColumn title="名称">
            <template #cell="{ record }">
              <div class="client-name-cell">
                <Avatar :size="28" :src="resolveServerUrl(record.icon_url) || undefined" class="client-avatar">
                  {{ record.name.charAt(0).toUpperCase() }}
                </Avatar>
                <div>
                  <div class="cell-title">{{ record.name }}</div>
                  <div class="cell-description">{{ record.description || '暂无描述' }}</div>
                </div>
              </div>
            </template>
          </TableColumn>
          <TableColumn v-if="manageAll" data-index="owner_username" title="所属用户" />
          <TableColumn data-index="client_id" title="Client ID">
            <template #cell="{ record }">
              <span class="mono">{{ record.client_id }}</span>
            </template>
          </TableColumn>
          <TableColumn data-index="client_type" title="类型">
            <template #cell="{ record }">
              <Tag :color="record.client_type === 'confidential' ? 'orange' : undefined" size="small">
                {{ record.client_type }}
              </Tag>
            </template>
          </TableColumn>
          <TableColumn title="Scope">
            <template #cell="{ record }">
              <Space>
                <Tag v-for="scope in record.scopes" :key="scope" size="small" round>
                  {{ scope }}
                </Tag>
              </Space>
            </template>
          </TableColumn>
          <TableColumn title="回调地址">
            <template #cell="{ record }">
              <div v-for="uri in record.redirect_uris" :key="uri" class="mono uri-line">
                {{ uri }}
              </div>
            </template>
          </TableColumn>
          <TableColumn title="操作">
            <template #cell="{ record }">
              <Space>
                <Button size="small" type="text" @click="openEditDrawer(record)">编辑</Button>
                <Popconfirm @ok="handleDelete(record)">
                  <Button size="small" type="text" status="danger">删除</Button>
                  <template #content>
                    删除后不可恢复，确认继续？
                  </template>
                </Popconfirm>
              </Space>
            </template>
          </TableColumn>
        </template>
      </Table>
    </Card>

    <ClientFormDrawer
      v-model:show="drawerVisible"
      v-model:saving="drawerSaving"
      :initial-client="activeClient"
      :manage-all="manageAll"
      @saved="handleSaved"
    />
  </section>
</template>

<style scoped>

.client-name-cell {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.client-avatar {
  border-radius: 8px;
  background: rgba(52, 159, 244, 0.18);
  color: var(--color-accent-blue);
  flex-shrink: 0;
}

.page-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.cell-title {
  font-weight: 600;
}

.cell-description {
  margin-top: 6px;
  color: var(--color-charcoal);
}

.uri-line + .uri-line {
  margin-top: 6px;
}

@media (max-width: 820px) {
  .page-alert {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
