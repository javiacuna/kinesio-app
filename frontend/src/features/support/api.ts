import { apiFetch, apiFetchBlob } from "@/shared/api/http";

export type SupportInfo = {
  email: string;
  phone: string;
};

export function getSupportInfo() {
  return apiFetch<SupportInfo>("/api/v1/support/info");
}

export function sendSupportContact(input: { subject: string; message: string; file?: File | null }) {
  const formData = new FormData();
  formData.set("subject", input.subject);
  formData.set("message", input.message);
  if (input.file) {
    formData.set("file", input.file);
  }
  return apiFetch<{ ok: boolean }>("/api/v1/support/contact", {
    method: "POST",
    body: formData,
  });
}

export type SupportTicket = {
  id: string;
  subject: string;
  message: string;
  attachment_file_name?: string;
  attachment_url?: string;
  created_at: string;
};

export function listSupportTickets() {
  return apiFetch<SupportTicket[]>("/api/v1/support/tickets");
}

export function downloadSupportAttachment(ticket: SupportTicket) {
  if (!ticket.attachment_url) {
    return Promise.reject(new Error("no_attachment"));
  }
  return apiFetchBlob(ticket.attachment_url);
}
