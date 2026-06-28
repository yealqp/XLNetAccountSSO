import { createApp } from 'vue'
import { createHead } from '@unhead/vue/client'

import App from './App.vue'
import './assets/main.css'
import router from './router'
import { registerAuthRouter } from './utils/authExpiry'
import { pinia } from './stores/pinia'
import { useSessionStore } from './stores/session'

const app = createApp(App)
const head = createHead()
app.use(head)
const sessionStore = useSessionStore(pinia)

registerAuthRouter(router)

window.addEventListener('auth-expired', () => {
  sessionStore.clear()
})

app.use(pinia)
app.use(router)
app.mount('#app')
