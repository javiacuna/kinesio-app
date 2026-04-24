import { apiFetch } from "@/shared/api/http";
import type { AuthRole } from "./types";

export type AdminUser = {
  uid: string;
  email: string;
  role: AuthRole | "";
  disabled: boolean;
};

export function listAdminUsers() {
  return apiFetch<AdminUser[]>("/api/v1/admin/users");
}

export function assignUserRole(input: { email: string; role: AuthRole }) {
  return apiFetch<{ uid: string; email: string; role: AuthRole }>("/api/v1/admin/users/role", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function inviteUserAccess(input: { email: string; role: AuthRole }) {
  return apiFetch<{
    uid: string;
    email: string;
    role: AuthRole;
    created: boolean;
    email_sent: boolean;
  }>("/api/v1/admin/users/invite", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
