import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      name: "MainPage",
      path: "/",
      component: () => import("../pages/MainPage/MainPage.vue"),
      meta: { authRequired: false },
      children: [
        {
          name: "EmptyPage",
          path: "",
          component: () => import("../pages/MainPage/Pages/EmptyPage.vue"),
        },
        {
          name: "ProfilePage",
          path: "profile/:id(\\d+)",
          component: () => import("../pages/MainPage/Pages/ProfilePage/ProfilePage.vue"),
        },
        {
          name: "EmployeesPage",
          path: "employees",
          component: () => import("../pages/MainPage/Pages/EmployeesPage/EmployeesPage.vue"),
        },
        {
          name: "DepartmentsPage",
          path: "departments",
          component: () => import("../pages/MainPage/Pages/DepartmentsPage/DepartmentsPage.vue"),
        },
      ],
    },
    {
      name: "AuthPage",
      path: "/login",
      component: () => import("../pages/AuthPage/AuthPage.vue"),
      meta: { authRequired: false },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFoundPage',
      component: () => import('../pages/NotFoundPage/NotFoundPage.vue')
    }
  ],
});

export default router
