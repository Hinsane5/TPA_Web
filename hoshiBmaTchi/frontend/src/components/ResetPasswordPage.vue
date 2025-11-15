<template>
  <div class="reset-password-container">
    <div class="lock-icon-section">
      <img src="/icons/lock-icon.png" alt="Lock" class="lock-icon" />
    </div>

    <div class="content-section">
      <h2 class="heading">Reset Password</h2>
      <p class="description">
        Enter your new password below. Make sure it's strong and secure.
      </p>
    </div>

    <form @submit.prevent="handleResetPassword" class="form">
      <div class="form-group">
        <label for="new-password" class="form-label">New Password</label>
        <div class="password-input-wrapper">
          <input
            v-model="newPassword"
            :type="showNewPassword ? 'text' : 'password'"
            id="new-password"
            placeholder="Enter new password"
            class="input-field"
            required
          />
          <button
            type="button"
            @click="showNewPassword = !showNewPassword"
            class="toggle-password-btn"
          >
            {{ showNewPassword ? "👁️" : "👁️‍🗨️" }}
          </button>
        </div>
        <div v-if="passwordValidation.criteria" class="password-criteria">
          <div
            :class="['criterion', { met: passwordValidation.criteria.length }]"
          >
            ✓ At least 8 characters
          </div>
          <div
            :class="[
              'criterion',
              { met: passwordValidation.criteria.uppercase },
            ]"
          >
            ✓ Uppercase letter
          </div>
          <div
            :class="[
              'criterion',
              { met: passwordValidation.criteria.lowercase },
            ]"
          >
            ✓ Lowercase letter
          </div>
          <div
            :class="['criterion', { met: passwordValidation.criteria.number }]"
          >
            ✓ Number
          </div>
          <div
            :class="['criterion', { met: passwordValidation.criteria.special }]"
          >
            ✓ Special character
          </div>
        </div>
        <p
          v-if="newPassword && !passwordValidation.isValid"
          class="error-message"
        >
          {{ passwordValidation.message }}
        </p>
      </div>

      <!-- Confirm Password Input -->
      <div class="form-group">
        <label for="confirm-password" class="form-label"
          >Confirm Password</label
        >
        <div class="password-input-wrapper">
          <input
            v-model="confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            id="confirm-password"
            placeholder="Confirm new password"
            class="input-field"
            required
          />
          <button
            type="button"
            @click="showConfirmPassword = !showConfirmPassword"
            class="toggle-password-btn"
          >
            {{ showConfirmPassword ? "👁️" : "👁️‍🗨️" }}
          </button>
        </div>
        <p
          v-if="confirmPassword && !confirmPasswordMatch"
          class="error-message"
        >
          Passwords do not match
        </p>
      </div>

      <button
        type="submit"
        class="btn btn-primary"
        :disabled="!isFormValid || isLoading"
      >
        {{ isLoading ? "Resetting..." : "Reset Password" }}
      </button>
    </form>

    <button type="button" class="back-button" @click="navigateTo('login')">
      Back to login
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { AuthPage } from "../types";
import { validatePassword, validateConfirmPassword } from "../utils/validation";
import { authApi } from "../services/apiService";

const route = useRoute();
const router = useRouter();

const newPassword = ref("");
const confirmPassword = ref("");
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);

const token = ref("");
const isLoading = ref(false);
const error = ref("");
const successMessage = ref("");

onMounted(() => {
  const tokenFromUrl = route.query.token;

  if (Array.isArray(tokenFromUrl)) {
    token.value = tokenFromUrl[0] || "";
  } else if (tokenFromUrl) {
    token.value = tokenFromUrl;
  } else {
    token.value = "";
  }

  if (!token.value) {
    error.value =
      "Invalid or missing reset token. Please use the link from your email.";
  }
});

const passwordValidation = computed(() => {
  return validatePassword(newPassword.value);
});

const confirmPasswordMatch = computed(() => {
  if (!confirmPassword.value) return true;
  return validateConfirmPassword(newPassword.value, confirmPassword.value)
    .isValid;
});

const isFormValid = computed(() => {
  return (
    passwordValidation.value.isValid &&
    confirmPasswordMatch.value &&
    newPassword.value &&
    confirmPassword.value
  );
});

const handleResetPassword = async () => {
  if (!isFormValid.value) {
    console.log("[v0] Form validation failed");
    return;
  }
  if (!token.value) {
    error.value = "Missing reset token. Please use the link from your email.";
    return;
  }

  isLoading.value = true;
  error.value = "";

  try {
    const data = {
      token: token.value,
      new_password: newPassword.value,
      confirm_password: confirmPassword.value,
    };

    const response = await authApi.resetPassword(data);

    successMessage.value = response.data.message || "Password has been reset successfully!";
  } catch (err: any) {
    console.error("Password reset failed:", err);
    if (err.response && err.response.data.error) {
      error.value = err.response.data.error;
    } else {
      error.value = "An unknown error occurred. Please try again.";
    }
  } finally {
    isLoading.value = false;
  }
};

const navigateTo = (page: AuthPage) => {
  if (page === "login") {
    router.push("/login");
  }
};
</script>

<style scoped>
.reset-password-container {
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
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-field {
  flex: 1;
  padding: 12px 16px;
  padding-right: 40px;
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

.toggle-password-btn {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 16px;
  cursor: pointer;
  padding: 4px;
  transition: color 0.2s ease;
}

.toggle-password-btn:hover {
  color: var(--text-primary);
}

.password-criteria {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  background-color: var(--surface-light);
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.criterion {
  font-size: 12px;
  color: var(--text-secondary);
  transition: color 0.2s ease;
}

.criterion.met {
  color: var(--success-color);
}

.error-message {
  font-size: 12px;
  color: var(--error-color);
  margin-top: -4px;
}

.btn {
  padding: 12px 16px;
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

.btn-primary:hover:not(:disabled) {
  background-color: var(--primary-hover);
}

.btn-primary:active:not(:disabled) {
  transform: scale(0.98);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
  .reset-password-container {
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
    padding-right: 36px;
    font-size: 13px;
  }

  .password-criteria {
    padding: 10px;
    gap: 5px;
  }

  .criterion {
    font-size: 11px;
  }
}

@media (max-width: 360px) {
  .reset-password-container {
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
