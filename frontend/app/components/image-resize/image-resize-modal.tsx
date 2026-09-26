"use client";

import {
  type RefObject,
  type SyntheticEvent,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import { ResizeSettings } from "./resize-settings";
import { BrowserImagePreview } from "../image-upload/browser-image-preview";
import {
  formatBytes,
  readDownloadName,
  readErrorMessage,
} from "../image-upload/image-file";
import { extensionForContentType } from "../image-upload/image-format";
import { CloseIcon, SpinnerIcon } from "../image-upload/icons";
import type { OptimizationResult } from "../image-upload/types";
import {
  dimensionError,
  targetDimensions,
  type ImageInfo,
  type ResizeOptions,
} from "./resize-options";

type Props = {
  file: File;
  previewUrl: string;
  returnFocusRef: RefObject<HTMLButtonElement | null>;
  onClose: () => void;
  onComplete: (result: OptimizationResult) => void;
};

const buttonClass =
  "inline-flex h-12 items-center justify-center gap-2 rounded-xl px-5 text-sm font-black transition focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:opacity-50";

export function ImageResizeModal({
  file,
  previewUrl,
  returnFocusRef,
  onClose,
  onComplete,
}: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const headingRef = useRef<HTMLHeadingElement>(null);
  const titleId = useId();
  const [info, setInfo] = useState<ImageInfo | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [processError, setProcessError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const submissionRef = useRef(false);
  const requestRef = useRef<AbortController | null>(null);
  const [options, setOptions] = useState<ResizeOptions>({
    mode: "pixels",
    width: "",
    height: "",
    axis: "width",
    reduction: 50,
    keepAspectRatio: true,
  });

  useEffect(() => {
    const dialog = dialogRef.current;
    dialog?.showModal();
    headingRef.current?.focus();
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      dialog?.close();
      document.body.style.overflow = previousOverflow;
      requestRef.current?.abort();
      window.requestAnimationFrame(() => {
        // A completed resize opens the result dialog, which owns its focus.
        if (!document.querySelector('[aria-modal="true"]'))
          returnFocusRef.current?.focus();
      });
    };
  }, [returnFocusRef]);

  useEffect(() => {
    const controller = new AbortController();
    setLoadError(null);
    const body = new FormData();
    body.append("image", file);
    async function inspect() {
      try {
        const response = await fetch("/api/images/resize/info", {
          method: "POST",
          body,
          signal: controller.signal,
        });
        if (!response.ok) throw new Error(await readErrorMessage(response));
        const metadata: ImageInfo = await response.json();
        if (
          ![metadata.width, metadata.height, metadata.frameCount].every(
            (n) => Number.isSafeInteger(n) && n > 0,
          )
        )
          throw new Error("Could not read image dimensions.");
        if (controller.signal.aborted) return;
        setInfo(metadata);
        setOptions((current) => ({
          ...current,
          width: String(metadata.width),
          height: String(metadata.height),
        }));
      } catch (error) {
        if (!controller.signal.aborted)
          setLoadError(
            error instanceof Error
              ? error.message
              : "Could not read image dimensions.",
          );
      }
    }
    void inspect();
    return () => controller.abort();
  }, [file, attempt]);

  const target = info ? targetDimensions(info, options) : null;
  const validationError = info ? dimensionError(info, target) : null;
  async function submit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault();
    event.stopPropagation();
    if (!info || !target || validationError || submissionRef.current) return;
    submissionRef.current = true;
    setIsSubmitting(true);
    setProcessError(null);
    const controller = new AbortController();
    requestRef.current = controller;
    const body = new FormData();
    body.append("image", file);
    for (const [key, value] of Object.entries(options)) {
      // Number inputs can contain exponent notation (e.g. 1e2). Send the
      // validated integer value in the decimal syntax expected by Go.
      body.append(
        key,
        key === "width" || key === "height" ? String(Number(value)) : String(value),
      );
    }
    try {
      const response = await fetch("/api/images/resize", {
        method: "POST",
        body,
        signal: controller.signal,
      });
      if (!response.ok) throw new Error(await readErrorMessage(response));
      const dimensions = {
        originalWidth: Number(response.headers.get("X-Original-Width")),
        originalHeight: Number(response.headers.get("X-Original-Height")),
        width: Number(response.headers.get("X-Image-Width")),
        height: Number(response.headers.get("X-Image-Height")),
      };
      if (
        !Object.values(dimensions).every(
          (n) => Number.isSafeInteger(n) && n > 0,
        )
      )
        throw new Error(
          "The server returned an incomplete resize result. Please try again.",
        );
      const blob = await response.blob();
      const name =
        readDownloadName(response.headers.get("Content-Disposition")) ??
        `${file.name.replace(/\.[^/.]+$/, "")}_resized${extensionForContentType(blob.type)}`;
      if (controller.signal.aborted) return;
      onComplete({
        operation: "resize",
        name,
        originalSize: file.size,
        size: blob.size,
        type: blob.type,
        url: URL.createObjectURL(blob),
        dimensions,
      });
    } catch (error) {
      if (!controller.signal.aborted)
        setProcessError(
          error instanceof Error
            ? error.message
            : "Could not resize this image. Please try again.",
        );
    } finally {
      submissionRef.current = false;
      setIsSubmitting(false);
    }
  }

  return createPortal(
    <dialog
      ref={dialogRef}
      aria-labelledby={titleId}
      aria-modal="true"
      onKeyDown={(event) => {
        if (event.key !== "Tab") return;
        const focusable = Array.from(
          event.currentTarget.querySelectorAll<HTMLElement>(
            "a[href], button, input, select, textarea, [tabindex]",
          ),
        ).filter(
          (element) =>
            element.tabIndex >= 0 &&
            !element.matches(":disabled") &&
            element.getClientRects().length > 0,
        );
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        if (!first) {
          event.preventDefault();
          headingRef.current?.focus();
        } else if (
          event.shiftKey &&
          (document.activeElement === first ||
            !focusable.includes(document.activeElement as HTMLElement))
        ) {
          event.preventDefault();
          last.focus();
        } else if (!event.shiftKey && document.activeElement === last) {
          event.preventDefault();
          first.focus();
        }
      }}
      onCancel={(event) => {
        event.preventDefault();
        if (!submissionRef.current) onClose();
      }}
      className="fixed inset-0 m-auto max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-[860px] overflow-hidden rounded-[22px] border-0 bg-white p-0 text-left text-[#081236] shadow-[0_34px_90px_rgba(0,0,0,0.35)] backdrop:bg-black/60"
    >
      <form
        onSubmit={submit}
        aria-busy={isSubmitting}
        className="flex max-h-[calc(100dvh-2rem)] flex-col"
      >
        <header className="flex shrink-0 items-start gap-4 border-b border-[#dbe8f1] px-5 py-4 sm:px-6">
          <div className="min-w-0 flex-1">
            <h2
              ref={headingRef}
              tabIndex={-1}
              id={titleId}
              className="text-xl font-black text-[#066f57] outline-none"
            >
              Resize image
            </h2>
            <p className="mt-1 text-sm leading-6 text-[#60708d]">
              Choose the dimensions that fit your needs.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={isSubmitting}
            aria-label="Close resize options"
            className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-[#d6e1ec] text-[#344464] transition hover:bg-[#f7fbff] focus-visible:outline focus-visible:outline-3 focus-visible:outline-[#08a87d] disabled:opacity-50"
          >
            <CloseIcon />
          </button>
        </header>
        <div className="grid min-h-0 gap-6 overflow-y-auto p-5 sm:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)] sm:p-6">
          <div className="min-w-0">
            <div className="rounded-2xl border border-[#dbe8f1] bg-[#f5f9fc] p-3">
              <BrowserImagePreview
                alt="Original image"
                className="overflow-hidden rounded-xl bg-white"
                imageClassName="h-36 w-full object-contain sm:h-60"
                fallbackClassName="h-36 sm:h-60"
                fileName={file.name}
                fileType={file.type}
                src={previewUrl}
              />
              <p className="mt-3 truncate text-sm font-bold" title={file.name}>
                {file.name}
              </p>
              <p className="mt-1 text-xs text-[#60708d]">
                {formatBytes(file.size)}
                {info ? ` · ${info.width} × ${info.height} px` : ""}
              </p>
            </div>
            <div
              aria-live="polite"
              className="mt-3 rounded-xl border border-[#bddfe1] bg-[#f0fbf8] px-4 py-3"
            >
              <p className="text-xs font-bold uppercase text-[#066f57]">
                Output dimensions
              </p>
              <p className="mt-1 text-xl font-black">
                {target && !validationError
                  ? `${target.width} × ${target.height} px`
                  : "—"}
              </p>
              {info && info.frameCount > 1 ? (
                <p className="mt-1 text-xs text-[#60708d]">
                  {info.frameCount} animation frames
                </p>
              ) : null}
            </div>
          </div>
          <div className="min-w-0">
            {!info ? (
              <div
                role={loadError ? "alert" : "status"}
                className="flex min-h-40 flex-col items-center justify-center gap-3 text-center text-sm text-[#60708d]"
              >
                {loadError ? (
                  <>
                    <p>{loadError}</p>
                    <button
                      type="button"
                      className={`${buttonClass} border border-[#bddfe1] text-[#066f57]`}
                      onClick={() => setAttempt((n) => n + 1)}
                    >
                      Try again
                    </button>
                  </>
                ) : (
                  <>
                    <SpinnerIcon />
                    <p>Reading image dimensions…</p>
                  </>
                )}
              </div>
            ) : (
              <ResizeSettings
                info={info}
                options={options}
                isSubmitting={isSubmitting}
                onOptionsChange={(next) => {
                  setOptions(next);
                  setProcessError(null);
                }}
              />
            )}
            {processError ? (
              <p
                role="alert"
                className="mt-4 rounded-xl bg-[#fff5f5] px-3 py-3 text-sm text-[#991b1b]"
              >
                {processError}
              </p>
            ) : null}
          </div>
        </div>
        <footer className="flex shrink-0 flex-col-reverse gap-3 border-t border-[#dbe8f1] bg-white px-5 py-4 sm:flex-row sm:items-center sm:justify-end sm:px-6">
          {isSubmitting ? (
            <span role="status" className="text-sm text-[#60708d]">
              Resizing image…
            </span>
          ) : null}
          <button
            type="button"
            disabled={isSubmitting}
            onClick={onClose}
            className={`${buttonClass} border border-[#bddfe1] text-[#066f57] hover:bg-[#f8fffd]`}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={!info || !!validationError || isSubmitting}
            className={`${buttonClass} bg-[#081236] text-white hover:bg-[#14234a]`}
          >
            {isSubmitting ? <SpinnerIcon /> : null}Resize image
          </button>
        </footer>
      </form>
    </dialog>,
    document.body,
  );
}
