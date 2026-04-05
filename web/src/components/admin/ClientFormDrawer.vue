<script setup lang="ts">
import {
  NAvatar,
  NButton,
  NCheckbox,
  NCheckboxGroup,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NSpace,
  NSwitch,
  NText,
  NUpload,
  useMessage,
} from 'naive-ui'
import type { UploadCustomRequestOptions } from 'naive-ui'
import { computed, onBeforeUnmount, reactive, shallowRef, watch } from 'vue'

import { createClient, updateClient, uploadClientIcon } from '@/api/admin'
import { ApiError } from '@/api/http'
import { getServerBaseUrl } from '@/config/endpoints'
import { defaultScopeKeys, scopeOptions } from '@/constants/scopes'
import { useViewport } from '@/composables/useViewport'
import type { OAuthClientRecord } from '@/types/api'

interface Props {
  initialClient?: OAuthClientRecord | null
}

const props = defineProps<Props>()
const emit = defineEmits<{ saved: [client: OAuthClientRecord] }>()
const show = defineModel<boolean>('show', { required: true })
const isSaving = defineModel<boolean>('saving', { default: false })

const message = useMessage()
const { width } = useViewport()
const drawerWidth = computed(() => Math.min(520, Math.max(300, width.value - 16)))
const isUploadingIcon = shallowRef(false)
const localPreviewUrl = shallowRef('')
const previewLoadFailed = shallowRef(false)

const formState = reactive({
  name: '',
  description: '',
  iconUrl: '',
  clientId: '',
  clientType: 'public',
  redirectUris: '',
  scopes: [...defaultScopeKeys],
  trusted: false,
})

const isEditing = computed(() => Boolean(props.initialClient?.id))
const previewIconUrl = computed(() => {
  if (localPreviewUrl.value) {
    return localPreviewUrl.value
  }

  const raw = formState.iconUrl.trim()
  if (!raw) {
    return ''
  }

  if (/^https?:\/\//i.test(raw) || raw.startsWith('blob:') || raw.startsWith('data:')) {
    return raw
  }

  const baseUrl = getServerBaseUrl()
  return new URL(raw.replace(/^\/+/, '/'), `${baseUrl}/`).toString()
})
const previewInitial = computed(() => (formState.name.trim().charAt(0) || 'C').toUpperCase())

watch(
  () => [show.value, props.initialClient] as const,
  () => {
    if (!show.value) {
      return
    }

    formState.name = props.initialClient?.name ?? ''
    formState.description = props.initialClient?.description ?? ''
    formState.iconUrl = props.initialClient?.icon_url ?? ''
    formState.clientId = props.initialClient?.client_id ?? ''
    formState.clientType = props.initialClient?.client_type ?? 'public'
    formState.redirectUris = props.initialClient?.redirect_uris.join('\n') ?? ''
    formState.scopes = props.initialClient?.scopes.length
      ? [...props.initialClient.scopes]
      : [...defaultScopeKeys]
    formState.trusted = props.initialClient?.trusted ?? false
    previewLoadFailed.value = false
    resetLocalPreview()
  },
  { immediate: true },
)

watch(() => formState.iconUrl, () => {
  previewLoadFailed.value = false
})

onBeforeUnmount(() => {
  resetLocalPreview()
})

async function handleSubmit() {
  isSaving.value = true

  const payload = {
    name: formState.name,
    description: formState.description,
    icon_url: formState.iconUrl,
    client_id: formState.clientId,
    client_type: formState.clientType,
    redirect_uris: splitByLine(formState.redirectUris),
    scopes: formState.scopes.length ? [...formState.scopes] : [...defaultScopeKeys],
    trusted: formState.trusted,
  }

  try {
    const result = isEditing.value && props.initialClient
      ? await updateClient(props.initialClient.id, payload)
      : await createClient(payload)

    message.success(isEditing.value ? '客户端已更新' : '客户端已创建')
    emit('saved', result)
    show.value = false
  }
  finally {
    isSaving.value = false
  }
}

function splitByLine(value: string) {
  return value
    .split('\n')
    .map(item => item.trim())
    .filter(Boolean)
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
		const result = await uploadClientIcon(file)
		formState.iconUrl = result.icon_url
		message.success('图标已上传')
		options.onFinish?.()
	}
	catch (error) {
		message.error(error instanceof ApiError ? error.message : '上传图标失败')
		options.onError?.()
	}
	finally {
		isUploadingIcon.value = false
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
  <NDrawer v-model:show="show" :width="drawerWidth">
    <NDrawerContent :title="isEditing ? '编辑客户端' : '新建客户端'" closable>
      <NForm label-placement="top">
        <NFormItem label="客户端名称">
          <NInput v-model:value="formState.name" placeholder="例如：XLNet Console" />
        </NFormItem>

        <NFormItem label="图标地址">
          <NSpace vertical :size="10" style="width: 100%;">
            <NInput v-model:value="formState.iconUrl" placeholder="https://example.com/icon.png" />
            <NUpload
              accept="image/*"
              :show-file-list="false"
              :custom-request="handleIconUpload"
            >
              <NButton secondary :loading="isUploadingIcon">上传图标</NButton>
            </NUpload>
          </NSpace>
        </NFormItem>

        <NFormItem label="图标预览">
          <div class="icon-preview">
            <div class="icon-preview-avatar">
              <img
                v-if="previewIconUrl && !previewLoadFailed"
                :src="previewIconUrl"
                alt="客户端图标预览"
                class="icon-preview-image"
                @error="previewLoadFailed = true"
              >
              <NAvatar v-else :size="48" :round="false" class="icon-preview-fallback">
                {{ previewInitial }}
              </NAvatar>
            </div>
            <div class="icon-preview-meta">
              <strong>{{ formState.name || '客户端预览' }}</strong>
              <NText depth="3">支持远程地址与本地上传。</NText>
            </div>
          </div>
        </NFormItem>

        <NFormItem label="客户端 ID">
          <NInput
            v-model:value="formState.clientId"
            :disabled="isEditing"
            placeholder="留空则自动生成"
          />
        </NFormItem>

        <NFormItem label="客户端类型">
          <NSelect
            v-model:value="formState.clientType"
            :disabled="isEditing"
            :options="[
              { label: 'Public', value: 'public' },
              { label: 'Confidential', value: 'confidential' },
            ]"
          />
        </NFormItem>

        <NFormItem label="回调地址">
          <NInput
            v-model:value="formState.redirectUris"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 6 }"
            placeholder="每行一个 redirect_uri"
          />
        </NFormItem>

        <NFormItem label="允许 Scope">
          <NCheckboxGroup v-model:value="formState.scopes">
            <NSpace vertical size="small">
              <div v-for="scope in scopeOptions" :key="scope.key" class="scope-option">
                <NCheckbox :value="scope.key">
                  {{ scope.label }}
                </NCheckbox>
                <NText depth="3">{{ scope.description }}</NText>
              </div>
            </NSpace>
          </NCheckboxGroup>
        </NFormItem>

        <NFormItem label="说明">
          <NInput
            v-model:value="formState.description"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 5 }"
            placeholder="客户端用途说明"
          />
        </NFormItem>

        <NFormItem label="跳过授权确认">
          <NSwitch v-model:value="formState.trusted" />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="show = false">取消</NButton>
          <NButton type="primary" :loading="isSaving" @click="handleSubmit">
            {{ isEditing ? '保存' : '创建' }}
          </NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
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
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.icon-preview-fallback {
  background: rgba(52, 159, 244, 0.18);
  color: #9fd6ff;
}

.icon-preview-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.scope-option {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
</style>
