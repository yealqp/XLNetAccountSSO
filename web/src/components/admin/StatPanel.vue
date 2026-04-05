<script setup lang="ts">
import { NCard, NText } from 'naive-ui'

interface Props {
  label: string
  value: string | number
  detail: string
}

defineProps<Props>()
</script>

<template>
  <NCard size="small">
    <div class="stat-panel">
      <NText depth="3">{{ label }}</NText>
      <strong class="stat-value">{{ value }}</strong>
      <NText depth="3">{{ detail }}</NText>
    </div>
  </NCard>
</template>

<style scoped>
.stat-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stat-value {
  font-size: 34px;
  line-height: 1;
}
</style>
