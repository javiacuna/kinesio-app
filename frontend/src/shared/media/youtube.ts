export function getYouTubeEmbedUrl(rawUrl: string) {
  try {
    const url = new URL(rawUrl);
    const host = url.hostname.replace(/^www\./, "").replace(/^m\./, "").toLowerCase();
    let videoID = "";

    if (host === "youtu.be") {
      videoID = url.pathname.split("/").filter(Boolean)[0] ?? "";
    }

    if (host === "youtube.com" || host === "youtube-nocookie.com") {
      if (url.pathname === "/watch") {
        videoID = url.searchParams.get("v") ?? "";
      } else {
        const parts = url.pathname.split("/").filter(Boolean);
        if ((parts[0] === "embed" || parts[0] === "shorts") && parts[1]) {
          videoID = parts[1];
        }
      }
    }

    if (!/^[A-Za-z0-9_-]{6,}$/.test(videoID)) return null;
    return `https://www.youtube-nocookie.com/embed/${videoID}`;
  } catch {
    return null;
  }
}
