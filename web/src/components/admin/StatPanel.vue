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
  <NCard size="small" :bordered="true">
    <div class="stat-panel">
      <NText depth="3" class="stat-label">{{ label }}</NText>
      <strong class="stat-value">{{ value }}</strong>
      <NText depth="3" class="stat-detail">{{ detail }}</NText>
    </div>
  </NCard>
</template>

<style scoped>
.stat-panel {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-label {
  font-size: 12px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  font-weight: 600;
}

.stat-value {
  font-size: 40px;
  font-weight: 650;
  line-height: 1;
  letter-spacing: -0.03em;
  color: var(--color-ink);
}

.stat-detail {
  font-size: 12px;
  line-height: 1.4;
}
</style>
