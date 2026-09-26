import { type Dispatch, type SetStateAction, useId } from "react";
import {
  dimensionError,
  targetDimensions,
  type ImageInfo,
  type ResizeOptions,
} from "./resize-options";

const fieldClass =
  "mt-2 h-12 w-full rounded-xl border border-[#cfe0ec] bg-white px-3 text-base font-bold text-[#081236] outline-none focus:border-[#08a87d] focus:ring-2 focus:ring-[#08a87d]/20 disabled:bg-[#f1f5f9]";

type Props = {
  info: ImageInfo;
  options: ResizeOptions;
  isSubmitting: boolean;
  onOptionsChange: Dispatch<SetStateAction<ResizeOptions>>;
};

export function ResizeSettings({
  info,
  options,
  isSubmitting,
  onOptionsChange: setOptions,
}: Props) {
  const titleId = useId();
  const target = targetDimensions(info, options);
  const validationError = dimensionError(info, target);
  const unchanged =
    target && target.width === info.width && target.height === info.height;
  const enlarging =
    target && (target.width > info.width || target.height > info.height);
  function changeDimension(axis: "width" | "height", value: string) {
    setOptions((current) => {
      const next = { ...current, [axis]: value, axis };
      const n = Number(value);
      if (info && current.keepAspectRatio && Number.isSafeInteger(n) && n > 0) {
        const other = axis === "width" ? "height" : "width";
        next[other] = String(
          Math.max(1, Math.round((n * info[other]) / info[axis])),
        );
      }
      return next;
    });
  }

  return (
    <fieldset disabled={isSubmitting} className="min-w-0">
      <legend className="sr-only">Resize options</legend>
      <div
        className="grid grid-cols-2 gap-1 rounded-xl bg-[#edf3f8] p-1"
        role="tablist"
        aria-label="Resize mode"
      >
        {(["pixels", "percentage"] as const).map((mode) => (
          <button
            key={mode}
            id={`${titleId}-${mode}`}
            aria-controls={`${titleId}-panel`}
            type="button"
            role="tab"
            aria-selected={options.mode === mode}
            tabIndex={options.mode === mode ? 0 : -1}
            onKeyDown={(event) => {
              if (
                ["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)
              ) {
                event.preventDefault();
                const next =
                  event.key === "Home"
                    ? "pixels"
                    : event.key === "End"
                      ? "percentage"
                      : mode === "pixels"
                        ? "percentage"
                        : "pixels";
                setOptions({ ...options, mode: next });
                document.getElementById(`${titleId}-${next}`)?.focus();
              }
            }}
            onClick={() => {
              setOptions({ ...options, mode });
            }}
            className={`rounded-lg px-3 py-3 text-sm font-black outline-offset-2 focus-visible:outline-[#08a87d] ${options.mode === mode ? "bg-white text-[#066f57] shadow-sm" : "text-[#60708d]"}`}
          >
            {mode === "pixels" ? "Pixels" : "Percentage"}
          </button>
        ))}
      </div>
      <div
        role="tabpanel"
        id={`${titleId}-panel`}
        aria-labelledby={`${titleId}-${options.mode}`}
        className="mt-5"
      >
        {options.mode === "pixels" ? (
          <>
            <p className="text-sm leading-6 text-[#60708d]">
              Set your output size in pixels.
            </p>
            <div className="mt-3 grid grid-cols-2 gap-3">
              <label className="text-sm font-bold">
                Width (px)
                <input
                  className={fieldClass}
                  inputMode="numeric"
                  type="number"
                  min="1"
                  max="32000000"
                  step="1"
                  value={options.width}
                  onChange={(e) => changeDimension("width", e.target.value)}
                  required
                />
              </label>
              <label className="text-sm font-bold">
                Height (px)
                <input
                  className={fieldClass}
                  inputMode="numeric"
                  type="number"
                  min="1"
                  max="32000000"
                  step="1"
                  value={options.height}
                  onChange={(e) => changeDimension("height", e.target.value)}
                  required
                />
              </label>
            </div>
            <label className="mt-5 flex items-center gap-3 text-sm font-bold">
              <input
                type="checkbox"
                className="h-4 w-4 accent-[#08a87d]"
                checked={options.keepAspectRatio}
                onChange={(e) => {
                  const checked = e.target.checked;
                  setOptions((current) => {
                    const next = {
                      ...current,
                      keepAspectRatio: checked,
                    };
                    if (checked) {
                      const other = next.axis === "width" ? "height" : "width";
                      const n = Number(next[next.axis]);
                      if (Number.isSafeInteger(n) && n > 0)
                        next[other] = String(
                          Math.max(
                            1,
                            Math.round((n * info[other]) / info[next.axis]),
                          ),
                        );
                    }
                    return next;
                  });
                }}
              />
              Keep aspect ratio
            </label>
            <label className="mt-4 flex items-center gap-3 text-sm font-bold">
              <input
                type="checkbox"
                className="h-4 w-4 accent-[#08a87d]"
                checked={options.withoutEnlargement}
                onChange={(e) =>
                  setOptions({
                    ...options,
                    withoutEnlargement: e.target.checked,
                  })
                }
              />
              Don’t enlarge
            </label>
            <p className="mt-2 text-xs leading-5 text-[#60708d]">
              {options.keepAspectRatio
                ? "The other dimension adjusts automatically."
                : "Changing the proportions will stretch the image."}{" "}
              {options.withoutEnlargement
                ? "Output is capped at the original dimensions."
                : ""}
            </p>
          </>
        ) : (
          <>
            <p className="mb-3 text-sm leading-6 text-[#60708d]">
              Reduce width and height by the same percentage.
            </p>
            <div className="grid gap-2">
              {[25, 50, 75].map((reduction) => (
                <label
                  key={reduction}
                  className={`flex cursor-pointer items-center justify-between rounded-xl border px-4 py-3 ${options.reduction === reduction ? "border-[#08a87d] bg-[#f0fbf8]" : "border-[#dbe8f1]"}`}
                >
                  <span>
                    <span className="text-sm font-bold">
                      {reduction}% smaller
                    </span>
                    <span className="mt-1 block text-xs text-[#60708d]">
                      {Math.max(
                        1,
                        Math.round((info.width * (100 - reduction)) / 100),
                      )}{" "}
                      ×{" "}
                      {Math.max(
                        1,
                        Math.round((info.height * (100 - reduction)) / 100),
                      )}{" "}
                      px
                    </span>
                  </span>
                  <input
                    type="radio"
                    name="resize-reduction"
                    value={reduction}
                    checked={options.reduction === reduction}
                    onChange={() => {
                      setOptions({ ...options, reduction });
                    }}
                    className="h-4 w-4 accent-[#08a87d]"
                  />
                </label>
              ))}
            </div>
          </>
        )}
      </div>
      {validationError ? (
        <p role="alert" className="mt-3 text-sm font-bold text-[#991b1b]">
          {validationError}
        </p>
      ) : null}
      {unchanged && !validationError ? (
        <p className="mt-3 text-xs leading-5 text-[#60708d]">
          The dimensions are unchanged. The original file will be returned.
        </p>
      ) : null}
      {enlarging && !validationError ? (
        <p className="mt-3 text-xs leading-5 text-[#875a07]">
          Enlarging increases dimensions, but does not add detail.
        </p>
      ) : null}
    </fieldset>
  );
}
