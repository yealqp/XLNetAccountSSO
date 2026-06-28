<script setup lang="ts">
import { NAlert, NAvatar, NButton, NCard, NForm, NFormItem, NInput, NSpace, NSwitch, NText, NUpload, useMessage } from 'naive-ui'
import type { UploadCustomRequestOptions } from 'naive-ui'
import { computed, onBeforeUnmount, onMounted, reactive, shallowRef } from 'vue'

import { ApiError } from '@/api/http'
import { sendTestEmail, uploadWebIcon } from '@/api/settings'
import { resolveServerUrl } from '@/config/endpoints'
import { usePlatformStore } from '@/stores/platform'

const message = useMessage()
const platformStore = usePlatformStore()

const isSaving = shallowRef(false)
const isSendingTestEmail = shallowRef(false)
const isUploadingIcon = shallowRef(false)
const loadError = shallowRef('')
const localPreviewUrl = shallowRef('')
const previewLoadFailed = shallowRef(false)
const smtpConfigured = shallowRef(false)
const capConfigured = shallowRef(false)
const formState = reactive({
  platformName: '',
  allowRegistration: false,
  webIconUrl: '',
  testEmail: '',
})

const previewIconUrl = computed(() => {
  if (localPreviewUrl.value) {
    return localPreviewUrl.value
  }
  return resolveServerUrl(formState.webIconUrl)
})

onMounted(async () => {
  await loadSettings()
})

onBeforeUnmount(() => {
  resetLocalPreview()
})

async function loadSettings() {
  loadError.value = ''
  try {
    const settings = await platformStore.loadAdminSettings()
    formState.platformName = settings.platform_name
    formState.allowRegistration = settings.allow_registration
    formState.webIconUrl = settings.web_icon_url ?? ''
    smtpConfigured.value = settings.smtp_configured ?? false
    capConfigured.value = settings.cap_configured ?? false
    previewLoadFailed.value = false
    resetLocalPreview()
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '加载系统设置失败'
    message.error(loadError.value)
  }
}

async function handleIconUpload(options: UploadCustomRequestOptions) {
  const file = options.file.file
  if (!(file instanceof File)) {
    options.onError?.()
    message.error('无法读取上传文件')
    return
  }

  isUploadingIcon.value = true
  setLocalPreview(file)

  try {
    const result = await uploadWebIcon(file)
    formState.webIconUrl = result.web_icon_url
    previewLoadFailed.value = false
    message.success('网页图标已上传')
    options.onFinish?.()
  }
  catch (error) {
    message.error(error instanceof ApiError ? error.message : '上传网页图标失败')
    options.onError?.()
  }
  finally {
    isUploadingIcon.value = false
  }
}

async function handleSubmit() {
  isSaving.value = true
  loadError.value = ''
  try {
    await platformStore.savePlatformSettings({
      platform_name: formState.platformName,
      allow_registration: formState.allowRegistration,
      web_icon_url: formState.webIconUrl,
    })
    message.success('系统设置已保存')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '保存系统设置失败'
    message.error(loadError.value)
  }
  finally {
    isSaving.value = false
  }
}

async function handleSendTestEmail() {
  isSendingTestEmail.value = true
  loadError.value = ''
  try {
    await sendTestEmail({ to: formState.testEmail })
    message.success('测试邮件已发送')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '发送测试邮件失败'
    message.error(loadError.value)
  }
  finally {
    isSendingTestEmail.value = false
  }
}

function setLocalPreview(file: File) {
  resetLocalPreview()
  localPreviewUrl.value = URL.createObjectURL(file)
  previewLoadFailed.value = false
}

function resetLocalPreview() {
  if (localPreviewUrl.value) {
    URL.revokeObjectURL(localPreviewUrl.value)
  }
  localPreviewUrl.value = ''
}
</script>

<template>
  <section class="page-stack">
    <header class="page-header">
      <div>
        <h1 class="page-title">系统设置</h1>
        <p class="page-subtitle">平台、注册与网页图标配置。</p>
      </div>
    </header>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      {{ loadError }}
    </NAlert>

    <NCard title="平台设置">
      <NForm label-placement="top" @submit.prevent="handleSubmit">
        <NFormItem label="平台名称">
          <NInput v-model:value="formState.platformName" placeholder="请输入平台名称" />
        </NFormItem>

        <NFormItem label="开放注册">
          <NSwitch v-model:value="formState.allowRegistration" />
        </NFormItem>

        <NFormItem label="邮件服务">
          <NText v-if="smtpConfigured" type="success">已配置</NText>
          <NText v-else type="warning">未配置（请在环境变量中设置 SMTP_HOST / SMTP_USER / SMTP_PWD）</NText>
        </NFormItem>

        <NFormItem label="人机验证 (CAPTCHA)">
          <NText v-if="capConfigured" type="success">已配置</NText>
          <NText v-else type="warning">未配置（请在环境变量中设置 CAP_API_ENDPOINT / CAP_SITE_KEY / CAP_SECRET_KEY）</NText>
        </NFormItem>

        <NFormItem label="网页图标地址">
          <NSpace vertical :size="10" style="width: 100%;">
            <NInput v-model:value="formState.webIconUrl" placeholder="https://example.com/favicon.ico" />
            <NUpload
              accept="image/*"
              :show-file-list="false"
              :custom-request="handleIconUpload"
            >
              <NButton secondary :loading="isUploadingIcon">上传网页图标</NButton>
            </NUpload>
          </NSpace>
        </NFormItem>

        <NFormItem label="图标预览">
          <div class="icon-preview">
            <div class="icon-preview-avatar">
              <img
                v-if="previewIconUrl && !previewLoadFailed"
                :src="previewIconUrl"
                alt="网页图标预览"
                class="icon-preview-image"
                @error="previewLoadFailed = true"
              >
              <NAvatar v-else :size="48" :round="false" class="icon-preview-fallback">
                X
              </NAvatar>
            </div>
            <div class="icon-preview-meta">
              <strong>{{ formState.platformName || '网页图标预览' }}</strong>
              <NText depth="3">保存后会同步更新站点 favicon。</NText>
            </div>
          </div>
        </NFormItem>

        <NFormItem label="测试收件邮箱">
          <NInput v-model:value="formState.testEmail" placeholder="test@example.com" />
        </NFormItem>

        <NSpace>
          <NButton type="primary" attr-type="submit" :loading="isSaving">
            保存
          </NButton>
          <NButton secondary :loading="isSendingTestEmail" @click="handleSendTestEmail">
            发送测试邮件
          </NButton>
        </NSpace>
      </NForm>
    </NCard>
  </section>
</template>

<style scoped>
.icon-preview {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-preview-avatar {
  width: 48px;
  height: 48px;
  flex-shrink: 0;
}

.icon-preview-image {
  display: block;
  width: 48px;
  height: 48px;
  object-fit: cover;
  border-radius: 10px;
  border: 1px solid var(--color-hairline);
}

.icon-preview-fallback {
  background: rgba(52, 159, 244, 0.18);
  color: var(--color-accent-blue);
}

.icon-preview-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
</style>
