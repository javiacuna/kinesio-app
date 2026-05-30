import { Link, isRouteErrorResponse, useRouteError } from "react-router-dom";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

export function ErrorPage() {
  const error = useRouteError();
  const { t } = useLanguage();
  const detail = errorDetail(error);

  return (
    <main className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
      <section className="w-full max-w-lg bg-white border rounded-lg p-6 space-y-4">
        <div>
          <p className="text-xs uppercase text-gray-500">{t("app.name")}</p>
          <h1 className="text-2xl font-semibold mt-1">{t("error.title")}</h1>
          <p className="text-sm text-gray-600 mt-2">{t("error.subtitle")}</p>
        </div>

        {detail && (
          <pre className="text-xs bg-gray-100 border rounded-lg p-3 overflow-auto text-gray-700">
            {detail}
          </pre>
        )}

        <div className="flex flex-wrap gap-3">
          <Link className="px-3 py-2 rounded-lg bg-black text-white text-sm" to="/">
            {t("error.goHome")}
          </Link>
          <button
            type="button"
            className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-100"
            onClick={() => window.location.reload()}
          >
            {t("error.reload")}
          </button>
        </div>
      </section>
    </main>
  );
}

function errorDetail(error: unknown) {
  if (isRouteErrorResponse(error)) {
    return `${error.status} ${error.statusText}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "";
}
