import { apiFetch } from "@/shared/api/http";
import type { StaffMember, StaffRole } from "./types";

export type SaveStaffMemberInput = {
  id?: string;
  first_name: string;
  last_name: string;
  email: string;
  role: StaffRole;
  phone?: string | null;
  active: boolean;
  firebase_uid?: string | null;
};

export function listStaffMembers(options?: { includeInactive?: boolean }) {
  const query = options?.includeInactive ? "?active=false" : "";
  return apiFetch<StaffMember[]>(`/api/v1/staff${query}`);
}

export function createStaffMember(input: SaveStaffMemberInput) {
  return apiFetch<StaffMember>("/api/v1/staff", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateStaffMember(input: SaveStaffMemberInput & { id: string }) {
  const { id, ...body } = input;
  return apiFetch<StaffMember>(`/api/v1/staff/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}
