import Image from "next/image";
import { ImageUploadForm } from "./components/image-upload-form";

export default function Home() {
  return (
    <main className="relative min-h-screen overflow-x-hidden bg-[#f6fbff] text-[#081236]">
      <DecorativeBackground />

      <section className="relative z-10 mx-auto flex min-h-screen w-full max-w-[1392px] flex-col px-5 py-8 sm:px-8 lg:px-12 lg:pb-0">
        <div className="grid flex-1 items-center gap-10 pt-8 md:pt-12 lg:grid-cols-[0.92fr_1.08fr] lg:gap-5 lg:pt-6">
          <HeroCopy />
          <ImageComparisonPreview />
        </div>

        <FeatureList />
        <Footer />
      </section>
    </main>
  );
}

function HeroCopy() {
  return (
    <div className="mx-auto w-full max-w-[620px] lg:mx-0 lg:pl-8 xl:pl-12">
      <p className="mb-7 text-xs font-bold uppercase tracking-normal text-[#00a97b] sm:text-sm">
        JPG & PNG &bull; SIMPLE &bull; OPEN SOURCE
      </p>

      <h1 className="max-w-[650px] text-5xl font-black leading-[0.98] tracking-normal text-[#081236] sm:text-6xl lg:text-[70px]">
        Optimize your images{" "}
        <span className="bg-gradient-to-r from-[#08a67f] to-[#16c39c] bg-clip-text text-transparent">
          with Go
        </span>
      </h1>

      <p className="mt-6 max-w-[520px] text-lg leading-8 tracking-normal text-[#60708d] sm:text-xl sm:leading-8">
        Upload a JPG or PNG, compress it in the Go backend and download the
        optimized result without storing the image on the server.
      </p>

      <ImageUploadForm />
    </div>
  );
}

function ImageComparisonPreview() {
  return (
    <div className="relative mx-auto h-[510px] w-full max-w-[650px] sm:h-[610px] lg:h-[700px] lg:max-w-none">
      <div
        aria-hidden="true"
        className="absolute left-[3%] top-[16%] h-[330px] w-[210px] rotate-[-10deg] overflow-hidden rounded-[28px] border-[6px] border-white bg-white opacity-45 shadow-[0_24px_70px_rgba(15,42,80,0.25)] blur-[3px] sm:h-[410px] sm:w-[260px] lg:left-[11%] lg:top-[20%]"
      >
        <Image
          alt=""
          className="h-full w-full object-cover saturate-50"
          fill
          priority
          sizes="(min-width: 1024px) 260px, 210px"
          src="/images/hero/mountain.png"
        />
      </div>

      <div
        aria-hidden="true"
        className="absolute right-[-4%] top-[23%] h-[360px] w-[230px] rotate-[10deg] overflow-hidden rounded-[28px] border-[6px] border-white bg-white opacity-45 shadow-[0_24px_70px_rgba(15,42,80,0.25)] blur-[4px] sm:h-[450px] sm:w-[280px] lg:right-[1%] lg:top-[27%]"
      >
        <Image
          alt=""
          className="h-full w-full object-cover saturate-50"
          fill
          priority
          sizes="(min-width: 1024px) 280px, 230px"
          src="/images/hero/mountain.png"
        />
      </div>

      <div className="absolute left-1/2 top-[8%] h-[410px] w-[300px] -translate-x-1/2 rotate-[12deg] rounded-[30px] bg-white p-[7px] shadow-[0_28px_78px_rgba(15,42,80,0.3)] sm:h-[535px] sm:w-[392px] lg:left-[53%] lg:top-[12%]">
        <div className="relative h-full overflow-hidden rounded-[24px]">
          <Image
            alt="Mountain lake image preview"
            className="h-full w-full object-cover"
            fill
            priority
            sizes="(min-width: 1024px) 392px, (min-width: 640px) 392px, 300px"
            src="/images/hero/mountain.png"
          />

          <div className="absolute inset-y-0 left-0 w-1/2 overflow-hidden">
            <div className="relative h-full w-[200%]">
              <Image
                alt=""
                className="h-full w-full object-cover brightness-[0.78] contrast-[0.86] saturate-[0.42]"
                fill
                sizes="(min-width: 1024px) 392px, (min-width: 640px) 392px, 300px"
                src="/images/hero/mountain.png"
              />
            </div>
            <div className="absolute inset-0 bg-[#243850]/20" />
          </div>

          <div className="absolute inset-y-0 left-1/2 w-[2px] -translate-x-1/2 bg-white/95 shadow-[0_0_16px_rgba(255,255,255,0.85)]" />
          <div className="absolute left-8 top-5 rounded-xl bg-[#203047]/80 px-5 py-3 text-sm font-bold text-white shadow-lg sm:left-9 sm:text-base">
            Before
          </div>
          <div className="absolute right-5 top-12 rotate-[-2deg] rounded-xl bg-white px-5 py-3 text-sm font-black text-[#00a87d] shadow-lg sm:right-6 sm:text-base">
            Optimized
          </div>
          <div className="absolute left-1/2 top-[56%] grid h-12 w-12 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full bg-white text-[#14234a] shadow-[0_12px_28px_rgba(8,18,54,0.22)]">
            <CompareIcon />
          </div>
        </div>
      </div>

      <div className="handwritten pointer-events-none absolute right-[3%] top-0 hidden rotate-[-6deg] text-[25px] font-bold leading-[1.05] text-[#6f7f9d] sm:block lg:right-[7%] lg:top-[2%] lg:text-[29px]">
        <p>Same dimensions</p>
        <p>Smaller when possible</p>
        <svg
          aria-hidden="true"
          className="absolute -right-12 top-11 h-24 w-20 rotate-[16deg] overflow-visible"
          fill="none"
          viewBox="0 0 76 96"
        >
          <path
            d="M7 5c35 1 54 18 52 49-.6 12-6 24-17 36"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="3"
          />
          <path
            d="M32 76l9 15 17-8"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="3"
          />
        </svg>
      </div>

      <div className="handwritten pointer-events-none absolute bottom-[5%] left-[10%] hidden rotate-[-9deg] text-[24px] font-bold leading-[1.05] text-[#6f7f9d] sm:block lg:bottom-[10%] lg:left-[18%] lg:text-[27px]">
        <svg
          aria-hidden="true"
          className="absolute -left-12 -top-10 h-20 w-20 -rotate-[14deg] overflow-visible"
          fill="none"
          viewBox="0 0 76 80"
        >
          <path
            d="M61 70C31 64 13 47 17 24c2-11 9-18 19-22"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="3"
          />
          <path
            d="M32 2l7-1 3 16"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="3"
          />
        </svg>
        <p>Measured sizes</p>
        <p>Clear result</p>
      </div>
    </div>
  );
}

function FeatureList() {
  const features = [
    {
      title: "Focused",
      description: "JPEG and PNG compression in the Go backend.",
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

function Footer() {
  return (
    <footer className="relative z-10 border-t border-[#dce7ee]/80 py-6 text-center text-sm text-[#60708d]">
      <span>Built with Go, Next.js and Tailwind CSS</span>
      <span aria-hidden="true" className="heart-mark ml-3 inline-block" />
    </footer>
  );
}

function DecorativeBackground() {
  return (
    <div aria-hidden="true" className="pointer-events-none absolute inset-0 overflow-hidden">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_12%,rgba(255,255,255,0.98),rgba(255,255,255,0)_36%),radial-gradient(circle_at_75%_5%,rgba(227,242,255,0.92),rgba(255,255,255,0)_35%),linear-gradient(180deg,#f7fcff_0%,#f8fcff_58%,#eef8f7_100%)]" />
      <svg
        className="absolute left-[-6%] top-[35%] h-[260px] w-[112%] text-[#9fe4d9]/45 sm:top-[40%]"
        fill="none"
        preserveAspectRatio="none"
        viewBox="0 0 1500 260"
      >
        <path
          d="M0 88C153 20 220 39 336 135c107 88 218 67 340 3 136-71 248-47 383 38 151 95 290 83 441-20"
          stroke="currentColor"
          strokeLinecap="round"
          strokeWidth="34"
        />
      </svg>
      <div className="hero-grass absolute bottom-0 left-[-7%] h-[43%] min-h-[280px] w-[78%] max-w-[820px] opacity-75 sm:h-[48%] lg:h-[45%]">
        <Image
          alt=""
          className="h-full w-full object-cover object-left-bottom"
          fill
          priority
          sizes="(min-width: 1024px) 820px, 78vw"
          src="/images/hero/grassfield.png"
        />
      </div>
      <div className="hero-grass absolute bottom-[-9%] right-[-24%] h-[35%] min-h-[220px] w-[58%] rotate-[-6deg] opacity-35 blur-[2px] sm:right-[-18%] lg:right-[-15%]">
        <Image
          alt=""
          className="h-full w-full object-cover object-left-bottom"
          fill
          sizes="58vw"
          src="/images/hero/grassfield.png"
        />
      </div>
    </div>
  );
}

function CompareIcon() {
  return (
    <svg aria-hidden="true" className="h-6 w-6" fill="none" viewBox="0 0 24 24">
      <path
        d="M8 7 4 11l4 4M4 11h7M16 17l4-4-4-4M20 13h-7"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.4"
      />
    </svg>
  );
}

function LightningIcon() {
  return (
    <svg aria-hidden="true" className="h-7 w-7" fill="currentColor" viewBox="0 0 24 24">
      <path d="M13.2 2.6 4.8 13.1a1 1 0 0 0 .8 1.6h5.1l-1.1 6.7a1 1 0 0 0 1.8.7l7.9-11.3a1 1 0 0 0-.8-1.6h-4.7l1.2-5.8a1 1 0 0 0-1.8-.8Z" />
    </svg>
  );
}

function ShieldIcon() {
  return (
    <svg aria-hidden="true" className="h-7 w-7" fill="none" viewBox="0 0 24 24">
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
    <svg aria-hidden="true" className="h-7 w-7" fill="none" viewBox="0 0 24 24">
      <path
        d="M6 19V9M12 19V5M18 19v-7"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="3"
      />
    </svg>
  );
}
