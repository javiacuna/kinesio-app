export type Kinesiologist = {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  license_number?: string | null;
  work_start_time: string;
  work_end_time: string;
  work_days: number[];
  active: boolean;
};
