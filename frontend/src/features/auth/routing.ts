import type { AuthRole } from "./types";

export function homePathForRole(role?: AuthRole | "") {
  switch (role) {
    case "paciente":
      return "/portal";
    case "kinesiologo":
      return "/agenda";
    case "recepcionista":
    case "admin":
    default:
      return "/agenda";
  }
}
