import { useLanguage } from "@/shared/i18n/LanguageProvider";

export function RequiredLabel({ children, required = false }: { children: string; required?: boolean }) {
  const { t } = useLanguage();
  return (
    <>
      {children}
      {required && <span className="text-red-600 ml-1" aria-label={t("common.required")}>*</span>}
    </>
  );
}
