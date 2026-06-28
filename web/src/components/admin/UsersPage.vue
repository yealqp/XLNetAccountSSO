<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { NAlert, NButton, NCard, NEmpty, NPopconfirm, NSpace, NTable, NTag, useMessage } from 'naive-ui'
import { onMounted, shallowRef } from 'vue'

useHead({ title: '用户管理 — XLNetAccount' })

import { deleteUser, fetchUsers } from '@/api/admin'
import { ApiError } from '@/api/http'
import UserFormDrawer from '@/components/admin/UserFormDrawer.vue'
import { useSessionStore } from '@/stores/session'
import type { UserRecord } from '@/types/api'

const message = useMessage()
const sessionStore = useSessionStore()

const users = shallowRef<UserRecord[]>([])
const drawerVisible = shallowRef(false)
const drawerSaving = shallowRef(false)
const activeUser = shallowRef<UserRecord | null>(null)
const loadError = shallowRef('')

onMounted(loadUsers)

async function loadUsers() {
  loadError.value = ''

  try {
    users.value = await fetchUsers()
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '加载用户失败'
    message.error(loadError.value)
  }
}

function openCreateDrawer() {
  activeUser.value = null
  drawerVisible.value = true
}

function openEditDrawer(user: UserRecord) {
  activeUser.value = user
  drawerVisible.value = true
}

async function handleSaved() {
  await loadUsers()
}

async function handleDelete(user: UserRecord) {
  try {
    await deleteUser(user.id)
    message.success('用户已删除')
    await loadUsers()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '删除用户失败')
  }
}

function isCurrentUser(user: UserRecord) {
  return sessionStore.user?.id === user.id
}
</script>

<template>
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">用户管理</h1>
        <p class="page-subtitle">维护账号、角色与状态。</p>
      </div>
      <NButton type="primary" @click="openCreateDrawer">新建用户</NButton>
    </header>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <NButton size="small" tertiary @click="loadUsers">重试</NButton>
      </div>
    </NAlert>

    <NCard>
      <NEmpty v-if="users.length === 0" description="暂无用户。" />

      <div v-else class="table-scroll">
        <NTable striped>
          <thead>
            <tr>
              <th>用户名</th>
              <th>邮箱</th>
              <th>角色</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td class="mono">{{ user.username }}</td>
              <td>{{ user.email || '-' }}</td>
              <td>
                <NTag size="small" :type="user.role === 'admin' ? 'warning' : 'default'">
                  {{ user.role }}
                </NTag>
              </td>
              <td>
                <NTag size="small" :type="user.status === 'active' ? 'success' : 'error'">
                  {{ user.status }}
                </NTag>
              </td>
              <td>
                <NSpace>
                  <NButton size="small" tertiary @click="openEditDrawer(user)">编辑</NButton>
                  <NPopconfirm @positive-click="handleDelete(user)">
                    <template #trigger>
                      <NButton size="small" tertiary type="error" :disabled="isCurrentUser(user)">删除</NButton>
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

    <UserFormDrawer
      v-model:show="drawerVisible"
      v-model:saving="drawerSaving"
      :initial-user="activeUser"
      @saved="handleSaved"
    />
  </section>
</template>

<style scoped>
.page-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

@media (max-width: 820px) {
  .page-alert {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
