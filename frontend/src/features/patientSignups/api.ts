import { apiFetch } from "@/shared/api/http";
import type { PendingPatientSignup, RegisterPatientAccountInput, RegisterPatientAccountResult } from "./types";

export function registerPatientAccount(input: RegisterPatientAccountInput) {
  return apiFetch<RegisterPatientAccountResult>("/api/v1/auth/patient-signup", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function listPendingPatientSignups(status: string = "pending") {
  return apiFetch<PendingPatientSignup[]>(`/api/v1/admin/patient-signups?status=${status}`);
}

export function approvePatientSignup(id: string) {
  return apiFetch<PendingPatientSignup>(`/api/v1/admin/patient-signups/${id}/approve`, {
    method: "POST",
  });
}

export function rejectPatientSignup(id: string, reason?: string) {
  return apiFetch<PendingPatientSignup>(`/api/v1/admin/patient-signups/${id}/reject`, {
    method: "POST",
    body: JSON.stringify({ reason: reason || undefined }),
  });
}
