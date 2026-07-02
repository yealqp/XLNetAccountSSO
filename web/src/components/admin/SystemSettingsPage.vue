<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Avatar, Button, Card, Form, FormItem, Input, Message, Space, Switch, TypographyText as Text, Upload } from '@arco-design/web-vue'
import type { RequestOption, UploadRequest } from '@arco-design/web-vue'
import { computed, onBeforeUnmount, onMounted, reactive, shallowRef } from 'vue'

useHead({ title: '系统设置 — XLNetAccount' })

import { ApiError } from '@/api/http'
import { sendTestEmail, uploadWebIcon } from '@/api/settings'
import { resolveServerUrl } from '@/config/endpoints'
import { usePlatformStore } from '@/stores/platform'

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
    Message.error(loadError.value)
  }
}

function handleIconUpload(options: RequestOption) {
  const file = options.fileItem.file
  if (!(file instanceof File)) {
    options.onError()
    Message.error('无法读取上传文件')
    return {} as UploadRequest
  }

  isUploadingIcon.value = true
  setLocalPreview(file)

  uploadWebIcon(file).then((result) => {
    formState.webIconUrl = result.web_icon_url
    previewLoadFailed.value = false
    Message.success('网页图标已上传')
    options.onSuccess()
  }).catch((error) => {
    Message.error(error instanceof ApiError ? error.message : '上传网页图标失败')
    options.onError()
  }).finally(() => {
    isUploadingIcon.value = false
  })

  return {} as UploadRequest
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
    Message.success('系统设置已保存')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '保存系统设置失败'
    Message.error(loadError.value)
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
    Message.success('测试邮件已发送')
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '发送测试邮件失败'
    Message.error(loadError.value)
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

    <Alert v-if="loadError" type="error" :show-icon="false">
      {{ loadError }}
    </Alert>

    <Card title="平台设置">
      <Form :model="formState" layout="vertical" @submit="handleSubmit">
        <FormItem label="平台名称">
          <Input v-model="formState.platformName" placeholder="请输入平台名称" />
        </FormItem>

        <FormItem label="开放注册">
          <Switch v-model="formState.allowRegistration" />
        </FormItem>

        <FormItem label="邮件服务">
          <Text v-if="smtpConfigured" type="success">已配置</Text>
          <Text v-else type="warning">未配置（请在环境变量中设置 SMTP_HOST / SMTP_USER / SMTP_PWD）</Text>
        </FormItem>

        <FormItem label="人机验证 (CAPTCHA)">
          <Text v-if="capConfigured" type="success">已配置</Text>
          <Text v-else type="warning">未配置（请在环境变量中设置 CAP_API_ENDPOINT / CAP_SITE_KEY / CAP_SECRET_KEY）</Text>
        </FormItem>

        <FormItem label="网页图标地址">
          <Space vertical :size="10" style="width: 100%;">
            <Input v-model="formState.webIconUrl" placeholder="https://example.com/favicon.ico" />
            <Upload
              accept="image/*"
              :show-file-list="false"
              :custom-request="handleIconUpload"
            >
              <Button type="secondary" :loading="isUploadingIcon">上传网页图标</Button>
            </Upload>
          </Space>
        </FormItem>

        <FormItem label="图标预览">
          <div class="icon-preview">
            <div class="icon-preview-avatar">
              <img
                v-if="previewIconUrl && !previewLoadFailed"
                :src="previewIconUrl"
                alt="网页图标预览"
                class="icon-preview-image"
                @error="previewLoadFailed = true"
              >
              <Avatar v-else :size="48" :round="false" class="icon-preview-fallback">
                X
              </Avatar>
            </div>
            <div class="icon-preview-meta">
              <strong>{{ formState.platformName || '网页图标预览' }}</strong>
              <Text type="secondary">保存后会同步更新站点 favicon。</Text>
            </div>
          </div>
        </FormItem>

        <FormItem label="测试收件邮箱">
          <Input v-model="formState.testEmail" placeholder="test@example.com" />
        </FormItem>

        <Space>
          <Button type="primary" html-type="submit" :loading="isSaving">
            保存
          </Button>
          <Button type="secondary" :loading="isSendingTestEmail" @click="handleSendTestEmail">
            发送测试邮件
          </Button>
        </Space>
      </Form>
    </Card>
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
