<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../services/apiService' 
import { validateEmail as validateEmailUtil } from '../utils/validation' 

const router = useRouter()

const email = ref('')
const isLoading = ref(false)
const error = ref('')
const message = ref('')

const validateEmail = () => {
  error.value = '' 
  const validation = validateEmailUtil(email.value.trim())
  if (!validation.isValid) {
    error.value = validation.message
    return false
  }
  return true
}

const handleResetPassword = async () => {
  error.value = ''
  message.value = ''
  
  if (!validateEmail()) {
    return
  }

  const trimmedEmail = email.value.trim() 

  isLoading.value = true
  try {
    const response = await authApi.forgotPassword(trimmedEmail) 
    console.log('RequestPasswordReset successful', response.data)
    message.value = response.data.message
  } catch (err: any) {
    console.error('Failed to send reset link:', err)
    if (err.response && err.response.data.error) {
      error.value = err.response.data.error
    } else {
      error.value = 'An unknown error occurred.'
    }
  } finally {
    isLoading.value = false
  }
}

const navigateToRegister = () => {
  router.push('/register') 
}

const navigateToLogin = () => {
  router.push('/login')
}
</script>

<template>
  <div class="forgot-password-container">
    <div class="lock-icon-section">
      <img src="/icons/lock-icon.png" alt="Lock" class="lock-icon" />
    </div>

    <div class="content-section">
      <h2 class="heading">Trouble logging in?</h2>
      <p class="description">
        Enter your email, phone, or username and we'll send you a link to get back into your account.
      </p>
    </div>

    <form class="form" @submit.prevent="handleResetPassword">
      <div class="form-group">
        <input 
          v-model="email"
          type="email" 
          placeholder="Enter your email" 
          class="input-field"
          required
        />
      </div>

      <button type="submit" class="btn btn-primary">Send login link</button>
    </form>

    <div class="divider">
      <span>OR</span>
    </div>

    <button type="button" class="action-button" @click="navigateToRegister">
      Create new account
    </button>

    <button type="button" class="back-button" @click="navigateToLogin">
      Back to login
    </button>
  </div>
</template>

<style scoped>
.forgot-password-container {
  display: flex;
  flex-direction: column;
  gap: 28px;
  padding: 40px 32px;
  background-color: var(--surface-dark);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.lock-icon-section {
  display: flex;
  justify-content: center;
}

.lock-icon {
  width: 80px;
  height: 80px;
  object-fit: contain;
}

.content-section {
  text-align: center;
}

.heading {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--text-primary);
}

.description {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
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
  background-color: #333333;
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
}

.btn-primary:hover {
  background-color: var(--primary-hover);
}

.btn-primary:active {
  transform: scale(0.98);
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

.action-button {
  padding: 10px 16px;
  background-color: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: color 0.2s ease;
  text-align: center;
}

.action-button:hover {
  color: var(--primary-color);
}

.back-button {
  padding: 10px 16px;
  background-color: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s ease;
  text-align: center;
}

.back-button:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.back-button:active {
  transform: scale(0.98);
}

/* Responsive */
@media (max-width: 480px) {
  .forgot-password-container {
    gap: 24px;
    padding: 32px 20px;
  }

  .lock-icon {
    width: 60px;
    height: 60px;
  }

  .heading {
    font-size: 18px;
  }

  .description {
    font-size: 12px;
  }

  .input-field {
    padding: 10px 14px;
    font-size: 13px;
  }
}

@media (max-width: 360px) {
  .forgot-password-container {
    padding: 24px 16px;
  }

  .lock-icon {
    width: 50px;
    height: 50px;
  }

  .heading {
    font-size: 16px;
  }

  .description {
    font-size: 11px;
  }
}
</style>
