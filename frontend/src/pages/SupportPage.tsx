import { useRef, useState, type FormEvent } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  downloadSupportAttachment,
  getSupportInfo,
  listSupportTickets,
  sendSupportContact,
  type SupportTicket,
} from "@/features/support/api";
import { useLanguage } from "@/shared/i18n/LanguageProvider";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";

export default function SupportPage() {
  const { t } = useLanguage();
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [success, setSuccess] = useState("");
  const [error, setError] = useState("");
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const queryClient = useQueryClient();

  const infoQ = useQuery({
    queryKey: ["support", "info"],
    queryFn: () => getSupportInfo(),
  });

  const ticketsQ = useQuery({
    queryKey: ["support", "tickets"],
    queryFn: () => listSupportTickets(),
  });

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSuccess("");
    setError("");

    if (!subject.trim() || !message.trim()) {
      setError(t("support.requiredFields"));
      return;
    }

    setIsSubmitting(true);
    try {
      await sendSupportContact({ subject: subject.trim(), message: message.trim(), file });
      setSubject("");
      setMessage("");
      setFile(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
      setSuccess(t("support.sent"));
      queryClient.invalidateQueries({ queryKey: ["support", "tickets"] });
    } catch {
      setError(t("support.sendFailed"));
    } finally {
      setIsSubmitting(false);
    }
  }

  async function openAttachment(ticket: SupportTicket) {
    const blob = await downloadSupportAttachment(ticket);
    const url = URL.createObjectURL(blob);
    window.open(url, "_blank", "noopener,noreferrer");
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000);
  }

  const info = infoQ.data;
  const tickets = ticketsQ.data ?? [];

  return (
    <main className="max-w-2xl mx-auto p-6 space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">{t("support.title")}</h1>
        <p className="text-sm text-gray-600">{t("support.subtitle")}</p>
      </header>

      <form onSubmit={submit} className="bg-white border rounded-lg p-4 space-y-4">
        <div>
          <label className="text-sm font-medium" htmlFor="support-subject">
            <RequiredLabel required>{t("support.subject")}</RequiredLabel>
          </label>
          <input
            id="support-subject"
            className="mt-1 w-full border rounded-lg p-2"
            type="text"
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            maxLength={150}
          />
        </div>

        <div>
          <label className="text-sm font-medium" htmlFor="support-message">
            <RequiredLabel required>{t("support.message")}</RequiredLabel>
          </label>
          <textarea
            id="support-message"
            className="mt-1 w-full border rounded-lg p-2 min-h-32"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            maxLength={4000}
          />
        </div>

        <label className="text-sm font-medium block">
          {t("support.attachment")}
          <input
            ref={fileInputRef}
            className="mt-1 w-full border rounded-lg p-2 bg-white text-sm"
            type="file"
            accept="image/*,.pdf"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          />
          <span className="mt-1 block text-xs text-gray-500">{t("support.attachmentHint")}</span>
        </label>

        {error && <p className="text-sm text-red-600">{error}</p>}
        {success && <p className="text-sm text-green-600">{success}</p>}

        <button
          type="submit"
          className="px-4 py-2 rounded-lg bg-black text-white text-sm hover:bg-gray-800 disabled:opacity-50"
          disabled={isSubmitting}
        >
          {isSubmitting ? t("support.sending") : t("support.send")}
        </button>
      </form>

      {info?.phone && (
        <section className="bg-white border rounded-lg p-4">
          <h2 className="text-sm font-semibold text-gray-700">{t("support.phoneTitle")}</h2>
          <p className="mt-1 text-lg font-medium">{info.phone}</p>
        </section>
      )}

      <section className="bg-white border rounded-lg p-4">
        <h2 className="text-sm font-semibold text-gray-700">{t("support.history")}</h2>
        {tickets.length === 0 ? (
          <p className="mt-2 text-sm text-gray-500">{t("support.historyEmpty")}</p>
        ) : (
          <div className="mt-3 overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="py-2 pr-4 font-medium">{t("support.date")}</th>
                  <th className="py-2 pr-4 font-medium">{t("support.subject")}</th>
                  <th className="py-2 pr-4 font-medium">{t("support.message")}</th>
                  <th className="py-2 pr-4 font-medium">{t("support.attachmentColumn")}</th>
                </tr>
              </thead>
              <tbody>
                {tickets.map((ticket) => (
                  <tr key={ticket.id} className="border-b last:border-0 align-top">
                    <td className="py-2 pr-4 whitespace-nowrap text-gray-600">
                      {new Date(ticket.created_at).toLocaleString()}
                    </td>
                    <td className="py-2 pr-4 font-medium">{ticket.subject}</td>
                    <td className="py-2 pr-4 text-gray-600 max-w-md">
                      <p className="line-clamp-3 whitespace-pre-wrap">{ticket.message}</p>
                    </td>
                    <td className="py-2 pr-4">
                      {ticket.attachment_url ? (
                        <button
                          type="button"
                          className="text-blue-600 hover:underline"
                          onClick={() => openAttachment(ticket)}
                        >
                          {ticket.attachment_file_name ?? t("support.attachment")}
                        </button>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
}
