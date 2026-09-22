import { type RefObject, useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";
import { BrowserImagePreview } from "./browser-image-preview";
import { formatBytes } from "./image-file";
import { CloseIcon, DownloadIcon } from "./icons";
import { ResultMetric } from "./result-metric";
import type { OptimizationResult } from "./types";

const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

type ImageResultModalProps = {
  isOpen: boolean;
  onClose: () => void;
  result: OptimizationResult;
  returnFocusRef: RefObject<HTMLButtonElement | null>;
};

export function ImageResultModal({
  isOpen,
  onClose,
  result,
  returnFocusRef,
}: ImageResultModalProps) {
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
