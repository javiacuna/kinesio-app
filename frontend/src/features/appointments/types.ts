export type Appointment = {
  id: string;
  patient_id: string;
  kinesiologist_id: string;
  start_at: string;
  end_at: string;
  status: "scheduled" | "cancelled";
  notes?: string | null;
};
