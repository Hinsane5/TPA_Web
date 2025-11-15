import { createRouter, createWebHistory } from "vue-router";
import LoginPage from "../components/LoginPage.vue";
import RegisterPage from "../components/RegisterPage.vue";
import ForgotPasswordPage from "../components/ForgotPasswordPage.vue";
import TwoFAVerificationPage from "../components/TwoFAVerificationPage.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      redirect: "/login",
    },
    {
      path: "/login",
      name: "login",
      component: LoginPage,
    },
    {
      path: "/register",
      name: "register",
      component: RegisterPage,
    },
    {
      path: "/forgot-password",
      name: "forgot-password",
      component: ForgotPasswordPage,
    },
    {
      path: "/verify-2fa",
      name: "verify-2fa",
      component: TwoFAVerificationPage,
    },
  ],
});

export default router;
