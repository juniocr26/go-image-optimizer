export const SUPPORTED_FORMAT_COPY =
  "JPG/JPEG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP and TIFF";

export const SUPPORTED_IMAGE_TYPES = new Set([
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

export const SUPPORTED_IMAGE_EXTENSIONS = new Set([
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

export const IMAGE_ACCEPT = [
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

export function extensionMatchesContentType(
  extension: string,
  contentType: string,
) {
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

export function extensionForContentType(contentType: string) {
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
