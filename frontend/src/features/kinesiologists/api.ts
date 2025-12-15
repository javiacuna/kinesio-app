import { apiFetch } from "@/shared/api/http";
import type { Kinesiologist } from "./types";

export function listKinesiologists() {
  return apiFetch<Kinesiologist[]>("/api/v1/kinesiologists");
}
