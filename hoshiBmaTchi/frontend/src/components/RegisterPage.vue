<script setup lang="ts">
import { ref, reactive } from "vue";
import { useRouter } from "vue-router";
import { authApi } from "../services/apiService";

import {
  validateName,
  validateUsername as validateUsernameUtil,
  validateEmail as validateEmailUtil,
  validatePassword as validatePasswordUtil,
  validateConfirmPassword as validateConfirmPasswordUtil,
  validateAge as validateAgeUtil,
  validateGender as validateGenderUtil,
  validateOTP as validateOTPUtil,
} from "../utils/validation";
import type { PasswordStrength } from "../utils/validation";

const router = useRouter();

const isLoading = ref(false);
const globalError = ref("");

const isSendingOtp = ref(false)
const otpMessage = ref('')

const formData = reactive({
  fullName: "",
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
  dateOfBirth: "",
  gender: "",
  profilePicture: null as File | null,
  otpCode: "",
  subscribeNewsletter: false,
  enable2FA: false,
});

const errors = reactive({
  fullName: "",
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
  dateOfBirth: "",
  gender: "",
  profilePicture: "",
  otpCode: "",
});

const passwordStrength = ref<PasswordStrength | null>(null);

const validateFullName = () => {
  const result = validateName(formData.fullName);
  errors.fullName = result.isValid ? '' : result.message
};

const validateUsernameFn = () => {
  const result = validateUsernameUtil(formData.username);
  errors.username = result.isValid ? '' : result.message
};

const validateEmailFn = () => {
  const result = validateEmailUtil(formData.email);
  errors.email = result.isValid ? '' : result.message
};

const validatePasswordFn = () => {
  const result = validatePasswordUtil(formData.password);
  passwordStrength.value = result;
  errors.password = result.isValid ? '' : result.message
};

const validateConfirmPasswordFn = () => {
  const result = validateConfirmPasswordUtil(
    formData.password,
    formData.confirmPassword
  );
  errors.confirmPassword = result.isValid ? '' : result.message
};

const validateAgeFn = () => {
  const result = validateAgeUtil(formData.dateOfBirth);
  errors.dateOfBirth = result.isValid ? '' : result.message
};

const validateGenderFn = () => {
  const result = validateGenderUtil(formData.gender);
  errors.gender = result.isValid ? '' : result.message
};

const validateOTPFn = () => {
  const result = validateOTPUtil(formData.otpCode);
  errors.otpCode = result.isValid ? '' : result.message
};

const handleProfilePictureUpload = (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (file) {
    if (file.size > 5 * 1024 * 1024) {
      errors.profilePicture = "File size must be less than 5MB";
      return;
    }
    formData.profilePicture = file;
    errors.profilePicture = "";
  }
};

const sendOTP = async () => {
  otpMessage.value = ''
  isSendingOtp.value = true

  validateEmailFn()
  if (errors.email) {
    isSendingOtp.value = false
    return
  }

  try {
    const response = await authApi.sendOtp(formData.email)
    otpMessage.value = response.data.message
  } catch (error: any) {
    console.error('Failed to send OTP:', error)
    if (error.response && error.response.data.error) {
      otpMessage.value = error.response.data.error
    } else {
      otpMessage.value = 'Failed to send OTP. Please try again.'
    }
  } finally {
    isSendingOtp.value = false
  }

}

const handleRegister = async () => {
  globalError.value = "";

  validateFullName();
  validateUsernameFn();
  validateEmailFn();
  validatePasswordFn();
  validateConfirmPasswordFn();
  validateAgeFn();
  validateGenderFn();
  validateOTPFn();

  const hasErrors = Object.values(errors).some((error) => error !== "");
  if (hasErrors) {
    console.log("Validation errors found");
    return;
  }

  isLoading.value = true;

  try {
    let profilePictureUrl = "";

    if (formData.profilePicture) {
      try {
        const uploadRes = await authApi.uploadAvatar(formData.profilePicture);
        profilePictureUrl = uploadRes.data.media_url;
      } catch (uploadError) {
        console.error("Failed to upload avatar:", uploadError);
        globalError.value = "Failed to upload profile picture. Please try again or skip it.";
        isLoading.value = false;
        return; 
      }
    }

    const apiData = {
      name: formData.fullName,
      username: formData.username,
      email: formData.email,
      password: formData.password,
      confirm_password: formData.confirmPassword,
      date_of_birth: new Date(formData.dateOfBirth).toISOString(),
      gender: formData.gender,
      profile_picture_url: profilePictureUrl, 
      subscribe_to_newsletter: formData.subscribeNewsletter,
      enable_2fa: formData.enable2FA,
      otp_code: formData.otpCode,
      turnstile_token: "dummy_token", 
    };

    const response = await authApi.register(apiData);
    console.log("Registration successful:", response.data);
    router.push("/login");
  } catch (error: any) {
    console.error("Registration failed:", error);
    if (error.response && error.response.data.error) {
      globalError.value = error.response.data.error;
    } else {
      globalError.value = "An unknown error occurred. Please try again.";
    }
  } finally {
    isLoading.value = false;
  }

  console.log("Registration attempt:", formData);
};

const getCriterionText = (key: string): string => {
  const texts: Record<string, string> = {
    length: "8+ characters",
    uppercase: "Uppercase letter",
    lowercase: "Lowercase letter",
    number: "Number",
    special: "Special character",
  };
  return texts[key] || key;
};

const validateUsername = validateUsernameFn;
const validateEmail = validateEmailFn;
const validatePassword = validatePasswordFn;
const validateConfirmPassword = validateConfirmPasswordFn;
const validateAge = validateAgeFn;
const validateGender = validateGenderFn;
const validateOTP = validateOTPFn;

const navigateToLogin = () => {
  router.push("/login");
};
</script>

<template>
  <div class="register-container">
    <div class="logo-section">
      <h1 class="brand-name">hoshiBmatchi</h1>
      <p class="subtitle">
        Sign up to see photos and videos from your friends.
      </p>
    </div>

    <form class="form" @submit.prevent="handleRegister">
      <!-- Full Name -->
      <div class="form-group">
        <label class="form-label">Full Name</label>
        <input
          v-model="formData.fullName"
          type="text"
          placeholder="Enter your full name"
          class="input-field"
          @blur="validateFullName"
        />
        <span v-if="errors.fullName" class="error-message">{{
          errors.fullName
        }}</span>
      </div>

      <!-- Username -->
      <div class="form-group">
        <label class="form-label">Username</label>
        <input
          v-model="formData.username"
          type="text"
          placeholder="Choose a username"
          class="input-field"
          @blur="validateUsername"
        />
        <span v-if="errors.username" class="error-message">{{
          errors.username
        }}</span>
      </div>

      <div class="form-group">
        <label class="form-label">Email</label>
        <input
          v-model="formData.email"
          type="email"
          placeholder="Enter your email"
          class="input-field"
          @blur="validateEmail"
        />
        <span v-if="errors.email" class="error-message">{{
          errors.email
        }}</span>
      </div>

      <div class="form-group">
        <label class="form-label">Password</label>
        <input
          v-model="formData.password"
          type="password"
          placeholder="Create a password"
          class="input-field"
          @input="validatePassword"
        />
        <!-- <div v-if="passwordStrength" class="password-strength">
          <div class="strength-bar">
            <div
              class="strength-fill"
              :class="`strength-${passwordStrength.strength}`"
              :style="{
                width: `${
                  (Object.values(passwordStrength.criteria).filter(Boolean)
                    .length /
                    5) *
                  100
                }%`,
              }"
            ></div>
          </div>
          <span
            class="strength-text"
            :class="`text-${passwordStrength.strength}`"
          >
            {{ passwordStrength.message }}
          </span>

          <div
            v-if="Object.keys(passwordStrength?.criteria || {}).length > 0"
            class="criteria-list"
          >
            <div
              v-for="(value, key) in passwordStrength.criteria"
              :key="key"
              class="criteria-item"
              :class="{ met: value }"
            >
              <span class="criterion-dot"></span>
              <span class="criterion-text">{{ getCriterionText(key) }}</span>
            </div>
          </div>
        </div> -->

        <span v-if="errors.password" class="error-message">{{
          errors.password
        }}</span>
      </div>

      <div class="form-group">
        <label class="form-label">Confirm Password</label>
        <input
          v-model="formData.confirmPassword"
          type="password"
          placeholder="Confirm your password"
          class="input-field"
          @blur="validateConfirmPassword"
        />
        <span v-if="errors.confirmPassword" class="error-message">{{
          errors.confirmPassword
        }}</span>
      </div>

      <!-- Date of Birth -->
      <div class="form-group">
        <label class="form-label">Date of Birth</label>
        <input
          v-model="formData.dateOfBirth"
          type="date"
          class="input-field"
          @blur="validateAge"
        />
        <span v-if="errors.dateOfBirth" class="error-message">{{
          errors.dateOfBirth
        }}</span>
      </div>

      <!-- Gender -->
      <div class="form-group">
        <label class="form-label">Gender</label>
        <select
          v-model="formData.gender"
          class="input-field"
          @blur="validateGender"
        >
          <option value="">Select your gender</option>
          <option value="male">Male</option>
          <option value="female">Female</option>
        </select>
        <span v-if="errors.gender" class="error-message">{{
          errors.gender
        }}</span>
      </div>

      <!-- Profile Picture -->
      <div class="form-group">
        <label class="form-label">Profile Picture</label>
        <input
          type="file"
          accept="image/*"
          class="input-field file-input"
          @change="handleProfilePictureUpload"
        />
        <span v-if="errors.profilePicture" class="error-message">{{
          errors.profilePicture
        }}</span>
      </div>

      <!-- OTP Code -->
      <div class="form-group">
        <label class="form-label">OTP Code (6-digit)</label>
        <div class="otp-input-group">
          <input
            v-model="formData.otpCode"
            type="text"
            placeholder="000000"
            maxlength="6"
            class="input-field otp-input"
            @blur="validateOTP"
          />
          <button type="button" class="btn btn-otp" @click="sendOTP">
            Send OTP
          </button>
        </div>
        <span v-if="errors.otpCode" class="error-message">{{
          errors.otpCode
        }}</span>
      </div>

      <!-- Newsletter Checkbox -->
      <div class="checkbox-group">
        <input
          id="newsletter"
          v-model="formData.subscribeNewsletter"
          type="checkbox"
          class="checkbox-input"
        />
        <label for="newsletter" class="checkbox-label"
          >Subscribe to our newsletter</label
        >
      </div>

      <!-- 2FA Checkbox -->
      <div class="checkbox-group">
        <input
          id="twofa"
          v-model="formData.enable2FA"
          type="checkbox"
          class="checkbox-input"
        />
        <label for="twofa" class="checkbox-label"
          >Enable Two-Factor Authentication (2FA)</label
        >
      </div>

      <button type="submit" class="btn btn-primary">Sign up</button>
    </form>

    <div class="login-section">
      <p class="login-text">
        Have an account?
        <button type="button" class="link-button" @click="navigateToLogin">
          Log in
        </button>
      </p>
    </div>
  </div>
</template>

<style scoped>
.register-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-height: 85vh;
  overflow-y: auto;
  padding-right: 8px;
}

.register-container::-webkit-scrollbar {
  width: 6px;
}

.register-container::-webkit-scrollbar-track {
  background: var(--surface-light);
  border-radius: 3px;
}

.register-container::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.register-container::-webkit-scrollbar-thumb:hover {
  background: var(--text-tertiary);
}

.logo-section {
  text-align: center;
  margin-bottom: 8px;
}

.brand-name {
  font-size: 28px;
  font-weight: 300;
  letter-spacing: 1px;
  font-style: italic;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 400;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.input-field {
  padding: 12px 16px;
  background-color: var(--surface-light);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 14px;
  transition: all 0.2s ease;
  font-family: inherit;
}

.input-field::placeholder {
  color: var(--text-tertiary);
}

.input-field:focus {
  outline: none;
  background-color: var(--surface-dark);
  border-color: var(--primary-color);
}

.input-field.otp-input {
  font-size: 16px;
  letter-spacing: 4px;
  text-align: center;
  font-weight: 600;
}

.otp-input-group {
  display: flex;
  gap: 8px;
}

.otp-input-group .input-field {
  flex: 1;
}

.btn-otp {
  padding: 12px 16px;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.btn-otp:hover {
  background-color: var(--primary-hover);
}

.password-strength {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.strength-bar {
  height: 4px;
  background-color: var(--surface-light);
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  border-radius: 2px;
  transition: all 0.3s ease;
}

.strength-weak {
  background-color: var(--error-color);
}

.strength-fair {
  background-color: var(--warning-color);
}

.strength-good {
  background-color: #8b5cf6;
}

.strength-strong {
  background-color: var(--success-color);
}

.strength-text {
  font-size: 12px;
  font-weight: 500;
}

.text-weak {
  color: var(--error-color);
}

.text-fair {
  color: var(--warning-color);
}

.text-good {
  color: #8b5cf6;
}

.text-strong {
  color: var(--success-color);
}

.criteria-list {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.criteria-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-tertiary);
  transition: color 0.2s ease;
}

.criteria-item.met {
  color: var(--success-color);
}

.criterion-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background-color: currentColor;
}

.criterion-text {
  flex: 1;
}

.checkbox-group {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
}

.checkbox-input {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--primary-color);
}

.checkbox-label {
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  user-select: none;
}

.error-message {
  font-size: 12px;
  color: var(--error-color);
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
  margin-top: 12px;
}

.btn-primary:hover {
  background-color: var(--primary-hover);
}

.btn-primary:active {
  transform: scale(0.98);
}

.login-section {
  text-align: center;
  padding-top: 12px;
}

.login-text {
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

/* Responsive */
@media (max-width: 480px) {
  .register-container {
    gap: 16px;
    max-height: 90vh;
  }

  .brand-name {
    font-size: 24px;
  }

  .subtitle {
    font-size: 12px;
  }

  .input-field {
    padding: 10px 14px;
    font-size: 13px;
  }

  .criteria-list {
    grid-template-columns: 1fr;
  }

  .otp-input-group {
    flex-direction: column;
  }

  .btn-otp {
    width: 100%;
  }
}
</style>
