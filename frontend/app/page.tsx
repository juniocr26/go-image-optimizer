import { ImageUploadForm } from "./components/image-upload-form";

export default function Home() {
  return (
    <main className="min-h-screen bg-[#f7f7f2] text-[#1f2933]">
      <section className="mx-auto flex min-h-screen w-full max-w-5xl flex-col justify-center px-6 py-12">
        <div className="max-w-3xl">
          <p className="mb-4 text-sm font-semibold uppercase tracking-normal text-[#0f766e]">
            Upload pipeline
          </p>
          <h1 className="text-4xl font-semibold tracking-normal text-[#111827] sm:text-5xl">
            Go Image Optimizer
          </h1>
          <p className="mt-5 max-w-2xl text-lg leading-8 text-[#46515f]">
            Send an image to the Go backend and download the returned file. This
            step validates the upload flow only; the image is not compressed yet.
          </p>
        </div>

        <ImageUploadForm />
      </section>
    </main>
  );
}
