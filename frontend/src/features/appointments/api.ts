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

export type UpdateAppointmentInput = {
  id: string;
  status?: "scheduled" | "cancelled";
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
