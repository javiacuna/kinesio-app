import { apiFetch } from "@/shared/api/http";
import type { Appointment } from "./types";

export function listAppointmentsDay(params: {
  date: string;
  kinesiologist_id: string;
}) {
  const q = new URLSearchParams(params).toString();
  return apiFetch<Appointment[]>(`/api/v1/appointments?${q}`);
}
