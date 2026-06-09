export type Appointment = {
  id: string;
  patient_id: string;
  kinesiologist_id: string;
  practice_id?: string | null;
  financier_id?: string | null;
  package_id?: string | null;
  package_session_number?: number | null;
  start_at: string;
  end_at: string;
  status: "scheduled" | "cancelled" | "completed";
  modality: "in_person" | "virtual";
  video_call_url?: string | null;
  video_provider?: string | null;
  video_meeting_id?: string | null;
  notes?: string | null;
  cancelled_reason?: string | null;
};
