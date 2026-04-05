import { useOsTheme } from 'naive-ui'
import { computed, ref, watch } from 'vue'

type ThemePreference = 'system' | 'light' | 'dark'

const storageKey = 'xlnetaccount-theme-preference'

const preference = ref<ThemePreference>(readPreference())

let persistenceBound = false

export function useThemeMode() {
  const osTheme = useOsTheme()

  if (!persistenceBound && typeof window !== 'undefined') {
    watch(preference, (value) => {
      window.localStorage.setItem(storageKey, value)
    })
    persistenceBound = true
  }

  const effectiveMode = computed<'light' | 'dark'>(() => {
    if (preference.value === 'system') {
      return osTheme.value === 'dark' ? 'dark' : 'light'
    }

    return preference.value
  })

  return {
    preference,
    effectiveMode,
    osTheme,
  }
}

function readPreference(): ThemePreference {
  if (typeof window === 'undefined') {
    return 'system'
  }

  const stored = window.localStorage.getItem(storageKey)
  if (stored === 'light' || stored === 'dark' || stored === 'system') {
    return stored
  }

  return 'system'
}
