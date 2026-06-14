import { apiFetch } from "@/shared/api/http";
import type { Patient } from "./types";

export type CreatePatientInput = {
  dni: string;
  first_name: string;
  last_name: string;
  email: string;
  financier_id?: string | null;
  financier_member_number?: string | null;
};

export function createPatient(input: CreatePatientInput) {
  return apiFetch<Patient>("/api/v1/patients", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export type UpdatePatientInput = CreatePatientInput & {
  id: string;
  phone?: string | null;
  birth_date?: string | null;
  clinical_notes?: string | null;
  active?: boolean;
};

export function updatePatient(input: UpdatePatientInput) {
  const { id, ...body } = input;
  return apiFetch<Patient>(`/api/v1/patients/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function getPatient(id: string) {
  return apiFetch<Patient>(`/api/v1/patients/${id}`);
}

export function archivePatient(id: string) {
  return apiFetch<void>(`/api/v1/patients/${id}`, {
    method: "DELETE",
  });
}

export async function listPatients(limit = 50, includeInactive = false) {
  const qs = new URLSearchParams({
    limit: String(limit),
  });
  if (includeInactive) qs.set("active", "false");

  return apiFetch<Patient[]>(`/api/v1/patients?${qs.toString()}`);
}

export async function searchPatients(query: string, limit = 20, includeInactive = false) {
  const qs = new URLSearchParams({
    query,
    limit: String(limit),
  });
  if (includeInactive) qs.set("active", "false");

  return apiFetch<Patient[]>(`/api/v1/patients?${qs.toString()}`);
}
