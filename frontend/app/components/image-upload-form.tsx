"use client";

import {
  type ChangeEvent,
  type DragEvent,
  type FormEvent,
  type RefObject,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";

const MAX_UPLOAD_BYTES = 50 * 1024 * 1024;
const SUPPORTED_FORMAT_COPY =
  "JPG/JPEG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP and TIFF";
const SUPPORTED_IMAGE_TYPES = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/avif",
  "image/heic",
  "image/heif",
  "image/gif",
  "image/bmp",
  "image/x-ms-bmp",
  "image/tiff",
]);
const SUPPORTED_IMAGE_EXTENSIONS = new Set([
  ".jpg",
  ".jpeg",
  ".png",
  ".webp",
  ".avif",
  ".heic",
  ".heif",
  ".gif",
  ".bmp",
  ".tif",
  ".tiff",
]);
const IMAGE_ACCEPT = [
  ".jpg",
  ".jpeg",
  ".png",
  ".webp",
  ".avif",
  ".heic",
  ".heif",
  ".gif",
  ".bmp",
  ".tif",
  ".tiff",
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/avif",
  "image/heic",
  "image/heif",
  "image/gif",
  "image/bmp",
  "image/tiff",
].join(",");
const IMAGE_ACTIONS = [{ id: "compress", label: "Compress" }] as const;
const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

type ImageActionId = (typeof IMAGE_ACTIONS)[number]["id"];

type OptimizationResult = {
  name: string;
  originalSize: number;
  size: number;
  type: string;
  url: string;
};

export function ImageUploadForm() {
  const inputRef = useRef<HTMLInputElement>(null);
  const runButtonRef = useRef<HTMLButtonElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [selectedAction, setSelectedAction] =
    useState<ImageActionId>("compress");
  const [isDragging, setIsDragging] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<OptimizationResult | null>(null);
  const [isResultModalOpen, setIsResultModalOpen] = useState(false);

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
    setIsResultModalOpen(false);

    if (!isSupportedImageFile(file)) {
      setSelectedFile(null);
      setPreviewUrl(null);
      setError(`Only ${SUPPORTED_FORMAT_COPY} images are supported.`);
      return;
    }

    if (file.size > MAX_UPLOAD_BYTES) {
      setSelectedFile(null);
      setPreviewUrl(null);
      setError("The selected image is larger than 50 MiB.");
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
    setIsResultModalOpen(false);

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
      setError("Choose an image before running an action.");
      return;
    }

    if (isSubmitting) {
      return;
    }

    if (selectedAction !== "compress") {
      setError("Choose a supported image action.");
      return;
    }

    const formData = new FormData();
    formData.append("image", selectedFile);

    setIsSubmitting(true);
    setError(null);
    setResult(null);
    setIsResultModalOpen(false);

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
        buildCompressedFilename(
          selectedFile.name,
          returnedImage.type || selectedFile.type,
        );

      setResult({
        name: downloadName,
        originalSize: selectedFile.size,
        size: returnedImage.size,
        type:
          returnedImage.type || selectedFile.type || "application/octet-stream",
        url: URL.createObjectURL(returnedImage),
      });
      setIsResultModalOpen(true);
    } catch {
      setError("Could not process the image right now.");
    } finally {
      setIsSubmitting(false);
    }
  }

  function closeResultModal() {
    setIsResultModalOpen(false);
  }

  return (
    <form
      aria-busy={isSubmitting}
      className="mt-8 rounded-[22px] bg-white/92 p-3 shadow-[0_24px_60px_rgba(10,32,70,0.08)] ring-1 ring-white/80 backdrop-blur"
      onSubmit={handleSubmit}
    >
      <div
        aria-hidden={isResultModalOpen ? true : undefined}
        className="grid gap-4"
      >
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
            accept={IMAGE_ACCEPT}
            aria-label="Choose a supported image"
            className="sr-only"
            disabled={isSubmitting}
            name="image"
            onChange={handleFileChange}
            type="file"
          />

          <div className="grid w-full justify-items-center">
            {previewUrl && selectedFile ? (
              <BrowserImagePreview
                alt={`Preview of ${selectedFile.name}`}
                className="w-full max-w-[420px] overflow-hidden rounded-2xl border border-[#dbe8f1] bg-white shadow-[0_16px_42px_rgba(15,42,80,0.12)]"
                fileName={selectedFile.name}
                fileType={selectedFile.type}
                imageClassName="h-48 w-full object-contain sm:h-56"
                fallbackClassName="h-48 sm:h-56"
                src={previewUrl}
              />
            ) : (
              <div className="grid h-16 w-16 place-items-center rounded-full bg-[#d9f4ec] text-[#08a87d] shadow-[0_16px_36px_rgba(0,168,125,0.16)]">
                <UploadIcon />
              </div>
            )}

            <p className="mt-5 text-lg font-black text-[#081236]">
              {selectedFile
                ? "Image ready to process"
                : "Drag and drop an image here"}
            </p>
            <p className="mt-2 text-base text-[#344464]">
              {selectedFile
                ? "Review the file and run an action"
                : "or click to select a file"}
            </p>
            <p className="mt-5 text-sm leading-6 text-[#60708d] sm:text-base">
              {SUPPORTED_FORMAT_COPY}, up to 50 MiB.
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

            <div className="mt-5 grid w-full max-w-[520px] gap-3">
              <div
                className={`grid w-full gap-3 ${
                  selectedFile ? "sm:grid-cols-2" : "justify-items-center"
                }`}
              >
                <button
                  className={`inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg bg-[#05a97f] px-7 text-base font-black text-white shadow-[0_16px_34px_rgba(0,168,125,0.25)] transition hover:bg-[#02976f] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:bg-[#8ccfbd] ${
                    selectedFile ? "" : "sm:w-[220px]"
                  }`}
                  disabled={isSubmitting}
                  onClick={openFilePicker}
                  type="button"
                >
                  <FolderIcon />
                  Choose file
                </button>

                {selectedFile ? (
                  <button
                    className="inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg border border-[#d6e1ec] bg-white px-7 text-base font-black text-[#344464] shadow-sm transition hover:bg-[#f7fbff] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#9eb6d8] disabled:cursor-not-allowed disabled:text-[#93a1b6]"
                    disabled={isSubmitting}
                    onClick={resetWorkflow}
                    type="button"
                  >
                    <ResetIcon />
                    Another image
                  </button>
                ) : null}
              </div>

              {selectedFile ? (
                <ImageOperationControls
                  isSubmitting={isSubmitting}
                  onActionChange={setSelectedAction}
                  runButtonRef={runButtonRef}
                  selectedAction={selectedAction}
                />
              ) : null}
            </div>
          </div>
        </div>

        {error ? (
          <div className="rounded-xl border border-[#fecaca] bg-[#fff5f5] px-4 py-3 text-sm font-medium text-[#991b1b]">
            {error}
          </div>
        ) : null}
      </div>

      {result ? (
        <ImageResultModal
          isOpen={isResultModalOpen}
          onClose={closeResultModal}
          result={result}
          returnFocusRef={runButtonRef}
        />
      ) : null}
    </form>
  );
}

function ImageOperationControls({
  isSubmitting,
  onActionChange,
  runButtonRef,
  selectedAction,
}: {
  isSubmitting: boolean;
  onActionChange: (action: ImageActionId) => void;
  runButtonRef: RefObject<HTMLButtonElement | null>;
  selectedAction: ImageActionId;
}) {
  const actionSelectId = useId();

  return (
    <div className="grid w-full gap-3 border-t border-[#dbe8f1]/80 pt-3 sm:grid-cols-[minmax(0,1fr)_124px]">
      <label className="sr-only" htmlFor={actionSelectId}>
        Select an image action
      </label>
      <div className="relative min-w-0">
        <select
          className="h-12 w-full appearance-none rounded-lg border border-[#cfe0ec] bg-white px-4 pr-11 text-base font-black text-[#081236] shadow-sm transition focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:bg-[#eef4f8] disabled:text-[#93a1b6]"
          disabled={isSubmitting}
          id={actionSelectId}
          onChange={(event) =>
            onActionChange(event.currentTarget.value as ImageActionId)
          }
          value={selectedAction}
        >
          {IMAGE_ACTIONS.map((action) => (
            <option key={action.id} value={action.id}>
              {action.label}
            </option>
          ))}
        </select>
        <span className="pointer-events-none absolute inset-y-0 right-4 flex items-center text-[#60708d]">
          <ChevronDownIcon />
        </span>
      </div>

      <button
        ref={runButtonRef}
        className="inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg border border-[#08a87d]/35 bg-white px-5 text-base font-black text-[#078665] shadow-sm transition hover:bg-[#f0fbf8] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:border-[#c8d3df] disabled:text-[#93a1b6]"
        disabled={isSubmitting}
        type="submit"
      >
        {isSubmitting ? <SpinnerIcon /> : <PlayIcon />}
        Run
      </button>
    </div>
  );
}

function ImageResultModal({
  isOpen,
  onClose,
  result,
  returnFocusRef,
}: {
  isOpen: boolean;
  onClose: () => void;
  result: OptimizationResult;
  returnFocusRef: RefObject<HTMLButtonElement | null>;
}) {
  const dialogRef = useRef<HTMLElement>(null);
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  const titleId = useId();
  const descriptionId = useId();
  const reduction =
    ((result.originalSize - result.size) / result.originalSize) * 100;
  const hasReduction = result.size < result.originalSize;

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const previousBodyOverflow = document.body.style.overflow;
    const previousDocumentOverflow = document.documentElement.style.overflow;
    document.body.style.overflow = "hidden";
    document.documentElement.style.overflow = "hidden";

    const focusTimer = window.setTimeout(() => {
      closeButtonRef.current?.focus();
    }, 0);

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        event.stopPropagation();
        return;
      }

      if (event.key !== "Tab") {
        return;
      }

      const dialog = dialogRef.current;
      if (!dialog) {
        return;
      }

      const focusableElements = getFocusableElements(dialog);
      if (focusableElements.length === 0) {
        event.preventDefault();
        dialog.focus();
        return;
      }

      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];
      const activeElement = document.activeElement;

      if (!dialog.contains(activeElement)) {
        event.preventDefault();
        if (event.shiftKey) {
          lastElement.focus();
        } else {
          firstElement.focus();
        }
        return;
      }

      if (event.shiftKey) {
        if (activeElement === firstElement) {
          event.preventDefault();
          lastElement.focus();
        }
        return;
      }

      if (activeElement === lastElement) {
        event.preventDefault();
        firstElement.focus();
      }
    }

    document.addEventListener("keydown", handleKeyDown, true);

    return () => {
      window.clearTimeout(focusTimer);
      document.removeEventListener("keydown", handleKeyDown, true);
      document.body.style.overflow = previousBodyOverflow;
      document.documentElement.style.overflow = previousDocumentOverflow;
      window.requestAnimationFrame(() => {
        returnFocusRef.current?.focus();
      });
    };
  }, [isOpen, returnFocusRef]);

  if (!isOpen || typeof document === "undefined") {
    return null;
  }

  return createPortal(
    <div
      className="fixed inset-0 z-[1000] flex h-dvh w-screen items-center justify-center overflow-y-auto overscroll-contain bg-black/60 px-4 py-6 sm:px-6"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          event.preventDefault();
        }
      }}
    >
      <section
        ref={dialogRef}
        aria-describedby={descriptionId}
        aria-labelledby={titleId}
        aria-modal="true"
        className="flex max-h-[calc(100dvh-2rem)] w-full max-w-[720px] flex-col overflow-hidden rounded-[22px] bg-white text-left shadow-[0_34px_90px_rgba(0,0,0,0.35)] ring-1 ring-white/70"
        role="dialog"
        tabIndex={-1}
      >
        <header className="flex items-start gap-4 border-b border-[#dbe8f1] px-4 py-4 sm:px-6">
          <div className="min-w-0 flex-1">
            <h2
              className="text-lg font-black leading-7 text-[#066f57]"
              id={titleId}
            >
              Compression complete
            </h2>
            <p
              className="mt-1 truncate text-sm font-bold leading-6 text-[#344464]"
              id={descriptionId}
            >
              {result.name}
            </p>
            <p className="truncate text-sm leading-6 text-[#60708d]">
              {result.type}
            </p>
          </div>

          <button
            ref={closeButtonRef}
            aria-label="Close result"
            className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-[#d6e1ec] bg-white text-[#344464] shadow-sm transition hover:bg-[#f7fbff] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d]"
            onClick={onClose}
            type="button"
          >
            <CloseIcon />
          </button>
        </header>

        <div className="overflow-y-auto px-4 py-4 sm:px-6">
          <BrowserImagePreview
            alt="Optimized image preview"
            className="overflow-hidden rounded-xl border border-[#bddfe1] bg-white"
            fallbackClassName="h-60 sm:h-72"
            fileName={result.name}
            fileType={result.type}
            imageClassName="h-60 w-full object-contain sm:h-72"
            src={result.url}
          />

          <dl className="mt-4 grid gap-3 sm:grid-cols-3">
            <ResultMetric
              label="Original"
              value={formatBytes(result.originalSize)}
            />
            <ResultMetric label="Optimized" value={formatBytes(result.size)} />
            <ResultMetric
              label="Reduction"
              value={hasReduction ? `${reduction.toFixed(1)}%` : "No reduction"}
            />
          </dl>

          {!hasReduction ? (
            <p className="mt-4 rounded-xl border border-[#f6d49c] bg-[#fff9ec] px-4 py-3 text-sm font-bold text-[#875a07]">
              The optimized file is not smaller than the original.
            </p>
          ) : null}
        </div>

        <footer className="flex flex-col gap-3 border-t border-[#dbe8f1] bg-white px-4 py-4 sm:flex-row sm:justify-end sm:px-6">
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
            onClick={onClose}
            type="button"
          >
            <CloseIcon />
            Close
          </button>
        </footer>
      </section>
    </div>,
    document.body,
  );
}

function BrowserImagePreview({
  alt,
  className,
  fallbackClassName,
  fileName,
  fileType,
  imageClassName,
  src,
}: {
  alt: string;
  className: string;
  fallbackClassName: string;
  fileName: string;
  fileType: string;
  imageClassName: string;
  src: string;
}) {
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

function getFocusableElements(container: HTMLElement) {
  return Array.from(
    container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR),
  ).filter((element) => {
    const isVisible =
      element.offsetWidth > 0 ||
      element.offsetHeight > 0 ||
      element.getClientRects().length > 0;

    return !element.hasAttribute("disabled") && isVisible;
  });
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

function ImageIcon() {
  return (
    <svg aria-hidden="true" className="h-6 w-6" fill="none" viewBox="0 0 24 24">
      <path
        d="M4.5 6.2c0-.9.8-1.7 1.7-1.7h11.6c.9 0 1.7.8 1.7 1.7v11.6c0 .9-.8 1.7-1.7 1.7H6.2c-.9 0-1.7-.8-1.7-1.7V6.2Z"
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="2"
      />
      <path
        d="m5 16 4.2-4.2c.4-.4 1-.4 1.4 0l2.1 2.1 1.1-1.1c.4-.4 1-.4 1.4 0L19 16.6M8.5 8.5h.1"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
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

function PlayIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24">
      <path
        d="M8.5 5.8v12.4c0 .8.9 1.3 1.6.8l8.5-6.2c.6-.4.6-1.3 0-1.7L10.1 5c-.7-.5-1.6 0-1.6.8Z"
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="2.2"
      />
    </svg>
  );
}

function ChevronDownIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24">
      <path
        d="m7 10 5 5 5-5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.2"
      />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24">
      <path
        d="m7 7 10 10M17 7 7 17"
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
    <svg
      aria-hidden="true"
      className="h-5 w-5 animate-spin"
      fill="none"
      viewBox="0 0 24 24"
    >
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

  const encodedFilename = contentDisposition.match(
    /filename\*=UTF-8''([^;]+)/i,
  );
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
  const currentExtension = readExtension(filename);
  const extension = extensionMatchesContentType(currentExtension, contentType)
    ? currentExtension
    : extensionForContentType(contentType);
  const base = filename.replace(/\.[^/.]+$/, "") || "image";

  return `${base}_compressed${extension}`;
}

function isSupportedImageFile(file: File) {
  const type = file.type.toLowerCase();
  return (
    SUPPORTED_IMAGE_TYPES.has(type) ||
    SUPPORTED_IMAGE_EXTENSIONS.has(readExtension(file.name))
  );
}

function readExtension(filename: string) {
  const index = filename.lastIndexOf(".");
  return index >= 0 ? filename.slice(index).toLowerCase() : "";
}

function extensionMatchesContentType(extension: string, contentType: string) {
  switch (contentType) {
    case "image/jpeg":
      return extension === ".jpg" || extension === ".jpeg";
    case "image/png":
      return extension === ".png";
    case "image/webp":
      return extension === ".webp";
    case "image/avif":
      return extension === ".avif";
    case "image/heic":
    case "image/heif":
      return extension === ".heic" || extension === ".heif";
    case "image/gif":
      return extension === ".gif";
    case "image/bmp":
    case "image/x-ms-bmp":
      return extension === ".bmp";
    case "image/tiff":
      return extension === ".tif" || extension === ".tiff";
    default:
      return false;
  }
}

function extensionForContentType(contentType: string) {
  switch (contentType) {
    case "image/png":
      return ".png";
    case "image/webp":
      return ".webp";
    case "image/avif":
      return ".avif";
    case "image/heic":
      return ".heic";
    case "image/heif":
      return ".heif";
    case "image/gif":
      return ".gif";
    case "image/bmp":
    case "image/x-ms-bmp":
      return ".bmp";
    case "image/tiff":
      return ".tiff";
    default:
      return ".jpg";
  }
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
