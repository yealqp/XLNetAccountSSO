<script setup lang="ts">
import { NAlert, NButton, NCard, NForm, NFormItem, NInput, NInputNumber, NSwitch, useMessage } from 'naive-ui'
import { onMounted, reactive, shallowRef } from 'vue'

import { ApiError } from '@/api/http'
import { usePlatformStore } from '@/stores/platform'

const message = useMessage()
const platformStore = usePlatformStore()

const isSaving = shallowRef(false)
const loadError = shallowRef('')
const formState = reactive({
	platformName: '',
	allowRegistration: false,
	smtpHost: '',
	smtpUser: '',
	smtpPassword: '',
	smtpPort: 587,
	smtpTLS: true,
	capApiEndpoint: '',
	capSecretKey: '',
})

onMounted(async () => {
	await loadSettings()
})

async function loadSettings() {
	loadError.value = ''
	try {
		const settings = await platformStore.loadAdminSettings()
		formState.platformName = settings.platform_name
		formState.allowRegistration = settings.allow_registration
		formState.smtpHost = settings.smtp_host ?? ''
		formState.smtpUser = settings.smtp_user ?? ''
		formState.smtpPassword = settings.smtp_password ?? ''
		formState.smtpPort = Number(settings.smtp_port || 587)
		formState.smtpTLS = settings.smtp_tls ?? true
		formState.capApiEndpoint = settings.cap_api_endpoint ?? ''
		formState.capSecretKey = settings.cap_secret_key ?? ''
	}
	catch (error) {
		loadError.value = error instanceof ApiError ? error.message : '加载设置失败'
		message.error(loadError.value)
	}
}

async function handleSubmit() {
	isSaving.value = true
	loadError.value = ''
	try {
		await platformStore.savePlatformSettings({
			platform_name: formState.platformName,
			allow_registration: formState.allowRegistration,
			smtp_host: formState.smtpHost,
			smtp_user: formState.smtpUser,
			smtp_password: formState.smtpPassword,
			smtp_port: String(formState.smtpPort || ''),
			smtp_tls: formState.smtpTLS,
			cap_api_endpoint: formState.capApiEndpoint,
			cap_secret_key: formState.capSecretKey,
		})
		message.success('设置已保存')
	}
	catch (error) {
		loadError.value = error instanceof ApiError ? error.message : '保存设置失败'
		message.error(loadError.value)
	}
	finally {
		isSaving.value = false
	}
}
</script>

<template>
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">设置</h1>
        <p class="page-subtitle">平台基础设置。</p>
      </div>
    </header>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      {{ loadError }}
    </NAlert>

    <NCard title="平台名称">
      <NForm label-placement="top" @submit.prevent="handleSubmit">
        <NFormItem label="平台名称">
          <NInput v-model:value="formState.platformName" placeholder="请输入平台名称" />
        </NFormItem>

        <NFormItem label="开放注册">
          <NSwitch v-model:value="formState.allowRegistration" />
        </NFormItem>

        <NFormItem label="邮件服务器 Host">
          <NInput v-model:value="formState.smtpHost" placeholder="smtp.example.com" />
        </NFormItem>

        <NFormItem label="邮件账号">
          <NInput v-model:value="formState.smtpUser" placeholder="mailer@example.com" />
        </NFormItem>

        <NFormItem label="邮件密码">
          <NInput v-model:value="formState.smtpPassword" type="password" show-password-on="click" placeholder="邮件服务密码" />
        </NFormItem>

        <NFormItem label="邮件端口">
          <NInputNumber v-model:value="formState.smtpPort" :min="1" :max="65535" style="width: 100%;" />
        </NFormItem>

        <NFormItem label="启用 TLS">
          <NSwitch v-model:value="formState.smtpTLS" />
        </NFormItem>

        <NFormItem label="CAP.js API Endpoint">
          <NInput v-model:value="formState.capApiEndpoint" placeholder="https://cap.example.com/site-key/" />
        </NFormItem>

        <NFormItem label="CAP.js Secret Key">
          <NInput v-model:value="formState.capSecretKey" type="password" show-password-on="click" placeholder="CAP.js secret key" />
        </NFormItem>

        <NButton type="primary" attr-type="submit" :loading="isSaving">
          保存
        </NButton>
      </NForm>
    </NCard>
  </section>
</template>
