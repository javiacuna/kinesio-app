import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listNotifications, markAllNotificationsRead, markNotificationRead, type AppNotification } from "./api";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

export function NotificationBell({ variant = "sidebar" }: { variant?: "sidebar" | "header" }) {
  const [open, setOpen] = useState(false);
  const { language, t } = useLanguage();
  const queryClient = useQueryClient();
  const notificationsQ = useQuery({
    queryKey: ["notifications"],
    queryFn: () => listNotifications({ limit: 20 }),
    refetchInterval: 30_000,
  });

  const markRead = useMutation({
    mutationFn: markNotificationRead,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notifications"] }),
  });
  const markAllRead = useMutation({
    mutationFn: markAllNotificationsRead,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notifications"] }),
  });

  const items = notificationsQ.data?.items ?? [];
  const unreadCount = notificationsQ.data?.unread_count ?? 0;
  const dateFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat(language === "en" ? "en-US" : "es-AR", {
        day: "2-digit",
        month: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      }),
    [language],
  );

  function selectNotification(notification: AppNotification) {
    if (!notification.read_at) {
      markRead.mutate(notification.id);
    }
  }

  if (variant === "header") {
    return (
      <div className="relative">
        <button
          type="button"
          className="relative flex h-10 w-10 items-center justify-center rounded-full border bg-white text-gray-700 shadow-sm hover:bg-gray-100 transition-colors"
          onClick={() => setOpen((value) => !value)}
          aria-expanded={open}
          aria-label={t("notifications.title")}
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.8} strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5">
            <path d="M18 8a6 6 0 1 0-12 0c0 4-1.5 5.5-1.5 6.5h15C18 13.5 18 12 18 8Z" />
            <path d="M10.3 19a1.7 1.7 0 0 0 3.4 0" />
          </svg>
          {unreadCount > 0 && (
            <span className="absolute -top-1 -right-1 min-w-5 h-5 px-1 rounded-full bg-black text-white text-[11px] font-medium flex items-center justify-center border-2 border-white">
              {unreadCount > 99 ? "99+" : unreadCount}
            </span>
          )}
        </button>

        {open && (
          <div className="absolute right-0 mt-2 w-80 bg-white border rounded-lg shadow-lg overflow-hidden z-20">
            <div className="px-3 py-2 border-b flex items-center justify-between gap-3">
              <p className="text-sm font-semibold">{t("notifications.title")}</p>
              <button
                type="button"
                className="text-xs underline disabled:text-gray-400"
                onClick={() => markAllRead.mutate()}
                disabled={unreadCount === 0 || markAllRead.isPending}
              >
                {t("notifications.markAllRead")}
              </button>
            </div>

            <div className="max-h-96 overflow-auto">
              {notificationsQ.isLoading && <p className="p-3 text-sm text-gray-500">{t("notifications.loading")}</p>}
              {!notificationsQ.isLoading && items.length === 0 && (
                <p className="p-3 text-sm text-gray-500">{t("notifications.empty")}</p>
              )}
              {items.map((notification) => (
                <button
                  key={notification.id}
                  type="button"
                  className={`w-full text-left px-3 py-3 border-b last:border-b-0 hover:bg-gray-50 ${
                    notification.read_at ? "bg-white" : "bg-gray-50"
                  }`}
                  onClick={() => selectNotification(notification)}
                >
                  <span className="flex items-start justify-between gap-3">
                    <span className="text-sm font-semibold">{notification.title}</span>
                    {!notification.read_at && <span className="mt-1 h-2 w-2 rounded-full bg-black shrink-0" />}
                  </span>
                  <span className="mt-1 block text-sm text-gray-600">{notification.message}</span>
                  <span className="mt-2 block text-xs text-gray-500">
                    {dateFormatter.format(new Date(notification.created_at))}
                  </span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="relative">
      <button
        type="button"
        className="w-full flex items-center justify-between gap-2 px-3 py-2 rounded-lg border text-sm hover:bg-gray-100 bg-white"
        onClick={() => setOpen((value) => !value)}
        aria-expanded={open}
      >
        <span>{t("notifications.title")}</span>
        {unreadCount > 0 && (
          <span className="min-w-6 h-6 px-2 rounded-full bg-black text-white text-xs flex items-center justify-center">
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="mt-2 w-full lg:absolute lg:left-0 lg:bottom-full lg:mb-2 lg:w-80 bg-white border rounded-lg shadow-lg overflow-hidden z-20">
          <div className="px-3 py-2 border-b flex items-center justify-between gap-3">
            <p className="text-sm font-semibold">{t("notifications.title")}</p>
            <button
              type="button"
              className="text-xs underline disabled:text-gray-400"
              onClick={() => markAllRead.mutate()}
              disabled={unreadCount === 0 || markAllRead.isPending}
            >
              {t("notifications.markAllRead")}
            </button>
          </div>

          <div className="max-h-96 overflow-auto">
            {notificationsQ.isLoading && <p className="p-3 text-sm text-gray-500">{t("notifications.loading")}</p>}
            {!notificationsQ.isLoading && items.length === 0 && (
              <p className="p-3 text-sm text-gray-500">{t("notifications.empty")}</p>
            )}
            {items.map((notification) => (
              <button
                key={notification.id}
                type="button"
                className={`w-full text-left px-3 py-3 border-b last:border-b-0 hover:bg-gray-50 ${
                  notification.read_at ? "bg-white" : "bg-gray-50"
                }`}
                onClick={() => selectNotification(notification)}
              >
                <span className="flex items-start justify-between gap-3">
                  <span className="text-sm font-semibold">{notification.title}</span>
                  {!notification.read_at && <span className="mt-1 h-2 w-2 rounded-full bg-black shrink-0" />}
                </span>
                <span className="mt-1 block text-sm text-gray-600">{notification.message}</span>
                <span className="mt-2 block text-xs text-gray-500">
                  {dateFormatter.format(new Date(notification.created_at))}
                </span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
