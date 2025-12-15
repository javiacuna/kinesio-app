// Convierte fecha+hora local (del navegador) a RFC3339 en UTC
export function localDateTimeToUTC(dateISO: string, timeHHmm: string): string {
  // dateISO: "2025-12-15", timeHHmm: "14:00"
  const [y, m, d] = dateISO.split("-").map(Number);
  const [hh, mm] = timeHHmm.split(":").map(Number);

  // new Date(y, m-1, d, hh, mm) interpreta en horario local
  const local = new Date(y, (m - 1), d, hh, mm, 0, 0);

  return local.toISOString(); // UTC RFC3339
}

// suma minutos a una hora local y devuelve "HH:MM"
export function addMinutesToHHmm(timeHHmm: string, minutes: number): string {
  const [hh, mm] = timeHHmm.split(":").map(Number);
  const total = hh * 60 + mm + minutes;
  const h2 = Math.floor((total % (24 * 60)) / 60);
  const m2 = total % 60;
  return `${String(h2).padStart(2, "0")}:${String(m2).padStart(2, "0")}`;
}
