import { apiFetch } from "@/shared/api/http";

export type Material = {
  id: string;
  name: string;
  description?: string | null;
  total_qty: number;
  available_qty: number;
  created_at: string;
  updated_at: string;
  created_by_email?: string | null;
  created_by_role?: string | null;
  updated_by_email?: string | null;
  updated_by_role?: string | null;
};

export type MaterialLoan = {
  id: string;
  material_id: string;
  patient_id: string;
  kinesiologist_id: string;
  qty: number;
  notes?: string | null;
  loaned_at: string;
  due_date?: string | null;
  returned_at?: string | null;
  loaned_by_email?: string | null;
  loaned_by_role?: string | null;
  returned_by_email?: string | null;
  returned_by_role?: string | null;
};

export type CreateMaterialInput = {
  name: string;
  description?: string | null;
  total_qty: number;
};

export type UpdateMaterialInput = CreateMaterialInput & {
  id: string;
};

export type LoanMaterialInput = {
  material_id: string;
  patient_id: string;
  kinesiologist_id: string;
  qty: number;
  notes?: string | null;
  due_date?: string | null;
};

export function listMaterials(limit = 100) {
  const qs = new URLSearchParams({ limit: String(limit) });
  return apiFetch<Material[]>(`/api/v1/materials?${qs.toString()}`);
}

export function createMaterial(input: CreateMaterialInput) {
  return apiFetch<Material>("/api/v1/materials", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateMaterial(input: UpdateMaterialInput) {
  const { id, ...body } = input;
  return apiFetch<Material>(`/api/v1/materials/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function listMaterialLoans(activeOnly = false, limit = 100) {
  const qs = new URLSearchParams({ active: String(activeOnly), limit: String(limit) });
  return apiFetch<MaterialLoan[]>(`/api/v1/material-loans?${qs.toString()}`);
}

export function loanMaterial(input: LoanMaterialInput) {
  return apiFetch<MaterialLoan>("/api/v1/material-loans", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function returnMaterialLoan(loanId: string) {
  return apiFetch<MaterialLoan>(`/api/v1/material-loans/${loanId}/return`, {
    method: "POST",
  });
}

export function listPatientMaterialLoans(patientId: string, activeOnly = false) {
  const qs = new URLSearchParams({ active: String(activeOnly), limit: "100" });
  return apiFetch<MaterialLoan[]>(`/api/v1/patients/${patientId}/material-loans?${qs.toString()}`);
}
