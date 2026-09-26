export const IMAGE_ACTIONS = [
  { id: "compress", label: "Compress" },
  { id: "resize", label: "Resize" },
] as const;

export type ImageActionId = (typeof IMAGE_ACTIONS)[number]["id"];

export type OptimizationResult = {
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
