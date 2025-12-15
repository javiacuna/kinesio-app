import { apiFetch } from "../../shared/api/http";
import type { Appointment } from "./types";

export function listAppointmentsDay(params: { date: string; kinesiologist_id: string }) {
  const q = new URLSearchParams(params).toString();
  return apiFetch<Appointment[]>(`/api/v1/appointments?${q}`);
}

export type CreateAppointmentInput = {
  patient_id: string;
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
