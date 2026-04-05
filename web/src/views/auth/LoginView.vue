<script setup lang="ts">
import { NAlert, NButton, NCard, NForm, NFormItem, NInput, NSpace, NText, useMessage } from 'naive-ui'
import { computed, reactive, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const sessionStore = useSessionStore()

const isSubmitting = shallowRef(false)
const formState = reactive({
  username: 'admin',
  password: 'Admin123!',
})

const nextTargetLabel = computed(() => resolveNextTarget(route.query.next))

async function handleSubmit() {
  isSubmitting.value = true

  try {
    await sessionStore.signIn({
      username: formState.username,
      password: formState.password,
    })

    message.success('登录成功')
    await router.replace(resolveNextTarget(route.query.next))
  }
  catch (error) {
    const fallbackMessage = error instanceof ApiError ? error.message : '登录失败，请稍后重试'
    message.error(fallbackMessage)
  }
  finally {
    isSubmitting.value = false
  }
}

function resolveNextTarget(nextValue: unknown) {
  const next = Array.isArray(nextValue) ? nextValue[0] : nextValue

  if (typeof next === 'string' && next.startsWith('/') && !next.startsWith('//')) {
    return next
  }

  return '/admin'
}
</script>

<template>
  <NCard title="登录" class="login-card">
    <NSpace vertical :size="16">
      <NAlert type="info" :show-icon="false">
        进入 <span class="mono">{{ nextTargetLabel }}</span>
      </NAlert>

      <NForm label-placement="top" @submit.prevent="handleSubmit">
        <NFormItem label="用户名">
          <NInput v-model:value="formState.username" clearable placeholder="请输入用户名" />
        </NFormItem>

        <NFormItem label="密码">
          <NInput
            v-model:value="formState.password"
            type="password"
            show-password-on="click"
            placeholder="请输入密码"
          />
        </NFormItem>

        <NButton attr-type="submit" type="primary" block :loading="isSubmitting">
          登录
        </NButton>
      </NForm>

      <NText depth="3">默认账号：<span class="mono">admin / Admin123!</span></NText>
    </NSpace>
  </NCard>
</template>

<style scoped>
.login-card {
  max-width: 420px;
  margin: 0 auto;
}
</style>
