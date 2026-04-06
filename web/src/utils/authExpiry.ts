import type { Router } from 'vue-router'

import { clearAuthToken } from './authToken'

let routerInstance: Router | null = null
let redirecting = false

export function registerAuthRouter(router: Router) {
  routerInstance = router
}

export async function handleUnauthorizedRedirect() {
  clearAuthToken()

  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('auth-expired'))
  }

  if (!routerInstance || redirecting) {
    return
  }

  const currentRoute = routerInstance.currentRoute.value
  if (currentRoute.name === 'login') {
    return
  }

  redirecting = true
  try {
    await routerInstance.replace({
      name: 'login',
      query: { next: currentRoute.fullPath },
    })
  }
  finally {
    redirecting = false
  }
}
