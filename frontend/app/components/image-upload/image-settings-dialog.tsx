"use client";
import { type ReactNode, type RefObject, useEffect, useRef } from "react";
import { createPortal } from "react-dom";
type Props = { children: ReactNode; titleId: string; headingRef: RefObject<HTMLHeadingElement | null>; returnFocusRef: RefObject<HTMLButtonElement | null>; submissionRef: RefObject<boolean>; onClose: () => void; };
export function ImageSettingsDialog({children,titleId,headingRef,returnFocusRef,submissionRef,onClose}: Props) {
 const dialogRef = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const dialog = dialogRef.current;
    dialog?.showModal();
    headingRef.current?.focus();
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      dialog?.close();
      document.body.style.overflow = previousOverflow;

      window.requestAnimationFrame(() => {
        // A completed resize opens the result dialog, which owns its focus.
        if (!document.querySelector('[aria-modal="true"]'))
          returnFocusRef.current?.focus();
      });
    };
  }, [returnFocusRef, headingRef]);
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
{children}
</dialog>, document.body);
}
