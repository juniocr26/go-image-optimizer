import { useEffect, useState } from "react";
import { ImageIcon } from "./icons";

type BrowserImagePreviewProps = {
  alt: string;
  className: string;
  fallbackClassName: string;
  fileName: string;
  fileType: string;
  imageClassName: string;
  src: string;
};

export function BrowserImagePreview({
  alt,
  className,
  fallbackClassName,
  fileName,
  fileType,
  imageClassName,
  src,
}: BrowserImagePreviewProps) {
  const [previewFailed, setPreviewFailed] = useState(false);

  useEffect(() => {
    setPreviewFailed(false);
  }, [src]);

  return (
    <div className={className}>
      {previewFailed ? (
        <PreviewUnavailable
          className={fallbackClassName}
          fileName={fileName}
          fileType={fileType}
        />
      ) : (
        <img
          alt={alt}
          className={imageClassName}
          onError={() => setPreviewFailed(true)}
          src={src}
        />
      )}
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
          Preview not available for this format.
        </p>
      </div>
    </div>
  );
}
