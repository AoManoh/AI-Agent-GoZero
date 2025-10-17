import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createHead } from '@vueuse/head'
import ElementPlus from 'element-plus'
import ElementPlusX from 'vue-element-plus-x'
import 'element-plus/dist/index.css'
import './style.css'

const app = createApp(App)
const head = createHead()

app.use(router)
app.use(head)
app.use(ElementPlus)
app.use(ElementPlusX)
app.mount('#app')
