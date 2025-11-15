<template>
  <div class="app-container">
    <div class="auth-wrapper">
      <LoginPage v-if="currentPage === 'login'" @navigate="handleNavigate" />
      <RegisterPage v-else-if="currentPage === 'register'" @navigate="handleNavigate" />
      <ForgotPasswordPage v-else-if="currentPage === 'forgot-password'" @navigate="handleNavigate" />
      <!-- added 2FA verification page -->
      <TwoFAVerificationPage v-else-if="currentPage === '2fa-verification'" @navigate="handleNavigate" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { AuthPage } from './types'
import LoginPage from './components/LoginPage.vue'
import RegisterPage from './components/RegisterPage.vue'
import ForgotPasswordPage from './components/ForgotPasswordPage.vue'
import TwoFAVerificationPage from './components/TwoFAVerificationPage.vue'

const currentPage = ref<AuthPage>('login')

const handleNavigate = (page: AuthPage) => {
  currentPage.value = page
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  --primary-color: #5b5bff;
  --primary-hover: #4949ff;
  --background-dark: #0a0a0a;
  --surface-dark: #1a1a1a;
  --surface-light: #2a2a2a;
  --border-color: #3a3a3a;
  --text-primary: #ffffff;
  --text-secondary: #b0b0b0;
  --text-tertiary: #808080;
  --success-color: #22c55e;
  --error-color: #ef4444;
  --warning-color: #f59e0b;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: var(--background-dark);
  color: var(--text-primary);
}

.app-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
  background-color: var(--background-dark);
}

.auth-wrapper {
  width: 100%;
  max-width: 450px;
}

/* Animations */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.auth-wrapper > * {
  animation: slideIn 0.3s ease-out;
}

/* Responsive */
@media (max-width: 768px) {
  .app-container {
    min-height: auto;
    padding: 16px;
  }

  .auth-wrapper {
    max-width: 100%;
  }
}
</style>
