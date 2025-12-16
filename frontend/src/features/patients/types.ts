export type Patient = {
  id: string;
  dni: string;
  first_name: string;
  last_name: string;
  email: string;
  phone?: string | null;
};
