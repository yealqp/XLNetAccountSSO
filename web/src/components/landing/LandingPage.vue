<script setup lang="ts">
import { NButton, NCard, NGrid, NGridItem, NIcon, NSpace, NTag, NText } from 'naive-ui'
import { useHead } from '@unhead/vue'
import { computed, onMounted, onUnmounted, ref, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'
import { ChevronsDown, Fingerprint, KeyRound, LogIn, QrCode, ShieldCheck, Users } from 'lucide-vue-next'

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
const shellRef = useTemplateRef<HTMLElement>('shell')

// --- parallax & snap scroll ---
const parallaxOffset = ref(0)
const isScrolling = ref(false)
const isMobile = ref(false)
let isMounted = true

function smoothScrollTo(targetPosition: number) {
  const shell = shellRef.value
  if (!shell || isScrolling.value) return
  isScrolling.value = true
  const startPosition = shell.scrollTop
  const distance = targetPosition - startPosition
  const duration = 600
  let start: number | null = null

  function step(timestamp: number) {
    if (!isMounted) return
    const s = shellRef.value
    if (!s) { isScrolling.value = false; return }
    if (!start) start = timestamp
    const progress = Math.min((timestamp - start) / duration, 1)
    const ease = 1 - Math.pow(1 - progress, 4)
    s.scrollTop = startPosition + distance * ease
    if (progress < 1) {
      requestAnimationFrame(step)
    } else {
      isScrolling.value = false
    }
  }
  requestAnimationFrame(step)
}

function handleWheel(e: WheelEvent) {
  if (isMobile.value) return
  const shell = shellRef.value
  if (!shell) return
  const heroHeight = shell.clientHeight
  const scrollY = shell.scrollTop

  if (isScrolling.value) {
    e.preventDefault()
    return
  }

  if (scrollY < heroHeight) {
    e.preventDefault()
    if (e.deltaY > 0) {
      smoothScrollTo(heroHeight)
    } else if (e.deltaY < 0 && scrollY > 0) {
      smoothScrollTo(0)
    }
    return
  }

  if (scrollY < heroHeight + 120 && e.deltaY < 0) {
    e.preventDefault()
    smoothScrollTo(0)
  }
}

let touchStartY = 0

function handleTouchStart(e: TouchEvent) {
  if (isMobile.value) return
  touchStartY = e.touches[0].clientY
}

function handleTouchMove(e: TouchEvent) {
  if (isMobile.value) return
  const shell = shellRef.value
  if (!shell) return
  const heroHeight = shell.clientHeight
  const scrollY = shell.scrollTop
  if ((scrollY < heroHeight || scrollY < heroHeight + 120) && !isScrolling.value) {
    e.preventDefault()
  }
}

function handleTouchEnd(e: TouchEvent) {
  if (isMobile.value) return
  const shell = shellRef.value
  if (!shell) return
  const heroHeight = shell.clientHeight
  const scrollY = shell.scrollTop
  if (isScrolling.value) return

  const deltaY = touchStartY - e.changedTouches[0].clientY
  const threshold = 50

  if (scrollY < heroHeight) {
    if (deltaY > threshold) {
      smoothScrollTo(heroHeight)
    } else if (deltaY < -threshold && scrollY > 0) {
      smoothScrollTo(0)
    }
    return
  }

  if (scrollY < heroHeight + 120 && deltaY < -threshold) {
    smoothScrollTo(0)
  }
}

function updateParallax() {
  parallaxOffset.value = (shellRef.value?.scrollTop ?? 0) * 0.3
}

onMounted(() => {
  const shell = shellRef.value
  if (!shell) return

  isMobile.value = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) || window.innerWidth < 768

  shell.addEventListener('scroll', updateParallax, { passive: true })
  shell.addEventListener('wheel', handleWheel, { passive: false })
  shell.addEventListener('touchstart', handleTouchStart, { passive: true })
  shell.addEventListener('touchmove', handleTouchMove, { passive: false })
  shell.addEventListener('touchend', handleTouchEnd, { passive: true })
})

onUnmounted(() => {
  isMounted = false
  const shell = shellRef.value
  if (!shell) return
  shell.removeEventListener('scroll', updateParallax)
  shell.removeEventListener('wheel', handleWheel)
  shell.removeEventListener('touchstart', handleTouchStart)
  shell.removeEventListener('touchmove', handleTouchMove)
  shell.removeEventListener('touchend', handleTouchEnd)
})
</script>

<template>
  <div ref="shell" class="landing-shell">
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
      <!-- Hero Section with snap + parallax -->
      <section class="hero-section">
        <div class="hero-bg-glow">
          <div
            class="glow-sphere sphere-1"
            :style="{ transform: `translateY(${parallaxOffset * 0.3}px)` }"
          ></div>
          <div
            class="glow-sphere sphere-2"
            :style="{ transform: `translateY(${parallaxOffset * 0.2}px)` }"
          ></div>
          <div
            class="glow-sphere sphere-3"
            :style="{ transform: `translateY(${parallaxOffset * 0.4}px)` }"
          ></div>
        </div>

        <div class="hero-content">
          <NTag round size="small" class="hero-tag">OAuth2 / OIDC</NTag>
          <h1 class="hero-title">
            <span class="title-gradient">统一账号<br />与授权中心</span>
          </h1>
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

        <div class="scroll-indicator">
          <NIcon size="24"><ChevronsDown /></NIcon>
        </div>
      </section>

      <!-- Features -->
      <section class="features-section page-container">
        <div class="features-header">
          <h2 class="features-title">核心能力</h2>
          <p class="features-subtitle">标准协议、安全优先、开箱即用。</p>
        </div>
        <NGrid cols="1 s:2 l:3" responsive="screen" :x-gap="12" :y-gap="12">
          <NGridItem v-for="feature in features" :key="feature.title">
            <NCard size="small" class="feature-card">
              <div class="feature-card-inner">
                <div class="feature-icon-box">
                  <NIcon size="22">
                    <component :is="feature.icon" />
                  </NIcon>
                </div>
                <div class="feature-text">
                  <NText class="feature-card-title">{{ feature.title }}</NText>
                  <NText depth="3" class="feature-card-desc">{{ feature.description }}</NText>
                </div>
              </div>
            </NCard>
          </NGridItem>
        </NGrid>
      </section>

      <!-- CTA -->
      <section class="cta-section">
        <div class="cta-banner">
          <div class="cta-bg-glow">
            <div class="cta-glow glow-1"></div>
            <div class="cta-glow glow-2"></div>
          </div>
          <div class="cta-body page-container">
            <h2 class="cta-title">准备好开始了吗？</h2>
            <p class="cta-description">创建账号或集成您的应用，几分钟即可完成。</p>
            <NSpace class="cta-actions">
              <NButton type="primary" size="large" @click="router.push({ name: 'register' })">创建账号</NButton>
              <NButton secondary size="large" @click="router.push({ name: 'login' })">登录</NButton>
            </NSpace>
          </div>
        </div>
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
  overflow-y: auto;
  scrollbar-gutter: stable;
  background: var(--color-canvas);
}

/* --- Header --- */
.landing-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(12px);
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

/* --- Main --- */
.landing-main {
  flex: 1;
}

/* --- Hero --- */
.hero-section {
  position: relative;
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  text-align: center;
  overflow: hidden;
  background:
    radial-gradient(ellipse at 20% 50%, rgba(13, 148, 136, 0.08) 0%, transparent 60%),
    radial-gradient(ellipse at 80% 50%, rgba(59, 130, 246, 0.06) 0%, transparent 60%);
}

.hero-bg-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.glow-sphere {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.12;
  will-change: transform;
}

.sphere-1 {
  width: 500px;
  height: 500px;
  background: #14b8a6;
  top: -120px;
  right: -80px;
  animation: float-glow 18s ease-in-out infinite alternate;
}

.sphere-2 {
  width: 360px;
  height: 360px;
  background: #3b82f6;
  bottom: -60px;
  left: 5%;
  animation: float-glow 22s ease-in-out infinite alternate-reverse;
}

.sphere-3 {
  width: 280px;
  height: 280px;
  background: #a855f7;
  top: 45%;
  left: 50%;
  animation: float-glow 20s ease-in-out infinite alternate;
}

@keyframes float-glow {
  0% {
    transform: translate(0, 0) scale(1);
  }
  100% {
    transform: translate(30px, -40px) scale(1.15);
  }
}

.hero-content {
  position: relative;
  z-index: 1;
  max-width: 640px;
}

.hero-tag {
  margin-bottom: 20px;
}

.hero-title {
  margin: 0;
  font-size: clamp(36px, 6vw, 64px);
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.08;
}

.title-gradient {
  background: linear-gradient(135deg, #2dd4bf, #60a5fa, #c084fc);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
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

.scroll-indicator {
  position: absolute;
  bottom: 32px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  z-index: 1;
  animation: bounce-down 2s ease-in-out infinite;
  color: rgba(252, 253, 255, 0.4);
}

@keyframes bounce-down {
  0%, 100% { transform: translateY(0); opacity: 0.4; }
  50% { transform: translateY(8px); opacity: 0.8; }
}

/* --- Features --- */
.features-section {
  padding: 0 20px 80px;
  position: relative;
  z-index: 1;
}

.features-header {
  text-align: center;
  margin-bottom: 32px;
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

.feature-card {
  border-radius: var(--radius-md);
}

.feature-card-inner {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.feature-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 10px;
  background: rgba(45, 212, 191, 0.12);
  color: #2dd4bf;
}

.feature-text {
  flex: 1;
  min-width: 0;
}

.feature-card-title {
  display: block;
  font-weight: 600;
  font-size: 14px;
  color: var(--color-ink);
  margin-bottom: 4px;
}

.feature-card-desc {
  display: block;
  font-size: 13px;
  line-height: 1.5;
}

/* --- CTA --- */
.cta-section {
  padding: 0;
}

.cta-banner {
  position: relative;
  overflow: hidden;
  padding: 64px 0;
  border-top: 1px solid var(--color-hairline);
}

.cta-bg-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.cta-glow {
  position: absolute;
  width: 400px;
  height: 400px;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.08;
}

.cta-glow.glow-1 {
  top: -100px;
  right: -80px;
  background: #14b8a6;
}

.cta-glow.glow-2 {
  bottom: -120px;
  left: -60px;
  background: #3b82f6;
}

.cta-body {
  position: relative;
  z-index: 1;
  text-align: center;
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

.cta-actions {
  justify-content: center;
}

/* --- Footer --- */
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

/* --- Responsive --- */
@media (max-width: 720px) {
  .hero-section {
    padding: 60px 16px;
  }

  .glow-sphere {
    filter: blur(60px);
  }

  .sphere-1 {
    width: 300px;
    height: 300px;
  }

  .sphere-2 {
    width: 220px;
    height: 220px;
  }

  .sphere-3 {
    width: 180px;
    height: 180px;
  }

  .features-section {
    padding: 0 16px 48px;
  }

  .cta-banner {
    padding: 48px 0;
  }

  .landing-header-inner {
    padding: 0 12px;
  }
}
</style>
