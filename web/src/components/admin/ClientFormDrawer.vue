<script setup lang="ts">
import {
  Avatar,
  Button,
  Drawer,
  Form,
  FormItem,
  Input,
  Message,
  Popconfirm,
  Select,
  Space,
  Switch,
  Tooltip,
  TypographyText as Text,
  Upload,
} from '@arco-design/web-vue'
import type { RequestOption, UploadRequest } from '@arco-design/web-vue'
import Textarea from '@/components/ui/Textarea.vue'
import { computed, onBeforeUnmount, reactive, shallowRef, watch } from 'vue'

import { createClient, updateClient, updateManagedClient, uploadClientIcon } from '@/api/admin'
import { ApiError } from '@/api/http'
import { resolveServerUrl } from '@/config/endpoints'
import { defaultScopeKeys, scopeOptions } from '@/constants/scopes'
import { useViewport } from '@/composables/useViewport'
import type { OAuthClientRecord } from '@/types/api'

const clientTypeOptions = [
  { label: 'Public', value: 'public', description: '无法安全保管 secret，依赖 PKCE 保护，适用于 SPA/移动端' },
  { label: 'Confidential', value: 'confidential', description: '可安全保管 secret，授权码交换时需校验，适用于后端服务' },
]

interface Props {
  initialClient?: OAuthClientRecord | null
  manageAll?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{ saved: [client: OAuthClientRecord] }>()
const show = defineModel<boolean>('show', { required: true })
const isSaving = defineModel<boolean>('saving', { default: false })

const { width } = useViewport()
const drawerWidth = computed(() => Math.min(520, Math.max(300, width.value - 16)))
const isUploadingIcon = shallowRef(false)
const localPreviewUrl = shallowRef('')
const previewLoadFailed = shallowRef(false)
const showTrustedConfirm = shallowRef(false)

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

const selectedScopes = computed(() => new Set(formState.scopes))

function toggleScope(key: string) {
  const idx = formState.scopes.indexOf(key)
  if (idx >= 0) {
    formState.scopes.splice(idx, 1)
  }
  else {
    formState.scopes.push(key)
  }
}

const isEditing = computed(() => Boolean(props.initialClient?.id))
const previewIconUrl = computed(() => {
  if (localPreviewUrl.value) {
    return localPreviewUrl.value
  }

  return resolveServerUrl(formState.iconUrl)
})
const previewInitial = computed(() => (formState.name.trim().charAt(0) || 'C').toUpperCase())

const hasRequiredFields = computed(() => {
  return formState.name.trim() !== '' && formState.redirectUris.trim() !== ''
})

watch(
  () => [show.value, props.initialClient] as const,
  () => {
    if (!show.value) {
      showTrustedConfirm.value = false
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
      ? await (props.manageAll ? updateManagedClient(props.initialClient.id, payload) : updateClient(props.initialClient.id, payload))
      : await createClient(payload)

    Message.success(isEditing.value ? '客户端已更新' : '客户端已创建')
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

function handleIconUpload(options: RequestOption) {
	const file = options.fileItem.file
	if (!(file instanceof File)) {
		options.onError()
		Message.error('无法读取上传文件')
		return {} as UploadRequest
	}

	isUploadingIcon.value = true
	setLocalPreview(file)

	uploadClientIcon(file).then((result) => {
		formState.iconUrl = result.icon_url
		Message.success('图标已上传')
		options.onSuccess()
	}).catch((error) => {
		Message.error(error instanceof ApiError ? error.message : '上传图标失败')
		options.onError()
	}).finally(() => {
		isUploadingIcon.value = false
	})

	return {} as UploadRequest
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

function onTrustedClick() {
  if (formState.trusted) {
    formState.trusted = false
  } else {
    showTrustedConfirm.value = true
  }
}

function onTrustedConfirm() {
  formState.trusted = true
  showTrustedConfirm.value = false
}

function onTrustedCancel() {
  showTrustedConfirm.value = false
}
</script>

<template>
  <Drawer v-model:visible="show" :width="drawerWidth" :title="isEditing ? '编辑客户端' : '新建客户端'" closable>
      <Form :model="formState" layout="vertical">
        <FormItem label="客户端名称" required>
          <Input v-model="formState.name" placeholder="例如：XLNet Console" />
        </FormItem>

        <FormItem label="图标地址">
          <Space vertical :size="10" style="width: 100%;">
            <Input v-model="formState.iconUrl" placeholder="https://example.com/icon.png" />
            <Upload
              accept="image/*"
              :show-file-list="false"
              :custom-request="handleIconUpload"
            >
              <Button type="secondary" :loading="isUploadingIcon">上传图标</Button>
            </Upload>
          </Space>
        </FormItem>

        <FormItem label="图标预览">
          <div class="icon-preview">
            <div class="icon-preview-avatar">
              <img
                v-if="previewIconUrl && !previewLoadFailed"
                :src="previewIconUrl"
                alt="客户端图标预览"
                class="icon-preview-image"
                @error="previewLoadFailed = true"
              >
              <Avatar v-else :size="48" :round="false" class="icon-preview-fallback">
                {{ previewInitial }}
              </Avatar>
            </div>
            <div class="icon-preview-meta">
              <strong>{{ formState.name || '客户端预览' }}</strong>
              <Text type="secondary">支持远程地址与本地上传。</Text>
            </div>
          </div>
        </FormItem>

        <FormItem label="客户端 ID">
          <Input
            v-model="formState.clientId"
            :disabled="isEditing"
            placeholder="留空则自动生成"
          />
        </FormItem>

        <FormItem label="客户端类型">
          <Select
            v-model="formState.clientType"
            :disabled="isEditing"
            :options="clientTypeOptions"
          >
            <template #option="{ data }">
              <div class="client-type-option">
                <div class="client-type-label">{{ data.label }}</div>
                <div class="client-type-desc">{{ data.description }}</div>
              </div>
            </template>
          </Select>
        </FormItem>

        <FormItem label="回调地址" required>
          <Textarea
            v-model:value="formState.redirectUris"
            placeholder="每行一个 redirect_uri"
            :auto-size="{ minRows: 3, maxRows: 6 }"
          />
        </FormItem>

        <FormItem label="允许 Scope">
          <div class="scope-grid">
            <div
              v-for="scope in scopeOptions"
              :key="scope.key"
              class="scope-option"
              :class="{ selected: selectedScopes.has(scope.key) }"
              @click="toggleScope(scope.key)"
            >
              <div class="scope-check-col">
                <div class="scope-check-indicator">
                  <span v-if="selectedScopes.has(scope.key)" class="scope-check-mark">✓</span>
                </div>
              </div>
              <div class="scope-info">
                <span class="scope-label">{{ scope.label }}</span>
                <span class="scope-desc">{{ scope.description }}</span>
              </div>
            </div>
          </div>
        </FormItem>

        <FormItem label="说明">
          <Textarea
            v-model:value="formState.description"
            :auto-size="{ minRows: 3, maxRows: 5 }"
            placeholder="客户端用途说明"
          />
        </FormItem>

        <FormItem label="跳过授权确认">
          <Popconfirm
            :popup-visible="showTrustedConfirm"
            content="启用后该客户端将跳过用户授权确认页面，用户无法看到具体权限请求，请谨慎操作。仅在完全信任该客户端时启用此项。"
            ok-text="确认启用"
            cancel-text="取消"
            @ok="onTrustedConfirm"
            @cancel="onTrustedCancel"
          >
            <Switch :model-value="formState.trusted" @click="onTrustedClick" />
          </Popconfirm>
        </FormItem>
      </Form>

      <template #footer>
        <Space justify="end">
          <Button @click="show = false">取消</Button>
          <Tooltip v-if="!hasRequiredFields" content="请填写客户端名称和回调地址">
            <Button type="primary" :loading="isSaving" disabled @click="handleSubmit">
              {{ isEditing ? '保存' : '创建' }}
            </Button>
          </Tooltip>
          <Button v-else type="primary" :loading="isSaving" @click="handleSubmit">
            {{ isEditing ? '保存' : '创建' }}
          </Button>
        </Space>
      </template>
      </Drawer>
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

.scope-grid {
  display: grid;
  gap: 6px;
}

.scope-option {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md);
  background: rgba(0, 0, 0, 0.02);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
  user-select: none;
}

.scope-option:hover {
  border-color: var(--color-accent-blue);
  background: rgba(37, 99, 235, 0.04);
}

.scope-option.selected {
  border-color: var(--color-accent-blue);
  background: rgba(37, 99, 235, 0.08);
}

.scope-check-col {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  margin-top: 2px;
  flex-shrink: 0;
}

.scope-check-indicator {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 1px solid var(--color-hairline-strong);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-surface-card);
  transition: border-color 0.15s, background 0.15s;
}

.scope-option.selected .scope-check-indicator {
  border-color: var(--color-accent-blue);
  background: var(--color-accent-blue);
}

.scope-check-mark {
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
}

.scope-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.scope-label {
  font-weight: 500;
  font-size: 14px;
  color: var(--color-ink);
}

.scope-desc {
  font-size: 12px;
  color: var(--color-charcoal);
}

.client-type-option {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 4px 0;
}

.client-type-label {
  font-weight: 500;
  font-size: 14px;
  line-height: 1.4;
}

.client-type-desc {
  font-size: 12px;
  color: var(--color-charcoal);
  line-height: 1.3;
}
</style>
