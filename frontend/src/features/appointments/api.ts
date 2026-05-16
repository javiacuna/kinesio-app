import { apiFetch } from "../../shared/api/http";
import type { Appointment } from "./types";

export function listAppointmentsDay(params: { date: string; kinesiologist_id: string }) {
  const q = new URLSearchParams(params).toString();
  return apiFetch<Appointment[]>(`/api/v1/appointments?${q}`);
}

export function listPatientAppointments(params: { from: string; to: string; patient_id?: string }) {
  const q = new URLSearchParams(params).toString();
  return apiFetch<Appointment[]>(`/api/v1/appointments/patient?${q}`);
}

export type CreateAppointmentInput = {
  patient_id?: string;
  kinesiologist_id: string;
  practice_id?: string;
  financier_id?: string;
  start_at: string;
  end_at: string;
  notes?: string;
};

export function createAppointment(input: CreateAppointmentInput) {
  return apiFetch<Appointment>("/api/v1/appointments", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export type AppointmentPackage = {
  id: string;
  patient_id: string;
  kinesiologist_id: string;
  practice_id?: string | null;
  financier_id?: string | null;
  sessions_count: number;
  duration_min: number;
  start_date: string;
  start_time: string;
  weekdays_only: boolean;
  work_days?: number[] | null;
  notes?: string | null;
  created_at: string;
  updated_at: string;
};

export type AppointmentPackageWriteResponse = {
  package: AppointmentPackage;
  appointments: Appointment[];
};

export type CreateAppointmentPackageInput = {
  patient_id: string;
  kinesiologist_id: string;
  practice_id?: string;
  financier_id?: string;
  start_date: string;
  start_time: string;
  duration_min: number;
  sessions_count: number;
  weekdays_only: boolean;
  notes?: string;
};

export function createAppointmentPackage(input: CreateAppointmentPackageInput) {
  return apiFetch<AppointmentPackageWriteResponse>("/api/v1/appointment-packages", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export type UpdateAppointmentPackageInput = {
  id: string;
  start_date?: string;
  start_time?: string;
  duration_min?: number;
  practice_id?: string;
  financier_id?: string;
  work_days?: number[];
  notes?: string;
};

export function updateAppointmentPackage(input: UpdateAppointmentPackageInput) {
  const { id, ...patch } = input;
  return apiFetch<AppointmentPackageWriteResponse>(`/api/v1/appointment-packages/${id}`, {
    method: "PUT",
    body: JSON.stringify(patch),
  });
}

export type UpdateAppointmentInput = {
  id: string;
  status?: "scheduled" | "cancelled" | "completed";
  practice_id?: string;
  financier_id?: string;
  cancelled_reason?: string;
  notes?: string;
  start_at?: string;
  end_at?: string;
};

export function updateAppointment(input: UpdateAppointmentInput) {
  const { id, ...patch } = input;
  return apiFetch<Appointment>(`/api/v1/appointments/${id}`, {
    method: "PUT",
    body: JSON.stringify(patch),
  });
}

export function cancelAppointment(input: { id: string; reason?: string }) {
  return apiFetch<Appointment>(`/api/v1/appointments/${input.id}`, {
    method: "DELETE",
    body: JSON.stringify({ reason: input.reason }),
  });
}
