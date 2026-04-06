<script setup lang="ts">
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NSpace,
  useMessage,
} from 'naive-ui'
import { computed, reactive, watch } from 'vue'

import { createUser, updateUser } from '@/api/admin'
import { useViewport } from '@/composables/useViewport'
import type { UserRecord } from '@/types/api'

interface Props {
  initialUser?: UserRecord | null
}

const props = defineProps<Props>()
const emit = defineEmits<{ saved: [user: UserRecord] }>()
const show = defineModel<boolean>('show', { required: true })
const isSaving = defineModel<boolean>('saving', { default: false })
const message = useMessage()
const { width } = useViewport()
const drawerWidth = computed(() => Math.min(460, Math.max(280, width.value - 16)))

const formState = reactive({
  username: '',
  password: '',
  email: '',
  role: 'user',
  status: 'active',
})

const isEditing = computed(() => Boolean(props.initialUser?.id))

watch(
  () => [show.value, props.initialUser] as const,
  () => {
    if (!show.value) {
      return
    }

    formState.username = props.initialUser?.username ?? ''
    formState.password = ''
    formState.email = props.initialUser?.email ?? ''
    formState.role = props.initialUser?.role ?? 'user'
    formState.status = props.initialUser?.status ?? 'active'
  },
  { immediate: true },
)

async function handleSubmit() {
  isSaving.value = true

  const payload = {
    username: formState.username,
    password: formState.password,
    email: formState.email,
    role: formState.role,
    status: formState.status,
  }

  try {
    const result = isEditing.value && props.initialUser
      ? await updateUser(props.initialUser.id, payload)
      : await createUser(payload)

    message.success(isEditing.value ? '用户已更新' : '用户已创建')
    emit('saved', result)
    show.value = false
  }
  finally {
    isSaving.value = false
  }
}
</script>

<template>
  <NDrawer v-model:show="show" :width="drawerWidth">
    <NDrawerContent :title="isEditing ? '编辑用户' : '新建用户'" closable>
      <NForm label-placement="top">
        <NFormItem label="用户名">
          <NInput v-model:value="formState.username" />
        </NFormItem>

        <NFormItem :label="isEditing ? '重置密码' : '初始密码'">
          <NInput v-model:value="formState.password" type="password" show-password-on="click" />
        </NFormItem>

        <NFormItem label="邮箱">
          <NInput v-model:value="formState.email" />
        </NFormItem>

        <NFormItem label="角色">
          <NSelect
            v-model:value="formState.role"
            :options="[
              { label: 'User', value: 'user' },
              { label: 'Admin', value: 'admin' },
            ]"
          />
        </NFormItem>

        <NFormItem label="状态">
          <NSelect
            v-model:value="formState.status"
            :options="[
              { label: 'Active', value: 'active' },
              { label: 'Disabled', value: 'disabled' },
            ]"
          />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="show = false">取消</NButton>
          <NButton type="primary" :loading="isSaving" @click="handleSubmit">
            {{ isEditing ? '保存修改' : '创建用户' }}
          </NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
