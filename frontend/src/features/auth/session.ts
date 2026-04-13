import type { AuthUser } from "./types";

const TOKEN_KEY = "auth_id_token";
const REFRESH_TOKEN_KEY = "auth_refresh_token";
const USER_KEY = "auth_user";

export function getAuthToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function getStoredUser(): AuthUser | null {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;

  try {
    return JSON.parse(raw) as AuthUser;
  } catch {
    clearAuthSession();
    return null;
  }
}

export function saveAuthTokens(idToken: string, refreshToken: string) {
  localStorage.setItem(TOKEN_KEY, idToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
}

export function saveAuthUser(user: AuthUser) {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearAuthSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}
