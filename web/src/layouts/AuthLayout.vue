<script setup lang="ts">
import { Layout, LayoutContent } from '@arco-design/web-vue'
import { RouterView } from 'vue-router'

import { usePlatformStore } from '@/stores/platform'

const platformStore = usePlatformStore()
void platformStore.ensureLoaded().catch(() => {})
</script>

<template>
  <Layout class="page-shell auth-layout">
    <LayoutContent class="auth-content">
      <div class="auth-shell page-container">
        <header class="auth-brand">
          <p class="auth-eyebrow">{{ platformStore.displayName }}</p>
          <h1 class="auth-headline">统一账号与授权中心</h1>
          <p class="auth-description">登录、授权确认与令牌管理统一收口。</p>
        </header>

        <section class="auth-panel">
          <div class="auth-panel-card">
            <RouterView />
          </div>
        </section>
      </div>
    </LayoutContent>
  </Layout>
</template>

<style scoped>
.auth-layout {
  background: var(--color-canvas);
}

.auth-layout :deep(.arco-layout-content) {
  min-height: 100%;
}

.auth-content {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--color-canvas);
}

.auth-shell {
  min-height: 100vh;
  min-height: 100dvh;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(360px, 440px);
  align-items: center;
  gap: clamp(40px, 6vw, 80px);
  padding: clamp(24px, 4vw, 48px);
}

.auth-brand {
  max-width: 520px;
}

.auth-eyebrow {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  border-radius: 9999px;
  margin: 0 0 20px;
  background: var(--color-surface-card);
  border: 1px solid var(--color-hairline-strong);
  color: var(--color-charcoal);
  font-size: 12px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.auth-headline {
  margin: 0;
  font-size: clamp(32px, 4.4vw, 52px);
  line-height: 1.08;
  font-weight: 650;
  letter-spacing: -0.03em;
  color: var(--color-ink);
}

.auth-description {
  max-width: 440px;
  margin: 18px 0 0;
  font-size: clamp(15px, 1.4vw, 17px);
  line-height: 1.65;
  color: var(--color-charcoal);
}

.auth-panel {
  display: flex;
  justify-content: flex-end;
  align-items: stretch;
  min-height: 0;
}

.auth-panel-card {
  width: min(100%, 460px);
  padding: clamp(24px, 3vw, 32px);
  max-height: calc(100dvh - (clamp(24px, 4vw, 48px) * 2));
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-hairline-strong);
  background: var(--color-surface-card);
  overflow-y: auto;
  overscroll-behavior: contain;
}

@media (max-width: 1080px) {
  .auth-shell {
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(0, 1fr);
    gap: 28px;
  }

  .auth-panel {
    justify-content: center;
  }

  .auth-panel-card {
    width: min(100%, 520px);
    max-height: 100%;
  }
}

@media (max-width: 720px) {
  .auth-shell {
    --auth-shell-padding: 16px;
    gap: 16px;
    padding: 16px;
  }

  .auth-headline {
    font-size: 28px;
  }

  .auth-description {
    font-size: 14px;
  }

  .auth-panel-card {
    width: 100%;
    max-height: none;
    padding: 20px;
    overflow-y: visible;
  }

  .auth-content {
    overflow-y: auto;
  }
}
</style>
