<script setup lang="ts">
import { useHead } from '@unhead/vue'
import {
  NAlert,
  NAvatar,
  NButton,
  NCard,
  NCode,
  NEmpty,
  NPopconfirm,
  NSpace,
  NTable,
  NTag,
  useDialog,
  useMessage,
} from 'naive-ui'
import { computed, h, onMounted, shallowRef, watch } from 'vue'
import { useRoute } from 'vue-router'

useHead({ title: '应用管理 — XLNetAccount' })

import { ApiError } from '@/api/http'
import { deleteClient, deleteManagedClient, fetchClients, fetchManagedClients } from '@/api/admin'
import ClientFormDrawer from '@/components/admin/ClientFormDrawer.vue'
import { resolveServerUrl } from '@/config/endpoints'
import type { OAuthClientRecord } from '@/types/api'

const message = useMessage()
const dialog = useDialog()
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
    message.error(loadError.value)
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
    message.success('客户端已删除')
    await loadClients()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '删除客户端失败')
  }
}

function openSecretDialog(secret: string) {
  dialog.warning({
    title: 'Client Secret',
    content: () => h('div', { style: { display: 'grid', gap: '12px' } }, [
      h('span', '仅展示一次，请立即保存。'),
      h(NCode, {
        code: secret,
        language: 'text',
        wordWrap: true,
      }),
    ]),
    positiveText: '我已保存',
    negativeText: '复制',
    onNegativeClick: async () => {
      try {
        await navigator.clipboard.writeText(secret)
        message.success('已复制 Client Secret')
      }
      catch {
        message.error('复制失败，请手动复制')
      }
      return false
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
      <NButton type="primary" @click="openCreateDrawer">新建客户端</NButton>
    </header>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <NButton size="small" tertiary @click="loadClients">重试</NButton>
      </div>
    </NAlert>

    <NCard>
      <NEmpty v-if="clients.length === 0" description="暂无客户端。" />

      <div v-else class="table-scroll">
        <NTable striped>
          <thead>
            <tr>
              <th>名称</th>
              <th v-if="manageAll">所属用户</th>
              <th>Client ID</th>
              <th>类型</th>
              <th>Scope</th>
              <th>回调地址</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="client in clients" :key="client.id">
              <td>
                <div class="client-name-cell">
                  <NAvatar :size="28" :src="resolveServerUrl(client.icon_url) || undefined" :round="false" class="client-avatar">
                    {{ client.name.charAt(0).toUpperCase() }}
                  </NAvatar>
                  <div>
                    <div class="cell-title">{{ client.name }}</div>
                    <div class="cell-description">{{ client.description || '暂无描述' }}</div>
                  </div>
                </div>
              </td>
              <td v-if="manageAll">{{ client.owner_username || '-' }}</td>
              <td class="mono">{{ client.client_id }}</td>
              <td>
                <NTag :type="client.client_type === 'confidential' ? 'warning' : 'info'" size="small">
                  {{ client.client_type }}
                </NTag>
              </td>
              <td>
                <NSpace>
                  <NTag v-for="scope in client.scopes" :key="scope" size="small" round>
                    {{ scope }}
                  </NTag>
                </NSpace>
              </td>
              <td>
                <div v-for="uri in client.redirect_uris" :key="uri" class="mono uri-line">
                  {{ uri }}
                </div>
              </td>
              <td>
                <NSpace>
                  <NButton size="small" tertiary @click="openEditDrawer(client)">编辑</NButton>
                  <NPopconfirm @positive-click="handleDelete(client)">
                    <template #trigger>
                      <NButton size="small" tertiary type="error">删除</NButton>
                    </template>
                    删除后不可恢复，确认继续？
                  </NPopconfirm>
                </NSpace>
              </td>
            </tr>
          </tbody>
        </NTable>
      </div>
    </NCard>

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
