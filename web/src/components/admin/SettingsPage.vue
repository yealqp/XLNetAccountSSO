<script setup lang="ts">
import { NAlert, NButton, NCard, NForm, NFormItem, NInput, useMessage } from 'naive-ui'
import { onMounted, reactive, shallowRef } from 'vue'

import { ApiError } from '@/api/http'
import { usePlatformStore } from '@/stores/platform'

const message = useMessage()
const platformStore = usePlatformStore()

const isSaving = shallowRef(false)
const loadError = shallowRef('')
const formState = reactive({
	platformName: '',
})

onMounted(async () => {
	await loadSettings()
})

async function loadSettings() {
	loadError.value = ''
	try {
		formState.platformName = await platformStore.loadAdminSettings()
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
		await platformStore.savePlatformName(formState.platformName)
		message.success('平台名称已保存')
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

        <NButton type="primary" attr-type="submit" :loading="isSaving">
          保存
        </NButton>
      </NForm>
    </NCard>
  </section>
</template>
