import './assets/styles/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router/router'

import textButton from './lib/textButton.vue'
import iconButton from './lib/iconButton.vue'
import iconTextButton from './lib/iconTextButton.vue'
import searchBar from './lib/searchBar.vue'
import userAvatar from './lib/userAvatar.vue'

const app = createApp(App)
app.use(createPinia())
app.use(router)

app.component("TextButton", textButton)
app.component("IconButton", iconButton)
app.component("IconTextButton", iconTextButton)
app.component("SearchBar", searchBar)
app.component("UserAvatar", userAvatar)

app.mount('#app')
