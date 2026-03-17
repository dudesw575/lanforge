import { createRouter, createWebHistory } from "vue-router";
import { routes, handleHotUpdate } from "vue-router/auto-routes";
import { useAuthStore } from "./stores/auth";

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const authStore = useAuthStore()

  if (!authStore.ready) return false

  const requiresAuth = to.matched.some(
    (record) => record.meta.requiresAuth
  )

  if (requiresAuth && !authStore.isAuthenticated) {
    console.log("❌ User not authenticated or token expired → redirecting")
    return { path: "/" }
  }

  if (to.path === "/" && authStore.isAuthenticated) {
    return { path: "/dashboard" }
  }
})

if (import.meta.hot) {
  handleHotUpdate(router);
}
