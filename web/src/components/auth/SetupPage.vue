<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Button, Form, FormItem, Input, Message } from '@arco-design/web-vue'
import { reactive, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

useHead({ title: '初始化 — XLNetAccount' })

import { ApiError } from '@/api/http'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'
import { useSetupStore } from '@/stores/setup'
import { resolveNextTarget } from '@/utils/authNext'

const route = useRoute()
const router = useRouter()
const platformStore = usePlatformStore()
const sessionStore = useSessionStore()
const setupStore = useSetupStore()

const isSubmitting = shallowRef(false)
const submitError = shallowRef('')
const formState = reactive({
	username: '',
	password: '',
	confirmPassword: '',
	email: '',
})

void platformStore.ensureLoaded().catch(() => {})

async function handleSubmit() {
	if (formState.password !== formState.confirmPassword) {
		submitError.value = '两次输入的密码不一致'
		return
	}

	isSubmitting.value = true
	submitError.value = ''

	try {
		await setupStore.completeSetup({
			username: formState.username,
			password: formState.password,
			email: formState.email,
		})
		await sessionStore.signIn({
			username: formState.username,
			password: formState.password,
		})
		Message.success('初始化完成')
		await router.replace(resolveNextTarget(route.query.next))
	}
	catch (error) {
		submitError.value = error instanceof ApiError ? error.message : '初始化失败，请稍后重试'
		Message.error(submitError.value)
	}
	finally {
		isSubmitting.value = false
	}
}

</script>

<template>
  <div class="auth-panel-view">
    <div class="auth-panel-header">
      <p class="auth-panel-kicker">首次初始化</p>
      <h2 class="auth-panel-title">创建 {{ platformStore.displayName }} 管理员</h2>
      <p class="auth-panel-subtitle">首次使用前需要先创建平台管理员账号。</p>
    </div>

    <Form :model="formState" layout="vertical" class="auth-form" @submit="handleSubmit">
      <FormItem label="用户名">
        <Input v-model="formState.username" clearable placeholder="管理员用户名" size="large" @input="submitError = ''" />
      </FormItem>

		<FormItem label="邮箱">
		  <Input v-model="formState.email" clearable placeholder="邮箱" size="large" @input="submitError = ''" />
		</FormItem>

      <FormItem label="密码">
        <Input
          v-model="formState.password"
          type="password"
                    placeholder="密码"
          size="large"
          @input="submitError = ''"
        />
      </FormItem>

      <FormItem label="确认密码">
        <Input
          v-model="formState.confirmPassword"
          type="password"
                    placeholder="再次输入密码"
          size="large"
          @input="submitError = ''"
        />
      </FormItem>

      <Button type="primary" size="large" long :loading="isSubmitting" html-type="submit">
        完成初始化
      </Button>
    </Form>
  </div>
</template>

<style scoped>
.auth-panel-view {
  color: var(--color-ink);
}

.auth-panel-header {
  margin-bottom: 24px;
}

.auth-panel-kicker {
  margin: 0 0 10px;
  color: var(--color-charcoal);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.auth-panel-title {
  margin: 0;
  font-size: clamp(26px, 2.8vw, 32px);
  font-weight: 650;
  letter-spacing: -0.03em;
}

.auth-panel-subtitle {
  margin: 10px 0 0;
  color: var(--color-charcoal);
  font-size: 14px;
  line-height: 1.65;
}

.auth-form {
  display: grid;
  gap: 4px;
}
</style>
