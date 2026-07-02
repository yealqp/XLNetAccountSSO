<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Button, Card, Message, Popconfirm, Space, Table, TableColumn, Tag } from '@arco-design/web-vue'
import { onMounted, shallowRef } from 'vue'

useHead({ title: '用户管理 — XLNetAccount' })

import { deleteUser, fetchUsers } from '@/api/admin'
import { ApiError } from '@/api/http'
import UserFormDrawer from '@/components/admin/UserFormDrawer.vue'
import { useSessionStore } from '@/stores/session'
import type { UserRecord } from '@/types/api'

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
    Message.error(loadError.value)
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
    Message.success('用户已删除')
    await loadUsers()
  }
  catch (error) {
    Message.error(error instanceof ApiError ? error.message : '删除用户失败')
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
      <Button type="primary" @click="openCreateDrawer">新建用户</Button>
    </header>

    <Alert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <Button size="small" type="text" @click="loadUsers">重试</Button>
      </div>
    </Alert>

    <Card>
      <Table :data="users" stripe row-key="id">
        <template #columns>
          <TableColumn data-index="id" title="ID" :width="80">
            <template #cell="{ record }">
              <span class="mono">{{ record.id }}</span>
            </template>
          </TableColumn>
          <TableColumn data-index="username" title="用户名">
            <template #cell="{ record }">
              <span class="mono">{{ record.username }}</span>
            </template>
          </TableColumn>
          <TableColumn data-index="email" title="邮箱">
            <template #cell="{ record }">
              {{ record.email || '-' }}
            </template>
          </TableColumn>
          <TableColumn data-index="role" title="角色">
            <template #cell="{ record }">
              <Tag size="small" :color="record.role === 'admin' ? 'orange' : undefined">{{ record.role }}</Tag>
            </template>
          </TableColumn>
          <TableColumn data-index="status" title="状态">
            <template #cell="{ record }">
              <Tag size="small" :color="record.status === 'active' ? 'green' : 'red'">{{ record.status }}</Tag>
            </template>
          </TableColumn>
          <TableColumn title="操作">
            <template #cell="{ record }">
              <Space>
                <Button size="small" type="text" @click="openEditDrawer(record)">编辑</Button>
                <Popconfirm @ok="handleDelete(record)">
                  <Button size="small" type="text" status="danger" :disabled="isCurrentUser(record)">删除</Button>
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
