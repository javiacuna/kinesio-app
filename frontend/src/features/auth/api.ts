import { apiFetch } from "@/shared/api/http";
import type { AuthUser, LoginInput, LoginResponse } from "./types";

export function login(input: LoginInput) {
  return apiFetch<LoginResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getMe() {
  return apiFetch<AuthUser>("/api/v1/auth/me");
}
