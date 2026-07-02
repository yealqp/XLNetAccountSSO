<script setup lang="ts">
import {
  Button,
  Drawer,
  Form,
  FormItem,
  Input,
  Message,
  Select,
  Space,
} from '@arco-design/web-vue'
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

    Message.success(isEditing.value ? '用户已更新' : '用户已创建')
    emit('saved', result)
    show.value = false
  }
  finally {
    isSaving.value = false
  }
}
</script>

<template>
  <Drawer v-model:visible="show" :width="drawerWidth" :title="isEditing ? '编辑用户' : '新建用户'" closable>
      <Form :model="formState" layout="vertical">
        <FormItem label="用户名">
          <Input v-model="formState.username" />
        </FormItem>

        <FormItem :label="isEditing ? '重置密码' : '初始密码'">
          <Input v-model="formState.password" type="password" />
        </FormItem>

        <FormItem label="邮箱">
          <Input v-model="formState.email" />
        </FormItem>

        <FormItem label="角色">
          <Select
            v-model="formState.role"
            :options="[
              { label: 'User', value: 'user' },
              { label: 'Admin', value: 'admin' },
            ]"
          />
        </FormItem>

        <FormItem label="状态">
          <Select
            v-model="formState.status"
            :options="[
              { label: 'Active', value: 'active' },
              { label: 'Disabled', value: 'disabled' },
            ]"
          />
        </FormItem>
      </Form>

      <template #footer>
        <Space justify="end">
          <Button @click="show = false">取消</Button>
          <Button type="primary" :loading="isSaving" @click="handleSubmit">
            {{ isEditing ? '保存修改' : '创建用户' }}
          </Button>
        </Space>
      </template>
      </Drawer>
</template>
