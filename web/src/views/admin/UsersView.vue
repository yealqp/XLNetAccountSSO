<script setup lang="ts">
import { NButton, NCard, NEmpty, NTable, NTag, useMessage } from 'naive-ui'
import { onMounted, shallowRef } from 'vue'

import { fetchUsers } from '@/api/admin'
import { ApiError } from '@/api/http'
import UserFormDrawer from '@/components/admin/UserFormDrawer.vue'
import type { UserRecord } from '@/types/api'

const message = useMessage()

const users = shallowRef<UserRecord[]>([])
const drawerVisible = shallowRef(false)
const drawerSaving = shallowRef(false)
const activeUser = shallowRef<UserRecord | null>(null)

onMounted(loadUsers)

async function loadUsers() {
  try {
    users.value = await fetchUsers()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '加载用户失败')
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
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <h1 class="page-title">用户管理</h1>
        <p class="page-subtitle">维护账号、角色与状态。</p>
      </div>
      <NButton type="primary" @click="openCreateDrawer">新建用户</NButton>
    </header>

    <NCard>
      <NEmpty v-if="users.length === 0" description="暂无用户。" />

      <div v-else class="table-scroll">
        <NTable striped>
          <thead>
            <tr>
              <th>用户名</th>
              <th>显示名称</th>
              <th>邮箱</th>
              <th>角色</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td class="mono">{{ user.username }}</td>
              <td>{{ user.display_name }}</td>
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
                <NButton size="small" tertiary @click="openEditDrawer(user)">编辑</NButton>
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
.table-scroll {
  overflow-x: auto;
}
</style>
