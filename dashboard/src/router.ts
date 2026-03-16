// import { createRouter, createWebHistory } from "vue-router";
// import { routes, handleHotUpdate } from "vue-router/auto-routes";
// import { useAuthStore } from "./stores/auth";

// export const router = createRouter({
//   history: createWebHistory(),
//   routes,
// });

// router.beforeEach((to ) => {
//   const authStore = useAuthStore();

//   const requiresAuth = to.meta.requiresAuth;

//   if (requiresAuth && !authStore.authenticated) {
//     return ({ path: "/" });
//   }

// });

// if (import.meta.hot) {
//   handleHotUpdate(router);
// }
import { createRouter, createWebHistory } from "vue-router";
import { routes, handleHotUpdate } from "vue-router/auto-routes";
import { useAuthStore } from "./stores/auth";

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const authStore = useAuthStore();
  const requiresAuth = to.meta.requiresAuth;

  // Redirect unauthenticated users trying to access protected pages
  if (requiresAuth && !authStore.authenticated) {
    return { path: "/" };
  }

  // Redirect authenticated users away from the landing page
  if (to.path === "/" && authStore.authenticated) {
    return { path: "/dashboard" };
  }
});

if (import.meta.hot) {
  handleHotUpdate(router);
}