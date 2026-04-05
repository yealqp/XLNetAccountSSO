import { computed, onMounted, onUnmounted, readonly, shallowRef } from 'vue'

export function useViewport(mobileBreakpoint = 960) {
  const width = shallowRef(typeof window === 'undefined' ? 1280 : window.innerWidth)

  function syncWidth() {
    width.value = window.innerWidth
  }

  onMounted(() => {
    syncWidth()
    window.addEventListener('resize', syncWidth, { passive: true })
  })

  onUnmounted(() => {
    window.removeEventListener('resize', syncWidth)
  })

  const isMobile = computed(() => width.value < mobileBreakpoint)

  return {
    width: readonly(width),
    isMobile,
  }
}
