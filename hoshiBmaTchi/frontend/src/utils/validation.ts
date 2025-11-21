
export interface ValidationResult {
  isValid: boolean;
  message: string;
}

export const validateName = (name: string): ValidationResult => {
  if (name.length < 5) {
    return {
      isValid: false,
      message: "Name must be at least 5 characters long",
    };
  }
  if (!/^[a-zA-Z\s]+$/.test(name)) {
    return {
      isValid: false,
      message: "Name can only contain letters and spaces",
    };
  }
  return { isValid: true, message: "" };
};

export const validateUsername = (username: string): ValidationResult => {
  if (username.length < 3) {
    return {
      isValid: false,
      message: "Username must be at least 3 characters",
    };
  }
  if (!/^[a-zA-Z0-9_]+$/.test(username)) {
    return {
      isValid: false,
      message: "Username can only contain letters, numbers, and underscores",
    };
  }
  return { isValid: true, message: "" };
};

export const validateEmail = (email: string): ValidationResult => {
  const emailPattern = /^[^\s@]+@[^\s@]+\.[a-zA-Z]{2,}$/;
  if (!emailPattern.test(email)) {
    return { isValid: false, message: "Please enter a valid email address" };
  }
  return { isValid: true, message: "" };
};

export interface PasswordStrength {
  isValid: boolean;
  strength: "weak" | "fair" | "good" | "strong";
  criteria: {
    length: boolean;
    uppercase: boolean;
    lowercase: boolean;
    number: boolean;
    special: boolean;
  };
  message: string;
}

export const validatePassword = (password: string): PasswordStrength => {
  const criteria = {
    length: password.length >= 8,
    uppercase: /[A-Z]/.test(password),
    lowercase: /[a-z]/.test(password),
    number: /[0-9]/.test(password),
    special: /[!@#$%^&*(),.?":{}|<>]/.test(password),
  };

  const meetsCount = Object.values(criteria).filter(Boolean).length;

  let strength: "weak" | "fair" | "good" | "strong" = "weak";
  if (meetsCount >= 4) strength = "strong";
  else if (meetsCount === 3) strength = "good";
  else if (meetsCount === 2) strength = "fair";

  const isValid = meetsCount >= 4;

  const getMessage = () => {
    const missing = [];
    if (!criteria.length) missing.push("at least 8 characters");
    if (!criteria.uppercase) missing.push("uppercase letter");
    if (!criteria.lowercase) missing.push("lowercase letter");
    if (!criteria.number) missing.push("number");
    if (!criteria.special) missing.push("special character");

    if (isValid) return "Password is strong";
    return `Password needs: ${missing.join(", ")}`;
  };

  return {
    isValid,
    strength,
    criteria,
    message: getMessage(),
  };
};

export const validateConfirmPassword = (
  password: string,
  confirmPassword: string
): ValidationResult => {
  if (password !== confirmPassword) {
    return { isValid: false, message: "Passwords do not match" };
  }
  return { isValid: true, message: "" };
};

export const validateGender = (gender: string): ValidationResult => {
  if (!["male", "female"].includes(gender.toLowerCase())) {
    return { isValid: false, message: "Please select a valid gender" };
  }
  return { isValid: true, message: "" };
};

export const validateAge = (dateOfBirth: string): ValidationResult => {
  if (!dateOfBirth) {
    return { isValid: false, message: "Please select your date of birth" };
  }

  const birthDate = new Date(dateOfBirth);
  const today = new Date();
  let age = today.getFullYear() - birthDate.getFullYear();
  const monthDiff = today.getMonth() - birthDate.getMonth();

  if (
    monthDiff < 0 ||
    (monthDiff === 0 && today.getDate() < birthDate.getDate())
  ) {
    age--;
  }

  if (age < 13) {
    return { isValid: false, message: "You must be at least 13 years old" };
  }

  return { isValid: true, message: "" };
};

export const validateOTP = (otp: string): ValidationResult => {
  if (!/^\d{6}$/.test(otp)) {
    return { isValid: false, message: "OTP must be 6 digits" };
  }
  return { isValid: true, message: "" };
};

export const validate2FACode = (code: string): ValidationResult => {
  if (!/^\d{6}$/.test(code)) {
    return { isValid: false, message: "2FA code must be 6 digits" };
  }
  return { isValid: true, message: "" };
};
