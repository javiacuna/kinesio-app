import { getYouTubeEmbedUrl } from "@/shared/media/youtube";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

export type VideoPreview = {
  title: string;
  url: string;
};

export function VideoPreviewModal({
  preview,
  onClose,
}: {
  preview: VideoPreview | null;
  onClose: () => void;
}) {
  const { t } = useLanguage();

  if (!preview) return null;

  const embedUrl = getYouTubeEmbedUrl(preview.url);
  if (!embedUrl) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4" role="dialog" aria-modal="true">
      <div className="w-full max-w-4xl rounded-lg bg-white shadow-xl">
        <div className="flex items-start justify-between gap-3 border-b p-4">
          <div>
            <h2 className="text-lg font-semibold">{preview.title}</h2>
            <a className="text-sm text-blue-700 underline" href={preview.url} target="_blank" rel="noreferrer">
              {t("common.openInYoutube")}
            </a>
          </div>
          <button type="button" className="rounded-lg border px-3 py-1 text-sm hover:bg-gray-50" onClick={onClose}>
            {t("common.close")}
          </button>
        </div>
        <div className="aspect-video bg-black">
          <iframe
            className="h-full w-full"
            src={embedUrl}
            title={preview.title}
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
            allowFullScreen
          />
        </div>
      </div>
    </div>
  );
}
