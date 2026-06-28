<script setup lang="ts">
import { NButton, NCard, NGrid, NGridItem, NIcon, NSpace, NTag, NText } from 'naive-ui'
import { useHead } from '@unhead/vue'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Fingerprint, KeyRound, LogIn, QrCode, ShieldCheck, Users } from 'lucide-vue-next'

import { usePlatformStore } from '@/stores/platform'

const router = useRouter()
const platformStore = usePlatformStore()
void platformStore.ensureLoaded().catch(() => {})

useHead({
  title: 'XLNetAccount — 统一账号与授权中心',
  meta: [
    { name: 'description', content: 'OAuth2 / OIDC 认证服务，提供账号管理、令牌管理、两步验证与通行密钥支持。' },
    { property: 'og:title', content: 'XLNetAccount — 统一账号与授权中心' },
  ],
})

const features = [
  { icon: KeyRound, title: 'OAuth2 授权', description: '支持授权码、隐式、密码与客户端凭证四种流程。' },
  { icon: Fingerprint, title: 'OIDC 身份认证', description: '基于 OpenID Connect 的身份层，兼容标准客户端。' },
  { icon: QrCode, title: '两步验证', description: 'TOTP 动态码，增强账号安全性。' },
  { icon: ShieldCheck, title: '通行密钥', description: 'WebAuthn 无密码认证，快速安全。' },
  { icon: Users, title: '账号管理', description: '统一管理用户、应用与令牌生命周期。' },
]

const platformName = computed(() => platformStore.displayName)
</script>

<template>
  <div class="landing-shell">
    <header class="landing-header">
      <div class="landing-header-inner page-container">
        <div class="landing-brand">
          <div class="brand-icon">{{ platformName.charAt(0) }}</div>
          <span class="brand-name">{{ platformName }}</span>
        </div>
        <nav class="landing-nav">
          <NButton quaternary @click="router.push({ name: 'login' })">登录</NButton>
          <NButton type="primary" @click="router.push({ name: 'register' })">注册</NButton>
        </nav>
      </div>
    </header>

    <main class="landing-main">
      <section class="hero-section page-container">
        <div class="hero-content">
          <NTag round size="small" type="info" class="hero-tag">OAuth2 / OIDC</NTag>
          <h1 class="hero-title">统一账号<br />与授权中心</h1>
          <p class="hero-description">
            安全、标准化的认证与授权基础设施。支持 OAuth2、OpenID Connect、
            两步验证与通行密钥，为您的应用提供可靠的账号体系。
          </p>
          <NSpace class="hero-actions">
            <NButton type="primary" size="large" @click="router.push({ name: 'login' })">
              <template #icon><NIcon><LogIn /></NIcon></template>
              立即登录
            </NButton>
            <NButton secondary size="large" @click="router.push({ name: 'register' })">创建账号</NButton>
          </NSpace>
        </div>
      </section>

      <section class="features-section page-container">
        <div class="features-header">
          <h2 class="features-title">核心能力</h2>
          <p class="features-subtitle">标准协议、安全优先、开箱即用。</p>
        </div>
        <NGrid cols="1 s:2 l:3" responsive="screen" :x-gap="12" :y-gap="12">
          <NGridItem v-for="feature in features" :key="feature.title">
            <NCard :title="feature.title" size="small" class="feature-card">
              <template #header-extra>
                <NIcon size="20" color="var(--color-accent-blue)">
                  <component :is="feature.icon" />
                </NIcon>
              </template>
              <NText depth="3">{{ feature.description }}</NText>
            </NCard>
          </NGridItem>
        </NGrid>
      </section>

      <section class="cta-section page-container">
        <NCard class="cta-card">
          <h2 class="cta-title">准备好开始了吗？</h2>
          <p class="cta-description">创建账号或集成您的应用，几分钟即可完成。</p>
          <NSpace>
            <NButton type="primary" size="large" @click="router.push({ name: 'register' })">创建账号</NButton>
            <NButton secondary size="large" @click="router.push({ name: 'login' })">登录</NButton>
          </NSpace>
        </NCard>
      </section>
    </main>

    <footer class="landing-footer">
      <div class="page-container footer-inner">
        <NText depth="3">{{ platformName }} &mdash; 统一账号与授权中心</NText>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.landing-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--color-canvas);
}

.landing-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--color-hairline);
}

.landing-header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 20px;
}

.landing-brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: rgba(52, 159, 244, 0.18);
  color: var(--color-accent-blue);
  font-weight: 700;
  font-size: 15px;
}

.brand-name {
  font-size: 16px;
  font-weight: 600;
}

.landing-nav {
  display: flex;
  align-items: center;
  gap: 8px;
}

.landing-main {
  flex: 1;
}

.hero-section {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: clamp(48px, 10vw, 96px) 20px;
  text-align: center;
}

.hero-content {
  max-width: 640px;
}

.hero-tag {
  margin-bottom: 16px;
}

.hero-title {
  margin: 0;
  font-size: clamp(40px, 6vw, 64px);
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1.08;
  color: var(--color-ink);
}

.hero-description {
  max-width: 480px;
  margin: 20px auto 0;
  font-size: clamp(15px, 1.4vw, 17px);
  line-height: 1.7;
  color: var(--color-charcoal);
}

.hero-actions {
  margin-top: 32px;
  justify-content: center;
}

.features-section {
  padding: 0 20px 64px;
}

.features-header {
  text-align: center;
  margin-bottom: 28px;
}

.features-title {
  margin: 0;
  font-size: 28px;
  font-weight: 650;
  color: var(--color-ink);
}

.features-subtitle {
  margin: 8px 0 0;
  color: var(--color-charcoal);
  font-size: 15px;
}

.feature-card :deep(.n-card-header__extra) {
  display: flex;
  align-items: center;
}

.cta-section {
  padding: 0 20px 64px;
}

.cta-card {
  text-align: center;
  padding: 48px 20px;
}

.cta-title {
  margin: 0;
  font-size: 28px;
  font-weight: 650;
  color: var(--color-ink);
}

.cta-description {
  margin: 10px 0 24px;
  color: var(--color-charcoal);
  font-size: 15px;
}

.landing-footer {
  border-top: 1px solid var(--color-hairline);
}

.footer-inner {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 56px;
  padding: 0 20px;
  text-align: center;
}

@media (max-width: 720px) {
  .hero-section {
    padding: 40px 16px;
  }

  .features-section {
    padding: 0 16px 40px;
  }

  .cta-section {
    padding: 0 16px 40px;
  }

  .landing-header-inner {
    padding: 0 12px;
  }
}
</style>
