<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authApi } from '../services/apiService'
import { validate2FACode } from '../utils/validation'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const otpCode = ref('')
const email = ref('')
const isLoading = ref(false)
const error = ref('')

onMounted(() => {
  email.value = route.query.email as string || ''
  if (!email.value) {
    router.push('/login')
  }
})

const handleSubmit = async () => {
  error.value = ''
  isLoading.value = true

  try {
    const data = {
      email: email.value,
      otp_code: otpCode.value.trim()
    }

    const response = await authApi.verify2FA(data)

    if (response.data.access_token) {
        console.log("2FA successful, tokens:", response.data);

        if (response.data.access_token && response.data.refresh_token) {
          localStorage.setItem('accessToken', response.data.access_token)
          localStorage.setItem('refreshToken', response.data.refresh_token)
        }

        router.push("/dashboard");

      } else {
        error.value = "Invalid token response from server.";
      }

  } catch (err: any) {
    console.error('2FA verification failed:', err)
    if (err.response && err.response.data.error) {
      error.value = err.response.data.error
    } else {
      error.value = 'An unknown error occurred. Please try again.'
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
  <div class="twofa-container">
    <div class="lock-icon-section">
      <img src="/icons/lock-icon.png" alt="Lock" class="lock-icon" />
    </div>

    <div class="content-section">
      <h1 class="title">Verify Your Identity</h1>
      <p class="description">
        Enter the 6-digit code from your authenticator app or sent to your registered email.
      </p>
    </div>

    <form class="form" @submit.prevent="handleSubmit">
      <div class="form-group">
        <label class="form-label">2FA Code</label>
        <input 
          v-model="otpCode"
          type="text" 
          placeholder="000000" 
          maxlength="6"
          inputmode="numeric"
          class="input-field twofa-input"
        />
        <span v-if="error" class="error-message">{{ error }}</span>
      </div>

      <button type="submit" class="btn btn-primary">Verify 2FA Code</button>
    </form>

    <div class="divider">
      <span>OR</span>
    </div>

    <button type="button" class="create-account-link" @click="navigateToRegister">
      Create new account
    </button>

    <button type="button" class="back-button" @click="navigateToLogin">
      Back to login
    </button>
  </div>
</template>

<style scoped>
.twofa-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
  align-items: center;
  text-align: center;
}

.lock-icon-section {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;
}

.lock-icon {
  width: 100px;
  height: 100px;
  display: block;
  object-fit: contain;
  filter: brightness(0.9);
}

.content-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

.description {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.input-field {
  padding: 16px;
  background-color: var(--surface-light);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 16px;
  transition: all 0.2s ease;
  font-family: 'Courier New', monospace;
  text-align: center;
  letter-spacing: 6px;
}

.input-field::placeholder {
  color: var(--text-tertiary);
  letter-spacing: normal;
}

.input-field:focus {
  outline: none;
  background-color: var(--surface-dark);
  border-color: var(--primary-color);
}

.twofa-input {
  font-size: 32px;
  font-weight: bold;
  letter-spacing: 10px;
}

.error-message {
  font-size: 12px;
  color: var(--error-color);
}

.btn {
  padding: 12px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  width: 100%;
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
  width: 100%;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background-color: var(--border-color);
}

.create-account-link {
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 14px;
  cursor: pointer;
  transition: color 0.2s ease;
  padding: 8px 0;
  text-decoration: none;
  font-weight: 500;
}

.create-account-link:hover {
  color: var(--primary-color);
}

.back-button {
  background: none;
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  padding: 10px 16px;
  border-radius: 4px;
  width: 100%;
}

.back-button:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* Responsive */
@media (max-width: 480px) {
  .twofa-container {
    gap: 20px;
  }

  .lock-icon {
    width: 80px;
    height: 80px;
  }

  .title {
    font-size: 20px;
  }

  .description {
    font-size: 13px;
  }

  .twofa-input {
    font-size: 24px;
    letter-spacing: 8px;
  }
}
</style>
