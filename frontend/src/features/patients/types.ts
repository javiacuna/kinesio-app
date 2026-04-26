export type Patient = {
  id: string;
  dni: string;
  first_name: string;
  last_name: string;
  email: string;
  phone?: string | null;
  birth_date?: string | null;
  clinical_notes?: string | null;
  active: boolean;
  created_at?: string;
  updated_at?: string;
};
