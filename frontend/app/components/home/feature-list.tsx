export function FeatureList() {
  const features = [
    {
      title: "Focused",
      description: "Multi-format compression and resizing in Go.",
      icon: <LightningIcon />,
    },
    {
      title: "Ephemeral",
      description: "Images are returned directly to the browser.",
      icon: <ShieldIcon />,
    },
    {
      title: "Transparent",
      description: "Results show the actual measured file sizes.",
      icon: <ChartIcon />,
    },
  ];

  return (
    <div className="relative z-10 mx-auto mt-8 grid w-full max-w-[900px] gap-6 pb-8 sm:grid-cols-3 sm:gap-8 lg:mt-3">
      {features.map((feature) => (
        <article key={feature.title} className="flex items-center gap-4">
          <div className="grid h-14 w-14 shrink-0 place-items-center rounded-full bg-[#d9f4ec] text-[#08a87d] shadow-[0_12px_32px_rgba(0,168,125,0.12)]">
            {feature.icon}
          </div>
          <div>
            <h2 className="text-base font-black text-[#081236]">
              {feature.title}
            </h2>
            <p className="mt-1 text-sm leading-6 text-[#60708d]">
              {feature.description}
            </p>
          </div>
        </article>
      ))}
    </div>
  );
}

function LightningIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-7 w-7"
      fill="currentColor"
      viewBox="0 0 24 24"
    >
      <path d="M13.2 2.6 4.8 13.1a1 1 0 0 0 .8 1.6h5.1l-1.1 6.7a1 1 0 0 0 1.8.7l7.9-11.3a1 1 0 0 0-.8-1.6h-4.7l1.2-5.8a1 1 0 0 0-1.8-.8Z" />
    </svg>
  );
}

function ShieldIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-7 w-7"
      fill="none"
      viewBox="0 0 24 24"
    >
      <path
        d="M12 3.2 18.8 6v5.1c0 4.4-2.8 8-6.8 9.7-4-1.7-6.8-5.3-6.8-9.7V6L12 3.2Z"
        fill="currentColor"
        opacity="0.18"
      />
      <path
        d="M12 3.2 18.8 6v5.1c0 4.4-2.8 8-6.8 9.7-4-1.7-6.8-5.3-6.8-9.7V6L12 3.2Z"
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="2"
      />
    </svg>
  );
}

function ChartIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-7 w-7"
      fill="none"
      viewBox="0 0 24 24"
    >
      <path
        d="M6 19V9M12 19V5M18 19v-7"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="3"
      />
    </svg>
  );
}
