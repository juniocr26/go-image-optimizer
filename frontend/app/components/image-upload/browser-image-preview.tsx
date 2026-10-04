import { useEffect, useState } from "react";
import { acquirePreview, nativeDisplayBlob } from "./preview-cache";
import { ImageIcon } from "./icons";

type BrowserImagePreviewProps = {
  alt: string;
  className: string;
  fallbackClassName: string;
  fileName: string;
  fileType: string;
  imageClassName: string;
  src: string;
  source: Blob;
};

export function BrowserImagePreview({
  alt,
  className,
  fallbackClassName,
  fileName,
  fileType,
  imageClassName,
  source,
  src,
}: BrowserImagePreviewProps) {
  return <PreviewSession key={src} source={source} alt={alt} className={className} fallbackClassName={fallbackClassName} fileName={fileName} fileType={fileType} imageClassName={imageClassName} src={src} />;
}

function PreviewSession({source, alt, className, fallbackClassName, fileName, fileType, imageClassName}: BrowserImagePreviewProps) {
  const [display, setDisplay] = useState<{url: string; generated: boolean} | null>(null);
  const [fallback, setFallback] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    let url: string | undefined;
    const request = fallback ? acquirePreview(source, attempt > 0) : null;
    setDisplay(null);
    setLoaded(false);
    setError("");
    (request?.promise ?? nativeDisplayBlob(source)).then(blob => {
      if (!active) return;
      url = URL.createObjectURL(blob);
      setDisplay({url, generated: fallback});
    }).catch(reason => {
      if (active) setError(reason instanceof Error ? reason.message : "Preview unavailable.");
    });
    return () => {
      active = false;
      request?.release();
      if (url) URL.revokeObjectURL(url);
    };
  }, [source, fallback, attempt]);

  return (
    <div className={className}>
      {error ? <>
        <PreviewUnavailable className={fallbackClassName} fileName={fileName} fileType={fileType} />
        <p role="status" className="px-4 text-center text-sm">{error}</p>
        <button type="button" className="mx-auto block p-3 text-sm font-bold" onClick={() => { setFallback(true); setAttempt(value => value + 1); }}>Retry preview</button>
      </> : <>
        {!loaded && <p role="status" className="p-4 text-center text-sm">Loading preview…</p>}
        {display && <img alt={alt} className={imageClassName} style={{objectFit:"contain", display: loaded ? undefined : "none"}} src={display.url}
          onLoad={() => setLoaded(true)}
          onError={() => {
            if (display.generated) setError("The generated preview could not be displayed.");
            else setFallback(true);
          }} />}
      </>}
    </div>
  );
}

function PreviewUnavailable({
  className,
  fileName,
  fileType,
}: {
  className: string;
  fileName: string;
  fileType: string;
}) {
  return (
    <div
      className={`grid place-items-center px-5 py-6 text-center ${className}`}
    >
      <div className="min-w-0 max-w-full">
        <div className="mx-auto grid h-12 w-12 place-items-center rounded-full bg-[#eef5fb] text-[#60708d]">
          <ImageIcon />
        </div>
        <p className="mt-4 truncate text-sm font-black text-[#081236]">
          {fileName}
        </p>
        {fileType ? (
          <p className="mt-2 truncate text-sm leading-6 text-[#60708d]">
            {fileType}
          </p>
        ) : null}
        <p className="mt-2 text-sm font-bold leading-6 text-[#60708d]">
          Preview unavailable. Processing and download are still available.
        </p>
      </div>
    </div>
  );
}
