import { createRouter, createWebHistory } from "vue-router";
import LoginPage from "../components/LoginPage.vue";
import RegisterPage from "../components/RegisterPage.vue";
import ForgotPasswordPage from "../components/ForgotPasswordPage.vue";
import TwoFAVerificationPage from "../components/TwoFAVerificationPage.vue";
import ResetPasswordPage from "../components/ResetPasswordPage.vue";
import HomePage from "../components/HomePage.vue";
import ExplorePage from "../components/ExplorePage.vue";
import ReelsPage from "../components/ReelsPage.vue";
import MessagesPage from "../components/MessagesPage.vue";
import ProfilePage from "../components/ProfilePage.vue";
import DashboardLayout from "../components/DashboardLayout.vue";

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
    {
      path: "/reset-password",
      name: "reset-password",
      component: ResetPasswordPage,
    },
    {
      path: "/dashboard",
      component: DashboardLayout,
      redirect: "/dashboard/home",
      children: [
        {
          path: "home",
          name: "home",
          component: HomePage,
        },
        {
          path: "explore",
          name: "explore",
          component: ExplorePage,
        },
        {
          path: "reels",
          name: "reels",
          component: ReelsPage,
        },
        {
          path: "messages",
          name: "messages",
          component: MessagesPage,
        },
        {
          path: "profile",
          name: "profile",
          component: ProfilePage,
        },
      ],
    },
  ],
});

export default router;
