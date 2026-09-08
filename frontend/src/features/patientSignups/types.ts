export type SignupStatus = "pending" | "approved" | "rejected";

export type RegisterPatientAccountInput = {
  email: string;
  password: string;
  dni: string;
  first_name: string;
  last_name: string;
};

export type RegisterPatientAccountResult = {
  status: SignupStatus;
};

export type PossibleSignupMatch = {
  patient_id: string;
  first_name: string;
  last_name: string;
  email: string;
  email_matches: boolean;
};

export type PendingPatientSignup = {
  id: string;
  firebase_uid: string;
  dni: string;
  email: string;
  first_name: string;
  last_name: string;
  status: SignupStatus;
  matched_patient_id?: string;
  reviewed_by_email?: string;
  reviewed_at?: string;
  rejection_reason?: string;
  created_at: string;
  possible_match?: PossibleSignupMatch;
};
