import { type RefObject, useId } from "react";
import { ChevronDownIcon, PlayIcon, SpinnerIcon } from "./icons";
import { IMAGE_ACTIONS, type ImageActionId } from "./types";

type ImageOperationControlsProps = {
  isSubmitting: boolean;
  onActionChange: (action: ImageActionId) => void;
  runButtonRef: RefObject<HTMLButtonElement | null>;
  selectedAction: ImageActionId;
};

export function ImageOperationControls({
  isSubmitting,
  onActionChange,
  runButtonRef,
  selectedAction,
}: ImageOperationControlsProps) {
  const actionSelectId = useId();

  return (
    <div className="grid w-full gap-3 border-t border-[#dbe8f1]/80 pt-3 sm:grid-cols-[minmax(0,1fr)_124px]">
      <label className="sr-only" htmlFor={actionSelectId}>
        Select an image action
      </label>
      <div className="relative min-w-0">
        <select
          className="h-12 w-full appearance-none rounded-lg border border-[#cfe0ec] bg-white px-4 pr-11 text-base font-black text-[#081236] shadow-sm transition focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:bg-[#eef4f8] disabled:text-[#93a1b6]"
          disabled={isSubmitting}
          id={actionSelectId}
          onChange={(event) =>
            onActionChange(event.currentTarget.value as ImageActionId)
          }
          value={selectedAction}
        >
          {IMAGE_ACTIONS.map((action) => (
            <option key={action.id} value={action.id}>
              {action.label}
            </option>
          ))}
        </select>
        <span className="pointer-events-none absolute inset-y-0 right-4 flex items-center text-[#60708d]">
          <ChevronDownIcon />
        </span>
      </div>

      <button
        ref={runButtonRef}
        className="inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg border border-[#08a87d]/35 bg-white px-5 text-base font-black text-[#078665] shadow-sm transition hover:bg-[#f0fbf8] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:border-[#c8d3df] disabled:text-[#93a1b6]"
        disabled={isSubmitting}
        type="submit"
      >
        {isSubmitting ? <SpinnerIcon /> : <PlayIcon />}
        Run
      </button>
    </div>
  );
}
