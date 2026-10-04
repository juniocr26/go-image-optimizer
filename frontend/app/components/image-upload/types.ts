export const IMAGE_ACTIONS = [
  { id: "compress", label: "Compress" },
  { id: "resize", label: "Resize" },
  { id: "convert", label: "Convert format" },
] as const;

export type ImageActionId = (typeof IMAGE_ACTIONS)[number]["id"];

export type OptimizationResult = {
  blob: Blob;
  sourceFormat?: string;
  outputFormat?: string;
  operation?: ImageActionId;
  dimensions?: {
    originalWidth: number;
    originalHeight: number;
    width: number;
    height: number;
  };
  name: string;
  originalSize: number;
  size: number;
  type: string;
  url: string;
};
