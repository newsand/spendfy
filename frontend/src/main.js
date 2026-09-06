import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import SKUList from './views/SKUList.vue'
import SKUDetail from './views/SKUDetail.vue'
import SKUForm from './views/SKUForm.vue'

const routes = [
  { path: '/', component: SKUList },
  { path: '/skus/new', component: SKUForm },
  { path: '/skus/:id', component: SKUDetail },
  { path: '/skus/:id/edit', component: SKUForm },
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

createApp(App).use(router).mount('#app')
