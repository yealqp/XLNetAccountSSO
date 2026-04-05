<script setup lang="ts">
import { NButton, NCard, NEmpty, NTable, NTag, useMessage } from 'naive-ui'
import { onMounted, shallowRef } from 'vue'

import { fetchSessions, revokeSession } from '@/api/admin'
import { ApiError } from '@/api/http'
import type { SessionRecord } from '@/types/api'

const message = useMessage()
const sessions = shallowRef<SessionRecord[]>([])

onMounted(loadSessions)

async function loadSessions() {
  try {
    sessions.value = await fetchSessions()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '加载会话失败')
  }
}

async function handleRevoke(session: SessionRecord) {
  try {
    await revokeSession(session.id)
    message.success('会话已撤销')
    await loadSessions()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '撤销会话失败')
  }
}
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <h1 class="page-title">会话管理</h1>
        <p class="page-subtitle">查看并下线会话。</p>
      </div>
    </header>

    <NCard>
      <NEmpty v-if="sessions.length === 0" description="暂无会话记录。" />

      <div v-else class="table-scroll">
        <NTable striped>
          <thead>
            <tr>
              <th>状态</th>
              <th>IP</th>
              <th>User-Agent</th>
              <th>最后活跃</th>
              <th>过期时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="session in sessions" :key="session.id">
              <td>
                <NTag v-if="session.current" type="success" size="small">当前设备</NTag>
                <NTag v-else-if="session.revoked_at" type="error" size="small">已撤销</NTag>
                <NTag v-else type="info" size="small">活跃</NTag>
              </td>
              <td class="mono">{{ session.ip_address || '-' }}</td>
              <td>{{ session.user_agent || '-' }}</td>
              <td>{{ session.last_seen_at }}</td>
              <td>{{ session.expires_at }}</td>
              <td>
                <NButton
                  size="small"
                  tertiary
                  type="error"
                  :disabled="session.current || Boolean(session.revoked_at)"
                  @click="handleRevoke(session)"
                >
                  下线
                </NButton>
              </td>
            </tr>
          </tbody>
        </NTable>
      </div>
    </NCard>
  </section>
</template>

<style scoped>
.table-scroll {
  overflow-x: auto;
}
</style>
