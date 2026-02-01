import { createRouter, createWebHistory } from 'vue-router'
// Lazy load components
const Login = () => import('../views/Login.vue')
const Signup = () => import('../views/Signup.vue')
const Chat = () => import('../views/Chat.vue')

const routes = [
    { path: '/', redirect: '/chat' },
    { path: '/login', component: Login },
    { path: '/signup', component: Signup },
    { path: '/chat', component: Chat, meta: { requiresAuth: true } },
]

const router = createRouter({
    history: createWebHistory(),
    routes,
})

// Navigation Guard
router.beforeEach((to, from, next) => {
    const isAuthenticated = !!localStorage.getItem('access_token')
    if (to.meta.requiresAuth && !isAuthenticated) {
        next('/login')
    } else {
        next()
    }
})

export default router
