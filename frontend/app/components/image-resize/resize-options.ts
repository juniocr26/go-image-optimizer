export type ImageInfo = {
  width: number;
  height: number;
  contentType: string;
  format: string;
  frameCount: number;
};

export type ResizeOptions = {
  mode: "pixels" | "percentage";
  width: string;
  height: string;
  axis: "width" | "height";
  reduction: number;
  keepAspectRatio: boolean;
};

export function targetDimensions(info: ImageInfo, options: ResizeOptions) {
  let width = Number(options.width);
  let height = Number(options.height);
  if (
    options.mode === "pixels" &&
    (![width, height].every(Number.isSafeInteger) ||
      width < 1 ||
      height < 1 ||
      width > 32_000_000 ||
      height > 32_000_000)
  ) {
    return null;
  }
  if (options.mode === "percentage" || options.keepAspectRatio) {
    let scale = (100 - options.reduction) / 100;
    if (options.mode === "pixels") {
      scale =
        options.axis === "height" ? height / info.height : width / info.width;
    }
    width = Math.max(1, Math.round(info.width * scale));
    height = Math.max(1, Math.round(info.height * scale));
  }
  return { width, height };
}

export function dimensionError(
  info: ImageInfo,
  target: ReturnType<typeof targetDimensions>,
) {
  if (!target) return "Enter a positive whole number for each dimension.";
  if (target.width * target.height > 32_000_000)
    return "The output must not exceed 32 million pixels.";
  if (target.width * target.height * info.frameCount > 64_000_000)
    return "This animation exceeds the output frame pixel limit.";
  return null;
}
