import { useState, type InputHTMLAttributes } from "react";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

type PasswordInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type">;

export function PasswordInput({ className = "", ...props }: PasswordInputProps) {
  const { t } = useLanguage();
  const [visible, setVisible] = useState(false);

  return (
    <div className="relative">
      <input {...props} type={visible ? "text" : "password"} className={`${className} pr-10`} />
      <button
        type="button"
        onClick={() => setVisible((prev) => !prev)}
        className="absolute inset-y-0 right-0 flex items-center px-3 text-gray-500 hover:text-gray-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-black rounded-r-lg"
        aria-label={visible ? t("common.hidePassword") : t("common.showPassword")}
        aria-pressed={visible}
      >
        {visible ? <EyeOffIcon /> : <EyeIcon />}
      </button>
    </div>
  );
}

function EyeIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="M1.5 12S5 5 12 5s10.5 7 10.5 7-3.5 7-10.5 7S1.5 12 1.5 12Z" strokeLinecap="round" strokeLinejoin="round" />
      <circle cx="12" cy="12" r="3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function EyeOffIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path
        d="M3 3l18 18M10.6 10.7a3 3 0 0 0 4.2 4.2M6.6 6.7C4.1 8.3 2.5 10.6 1.5 12c0 0 3.5 7 10.5 7 2 0 3.7-.5 5.1-1.3M9.9 5.2A11.6 11.6 0 0 1 12 5c7 0 10.5 7 10.5 7-.4.8-1.4 2.5-3 4.1"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
