<script setup lang="ts">
import { Button } from '@arco-design/web-vue'
import { useHead } from '@unhead/vue'
import { computed, onMounted, onUnmounted, ref, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'
import { ChevronsDown, Code, Fingerprint, KeyRound, LogIn, QrCode, Shield, ShieldCheck, Users } from 'lucide-vue-next'

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

const stats = [
  { icon: Code, value: '4', label: '授权流程' },
  { icon: Shield, value: '2', label: '身份协议' },
  { icon: Fingerprint, value: '2', label: '安全方式' },
  { icon: KeyRound, value: '3', label: '令牌类型' },
]

const runtimeText = ref('')
const FOUNDING_DATE = '2024-06-01T00:00:00+08:00'
const ICP_BEIAN = '沪ICP备2025144886号-3'

function calculateRuntime() {
  const start = new Date(FOUNDING_DATE)
  const now = new Date()
  const diff = now.getTime() - start.getTime()
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)
  runtimeText.value = `${days} 天 ${hours} 小时 ${minutes} 分钟`
}

onMounted(() => {
  calculateRuntime()
  setInterval(calculateRuntime, 60000)
})

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
      <div class="header-inner">
        <div class="landing-brand">
          <div class="brand-icon">{{ platformName.charAt(0) }}</div>
          <span class="brand-name">{{ platformName }}</span>
        </div>
        <nav class="landing-nav">
          <Button type="primary" @click="router.push({ name: 'overview' })">控制台</Button>
        </nav>
      </div>
    </header>

    <main class="landing-main">
      <!-- Hero -->
      <section class="hero-section">
        <div class="hero-bg">
          <div class="hero-gradient"></div>
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

        <div class="hero-body">
          <div class="hero-content">
            <div class="hero-status">
              <span class="status-badge">
                <span class="status-dot dot-teal"></span>
                OAuth2 / OIDC
              </span>
              <span class="status-badge">
                <span class="status-dot dot-blue"></span>
                安全认证
              </span>
            </div>

            <h1 class="hero-title">
              <span class="title-gradient">统一账号<br />与授权中心</span>
            </h1>

            <p class="hero-desc">
              专注于 <span class="text-accent">OAuth2</span> 与
              <span class="text-accent">OpenID Connect</span> 标准协议，
              <br class="hide-mobile">
              提供安全、可靠的统一认证与授权基础设施。
            </p>

            <div class="hero-buttons">
              <Button type="primary" size="large" class="btn-primary" @click="router.push({ name: 'login' })">
                <LogIn :size="18" style="vertical-align: sub;" />
                立即登录
              </Button>
              <Button type="secondary" size="large" class="btn-secondary" @click="router.push({ name: 'register' })">
                创建账号
              </Button>
            </div>
          </div>
        </div>

        <div class="scroll-indicator">
          <ChevronsDown :size="24" />
        </div>
      </section>

      <!-- Features -->
      <section class="features-section">
        <div class="section-inner">
          <div class="section-header">
            <h2 class="section-title">核心能力</h2>
            <p class="section-subtitle">标准协议、安全优先、开箱即用。</p>
          </div>

          <div class="features-grid">
            <div v-for="feature in features" :key="feature.title" class="feature-card-wrapper">
              <div class="glass-card feature-card">
                <div class="feature-icon-box">
                  <component :is="feature.icon" :size="22" />
                </div>
                <h3 class="feature-title">{{ feature.title }}</h3>
                <p class="feature-desc">{{ feature.description }}</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Stats -->
      <section class="stats-section">
        <div class="section-inner">
          <div class="section-header">
            <h2 class="section-title">数据一览</h2>
            <p class="section-subtitle">标准协议支持与安全能力。</p>
          </div>
          <div class="stats-grid">
            <div v-for="stat in stats" :key="stat.label" class="stat-card-wrapper">
              <div class="glass-card stat-card">
                <div class="stat-icon-box">
                  <component :is="stat.icon" :size="28" />
                </div>
                <div class="stat-value">{{ stat.value }}</div>
                <div class="stat-label">{{ stat.label }}</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- CTA -->
      <section class="cta-section">
        <div class="cta-banner">
          <div class="cta-glow glow-1"></div>
          <div class="cta-glow glow-2"></div>
          <div class="cta-body">
            <h2 class="cta-heading">
              准备好<br class="show-mobile"> 开始了吗？
            </h2>
            <p class="cta-desc">创建账号或集成您的应用，几分钟即可完成。</p>
            <div class="cta-buttons">
              <Button type="primary" size="large" class="btn-primary" @click="router.push({ name: 'register' })">创建账号</Button>
              <Button type="secondary" size="large" class="btn-secondary" @click="router.push({ name: 'login' })">登录</Button>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="landing-footer">
      <div class="footer-inner">
        <div class="footer-left">
          <p class="footer-copyright">Copyright &copy; 2026 XLNet</p>
          <p class="footer-meta">{{ platformName }} &mdash; 已稳定运行 {{ runtimeText }}</p>
        </div>
        <div class="footer-right">
          <a
            href="https://beian.miit.gov.cn"
            target="_blank"
            rel="noopener noreferrer"
            class="footer-icp"
          >{{ ICP_BEIAN }}</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
/* ========================================
   Base
   ======================================== */
.landing-shell {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  overflow-y: auto;
  scrollbar-gutter: stable;
  background: #fafafa;
}

/* ========================================
   Header
   ======================================== */
.landing-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  background: rgba(250, 250, 250, 0.7);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 24px;
  max-width: 1200px;
  margin: 0 auto;
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
  background: rgba(45, 212, 191, 0.15);
  color: #0d9488;
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

/* ========================================
   Main
   ======================================== */
.landing-main {
  flex: 1;
}

/* ========================================
   Hero
   ======================================== */
.hero-section {
  position: relative;
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 24px;
  text-align: center;
  overflow: hidden;
}

.hero-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.hero-gradient {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.08), transparent 40%, rgba(59, 130, 246, 0.06));
}

.glow-sphere {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.08;
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
  top: 50%;
  left: 60%;
  animation: float-glow 20s ease-in-out infinite alternate;
}

@keyframes float-glow {
  0%   { transform: translate(0, 0) scale(1); }
  100% { transform: translate(30px, -40px) scale(1.15); }
}

.hero-body {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.hero-content {
  max-width: 720px;
  margin: 0 auto;
}

/* status badges */
.hero-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-bottom: 24px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.55);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-teal {
  background: #14b8a6;
  box-shadow: 0 0 8px rgba(20, 184, 166, 0.4);
  animation: pulse-dot 2s ease-in-out infinite;
}

.dot-blue {
  background: #3b82f6;
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.4);
  animation: pulse-dot 2s ease-in-out infinite 0.5s;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 0.6; transform: scale(0.9); }
  50%      { opacity: 1; transform: scale(1.1); }
}

.hero-title {
  margin: 0;
  font-size: clamp(36px, 6vw, 64px);
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.08;
  color: #0d0d11;
}

.title-gradient {
  background: linear-gradient(135deg, #0d9488, #2563eb, #7c3aed);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-desc {
  max-width: 560px;
  margin: 20px auto 0;
  font-size: clamp(15px, 1.4vw, 18px);
  line-height: 1.7;
  color: rgba(0, 0, 0, 0.5);
}

.hero-desc .text-accent {
  color: #0d9488;
  font-weight: 600;
}

.show-mobile {
  display: none;
}

.hero-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 36px;
}

.hero-buttons .btn-primary {
  height: 50px;
  padding: 0 28px;
  font-size: 15px;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(13, 148, 136, 0.15);
}

.hero-buttons .btn-secondary {
  height: 50px;
  font-size: 15px;
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
  color: rgba(0, 0, 0, 0.25);
}

@keyframes bounce-down {
  0%, 100% { transform: translateY(0); opacity: 0.3; }
  50%      { transform: translateY(8px); opacity: 0.7; }
}

/* ========================================
   Features
   ======================================== */
.features-section {
  padding: 80px 24px 100px;
  background: linear-gradient(180deg, #fafafa 0%, #f1f5f9 100%);
}

.section-inner {
  max-width: 1200px;
  margin: 0 auto;
}

.section-header {
  text-align: center;
  margin-bottom: 48px;
}

.section-title {
  margin: 0;
  font-size: 30px;
  font-weight: 800;
  color: #0d0d11;
}

.section-subtitle {
  margin: 10px 0 0;
  font-size: 16px;
  color: rgba(0, 0, 0, 0.5);
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.feature-card-wrapper {
  padding-top: 4px;
}

.glass-card {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 14px;
  padding: 24px;
  transition: all 0.3s ease;
  height: 100%;
}

.glass-card:hover {
  transform: translateY(-4px);
  border-color: rgba(13, 148, 136, 0.2);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.08);
}

.feature-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: rgba(13, 148, 136, 0.1);
  color: #0d9488;
  margin-bottom: 16px;
}

.feature-title {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 700;
  color: #0d0d11;
}

.feature-desc {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: rgba(0, 0, 0, 0.5);
}

/* ========================================
   Stats
   ======================================== */
.stats-section {
  padding: 80px 24px;
  background: #ffffff;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  text-align: center;
  padding: 32px 24px;
}

.stat-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 16px;
  margin: 0 auto 16px;
  background: rgba(13, 148, 136, 0.1);
  color: #0d9488;
}

.stat-value {
  font-size: 32px;
  font-weight: 800;
  color: #0d0d11;
  line-height: 1;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.5);
}

/* ========================================
   CTA
   ======================================== */
.cta-section {
  padding: 0;
}

.cta-banner {
  position: relative;
  overflow: hidden;
  padding: 80px 24px;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  border-top: none;
}

.cta-glow {
  position: absolute;
  width: 400px;
  height: 400px;
  border-radius: 50%;
  filter: blur(100px);
  pointer-events: none;
}

.cta-glow.glow-1 {
  top: -120px;
  right: -80px;
  background: rgba(45, 212, 191, 0.25);
}

.cta-glow.glow-2 {
  bottom: -140px;
  left: -60px;
  background: rgba(59, 130, 246, 0.15);
}

.cta-body {
  position: relative;
  z-index: 1;
  max-width: 640px;
  margin: 0 auto;
  text-align: center;
}

.cta-heading {
  margin: 0;
  font-size: clamp(26px, 4vw, 36px);
  font-weight: 800;
  color: #fcfdff;
  line-height: 1.2;
}

.cta-desc {
  margin: 14px 0 0;
  font-size: 16px;
  color: rgba(252, 253, 255, 0.55);
}

.cta-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 32px;
}

.cta-buttons .btn-primary {
  height: 50px;
  padding: 0 28px;
  font-size: 15px;
  border-radius: 10px;
  background: #fcfdff;
  color: #0f172a;
  font-weight: 700;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

.cta-buttons .btn-secondary {
  height: 50px;
  font-size: 15px;
}

/* ========================================
   Footer
   ======================================== */
.landing-footer {
  background: #f1f5f9;
  backdrop-filter: blur(8px);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.footer-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.footer-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.footer-copyright {
  margin: 0;
  font-size: 14px;
  color: rgba(0, 0, 0, 0.5);
}

.footer-meta {
  margin: 0;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.35);
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.footer-icp {
  font-size: 13px;
  color: rgba(0, 0, 0, 0.45);
  text-decoration: none;
  transition: color 0.2s;
}

.footer-icp:hover {
  color: #0d9488;
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 720px) {
  .hero-section {
    padding: 60px 16px;
  }

  .hero-status {
    flex-wrap: wrap;
    gap: 10px;
  }

  .hero-buttons {
    flex-direction: column;
    width: 100%;
    max-width: 280px;
    margin-left: auto;
    margin-right: auto;
  }

  .hero-buttons .btn-primary,
  .hero-buttons .btn-secondary {
    width: 100%;
  }

  .glow-sphere {
    filter: blur(60px);
  }
  .sphere-1 { width: 300px; height: 300px; }
  .sphere-2 { width: 220px; height: 220px; }
  .sphere-3 { width: 180px; height: 180px; }

  .features-section {
    padding: 48px 16px 64px;
  }

  .features-grid {
    grid-template-columns: 1fr;
  }

  .stats-section {
    padding: 48px 16px;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .cta-banner {
    padding: 56px 16px;
  }

  .cta-buttons {
    flex-direction: column;
    max-width: 280px;
    margin-left: auto;
    margin-right: auto;
  }

  .cta-buttons .btn-primary,
  .cta-buttons .btn-secondary {
    width: 100%;
  }

  .footer-inner {
    flex-direction: column;
    gap: 12px;
    text-align: center;
  }

  .header-inner {
    padding: 0 12px;
  }

  .hide-mobile {
    display: none;
  }

  .show-mobile {
    display: inline;
  }
}
</style>
