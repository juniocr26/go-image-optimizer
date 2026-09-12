"use client";

import { type ChangeEvent, type FormEvent, useEffect, useState } from "react";

type UploadResult = {
  name: string;
  size: number;
  type: string;
  url: string;
};

export function ImageUploadForm() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
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

    if (result?.url) {
      URL.revokeObjectURL(result.url);
    }

    setSelectedFile(file);
    setError(null);
    setResult(null);
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
      className="mt-10 border border-[#d6d3c8] bg-white p-5 shadow-sm sm:p-6"
      onSubmit={handleSubmit}
    >
      <div className="grid gap-5">
        <label className="grid gap-2">
          <span className="text-sm font-medium text-[#394150]">Image file</span>
          <input
            accept="image/*"
            className="block w-full border border-[#c8c6bb] bg-[#fafaf7] px-3 py-2 text-sm text-[#1f2933] file:mr-4 file:border-0 file:bg-[#1f2933] file:px-4 file:py-2 file:text-sm file:font-medium file:text-white"
            name="image"
            onChange={handleFileChange}
            type="file"
          />
        </label>

        <button
          className="w-full bg-[#0f766e] px-4 py-3 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:bg-[#9ca3af] sm:w-fit"
          disabled={isSubmitting}
          type="submit"
        >
          {isSubmitting ? "Processing..." : "Process image"}
        </button>

        {selectedFile ? (
          <p className="text-sm text-[#5f6c7b]">
            Selected: <span className="font-medium text-[#1f2933]">{selectedFile.name}</span>
          </p>
        ) : null}

        {error ? (
          <div className="border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm text-[#991b1b]">
            {error}
          </div>
        ) : null}

        {result ? (
          <div className="grid gap-4 border border-[#bfdbfe] bg-[#eff6ff] p-4">
            <div>
              <p className="text-sm font-medium text-[#1e3a8a]">Returned image</p>
              <p className="mt-1 text-sm text-[#394150]">
                {result.name} - {formatBytes(result.size)} - {result.type}
              </p>
            </div>

            {result.type.startsWith("image/") ? (
              <img
                alt="Returned image preview"
                className="max-h-72 w-full border border-[#bfdbfe] bg-white object-contain"
                src={result.url}
              />
            ) : null}

            <a
              className="w-full bg-[#1d4ed8] px-4 py-3 text-center text-sm font-semibold text-white sm:w-fit"
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
