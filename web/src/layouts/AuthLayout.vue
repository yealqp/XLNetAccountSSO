<script setup lang="ts">
import { NLayout, NLayoutContent } from 'naive-ui'
import { RouterView } from 'vue-router'

import { usePlatformStore } from '@/stores/platform'

const platformStore = usePlatformStore()
void platformStore.ensureLoaded().catch(() => {})
</script>

<template>
  <NLayout class="page-shell auth-layout">
    <NLayoutContent class="auth-content">
      <div class="auth-shell__veil"></div>
      <div class="auth-shell auth-shell__content page-container">
        <section class="auth-hero">
          <p class="auth-eyebrow">{{ platformStore.displayName }}</p>
          <h1 class="auth-headline">统一账号与授权中心</h1>
          <p class="auth-description">登录、授权确认与令牌管理统一收口。</p>
        </section>

        <section class="auth-panel">
          <div class="auth-panel-card">
            <RouterView />
          </div>
        </section>
      </div>
    </NLayoutContent>
  </NLayout>
</template>

<style scoped>
.auth-layout {
  background: transparent;
}

.auth-layout :deep(.n-layout-scroll-container) {
  min-height: 100%;
}

.auth-content {
  position: relative;
  min-height: 100vh;
  min-height: 100dvh;
  overflow: hidden;
  background:
    linear-gradient(110deg, rgba(8, 14, 26, 0.82) 0%, rgba(8, 14, 26, 0.62) 42%, rgba(8, 14, 26, 0.76) 100%),
    url('https://t.alcy.cc/ycy') center / cover no-repeat,
    #101014;
}

.auth-shell__veil {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgba(5, 10, 20, 0.22), rgba(5, 10, 20, 0.08));
  backdrop-filter: blur(2px);
}

.auth-shell {
  --auth-shell-padding: clamp(20px, 4vw, 40px);
  position: relative;
  z-index: 1;
  height: 100vh;
  height: 100dvh;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(340px, 440px);
  align-items: center;
  gap: clamp(28px, 5vw, 72px);
  padding: var(--auth-shell-padding);
}

.auth-hero {
  max-width: 520px;
  color: #f2f7ff;
}

.auth-eyebrow {
  display: inline-flex;
  align-items: center;
  padding: 7px 12px;
  border-radius: 999px;
  margin: 0 0 18px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.14);
  color: rgba(239, 246, 255, 0.86);
  font-size: 12px;
  letter-spacing: 0.14em;
}

.auth-headline {
  margin: 0;
  font-size: clamp(34px, 4.6vw, 54px);
  line-height: 1.08;
  font-weight: 650;
  letter-spacing: -0.03em;
}

.auth-description {
  max-width: 480px;
  margin: 18px 0 0;
  font-size: clamp(15px, 1.5vw, 17px);
  line-height: 1.7;
  color: rgba(230, 240, 252, 0.76);
}

.auth-panel {
  display: flex;
  justify-content: flex-end;
  align-items: stretch;
  min-height: 0;
}

.auth-panel-card {
  width: min(100%, 460px);
  padding: clamp(20px, 3vw, 28px);
  max-height: calc(100dvh - (var(--auth-shell-padding) * 2));
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(10, 18, 30, 0.42);
  box-shadow: 0 18px 48px rgba(4, 10, 22, 0.32);
  backdrop-filter: blur(22px);
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
  }

  .auth-headline {
    font-size: 32px;
  }

  .auth-description {
    font-size: 14px;
  }

  .auth-panel-card {
    width: 100%;
    padding: 20px 18px;
    border-radius: 20px;
  }
}
</style>
