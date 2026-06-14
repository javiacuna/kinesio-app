import { apiFetch } from "@/shared/api/http";

export type ReportPeriod = {
  from: string;
  to: string;
};

export type ReportTotals = {
  appointments_total: number;
  scheduled_appointments: number;
  cancelled_appointments: number;
  active_patients: number;
  low_stock_materials: number;
  pending_material_loans: number;
  billing_value_cents: number;
  collected_cents: number;
  pending_collection_cents: number;
  professional_fee_cents: number;
  center_amount_cents: number;
};

export type AppointmentDay = {
  date: string;
  scheduled: number;
  cancelled: number;
  total: number;
};

export type RevenueDay = {
  date: string;
  billing_value_cents: number;
  collected_cents: number;
};

export type LabelValue = {
  label: string;
  value: number;
};

export type LowStockMaterial = {
  id: string;
  name: string;
  available_qty: number;
  total_qty: number;
};

export type ReportsSummary = {
  period: ReportPeriod;
  totals: ReportTotals;
  appointments_by_day: AppointmentDay[];
  revenue_by_day: RevenueDay[];
  appointments_by_status: LabelValue[];
  appointments_by_kinesiologist: LabelValue[];
  low_stock_materials: LowStockMaterial[];
  pending_loans_by_material: LabelValue[];
};

export function getReportsSummary(input: {
  from: string;
  to: string;
  kinesiologist_id?: string;
  financier_id?: string;
  patient_id?: string;
}) {
  const qs = new URLSearchParams();
  if (input.from) qs.set("from", input.from);
  if (input.to) qs.set("to", input.to);
  if (input.kinesiologist_id) qs.set("kinesiologist_id", input.kinesiologist_id);
  if (input.financier_id) qs.set("financier_id", input.financier_id);
  if (input.patient_id) qs.set("patient_id", input.patient_id);
  return apiFetch<ReportsSummary>(`/api/v1/reports/summary?${qs.toString()}`);
}
