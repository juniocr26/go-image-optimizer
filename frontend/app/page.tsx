export const dynamic = "force-dynamic";

type BackendHealth = {
  ok: boolean;
  status: string;
};

async function getBackendHealth(): Promise<BackendHealth> {
  const backendUrl = process.env.BACKEND_URL ?? "http://localhost:8080";

  try {
    const response = await fetch(`${backendUrl}/health`, {
      cache: "no-store",
    });

    if (!response.ok) {
      return {
        ok: false,
        status: `Backend returned ${response.status}`,
      };
    }

    const data = (await response.json()) as { status?: string };

    return {
      ok: data.status === "ok",
      status: data.status ?? "unknown",
    };
  } catch {
    return {
      ok: false,
      status: "unreachable",
    };
  }
}

export default async function Home() {
  const backendHealth = await getBackendHealth();

  return (
    <main className="min-h-screen bg-[#f7f7f2] text-[#1f2933]">
      <section className="mx-auto flex min-h-screen w-full max-w-5xl flex-col justify-center px-6 py-12">
        <div className="max-w-3xl">
          <p className="mb-4 text-sm font-semibold uppercase tracking-normal text-[#0f766e]">
            Initial infrastructure
          </p>
          <h1 className="text-4xl font-semibold tracking-normal text-[#111827] sm:text-5xl">
            Go Image Optimizer is running.
          </h1>
          <p className="mt-5 max-w-2xl text-lg leading-8 text-[#46515f]">
            The frontend is online, Tailwind is loaded, and the backend health
            endpoint is ready for the first image-compression feature.
          </p>
        </div>

        <div className="mt-10 grid gap-4 sm:grid-cols-2">
          <div className="border border-[#d6d3c8] bg-white p-5 shadow-sm">
            <div className="text-sm font-medium text-[#5f6c7b]">Frontend</div>
            <div className="mt-3 flex items-center gap-3">
              <span className="h-3 w-3 bg-[#2563eb]" aria-hidden="true" />
              <span className="text-xl font-semibold text-[#111827]">ready</span>
            </div>
          </div>

          <div className="border border-[#d6d3c8] bg-white p-5 shadow-sm">
            <div className="text-sm font-medium text-[#5f6c7b]">Backend</div>
            <div className="mt-3 flex items-center gap-3">
              <span
                className={
                  backendHealth.ok ? "h-3 w-3 bg-[#16a34a]" : "h-3 w-3 bg-[#dc2626]"
                }
                aria-hidden="true"
              />
              <span className="text-xl font-semibold text-[#111827]">
                {backendHealth.status}
              </span>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
