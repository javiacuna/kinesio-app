import { apiFetch } from "@/shared/api/http";
import type { Patient } from "./types";

export async function searchPatients(query: string, limit = 20) {
  const qs = new URLSearchParams({
    query,
    limit: String(limit),
  });

  return apiFetch<Patient[]>(`/api/v1/patients?${qs.toString()}`);
}
