import { apiFetch } from "@/shared/api/http";

export type Financier = {
  id: string;
  name: string;
  kind: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type PracticeTariff = {
  id: string;
  practice_id: string;
  financier_id: string;
  billing_value_cents: number;
  copay_cents: number;
  currency: string;
  valid_from: string;
  valid_to?: string | null;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type ProfessionalFeeRule = {
  id: string;
  kinesiologist_id: string;
  practice_id: string;
  rule_type: "fixed" | "percentage";
  fixed_value_cents?: number | null;
  percentage?: number | null;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type FinancialMovement = {
  id: string;
  appointment_id: string;
  patient_id: string;
  kinesiologist_id: string;
  practice_id: string;
  financier_id: string;
  tariff_id: string;
  fee_rule_id?: string | null;
  billing_value_cents: number;
  copay_cents: number;
  payer_value_cents: number;
  professional_fee_cents: number;
  center_amount_cents: number;
  status: "pending" | "paid" | string;
  collection_status: "pending" | "collected" | "cancelled" | string;
  professional_payment_status: "pending" | "paid" | "cancelled" | string;
  collected_at?: string | null;
  professional_paid_at?: string | null;
  cancellation_reason?: string | null;
  created_at: string;
  updated_at: string;
};

export type SaveFinancierInput = {
  id?: string;
  name: string;
  kind: string;
  active: boolean;
};

export type SavePracticeTariffInput = {
  id?: string;
  practice_id: string;
  financier_id: string;
  billing_value_cents: number;
  copay_cents: number;
  currency: string;
  valid_from: string;
  valid_to?: string | null;
  active: boolean;
};

export type SaveProfessionalFeeRuleInput = {
  id?: string;
  kinesiologist_id: string;
  practice_id: string;
  rule_type: "fixed" | "percentage";
  fixed_value_cents?: number | null;
  percentage?: number | null;
  active: boolean;
};

export function listFinanciers(options?: { includeInactive?: boolean }) {
  const query = options?.includeInactive ? "?active=false" : "";
  return apiFetch<Financier[]>(`/api/v1/financiers${query}`);
}

export function saveFinancier(input: SaveFinancierInput) {
  const { id, ...body } = input;
  return apiFetch<Financier>(id ? `/api/v1/financiers/${id}` : "/api/v1/financiers", {
    method: id ? "PUT" : "POST",
    body: JSON.stringify(body),
  });
}

export function listPracticeTariffs(options?: { includeInactive?: boolean }) {
  const query = options?.includeInactive ? "?active=false" : "";
  return apiFetch<PracticeTariff[]>(`/api/v1/financial/tariffs${query}`);
}

export function savePracticeTariff(input: SavePracticeTariffInput) {
  const { id, ...body } = input;
  return apiFetch<PracticeTariff>(id ? `/api/v1/financial/tariffs/${id}` : "/api/v1/financial/tariffs", {
    method: id ? "PUT" : "POST",
    body: JSON.stringify(body),
  });
}

export function listProfessionalFeeRules(options?: { includeInactive?: boolean }) {
  const query = options?.includeInactive ? "?active=false" : "";
  return apiFetch<ProfessionalFeeRule[]>(`/api/v1/financial/fee-rules${query}`);
}

export function saveProfessionalFeeRule(input: SaveProfessionalFeeRuleInput) {
  const { id, ...body } = input;
  return apiFetch<ProfessionalFeeRule>(id ? `/api/v1/financial/fee-rules/${id}` : "/api/v1/financial/fee-rules", {
    method: id ? "PUT" : "POST",
    body: JSON.stringify(body),
  });
}

export function listFinancialMovements(params?: {
  from?: string;
  to?: string;
  status?: string;
  kinesiologist_id?: string;
  practice_id?: string;
  financier_id?: string;
  collection_status?: string;
  professional_payment_status?: string;
}) {
  const q = new URLSearchParams();
  if (params?.from) q.set("from", params.from);
  if (params?.to) q.set("to", params.to);
  if (params?.status) q.set("status", params.status);
  if (params?.kinesiologist_id) q.set("kinesiologist_id", params.kinesiologist_id);
  if (params?.practice_id) q.set("practice_id", params.practice_id);
  if (params?.financier_id) q.set("financier_id", params.financier_id);
  if (params?.collection_status) q.set("collection_status", params.collection_status);
  if (params?.professional_payment_status) q.set("professional_payment_status", params.professional_payment_status);
  const query = q.toString();
  return apiFetch<FinancialMovement[]>(`/api/v1/financial/movements${query ? `?${query}` : ""}`);
}

export function updateFinancialMovementStatus(input: {
  movement_id: string;
  collection_status?: "pending" | "collected" | "cancelled";
  professional_payment_status?: "pending" | "paid" | "cancelled";
  cancellation_reason?: string;
}) {
  const { movement_id, ...body } = input;
  return apiFetch<FinancialMovement>(`/api/v1/financial/movements/${movement_id}/status`, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}

export function completeAppointment(input: { appointment_id: string; practice_id: string; financier_id: string }) {
  return apiFetch<FinancialMovement>(`/api/v1/financial/appointments/${input.appointment_id}/complete`, {
    method: "POST",
    body: JSON.stringify({
      practice_id: input.practice_id,
      financier_id: input.financier_id,
    }),
  });
}
