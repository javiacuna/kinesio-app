import { apiFetch } from "@/shared/api/http";

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
    estimated_minutes: number;
    sets?: number | null;
    reps?: number | null;
  }>;
  created_at: string;
  updated_at: string;
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

export function listPatientPlans(patientId: string) {
  return apiFetch<ExercisePlan[]>(`/api/v1/patients/${patientId}/plans`);
}

export function listPatientEvolutions(patientId: string, limit = 50) {
  const qs = new URLSearchParams({ limit: String(limit) });
  return apiFetch<PatientEvolution[]>(`/api/v1/patients/${patientId}/evolutions?${qs.toString()}`);
}

export function listPatientMaterialLoans(patientId: string, activeOnly = false) {
  const qs = new URLSearchParams({ active: String(activeOnly) });
  return apiFetch<MaterialLoan[]>(`/api/v1/patients/${patientId}/material-loans?${qs.toString()}`);
}
