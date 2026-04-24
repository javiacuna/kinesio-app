import type { AuthRole } from "@/features/auth/types";

export type StaffRole = Extract<AuthRole, "admin" | "recepcionista" | "kinesiologo">;

export type StaffMember = {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: StaffRole;
  phone?: string | null;
  active: boolean;
  firebase_uid?: string | null;
  created_at: string;
  updated_at: string;
};
