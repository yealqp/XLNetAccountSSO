<script setup lang="ts">
import {
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
  useMessage,
} from 'naive-ui'
import { computed, reactive, watch } from 'vue'

import { createClient, updateClient } from '@/api/admin'
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

const formState = reactive({
  name: '',
  description: '',
  clientId: '',
  clientType: 'public',
  redirectUris: '',
  scopes: [...defaultScopeKeys],
  trusted: false,
})

const isEditing = computed(() => Boolean(props.initialClient?.id))

watch(
  () => [show.value, props.initialClient] as const,
  () => {
    if (!show.value) {
      return
    }

    formState.name = props.initialClient?.name ?? ''
    formState.description = props.initialClient?.description ?? ''
    formState.clientId = props.initialClient?.client_id ?? ''
    formState.clientType = props.initialClient?.client_type ?? 'public'
    formState.redirectUris = props.initialClient?.redirect_uris.join('\n') ?? ''
    formState.scopes = props.initialClient?.scopes.length
      ? [...props.initialClient.scopes]
      : [...defaultScopeKeys]
    formState.trusted = props.initialClient?.trusted ?? false
  },
  { immediate: true },
)

async function handleSubmit() {
  isSaving.value = true

  const payload = {
    name: formState.name,
    description: formState.description,
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
</script>

<template>
  <NDrawer v-model:show="show" :width="drawerWidth">
    <NDrawerContent :title="isEditing ? '编辑客户端' : '新建客户端'" closable>
      <NForm label-placement="top">
        <NFormItem label="客户端名称">
          <NInput v-model:value="formState.name" placeholder="例如：XLNet Console" />
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
          <NCheckboxGroup v-model:value="formState.scopes" class="scope-grid">
            <label v-for="scope in scopeOptions" :key="scope.key" class="scope-card">
              <div class="scope-card-head">
                <NCheckbox :value="scope.key">
                  {{ scope.label }}
                </NCheckbox>
              </div>
              <p>{{ scope.description }}</p>
            </label>
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
.scope-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.scope-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 18px;
  background: rgba(15, 23, 42, 0.56);
}

.scope-card-head {
  display: flex;
  align-items: center;
}

.scope-card p {
  margin: 0;
  color: rgba(255, 255, 255, 0.58);
  line-height: 1.5;
}

@media (max-width: 680px) {
  .scope-grid {
    grid-template-columns: 1fr;
  }
}
</style>
