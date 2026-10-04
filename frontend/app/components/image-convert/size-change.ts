export function conversionSizeChange(originalBytes: number, outputBytes: number) {
  if (originalBytes === outputBytes) return { label: "Size change", value: "No size change" };
  const percent = Math.abs(((originalBytes - outputBytes) / originalBytes) * 100);
  return { label: outputBytes < originalBytes ? "Reduction" : "Increase", value: `${percent < 0.1 ? "<0.1" : percent.toFixed(1)}%` };
}
