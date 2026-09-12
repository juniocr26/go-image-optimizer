"use client";

import {
  type ChangeEvent,
  type DragEvent,
  type FormEvent,
  useEffect,
  useRef,
  useState,
} from "react";

const MAX_UPLOAD_BYTES = 25 * 1024 * 1024;
const SUPPORTED_IMAGE_TYPES = new Set(["image/jpeg", "image/png"]);

type OptimizationResult = {
  name: string;
  originalSize: number;
  size: number;
  type: string;
  url: string;
};

export function ImageUploadForm() {
  const inputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<OptimizationResult | null>(null);

  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  useEffect(() => {
    return () => {
      if (result?.url) {
        URL.revokeObjectURL(result.url);
      }
    };
  }, [result?.url]);

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;
    if (file) {
      chooseFile(file);
    }
  }

  function chooseFile(file: File) {
    if (isSubmitting) {
      return;
    }

    setError(null);
    setResult(null);

    if (!SUPPORTED_IMAGE_TYPES.has(file.type)) {
      setSelectedFile(null);
      setPreviewUrl(null);
      setError("Only JPG and PNG images are supported.");
      return;
    }

    if (file.size > MAX_UPLOAD_BYTES) {
      setSelectedFile(null);
      setPreviewUrl(null);
      setError("The selected image is larger than 25 MiB.");
      return;
    }

    setSelectedFile(file);
    setPreviewUrl(URL.createObjectURL(file));
  }

  function resetWorkflow() {
    if (isSubmitting) {
      return;
    }

    setSelectedFile(null);
    setPreviewUrl(null);
    setError(null);
    setResult(null);

    if (inputRef.current) {
      inputRef.current.value = "";
    }
  }

  function openFilePicker() {
    if (isSubmitting) {
      return;
    }

    if (inputRef.current) {
      inputRef.current.value = "";
      inputRef.current.click();
    }
  }

  function handleDragOver(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();

    if (isSubmitting) {
      event.dataTransfer.dropEffect = "none";
      return;
    }

    event.dataTransfer.dropEffect = "copy";
    setIsDragging(true);
  }

  function handleDragLeave(event: DragEvent<HTMLDivElement>) {
    if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
      setIsDragging(false);
    }
  }

  function handleDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    setIsDragging(false);

    const file = event.dataTransfer.files?.[0] ?? null;
    if (file) {
      chooseFile(file);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!selectedFile) {
      setError("Choose an image before compressing.");
      return;
    }

    if (isSubmitting || result) {
      return;
    }

    const formData = new FormData();
    formData.append("image", selectedFile);

    setIsSubmitting(true);
    setError(null);
    setResult(null);

    try {
      const response = await fetch("/api/images/compress", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        setError(await readErrorMessage(response));
        return;
      }

      const returnedImage = await response.blob();
      const downloadName =
        readDownloadName(response.headers.get("Content-Disposition")) ??
        buildCompressedFilename(selectedFile.name, returnedImage.type || selectedFile.type);

      setResult({
        name: downloadName,
        originalSize: selectedFile.size,
        size: returnedImage.size,
        type: returnedImage.type || selectedFile.type || "application/octet-stream",
        url: URL.createObjectURL(returnedImage),
      });
    } catch {
      setError("Could not process the image right now.");
    } finally {
      setIsSubmitting(false);
    }
  }

  const reduction = result
    ? ((result.originalSize - result.size) / result.originalSize) * 100
    : 0;
  const hasReduction = result ? result.size < result.originalSize : false;

  return (
    <form
      aria-busy={isSubmitting}
      className="mt-8 rounded-[22px] bg-white/92 p-3 shadow-[0_24px_60px_rgba(10,32,70,0.08)] ring-1 ring-white/80 backdrop-blur"
      onSubmit={handleSubmit}
    >
      <div className="grid gap-4">
        <div
          className={`grid min-h-[248px] place-items-center rounded-[18px] border border-dashed px-4 py-7 text-center transition sm:min-h-[265px] sm:px-8 ${
            isDragging
              ? "border-[#08a87d] bg-[#ecfbf7]"
              : "border-[#9eb6d8] bg-white/54"
          }`}
          onDragLeave={handleDragLeave}
          onDragOver={handleDragOver}
          onDrop={handleDrop}
        >
          <input
            ref={inputRef}
            accept="image/jpeg,image/png"
            aria-label="Choose a JPG or PNG image"
            className="sr-only"
            disabled={isSubmitting}
            name="image"
            onChange={handleFileChange}
            type="file"
          />

          <div className="grid w-full justify-items-center">
            {previewUrl && selectedFile ? (
              <div className="w-full max-w-[420px] overflow-hidden rounded-2xl border border-[#dbe8f1] bg-white shadow-[0_16px_42px_rgba(15,42,80,0.12)]">
                <img
                  alt={`Preview of ${selectedFile.name}`}
                  className="h-56 w-full object-contain"
                  src={previewUrl}
                />
              </div>
            ) : (
              <div className="grid h-16 w-16 place-items-center rounded-full bg-[#d9f4ec] text-[#08a87d] shadow-[0_16px_36px_rgba(0,168,125,0.16)]">
                <UploadIcon />
              </div>
            )}

            <p className="mt-5 text-lg font-black text-[#081236]">
              {selectedFile ? "Image ready to compress" : "Drag and drop an image here"}
            </p>
            <p className="mt-2 text-base text-[#344464]">
              {selectedFile ? "Review the file and start when ready" : "or click to select a file"}
            </p>
            <p className="mt-5 text-sm leading-6 text-[#60708d] sm:text-base">
              JPG and PNG only, up to 25 MiB.
            </p>

            {selectedFile ? (
              <dl className="mt-5 grid w-full max-w-[520px] gap-3 rounded-2xl border border-[#dbe8f1] bg-white/80 p-4 text-left sm:grid-cols-[1.4fr_0.8fr]">
                <div className="min-w-0">
                  <dt className="text-xs font-black uppercase tracking-normal text-[#6f7f9d]">
                    File
                  </dt>
                  <dd className="mt-1 truncate text-sm font-bold text-[#081236]">
                    {selectedFile.name}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs font-black uppercase tracking-normal text-[#6f7f9d]">
                    Original
                  </dt>
                  <dd className="mt-1 text-sm font-bold text-[#081236]">
                    {formatBytes(selectedFile.size)}
                  </dd>
                </div>
              </dl>
            ) : null}

            {isSubmitting ? (
              <div className="mt-5 w-full max-w-[420px]" role="status">
                <div className="h-2 overflow-hidden rounded-full bg-[#dce7ee]">
                  <div className="indeterminate-progress h-full w-1/2 rounded-full bg-[#08a87d]" />
                </div>
                <p className="mt-3 text-sm font-bold text-[#078665]">
                  Compressing image...
                </p>
              </div>
            ) : null}

            <div className="mt-5 flex w-full flex-col items-center justify-center gap-3 sm:flex-row">
              <button
                className="inline-flex h-14 w-full items-center justify-center gap-3 rounded-lg bg-[#05a97f] px-8 text-base font-black text-white shadow-[0_16px_34px_rgba(0,168,125,0.25)] transition hover:bg-[#02976f] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:bg-[#8ccfbd] sm:w-[220px]"
                disabled={isSubmitting}
                onClick={openFilePicker}
                type="button"
              >
                <FolderIcon />
                Choose file
              </button>

              {selectedFile && !result ? (
                <button
                  className="inline-flex h-14 w-full items-center justify-center gap-3 rounded-lg border border-[#08a87d]/35 bg-white px-7 text-base font-black text-[#078665] shadow-sm transition hover:bg-[#f0fbf8] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:border-[#c8d3df] disabled:text-[#93a1b6] sm:w-auto"
                  disabled={isSubmitting}
                  type="submit"
                >
                  {isSubmitting ? (
                    <>
                      <SpinnerIcon />
                      Compressing
                    </>
                  ) : (
                    <>
                      <CompressIcon />
                      Compress
                    </>
                  )}
                </button>
              ) : null}

              {selectedFile || result ? (
                <button
                  className="inline-flex h-14 w-full items-center justify-center gap-3 rounded-lg border border-[#d6e1ec] bg-white px-7 text-base font-black text-[#344464] shadow-sm transition hover:bg-[#f7fbff] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#9eb6d8] disabled:cursor-not-allowed disabled:text-[#93a1b6] sm:w-auto"
                  disabled={isSubmitting}
                  onClick={resetWorkflow}
                  type="button"
                >
                  <ResetIcon />
                  Another image
                </button>
              ) : null}
            </div>
          </div>
        </div>

        {error ? (
          <div className="rounded-xl border border-[#fecaca] bg-[#fff5f5] px-4 py-3 text-sm font-medium text-[#991b1b]">
            {error}
          </div>
        ) : null}

        {result ? (
          <div className="grid gap-4 rounded-2xl border border-[#bddfe1] bg-[#f0fbf8] p-4">
            <div>
              <p className="text-sm font-black text-[#066f57]">
                Compression complete
              </p>
              <p className="mt-1 truncate text-sm leading-6 text-[#344464]">
                {result.name} - {result.type}
              </p>
            </div>

            {result.type.startsWith("image/") ? (
              <img
                alt="Optimized image preview"
                className="max-h-72 w-full rounded-xl border border-[#bddfe1] bg-white object-contain"
                src={result.url}
              />
            ) : null}

            <dl className="grid gap-3 sm:grid-cols-3">
              <ResultMetric label="Original" value={formatBytes(result.originalSize)} />
              <ResultMetric label="Optimized" value={formatBytes(result.size)} />
              <ResultMetric
                label="Reduction"
                value={hasReduction ? `${reduction.toFixed(1)}%` : "No reduction"}
              />
            </dl>

            {!hasReduction ? (
              <p className="rounded-xl border border-[#f6d49c] bg-[#fff9ec] px-4 py-3 text-sm font-bold text-[#875a07]">
                The optimized file is not smaller than the original.
              </p>
            ) : null}

            <div className="flex flex-col gap-3 sm:flex-row">
              <a
                className="inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg bg-[#081236] px-5 text-center text-sm font-black text-white transition hover:bg-[#14234a] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#081236] sm:w-fit"
                download={result.name}
                href={result.url}
              >
                <DownloadIcon />
                Download
              </a>

              <button
                className="inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg border border-[#bddfe1] bg-white px-5 text-sm font-black text-[#066f57] transition hover:bg-[#f8fffd] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] sm:w-fit"
                onClick={resetWorkflow}
                type="button"
              >
                <ResetIcon />
                Another image
              </button>
            </div>
          </div>
        ) : null}
      </div>
    </form>
  );
}

function ResultMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl bg-white px-4 py-3">
      <dt className="text-xs font-black uppercase tracking-normal text-[#6f7f9d]">
        {label}
      </dt>
      <dd className="mt-1 text-lg font-black text-[#081236]">{value}</dd>
    </div>
  );
}

function UploadIcon() {
  return (
    <svg aria-hidden="true" className="h-8 w-8" fill="none" viewBox="0 0 24 24">
      <path
        d="M12 15V4M12 4 8 8M12 4l4 4M5 14v4.2c0 .5.2.9.5 1.3.4.3.8.5 1.3.5h10.4c.5 0 .9-.2 1.3-.5.3-.4.5-.8.5-1.3V14"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.2"
      />
    </svg>
  );
}

function FolderIcon() {
  return (
    <svg aria-hidden="true" className="h-6 w-6" fill="none" viewBox="0 0 24 24">
      <path
        d="M3.8 7.5c0-1 .8-1.8 1.8-1.8h4.1l1.8 2.2h6.9c1 0 1.8.8 1.8 1.8v6.8c0 1-.8 1.8-1.8 1.8H5.6c-1 0-1.8-.8-1.8-1.8v-9Z"
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="2"
      />
    </svg>
  );
}

function CompressIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24">
      <path
        d="M5 7h14M7 12h10M10 17h4"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="2.2"
      />
    </svg>
  );
}

function DownloadIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24">
      <path
        d="M12 4v10M12 14l4-4M12 14l-4-4M5 19h14"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.2"
      />
    </svg>
  );
}

function ResetIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24">
      <path
        d="M5.5 8.2A7.5 7.5 0 1 1 4.8 16M5.5 8.2H10M5.5 8.2V3.8"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.2"
      />
    </svg>
  );
}

function SpinnerIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="9"
        stroke="currentColor"
        strokeWidth="3"
      />
      <path
        className="opacity-85"
        d="M21 12a9 9 0 0 0-9-9"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="3"
      />
    </svg>
  );
}

async function readErrorMessage(response: Response) {
  const contentType = response.headers.get("Content-Type") ?? "";

  if (contentType.includes("application/json")) {
    try {
      const data = (await response.json()) as { error?: string };
      return data.error ?? "The image could not be processed.";
    } catch {
      return "The image could not be processed.";
    }
  }

  return (await response.text()) || "The image could not be processed.";
}

function readDownloadName(contentDisposition: string | null) {
  if (!contentDisposition) {
    return null;
  }

  const encodedFilename = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (encodedFilename?.[1]) {
    try {
      return decodeURIComponent(encodedFilename[1]);
    } catch {
      return encodedFilename[1];
    }
  }

  const filename = contentDisposition.match(/filename="?([^";]+)"?/i);
  return filename?.[1] ?? null;
}

function buildCompressedFilename(filename: string, contentType: string) {
  const extension = contentType === "image/png" ? ".png" : ".jpg";
  const base = filename.replace(/\.[^/.]+$/, "") || "image";

  return `${base}_compressed${extension}`;
}

function formatBytes(bytes: number) {
  if (bytes < 1024) {
    return `${bytes} B`;
  }

  const kilobytes = bytes / 1024;
  if (kilobytes < 1024) {
    return `${kilobytes.toFixed(1)} KB`;
  }

  return `${(kilobytes / 1024).toFixed(1)} MB`;
}
