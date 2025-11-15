<template>
  <div class="login-container">
    <div class="logo-section">
      <img src="/icons/instagram-icon.png" alt="Instagram" class="instagram-logo" />
      <h1 class="brand-name">hoshiBmatchi</h1>
    </div>

    <form @submit.prevent="handleLogin" class="form">
      <div class="form-group">
        <input 
          v-model="email"
          type="text" 
          placeholder="Phone number, username, or email" 
          class="input-field"
          required
        />
      </div>

      <div class="form-group">
        <input 
          v-model="password"
          type="password" 
          placeholder="Password" 
          class="input-field"
          required
        />
      </div>

      <button type="submit" class="btn btn-primary">Log in</button>
    </form>

    <button type="button" class="btn btn-google" @click="handleGoogleLogin">
      <span class="google-icon">G</span>
      Log in with Google
    </button>

    <div class="divider">
      <span>OR</span>
    </div>

    <button type="button" class="forgot-password-link" @click="navigateToForgot">
      Forgot password?
    </button>

    <div class="signup-section">
      <p class="signup-text">
        Don't have an account? 
        <button type="button" class="link-button" @click="navigateToRegister">
          Sign up
        </button>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { AuthPage } from '../types'
import { useRouter } from 'vue-router'

const router = useRouter()

const email = ref('')
const password = ref('')

const handleLogin = () => {
  console.log('Login attempt:', { email: email.value, password: password.value })
}

const handleGoogleLogin = () => {
  console.log('Google login initiated')
}

const navigateToRegister = () => {
  router.push('/register')
}

const navigateToForgot = () => {
  router.push('/forgot-password')
}
</script>

<style scoped>
.login-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.logo-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;
}

.instagram-logo {
  width: 80px;
  height: 80px;
  display: block;
  object-fit: contain;
}

.brand-name {
  font-size: 28px;
  font-weight: 300;
  letter-spacing: 1px;
  font-style: italic;
  color: var(--text-primary);
}

.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.input-field {
  padding: 12px 16px;
  background-color: var(--surface-light);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 14px;
  transition: all 0.2s ease;
}

.input-field::placeholder {
  color: var(--text-tertiary);
}

.input-field:focus {
  outline: none;
  background-color: var(--surface-dark);
  border-color: var(--primary-color);
}

.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  margin-top: 8px;
}

.btn-primary:hover {
  background-color: var(--primary-hover);
}

.btn-primary:active {
  transform: scale(0.98);
}

.btn-google {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background-color: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  margin-top: 12px;
}

.btn-google:hover {
  background-color: var(--surface-light);
  border-color: var(--primary-color);
}

.google-icon {
  font-weight: bold;
  color: var(--primary-color);
}

.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background-color: var(--border-color);
}

.forgot-password-link {
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 14px;
  cursor: pointer;
  transition: color 0.2s ease;
  padding: 0;
  text-align: center;
}

.forgot-password-link:hover {
  color: var(--primary-color);
}

.signup-section {
  text-align: center;
  padding-top: 12px;
}

.signup-text {
  font-size: 14px;
  color: var(--text-secondary);
}

.link-button {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: 14px;
  cursor: pointer;
  font-weight: 600;
  transition: color 0.2s ease;
  padding: 0;
  margin-left: 4px;
}

.link-button:hover {
  color: var(--primary-hover);
}


@media (max-width: 480px) {
  .login-container {
    gap: 20px;
  }

  .instagram-logo {
    width: 60px;
    height: 60px;
  }

  .brand-name {
    font-size: 24px;
  }

  .input-field {
    padding: 10px 14px;
    font-size: 13px;
  }
}
</style>
