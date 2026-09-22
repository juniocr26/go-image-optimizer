export const IMAGE_ACTIONS = [{ id: "compress", label: "Compress" }] as const;

export type ImageActionId = (typeof IMAGE_ACTIONS)[number]["id"];

export type OptimizationResult = {
  name: string;
  originalSize: number;
  size: number;
  type: string;
  url: string;
};
