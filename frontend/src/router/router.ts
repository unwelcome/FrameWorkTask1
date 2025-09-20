import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // {
    //   name: "MainPage",
    //   path: "/",
    //   component: () => import("../")
    // },
    {
      name: "AuthPage",
      path: "/login",
      component: () => import("../pages/AuthPage/AuthPage.vue"),
      meta: { authRequired: false },
    }
  ],
});

export default router
