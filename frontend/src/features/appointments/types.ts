export type Appointment = {
  id: string;
  patient_id: string;
  kinesiologist_id: string;
  package_id?: string | null;
  package_session_number?: number | null;
  start_at: string;
  end_at: string;
  status: "scheduled" | "cancelled";
  notes?: string | null;
  cancelled_reason?: string | null;
};
