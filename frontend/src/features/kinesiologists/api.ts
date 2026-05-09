import { apiFetch } from "@/shared/api/http";
import type { Kinesiologist } from "./types";

export type SaveKinesiologistInput = {
  id?: string;
  first_name: string;
  last_name: string;
  email: string;
  license_number?: string | null;
  work_start_time: string;
  work_end_time: string;
  work_days: number[];
  active: boolean;
};

export function listKinesiologists(options?: { includeInactive?: boolean }) {
  const query = options?.includeInactive ? "?active=false" : "";
  return apiFetch<Kinesiologist[]>(`/api/v1/kinesiologists${query}`);
}

export function createKinesiologist(input: SaveKinesiologistInput) {
  return apiFetch<Kinesiologist>("/api/v1/kinesiologists", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateKinesiologist(input: SaveKinesiologistInput & { id: string }) {
  const { id, ...body } = input;
  return apiFetch<Kinesiologist>(`/api/v1/kinesiologists/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}
