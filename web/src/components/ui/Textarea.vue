<script setup lang="ts">
defineProps<{
  value?: string
  placeholder?: string
  disabled?: boolean
  autoSize?: boolean | { minRows?: number; maxRows?: number }
}>()

defineEmits<{ 'update:value': [value: string] }>()
</script>

<template>
  <div class="arco-textarea-wrapper">
    <textarea
      class="arco-textarea"
      :value="value"
      :placeholder="placeholder"
      :disabled="disabled"
      @input="$emit('update:value', ($event.target as HTMLTextAreaElement).value)"
    />
  </div>
</template>
