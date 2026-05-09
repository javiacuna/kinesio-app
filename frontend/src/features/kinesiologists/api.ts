import { apiFetch, apiFetchBlob } from "@/shared/api/http";
import type { Kinesiologist, Practice, Specialty } from "./types";

export type SaveKinesiologistInput = {
  id?: string;
  first_name: string;
  last_name: string;
  email: string;
  license_number?: string | null;
  work_start_time: string;
  work_end_time: string;
  work_days: number[];
  practice_ids: string[];
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

export type SaveSpecialtyInput = {
  id?: string;
  name: string;
  active: boolean;
};

export function listSpecialties(options?: { includeInactive?: boolean }) {
  const query = options?.includeInactive ? "?active=false" : "";
  return apiFetch<Specialty[]>(`/api/v1/specialties${query}`);
}

export function createSpecialty(input: SaveSpecialtyInput) {
  return apiFetch<Specialty>("/api/v1/specialties", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateSpecialty(input: SaveSpecialtyInput & { id: string }) {
  const { id, ...body } = input;
  return apiFetch<Specialty>(`/api/v1/specialties/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export type SavePracticeInput = {
  id?: string;
  specialty_id: string;
  name: string;
  description?: string | null;
  active: boolean;
};

export function listPractices(options?: { includeInactive?: boolean }) {
  const query = options?.includeInactive ? "?active=false" : "";
  return apiFetch<Practice[]>(`/api/v1/practices${query}`);
}

export function createPractice(input: SavePracticeInput) {
  return apiFetch<Practice>("/api/v1/practices", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updatePractice(input: SavePracticeInput & { id: string }) {
  const { id, ...body } = input;
  return apiFetch<Practice>(`/api/v1/practices/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export type KinesiologistAttachment = {
  id: string;
  kinesiologist_id: string;
  file_name: string;
  content_type: string;
  size_bytes: number;
  kind: "image" | "video" | "pdf" | string;
  category: string;
  notes?: string | null;
  uploaded_by_email?: string | null;
  uploaded_by_role?: string | null;
  download_url: string;
  created_at: string;
};

export function listKinesiologistAttachments(kinesiologistId: string) {
  return apiFetch<KinesiologistAttachment[]>(`/api/v1/kinesiologists/${kinesiologistId}/attachments`);
}

export function uploadKinesiologistAttachment(
  kinesiologistId: string,
  file: File,
  input?: { notes?: string; category?: string },
) {
  const body = new FormData();
  body.append("file", file);
  if (input?.notes?.trim()) body.append("notes", input.notes.trim());
  if (input?.category) body.append("category", input.category);

  return apiFetch<KinesiologistAttachment>(`/api/v1/kinesiologists/${kinesiologistId}/attachments`, {
    method: "POST",
    body,
  });
}

export function downloadKinesiologistAttachment(attachment: KinesiologistAttachment) {
  return apiFetchBlob(attachment.download_url);
}
