import { apiFetch, apiFetchBlob } from "@/shared/api/http";

export type ExercisePlan = {
  id: string;
  patient_id: string;
  kinesiologist_id: string;
  frequency: "daily" | "weekly" | string;
  duration_weeks: number;
  observations?: string | null;
  status: "active" | "closed" | string;
  items: Array<{
    id: string;
    name: string;
    description?: string | null;
    video_url?: string | null;
    guide_url?: string | null;
    estimated_minutes: number;
    sets?: number | null;
    reps?: number | null;
  }>;
  created_at: string;
  updated_at: string;
  created_by_email?: string | null;
  created_by_role?: string | null;
  updated_by_email?: string | null;
  updated_by_role?: string | null;
};

export type PatientEvolution = {
  id: string;
  patient_id: string;
  kinesiologist_id: string;
  appointment_id?: string | null;
  pain_level?: number | null;
  notes: string;
  created_at: string;
  updated_at: string;
};

export type CreatePatientEvolutionInput = {
  kinesiologist_id: string;
  appointment_id?: string | null;
  pain_level?: number | null;
  notes: string;
};

export type CreateExercisePlanInput = {
  kinesiologist_id: string;
  frequency: "daily" | "weekly";
  duration_weeks: number;
  observations?: string | null;
  items: Array<{
    name: string;
    estimated_minutes: number;
    sets?: number | null;
    reps?: number | null;
    description?: string | null;
  }>;
};

export type UpdateExercisePlanInput = CreateExercisePlanInput & {
  status: "active" | "closed";
};

export type MaterialLoan = {
  id: string;
  material_id: string;
  patient_id: string;
  kinesiologist_id: string;
  qty: number;
  notes?: string | null;
  loaned_at: string;
  returned_at?: string | null;
};

export type PatientAttachment = {
  id: string;
  patient_id: string;
  file_name: string;
  content_type: string;
  size_bytes: number;
  kind: "image" | "video" | "pdf" | string;
  category: string;
  patient_visible: boolean;
  notes?: string | null;
  uploaded_by_email?: string | null;
  uploaded_by_role?: string | null;
  updated_by_email?: string | null;
  updated_by_role?: string | null;
  updated_at?: string | null;
  download_url: string;
  created_at: string;
};

export type CIE10Code = {
  code: string;
  description: string;
  chapter?: string | null;
};

export type PatientDiagnosis = {
  id: string;
  patient_id: string;
  cie10_code: string;
  cie10_description: string;
  cie10_chapter?: string | null;
  kind: "primary" | "secondary" | string;
  status: "suspected" | "confirmed" | "resolved" | string;
  diagnosed_at: string;
  notes?: string | null;
  created_at: string;
  updated_at: string;
  created_by_email?: string | null;
  created_by_role?: string | null;
  updated_by_email?: string | null;
  updated_by_role?: string | null;
};

export type SavePatientDiagnosisInput = {
  cie10_code: string;
  kind: "primary" | "secondary";
  status: "suspected" | "confirmed" | "resolved";
  diagnosed_at: string;
  notes?: string | null;
};

export function searchCIE10Codes(query: string, limit = 30) {
  const qs = new URLSearchParams({ query, limit: String(limit) });
  return apiFetch<CIE10Code[]>(`/api/v1/cie10?${qs.toString()}`);
}

export function listPatientDiagnoses(patientId: string) {
  return apiFetch<PatientDiagnosis[]>(`/api/v1/patients/${patientId}/diagnoses`);
}

export function createPatientDiagnosis(patientId: string, input: SavePatientDiagnosisInput) {
  return apiFetch<PatientDiagnosis>(`/api/v1/patients/${patientId}/diagnoses`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updatePatientDiagnosis(patientId: string, diagnosisId: string, input: SavePatientDiagnosisInput) {
  return apiFetch<PatientDiagnosis>(`/api/v1/patients/${patientId}/diagnoses/${diagnosisId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deletePatientDiagnosis(patientId: string, diagnosisId: string) {
  return apiFetch<void>(`/api/v1/patients/${patientId}/diagnoses/${diagnosisId}`, {
    method: "DELETE",
  });
}

export function listPatientPlans(patientId: string) {
  return apiFetch<ExercisePlan[]>(`/api/v1/patients/${patientId}/plans`);
}

export function listMyPatientPlans() {
  return apiFetch<ExercisePlan[]>("/api/v1/patients/me/plans");
}

export function listPatientEvolutions(patientId: string, limit = 50) {
  const qs = new URLSearchParams({ limit: String(limit) });
  return apiFetch<PatientEvolution[]>(`/api/v1/patients/${patientId}/evolutions?${qs.toString()}`);
}

export function createPatientEvolution(patientId: string, input: CreatePatientEvolutionInput) {
  return apiFetch<PatientEvolution>(`/api/v1/patients/${patientId}/evolutions`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function createPatientPlan(patientId: string, input: CreateExercisePlanInput) {
  return apiFetch<ExercisePlan>(`/api/v1/patients/${patientId}/plans`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updatePatientPlan(planId: string, input: UpdateExercisePlanInput) {
  return apiFetch<ExercisePlan>(`/api/v1/plans/${planId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function listPatientMaterialLoans(patientId: string, activeOnly = false) {
  const qs = new URLSearchParams({ active: String(activeOnly) });
  return apiFetch<MaterialLoan[]>(`/api/v1/patients/${patientId}/material-loans?${qs.toString()}`);
}

export function listPatientAttachments(patientId: string) {
  return apiFetch<PatientAttachment[]>(`/api/v1/patients/${patientId}/attachments`);
}

export function listMyPatientAttachments() {
  return apiFetch<PatientAttachment[]>("/api/v1/patients/me/attachments");
}

export function uploadPatientAttachment(
  patientId: string,
  file: File,
  input?: { notes?: string; category?: string; patient_visible?: boolean },
) {
  const body = new FormData();
  body.append("file", file);
  if (input?.notes?.trim()) body.append("notes", input.notes.trim());
  if (input?.category) body.append("category", input.category);
  body.append("patient_visible", String(Boolean(input?.patient_visible)));

  return apiFetch<PatientAttachment>(`/api/v1/patients/${patientId}/attachments`, {
    method: "POST",
    body,
  });
}

export function updatePatientAttachment(
  attachmentId: string,
  input: { file_name: string; notes?: string | null; category: string; patient_visible: boolean },
) {
  return apiFetch<PatientAttachment>(`/api/v1/patient-attachments/${attachmentId}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deletePatientAttachment(attachmentId: string) {
  return apiFetch<void>(`/api/v1/patient-attachments/${attachmentId}`, {
    method: "DELETE",
  });
}

export function downloadPatientAttachment(attachment: PatientAttachment) {
  return apiFetchBlob(attachment.download_url);
}
