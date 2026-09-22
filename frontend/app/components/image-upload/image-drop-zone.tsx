import type {
  ChangeEvent,
  DragEvent,
  ReactNode,
  RefObject,
} from "react";
import { BrowserImagePreview } from "./browser-image-preview";
import { formatBytes } from "./image-file";
import { IMAGE_ACCEPT, SUPPORTED_FORMAT_COPY } from "./image-format";
import { FolderIcon, ResetIcon, UploadIcon } from "./icons";

type ImageDropZoneProps = {
  inputRef: RefObject<HTMLInputElement | null>;
  isDragging: boolean;
  isSubmitting: boolean;
  onDragLeave: (event: DragEvent<HTMLDivElement>) => void;
  onDragOver: (event: DragEvent<HTMLDivElement>) => void;
  onDrop: (event: DragEvent<HTMLDivElement>) => void;
  onFileChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onOpenFilePicker: () => void;
  onResetWorkflow: () => void;
  operationControls: ReactNode;
  previewUrl: string | null;
  selectedFile: File | null;
};

export function ImageDropZone({
  inputRef,
  isDragging,
  isSubmitting,
  onDragLeave,
  onDragOver,
  onDrop,
  onFileChange,
  onOpenFilePicker,
  onResetWorkflow,
  operationControls,
  previewUrl,
  selectedFile,
}: ImageDropZoneProps) {
  return (
    <div
      className={`grid min-h-[248px] place-items-center rounded-[18px] border border-dashed px-4 py-7 text-center transition sm:min-h-[265px] sm:px-8 ${
        isDragging
          ? "border-[#08a87d] bg-[#ecfbf7]"
          : "border-[#9eb6d8] bg-white/54"
      }`}
      onDragLeave={onDragLeave}
      onDragOver={onDragOver}
      onDrop={onDrop}
    >
      <input
        ref={inputRef}
        accept={IMAGE_ACCEPT}
        aria-label="Choose a supported image"
        className="sr-only"
        disabled={isSubmitting}
        name="image"
        onChange={onFileChange}
        type="file"
      />

      <div className="grid w-full justify-items-center">
        {previewUrl && selectedFile ? (
          <BrowserImagePreview
            alt={`Preview of ${selectedFile.name}`}
            className="w-full max-w-[420px] overflow-hidden rounded-2xl border border-[#dbe8f1] bg-white shadow-[0_16px_42px_rgba(15,42,80,0.12)]"
            fallbackClassName="h-48 sm:h-56"
            fileName={selectedFile.name}
            fileType={selectedFile.type}
            imageClassName="h-48 w-full object-contain sm:h-56"
            src={previewUrl}
          />
        ) : (
          <div className="grid h-16 w-16 place-items-center rounded-full bg-[#d9f4ec] text-[#08a87d] shadow-[0_16px_36px_rgba(0,168,125,0.16)]">
            <UploadIcon />
          </div>
        )}

        <p className="mt-5 text-lg font-black text-[#081236]">
          {selectedFile
            ? "Image ready to process"
            : "Drag and drop an image here"}
        </p>
        <p className="mt-2 text-base text-[#344464]">
          {selectedFile ? "Review the file and run an action" : "or click to select a file"}
        </p>
        <p className="mt-5 text-sm leading-6 text-[#60708d] sm:text-base">
          {SUPPORTED_FORMAT_COPY}, up to 50 MiB.
        </p>

        {selectedFile ? (
          <dl className="mt-5 grid w-full max-w-[520px] gap-3 rounded-2xl border border-[#dbe8f1] bg-white/80 p-4 text-left sm:grid-cols-[1.4fr_0.8fr]">
            <div className="min-w-0">
              <dt className="text-xs font-black uppercase tracking-normal text-[#6f7f9d]">
                File
              </dt>
              <dd className="mt-1 truncate text-sm font-bold text-[#081236]">
                {selectedFile.name}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-black uppercase tracking-normal text-[#6f7f9d]">
                Original
              </dt>
              <dd className="mt-1 text-sm font-bold text-[#081236]">
                {formatBytes(selectedFile.size)}
              </dd>
            </div>
          </dl>
        ) : null}

        {isSubmitting ? (
          <div className="mt-5 w-full max-w-[420px]" role="status">
            <div className="h-2 overflow-hidden rounded-full bg-[#dce7ee]">
              <div className="indeterminate-progress h-full w-1/2 rounded-full bg-[#08a87d]" />
            </div>
            <p className="mt-3 text-sm font-bold text-[#078665]">
              Compressing image...
            </p>
          </div>
        ) : null}

        <div className="mt-5 grid w-full max-w-[520px] gap-3">
          <div
            className={`grid w-full gap-3 ${
              selectedFile ? "sm:grid-cols-2" : "justify-items-center"
            }`}
          >
            <button
              className={`inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg bg-[#05a97f] px-7 text-base font-black text-white shadow-[0_16px_34px_rgba(0,168,125,0.25)] transition hover:bg-[#02976f] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#08a87d] disabled:cursor-not-allowed disabled:bg-[#8ccfbd] ${
                selectedFile ? "" : "sm:w-[220px]"
              }`}
              disabled={isSubmitting}
              onClick={onOpenFilePicker}
              type="button"
            >
              <FolderIcon />
              Choose file
            </button>

            {selectedFile ? (
              <button
                className="inline-flex h-12 w-full items-center justify-center gap-3 rounded-lg border border-[#d6e1ec] bg-white px-7 text-base font-black text-[#344464] shadow-sm transition hover:bg-[#f7fbff] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#9eb6d8] disabled:cursor-not-allowed disabled:text-[#93a1b6]"
                disabled={isSubmitting}
                onClick={onResetWorkflow}
                type="button"
              >
                <ResetIcon />
                Another image
              </button>
            ) : null}
          </div>

          {operationControls}
        </div>
      </div>
    </div>
  );
}
