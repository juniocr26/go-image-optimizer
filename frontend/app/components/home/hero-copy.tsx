import { ImageUploadForm } from "../image-upload/image-upload-form";

export function HeroCopy() {
  return (
    <div className="mx-auto w-full max-w-[620px] lg:mx-0 lg:pl-8 xl:pl-12">
      <p className="mb-7 text-xs font-bold uppercase tracking-normal text-[#00a97b] sm:text-sm">
        MULTI-FORMAT &bull; SIMPLE &bull; OPEN SOURCE
      </p>

      <h1 className="max-w-[650px] text-5xl font-black leading-[0.98] tracking-normal text-[#081236] sm:text-6xl lg:text-[70px]">
        Optimize your images{" "}
        <span className="bg-gradient-to-r from-[#08a67f] to-[#16c39c] bg-clip-text text-transparent">
          with Go
        </span>
      </h1>

      <p className="mt-6 max-w-[520px] text-lg leading-8 tracking-normal text-[#60708d] sm:text-xl sm:leading-8">
        Upload a supported image, compress it in the Go backend and download the
        optimized result without storing the image on the server.
      </p>

      <ImageUploadForm />
    </div>
  );
}
