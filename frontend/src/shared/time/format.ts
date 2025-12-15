export function formatLocalDateTime(iso: string): string {
  const d = new Date(iso);
  // Ej: 15/12/2025 10:30
  return d.toLocaleString(undefined, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatLocalTime(iso: string): string {
  const d = new Date(iso);
  // Ej: 10:30
  return d.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
}
