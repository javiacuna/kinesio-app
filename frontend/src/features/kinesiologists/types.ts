export type Specialty = {
  id: string;
  name: string;
  active: boolean;
};

export type Practice = {
  id: string;
  specialty_id: string;
  name: string;
  description?: string | null;
  active: boolean;
};

export type Kinesiologist = {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  license_number?: string | null;
  work_start_time: string;
  work_end_time: string;
  work_days: number[];
  practice_ids: string[];
  practices: Practice[];
  active: boolean;
};
