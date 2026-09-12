"use client";

import {
  type ChangeEvent,
  type DragEvent,
  type FormEvent,
  useEffect,
  useRef,
  useState,
} from "react";

type UploadResult = {
  name: string;
  size: number;
  type: string;
  url: string;
};

export function ImageUploadForm() {
  const inputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<UploadResult | null>(null);

  useEffect(() => {
    return () => {
      if (result?.url) {
        URL.revokeObjectURL(result.url);
      }
    };
  }, [result]);

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;
    chooseFile(file);
  }

  function chooseFile(file: File | null) {
    if (result?.url) {
      URL.revokeObjectURL(result.url);
    }

    if (file && !file.type.startsWith("image/")) {
      setSelectedFile(null);
      setError("Choose an image file before submitting.");
      setResult(null);
      return;
    }

    setSelectedFile(file);
    setError(null);
    setResult(null);
  }

  function openFilePicker() {
    if (inputRef.current) {
      inputRef.current.value = "";
      inputRef.current.click();
    }
  }

  function handleDragOver(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
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
    chooseFile(event.dataTransfer.files?.[0] ?? null);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!selectedFile) {
      setError("Choose an image before submitting.");
      return;
    }

    const formData = new FormData();
    formData.append("image", selectedFile);

    if (result?.url) {
      URL.revokeObjectURL(result.url);
    }

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

      setResult({
        name: selectedFile.name || "returned-image",
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

  return (
    <form
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
            accept="image/*"
            aria-label="Choose an image file"
            className="sr-only"
            name="image"
            onChange={handleFileChange}
            type="file"
          />

          <div className="grid justify-items-center">
            <div className="grid h-16 w-16 place-items-center rounded-full bg-[#d9f4ec] text-[#08a87d] shadow-[0_16px_36px_rgba(0,168,125,0.16)]">
              <UploadIcon />
            </div>

            <p className="mt-5 text-lg font-black text-[#081236]">
              Drag and drop an image here
            </p>
            <p className="mt-2 text-base text-[#344464]">
              or click to select a file
            </p>
            <p className="mt-5 text-sm leading-6 text-[#60708d] sm:text-base">
              Supports common formats (JPG, PNG, WebP, etc.)
            </p>

            <div className="mt-5 flex w-full flex-col items-center justify-center gap-3 sm:flex-row">
              <button
                className="inline-flex h-14 w-full items-center justify-center gap-3 rounded-lg bg-[#05a97f] px-8 text-base font-black text-white shadow-[0_16px_34px_rgba(0,168,125,0.25)] transition hover:bg-[#02976f] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] sm:w-[220px]"
                onClick={openFilePicker}
                type="button"
              >
                <FolderIcon />
                Choose file
              </button>

              {selectedFile ? (
                <button
                  className="inline-flex h-14 w-full items-center justify-center rounded-lg border border-[#08a87d]/35 bg-white px-7 text-base font-black text-[#078665] shadow-sm transition hover:bg-[#f0fbf8] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:border-[#c8d3df] disabled:text-[#93a1b6] sm:w-auto"
                  disabled={isSubmitting}
                  type="submit"
                >
                  {isSubmitting ? "Sending..." : "Send to backend"}
                </button>
              ) : null}
            </div>

            {selectedFile ? (
              <p className="mt-4 max-w-full truncate text-sm text-[#60708d]">
                Selected:{" "}
                <span className="font-bold text-[#081236]">{selectedFile.name}</span>
              </p>
            ) : null}
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
                Returned by backend
              </p>
              <p className="mt-1 text-sm leading-6 text-[#344464]">
                {result.name} - {formatBytes(result.size)} - {result.type}
              </p>
            </div>

            {result.type.startsWith("image/") ? (
              <img
                alt="Returned image preview"
                className="max-h-72 w-full rounded-xl border border-[#bddfe1] bg-white object-contain"
                src={result.url}
              />
            ) : null}

            <a
              className="inline-flex h-12 w-full items-center justify-center rounded-lg bg-[#081236] px-5 text-center text-sm font-black text-white transition hover:bg-[#14234a] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#081236] sm:w-fit"
              download={result.name}
              href={result.url}
            >
              Download returned image
            </a>
          </div>
        ) : null}
      </div>
    </form>
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

async function readErrorMessage(response: Response) {
  const contentType = response.headers.get("Content-Type") ?? "";

  if (contentType.includes("application/json")) {
    const data = (await response.json()) as { error?: string };
    return data.error ?? "The image could not be processed.";
  }

  return (await response.text()) || "The image could not be processed.";
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
