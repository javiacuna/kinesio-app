import { getAuthToken } from "@/features/auth/session";

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);

  if (!headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const token = getAuthToken();
  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(path, { ...init, headers });

  if (!res.ok) {
    let body: any = null;
    try { body = await res.json(); } catch {}
    const msg = body?.error ?? `HTTP ${res.status}`;
    const e = new Error(msg);
    (e as any).status = res.status;
    (e as any).body = body;
    throw e;
  }

  return (await res.json()) as T;
}
