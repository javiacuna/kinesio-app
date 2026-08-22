import { apiFetch } from "@/shared/api/http";
import type {
  AuthUser,
  ChangePasswordInput,
  ConfirmPasswordResetInput,
  LoginInput,
  LoginResponse,
  PasswordResetInput,
} from "./types";

export function login(input: LoginInput) {
  return apiFetch<LoginResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getMe() {
  return apiFetch<AuthUser>("/api/v1/auth/me");
}

export function requestPasswordReset(input: PasswordResetInput) {
  return apiFetch<{ email_sent: boolean }>("/api/v1/auth/password-reset", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function changePassword(input: ChangePasswordInput) {
  return apiFetch<LoginResponse>("/api/v1/auth/change-password", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function confirmPasswordReset(input: ConfirmPasswordResetInput) {
  return apiFetch<{ reset: boolean }>("/api/v1/auth/confirm-password-reset", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
