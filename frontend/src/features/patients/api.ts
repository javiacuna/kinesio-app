import { apiFetch } from "@/shared/api/http";
import type { Patient } from "./types";

export async function searchPatients(query: string, limit = 20) {
  const qs = new URLSearchParams({
    query,
    limit: String(limit),
  });

  // Si back usa /api/v1/patients?query=... cambiar esta línea por:
  // return apiFetch<Patient[]>(`/api/v1/patients?${qs.toString()}`);
  return apiFetch<Patient[]>(`/api/v1/patients/search?${qs.toString()}`);
}
