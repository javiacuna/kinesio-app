import type { AuthRole } from "./types";

export function homePathForRole(role?: AuthRole | "") {
  switch (role) {
    case "paciente":
      return "/portal";
    case "kinesiologo":
    case "recepcionista":
    case "admin":
      return "/agenda";
    default:
      return "/unauthorized";
  }
}
