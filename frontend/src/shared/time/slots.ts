import { hhmmToMinutes } from "./local";

export function minutesToHHmm(total: number): string {
  const h = Math.floor(total / 60);
  const m = total % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`;
}

export function generateSlots(startHHmm = "08:00", endHHmm = "20:00", stepMin = 15) {
  const start = hhmmToMinutes(startHHmm);
  const end = hhmmToMinutes(endHHmm);

  const out: string[] = [];
  for (let t = start; t <= end; t += stepMin) {
    out.push(minutesToHHmm(t));
  }
  return out;
}
