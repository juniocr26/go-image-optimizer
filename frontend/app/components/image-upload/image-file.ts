import {
  extensionForContentType,
  extensionMatchesContentType,
  SUPPORTED_IMAGE_EXTENSIONS,
  SUPPORTED_IMAGE_TYPES,
} from "./image-format";

export const MAX_UPLOAD_BYTES = 50 * 1024 * 1024;

export async function readErrorMessage(response: Response) {
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

export function readDownloadName(contentDisposition: string | null) {
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

export function buildCompressedFilename(filename: string, contentType: string) {
  const currentExtension = readExtension(filename);
  const extension = extensionMatchesContentType(currentExtension, contentType)
    ? currentExtension
    : extensionForContentType(contentType);
  const base = filename.replace(/\.[^/.]+$/, "") || "image";

  return `${base}_compressed${extension}`;
}

export function isSupportedImageFile(file: File) {
  const type = file.type.toLowerCase();
  return (
    SUPPORTED_IMAGE_TYPES.has(type) ||
    SUPPORTED_IMAGE_EXTENSIONS.has(readExtension(file.name))
  );
}

export function readExtension(filename: string) {
  const index = filename.lastIndexOf(".");
  return index >= 0 ? filename.slice(index).toLowerCase() : "";
}

export function formatBytes(bytes: number) {
  if (bytes < 1024) {
    return `${bytes} B`;
  }

  const kilobytes = bytes / 1024;
  if (kilobytes < 1024) {
    return `${kilobytes.toFixed(1)} KB`;
  }

  return `${(kilobytes / 1024).toFixed(1)} MB`;
}
