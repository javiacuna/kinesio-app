import { apiFetch } from "@/shared/api/http";

export type AppNotification = {
  id: string;
  recipient_email: string;
  recipient_role?: string;
  type: string;
  title: string;
  message: string;
  entity_type?: string;
  entity_id?: string;
  read_at?: string;
  created_at: string;
};

export type NotificationsResponse = {
  items: AppNotification[];
  unread_count: number;
};

export function listNotifications(input?: { limit?: number; unreadOnly?: boolean }) {
  const qs = new URLSearchParams();
  if (input?.limit) qs.set("limit", String(input.limit));
  if (input?.unreadOnly) qs.set("unread_only", "true");
  const query = qs.toString();
  return apiFetch<NotificationsResponse>(`/api/v1/notifications${query ? `?${query}` : ""}`);
}

export function markNotificationRead(id: string) {
  return apiFetch<void>(`/api/v1/notifications/${id}/read`, { method: "POST" });
}

export function markAllNotificationsRead() {
  return apiFetch<void>("/api/v1/notifications/read-all", { method: "POST" });
}
