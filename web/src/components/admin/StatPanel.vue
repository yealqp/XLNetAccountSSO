<script setup lang="ts">
interface Props {
  label: string
  value: string | number
  detail: string
}

defineProps<Props>()
</script>

<template>
  <article class="stat-panel">
    <span class="stat-label">{{ label }}</span>
    <strong class="stat-value">{{ value }}</strong>
    <p class="stat-detail">{{ detail }}</p>
  </article>
</template>

<style scoped>
.stat-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 20px;
  background: rgba(17, 24, 39, 0.7);
}

.stat-label {
  color: rgba(255, 255, 255, 0.58);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 12px;
}

.stat-value {
  font-size: 34px;
  line-height: 1;
}

.stat-detail {
  margin: 0;
  color: rgba(255, 255, 255, 0.58);
}
</style>
