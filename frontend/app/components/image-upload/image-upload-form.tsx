"use client";

import {
  type ChangeEvent,
  type DragEvent,
  type SyntheticEvent,
  useEffect,
  useRef,
  useState,
} from "react";
import { ImageDropZone } from "./image-drop-zone";
import {
  buildCompressedFilename,
  isSupportedImageFile,
  MAX_UPLOAD_BYTES,
  readDownloadName,
  readErrorMessage,
} from "./image-file";
import { SUPPORTED_FORMAT_COPY } from "./image-format";
import { ImageOperationControls } from "./image-operation-controls";
import { ImageResultModal } from "./image-result-modal";
import type { ImageActionId, OptimizationResult } from "./types";

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

  async function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
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
        <ImageDropZone
          inputRef={inputRef}
          isDragging={isDragging}
          isSubmitting={isSubmitting}
          onDragLeave={handleDragLeave}
          onDragOver={handleDragOver}
          onDrop={handleDrop}
          onFileChange={handleFileChange}
          onOpenFilePicker={openFilePicker}
          onResetWorkflow={resetWorkflow}
          operationControls={
            selectedFile ? (
              <ImageOperationControls
                isSubmitting={isSubmitting}
                onActionChange={setSelectedAction}
                runButtonRef={runButtonRef}
                selectedAction={selectedAction}
              />
            ) : null
          }
          previewUrl={previewUrl}
          selectedFile={selectedFile}
        />

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
