<script setup lang="ts">
import {
  NAlert,
  NButton,
  NCard,
  NEmpty,
  NSpace,
  NTable,
  NTag,
  useMessage,
} from 'naive-ui'
import { onMounted, shallowRef } from 'vue'

import { ApiError } from '@/api/http'
import { fetchClients } from '@/api/admin'
import ClientFormDrawer from '@/components/admin/ClientFormDrawer.vue'
import type { OAuthClientRecord } from '@/types/api'

const message = useMessage()

const clients = shallowRef<OAuthClientRecord[]>([])
const drawerVisible = shallowRef(false)
const drawerSaving = shallowRef(false)
const activeClient = shallowRef<OAuthClientRecord | null>(null)
const latestSecret = shallowRef('')

onMounted(loadClients)

async function loadClients() {
  try {
    clients.value = await fetchClients()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '加载客户端失败')
  }
}

function openCreateDrawer() {
  activeClient.value = null
  latestSecret.value = ''
  drawerVisible.value = true
}

function openEditDrawer(client: OAuthClientRecord) {
  activeClient.value = client
  latestSecret.value = ''
  drawerVisible.value = true
}

async function handleSaved(client: OAuthClientRecord) {
  latestSecret.value = client.client_secret ?? ''
  await loadClients()
}
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <h1 class="page-title">客户端管理</h1>
        <p class="page-subtitle">维护应用、回调地址与 scope。</p>
      </div>
      <NButton type="primary" @click="openCreateDrawer">新建客户端</NButton>
    </header>

    <NAlert v-if="latestSecret" type="warning" title="仅展示一次的 Client Secret" class="secret-alert">
      <span class="mono">{{ latestSecret }}</span>
    </NAlert>

    <NCard>
      <NEmpty v-if="clients.length === 0" description="暂无客户端。" />

      <div v-else class="table-scroll">
        <NTable striped>
          <thead>
            <tr>
              <th>名称</th>
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
                <div class="cell-title">{{ client.name }}</div>
                <div class="cell-description">{{ client.description || '暂无描述' }}</div>
              </td>
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
                <NButton size="small" tertiary @click="openEditDrawer(client)">编辑</NButton>
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
      @saved="handleSaved"
    />
  </section>
</template>

<style scoped>
.secret-alert {
  margin-bottom: 18px;
}

.table-scroll {
  overflow-x: auto;
}

.cell-title {
  font-weight: 600;
}

.cell-description {
  margin-top: 6px;
  color: rgba(255, 255, 255, 0.55);
}

.uri-line + .uri-line {
  margin-top: 6px;
}
</style>
