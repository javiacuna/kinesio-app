const DEMO_TOKEN = "demo-recepcionista-token";

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);

  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  headers.set("Authorization", `Bearer ${DEMO_TOKEN}`);

  const res = await fetch(path, { ...init, headers });

  if (!res.ok) {
    let body: any = null;
    try {
      body = await res.json();
    } catch {}
    throw new Error(body?.error ?? `HTTP ${res.status}`);
  }

  return (await res.json()) as T;
}
